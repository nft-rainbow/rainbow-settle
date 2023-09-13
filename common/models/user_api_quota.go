package models

import (
	"encoding/json"
	"math"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserApiQuota struct {
	BaseModel
	UserId             uint           `json:"user_id"`
	CostType           enums.CostType `json:"cost_type"`
	CountReset         int            `json:"count_reset"`           // 会在指定重置时间后重置
	NextResetCountTime time.Time      `json:"next_reset_count_time"` // 下一次重置时间
	CountRollover      int            `json:"count_rollover"`        // 下个月顺延
}

func (u *UserApiQuota) Total() int {
	if u == nil {
		return 0
	}
	return u.CountReset + u.CountRollover
}

func InitUserUserApiQuota() {
	userIds := MustGetAllUserIds()
	costTypes := MustGetAllCostTypes()

	if err := GetUserQuotaOperator().CreateIfNotExists(GetDB(), userIds, costTypes); err != nil {
		panic(err)
	}
}

var (
	userQuotaOperator UserQuotaOperator
)

func GetUserQuotaOperator() *UserQuotaOperator {
	return &userQuotaOperator
}

type UserQuotaOperator struct {
}

func (u *UserQuotaOperator) GetUserQuotasMap(userId uint) (map[enums.CostType]*UserApiQuota, error) {
	uqs, err := u.GetUserQuotas(userId, 0, math.MaxInt32)
	if err != nil {
		return nil, err
	}

	uqsMap := lo.SliceToMap(uqs.Items, func(v *UserApiQuota) (enums.CostType, *UserApiQuota) {
		return v.CostType, v
	})
	return uqsMap, nil
}

func (*UserQuotaOperator) GetUserQuotas(userId uint, offset int, limit int) (*ginutils.List[*UserApiQuota], error) {
	var quotas []*UserApiQuota
	var count int64
	if err := GetDB().Model(&UserApiQuota{}).Where("user_id = ?", userId).Count(&count).Offset(offset).Limit(limit).Find(&quotas).Error; err != nil {
		return nil, err
	}

	return &ginutils.List[*UserApiQuota]{Count: count, Items: quotas}, nil
}

func (u *UserQuotaOperator) CreateIfNotExists(tx *gorm.DB, userIds []uint, costTypes []enums.CostType) error {
	var quotas []*UserApiQuota
	if err := tx.Where("user_id in ?", userIds).Where("cost_type in ?", costTypes).Find(&quotas).Error; err != nil {
		return err
	}

	exists := make(map[uint]map[enums.CostType]*UserApiQuota)
	for _, q := range quotas {
		if _, ok := exists[q.UserId]; !ok {
			exists[q.UserId] = make(map[enums.CostType]*UserApiQuota)
		}
		exists[q.UserId][q.CostType] = q
	}

	var unexists []*UserApiQuota
	for _, userId := range userIds {
		for _, costType := range costTypes {
			if _, ok := exists[userId][costType]; !ok {
				unexists = append(unexists, &UserApiQuota{
					UserId:             userId,
					CostType:           costType,
					NextResetCountTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local),
				})
			}
		}
	}
	if len(unexists) == 0 {
		return nil
	}

	return tx.Save(&unexists).Error
}

// force为 true 则即使 此时在nextResetTime之前 也会强制更新（当升级套餐时需要force reset）
func (u *UserQuotaOperator) Reset(tx *gorm.DB, userIds []uint, resetQuotas map[enums.CostType]int, nextResetTime time.Time, force bool) error {
	var flcs []*FiatLogCache
	err := func() error {
		var matched []*UserApiQuota

		where := tx.
			Where("cost_type in ?", utils.GetMapKeys(resetQuotas)).
			Where("user_id in ?", userIds)

		if !force {
			where = where.Where("next_reset_count_time<?", nextResetTime)
		}

		if err := tx.Table("user_api_quota").
			Where(where).
			Find(&matched).Error; err != nil {
			return err
		}
		if len(matched) == 0 {
			return nil
		}

		for _, uaq := range matched {
			uaq.CountReset = resetQuotas[uaq.CostType]
			uaq.NextResetCountTime = nextResetTime
		}
		if err := tx.Save(&matched).Error; err != nil {
			return err
		}

		for _, uaq := range matched {
			meta, _ := json.Marshal(FiatMetaResetQuota{uaq.CostType, resetQuotas[uaq.CostType]})
			flcs = append(flcs, &FiatLogCache{
				FiatLogCore: FiatLogCore{
					UserId:  uaq.UserId,
					Type:    FIAT_LOG_TYPE_RESET_API_QUOTA,
					Meta:    meta,
					OrderNO: RandomOrderNO(),
				},
			})

		}
		if err := tx.Debug().Save(&flcs).Error; err != nil {
			return err
		}

		// logrus.WithField("len", len(matched)).Info("reset api quota completed")
		return nil
	}()

	logrus.WithError(err).WithField("fiat log chace ids", GetIds(flcs)).WithField("user ids", userIds).WithField("reset quotas", resetQuotas).WithField("next reset time", nextResetTime).Info("reset users api quota completed")
	return err
}

func (u *UserQuotaOperator) DepositDataBundle(tx *gorm.DB, udb *UserDataBundle) error {

	var flcs []*FiatLogCache
	err := func() error {
		if udb.IsConsumed {
			return nil
		}

		dataBundle, err := GetDataBundleById(udb.DataBundleId)
		if err != nil {
			return err
		}

		rolloverQuotas := lo.SliceToMap(dataBundle.DataBundleDetails, func(d *DataBundleDetail) (enums.CostType, int) {
			return d.CostType, d.Count * int(udb.Count)
		})

		quotas, err := u.depositRollover(tx, udb.UserId, rolloverQuotas)
		if err != nil {
			return err
		}

		udb.IsConsumed = true
		if err := tx.Save(&udb).Error; err != nil {
			return err
		}

		for _, v := range quotas {
			meta, _ := json.Marshal(FiatMetaDepositDataBundle{udb.DataBundleId, *v})
			flcs = append(flcs, &FiatLogCache{
				FiatLogCore: FiatLogCore{
					UserId:  udb.UserId,
					Type:    FIAT_LOG_TYPE_DEPOSITE_DATABUNDLE,
					Meta:    meta,
					OrderNO: RandomOrderNO(),
				},
			})
		}

		if err := tx.Save(&flcs).Error; err != nil {
			return err
		}
		return nil
	}()

	logrus.WithError(err).WithField("user data bundle", udb).WithField("fiat log cache ids", GetIds(flcs)).Info("deposit data bundle completed")
	return nil
}

func (u *UserQuotaOperator) depositRollover(tx *gorm.DB, userId uint, rolloverQuotas map[enums.CostType]int) ([]*Quota, error) {
	var costCounts []*Quota
	err := func() error {
		var matched []*UserApiQuota
		if err := tx.Table("user_api_quota").
			Where("cost_type in ?", utils.GetMapKeys(rolloverQuotas)).
			Where("user_id = ?", userId).
			Find(&matched).Error; err != nil {
			return err
		}
		// if len(matched) == 0 {
		// 	return nil
		// }

		for _, uaq := range matched {
			// uaq.CountReset = rolloverQuotas[uaq.CostType]
			err := tx.Table("user_api_quota").
				Where("cost_type = ? and user_id = ?", uaq.CostType, uaq.UserId).
				Update("count_rollover", gorm.Expr("count_rollover+?", rolloverQuotas[uaq.CostType])).Error

			if err != nil {
				return err
			}
		}
		// if err := tx.Save(&matched).Error; err != nil {
		// 	return err
		// }

		// var flcs []*FiatLogCache

		for _, uaq := range matched {
			// meta, _ := json.Marshal(FiatMetaDepositRollover{uaq.CostType, rolloverQuotas[uaq.CostType]})
			// flcs = append(flcs, &FiatLogCache{
			// 	FiatLogCore: FiatLogCore{
			// 		UserId:  uaq.UserId,
			// 		Type:    FIAT_LOG_TYPE_RESET_API_QUOTA,
			// 		Meta:    meta,
			// 		OrderNO: RandomOrderNO(),
			// 	},
			// })
			costCounts = append(costCounts, &Quota{uaq.CostType, rolloverQuotas[uaq.CostType]})
		}
		// if err := tx.Save(&flcs).Error; err != nil {
		// 	return err
		// }

		// logrus.WithField("len", len(matched)).Info("reset api quota completed")
		return nil
	}()

	logrus.WithError(err).WithField("user id", userId).WithField("reset quotas", rolloverQuotas).Info("deposite user rollover quota completed")
	return costCounts, err
}

func (u *UserQuotaOperator) Refund(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {

	if err := tx.Model(&UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset+?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover+?", countRollover)).
		Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(FiatMetaRefundApiQuota{costType, countReset})
	fl := FiatLogCache{
		FiatLogCore: FiatLogCore{
			UserId:  userId,
			Type:    FIAT_LOG_TYPE_REFUND_API_QUOTA,
			Meta:    meta,
			OrderNO: RandomOrderNO(),
		},
	}
	if err := tx.Create(&fl).Error; err != nil {
		return 0, err
	}

	return fl.ID, nil
}

func (u *UserQuotaOperator) Pay(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {
	if err := tx.Model(&UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset-?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover-?", countRollover)).
		Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(FiatMetaPayApiQuota{costType, countReset})
	fl := FiatLogCache{
		FiatLogCore: FiatLogCore{
			UserId:  userId,
			Type:    FIAT_LOG_TYPE_PAY_API_QUOTA,
			Meta:    meta,
			OrderNO: RandomOrderNO(),
		},
	}
	if err := tx.Create(&fl).Error; err != nil {
		return 0, err
	}

	return fl.ID, nil
}
