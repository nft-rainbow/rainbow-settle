package services

import (
	"encoding/json"
	"math"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	userQuotaOperator UserQuotaOperator
)

func InitUserApiQuota() {
	userIds := models.MustGetAllUserIds()
	costTypes := models.MustGetAllCostTypes()

	if err := GetUserQuotaOperator().CreateIfNotExists(models.GetDB(), userIds, costTypes); err != nil {
		panic(err)
	}
}

func GetUserQuotaOperator() *UserQuotaOperator {
	return &userQuotaOperator
}

type UserQuotaOperator struct {
}

func (u *UserQuotaOperator) GetUserQuotasMap(userId uint) (map[enums.CostType]*models.UserApiQuota, error) {
	uqs, err := u.GetUserQuotas(userId, 0, math.MaxInt32)
	if err != nil {
		return nil, err
	}

	uqsMap := lo.SliceToMap(uqs.Items, func(v *models.UserApiQuota) (enums.CostType, *models.UserApiQuota) {
		return v.CostType, v
	})
	return uqsMap, nil
}

func (*UserQuotaOperator) GetUserQuotas(userId uint, offset int, limit int) (*ginutils.List[*models.UserApiQuota], error) {
	var quotas []*models.UserApiQuota
	var count int64
	if err := models.GetDB().Model(&models.UserApiQuota{}).Where("user_id = ?", userId).Count(&count).Offset(offset).Limit(limit).Find(&quotas).Error; err != nil {
		return nil, err
	}

	return &ginutils.List[*models.UserApiQuota]{Count: count, Items: quotas}, nil
}

func (u *UserQuotaOperator) CreateIfNotExists(tx *gorm.DB, userIds []uint, costTypes []enums.CostType) error {
	logrus.WithField("user ids", userIds).WithField("cost types", costTypes).Info("create user api quota")
	var quotas []*models.UserApiQuota
	if err := tx.Where("user_id in ?", userIds).Where("cost_type in ?", costTypes).Find(&quotas).Error; err != nil {
		return err
	}

	exists := make(map[uint]map[enums.CostType]*models.UserApiQuota)
	for _, q := range quotas {
		if _, ok := exists[q.UserId]; !ok {
			exists[q.UserId] = make(map[enums.CostType]*models.UserApiQuota)
		}
		exists[q.UserId][q.CostType] = q
	}

	var unexists []*models.UserApiQuota
	for _, userId := range userIds {
		for _, costType := range costTypes {
			if _, ok := exists[userId][costType]; !ok {
				unexists = append(unexists, &models.UserApiQuota{
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
	var flcs []*models.FiatLogCache
	err := func() error {
		var matched []*models.UserApiQuota

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
			meta, _ := json.Marshal(models.FiatMetaResetQuota{uaq.CostType, resetQuotas[uaq.CostType]})
			flcs = append(flcs, &models.FiatLogCache{
				FiatLogCore: models.FiatLogCore{
					UserId:  uaq.UserId,
					Type:    models.FIAT_LOG_TYPE_RESET_API_QUOTA,
					Meta:    meta,
					OrderNO: models.RandomOrderNO(),
				},
			})

		}
		if err := tx.Save(&flcs).Error; err != nil {
			return err
		}

		// logrus.WithField("len", len(matched)).Info("reset api quota completed")
		return nil
	}()

	logrus.WithError(err).WithField("fiat log chace ids", models.GetIds(flcs)).WithField("user ids", userIds).WithField("reset quotas", resetQuotas).WithField("next reset time", nextResetTime).Info("reset users api quota completed")
	return err
}

func (u *UserQuotaOperator) DepositDataBundle(tx *gorm.DB, udb *models.UserDataBundle) error {

	var flcs []*models.FiatLogCache
	err := func() error {
		if udb.IsConsumed {
			return nil
		}

		dataBundle, err := models.GetDataBundleById(udb.DataBundleId)
		if err != nil {
			return err
		}

		rolloverQuotas := lo.SliceToMap(dataBundle.DataBundleDetails, func(d *models.DataBundleDetail) (enums.CostType, int) {
			return d.CostType, d.Count * int(udb.Count)
		})

		quotas, err := u.depositRollover(tx, udb.UserId, rolloverQuotas)
		if err != nil {
			return err
		}

		ub, err := models.GetUserBalance(tx, udb.UserId)
		if err != nil {
			return err
		}

		udb.IsConsumed = true
		if err := tx.Save(&udb).Error; err != nil {
			return err
		}

		for _, v := range quotas {
			meta, _ := json.Marshal(models.FiatMetaDepositDataBundle{udb.DataBundleId, *v})
			flcs = append(flcs, &models.FiatLogCache{
				FiatLogCore: models.FiatLogCore{
					UserId:  udb.UserId,
					Type:    models.FIAT_LOG_TYPE_DEPOSITE_DATABUNDLE,
					Meta:    meta,
					OrderNO: models.RandomOrderNO(),
					Balance: ub.Balance,
				},
			})
		}

		if err := tx.Save(&flcs).Error; err != nil {
			return err
		}
		return nil
	}()

	logrus.WithError(err).WithField("user data bundle", udb).WithField("fiat log cache ids", models.GetIds(flcs)).Info("deposit data bundle completed")
	return nil
}

func (u *UserQuotaOperator) depositRollover(tx *gorm.DB, userId uint, rolloverQuotas map[enums.CostType]int) ([]*models.Quota, error) {
	var costCounts []*models.Quota
	err := func() error {
		var matched []*models.UserApiQuota
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
			costCounts = append(costCounts, &models.Quota{uaq.CostType, rolloverQuotas[uaq.CostType]})
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

	if err := tx.Model(&models.UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset+?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover+?", countRollover)).
		Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(models.FiatMetaRefundApiQuota{costType, countReset})
	fl := models.FiatLogCache{
		FiatLogCore: models.FiatLogCore{
			UserId:  userId,
			Type:    models.FIAT_LOG_TYPE_REFUND_API_QUOTA,
			Meta:    meta,
			OrderNO: models.RandomOrderNO(),
		},
	}
	if err := tx.Create(&fl).Error; err != nil {
		return 0, err
	}

	return fl.ID, nil
}

func (u *UserQuotaOperator) Pay(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {
	if err := tx.Model(&models.UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset-?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover-?", countRollover)).
		Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(models.FiatMetaPayApiQuota{costType, countReset})
	fl := models.FiatLogCache{
		FiatLogCore: models.FiatLogCore{
			UserId:  userId,
			Type:    models.FIAT_LOG_TYPE_PAY_API_QUOTA,
			Meta:    meta,
			OrderNO: models.RandomOrderNO(),
		},
	}
	if err := tx.Create(&fl).Error; err != nil {
		return 0, err
	}

	return fl.ID, nil
}
