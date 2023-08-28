package models

import (
	"encoding/json"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
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

func (*UserQuotaOperator) GetUserQuotas(userId uint) (map[enums.CostType]*UserApiQuota, error) {
	var quotas []*UserApiQuota
	if err := GetDB().Where("user_id = ?", userId).Find(&quotas).Error; err != nil {
		return nil, err
	}

	result := make(map[enums.CostType]*UserApiQuota)
	for _, q := range quotas {
		result[q.CostType] = q
	}
	return result, nil
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

func (u *UserQuotaOperator) Reset(tx *gorm.DB, userIds []uint, resetCounts map[enums.CostType]int, nextResetTime time.Time) error {
	err := func() error {
		var matched []*UserApiQuota
		if err := tx.Table("user_api_quota").
			Where("cost_type in ?", utils.GetMapKeys(resetCounts)).
			Where("user_id in ?", userIds).
			Where("next_reset_count_time<?", nextResetTime).Find(&matched).Error; err != nil {
			return err
		}
		if len(matched) == 0 {
			return nil
		}

		for _, uaq := range matched {
			uaq.CountReset = resetCounts[uaq.CostType]
			uaq.NextResetCountTime = nextResetTime
		}
		if err := tx.Save(&matched).Error; err != nil {
			return err
		}

		var flcs []*FiatLogCache
		for _, uaq := range matched {
			meta, _ := json.Marshal(map[string]interface{}{"cost_type": uaq.CostType, "count": resetCounts[uaq.CostType]})
			flcs = append(flcs, &FiatLogCache{
				FiatLogCore: FiatLogCore{
					UserId:  uaq.UserId,
					Type:    FIAT_LOG_TYPE_RESET_API_QUOTA,
					Meta:    meta,
					OrderNO: RandomOrderNO(),
				},
			})

		}
		if err := tx.Save(&flcs).Error; err != nil {
			return err
		}

		logrus.WithField("len", len(matched)).Info("reset api quota completed")
		return nil
	}()

	logrus.WithError(err).WithField("user ids", userIds).WithField("reset counts", resetCounts).WithField("next reset time", nextResetTime).Info("reset users api quota")
	return err
}

func (u *UserQuotaOperator) Refund(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {

	if err := tx.Model(&UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset+?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover+?", countRollover)).
		Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(map[string]interface{}{"cost_type": costType, "count": countReset})
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

	meta, _ := json.Marshal(map[string]interface{}{"cost_type": costType, "count": countReset})
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
