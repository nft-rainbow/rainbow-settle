package models

import (
	"encoding/json"

	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"gorm.io/gorm"
)

type UserApiQuota struct {
	BaseModel
	UserId        uint           `json:"user_id"`
	CostType      enums.CostType `json:"cost_type"`
	CountReset    int            `json:"count_reset"`    // 下个月重置
	CountRollover int            `json:"count_rollover"` // 下个月顺延
}

func (u *UserApiQuota) Total() int {
	return u.CountReset + u.CountRollover
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

func (u *UserQuotaOperator) Reset(tx *gorm.DB, userIds []uint, resetCounts map[enums.CostType]int) error {
	for costType, count := range resetCounts {
		if err := tx.Update("count_reset", count).Where("cost_type=?", costType).Where("user_id in ?", userIds).Error; err != nil {
			return err
		}

		meta, _ := json.Marshal(map[string]interface{}{"cost_type": costType, "count": count})
		for _, userId := range userIds {
			if err := tx.Create(FiatLogCache{
				UserId:  userId,
				Type:    FIAT_LOG_TYPE_RESET_API_QUOTA,
				Meta:    meta,
				OrderNO: RandomOrderNO(),
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *UserQuotaOperator) Refund(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {

	if err := tx.Update("count_reset", gorm.Expr("count_reset+?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover+?", countRollover)).
		Where("cost_type=?", costType).Where("user_id = ?", userId).Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(map[string]interface{}{"cost_type": costType, "count": countReset})
	fl := FiatLogCache{
		UserId:  userId,
		Type:    FIAT_LOG_TYPE_REFUND_API_QUOTA,
		Meta:    meta,
		OrderNO: RandomOrderNO(),
	}
	if err := tx.Create(&fl).Error; err != nil {
		return 0, err
	}

	return fl.ID, nil
}

func (u *UserQuotaOperator) Pay(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {
	if err := tx.Update("count_reset", gorm.Expr("count_reset-?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover-?", countRollover)).
		Where("cost_type=?", costType).Where("user_id = ?", userId).Error; err != nil {
		return 0, err
	}

	meta, _ := json.Marshal(map[string]interface{}{"cost_type": costType, "count": countReset})
	fl := FiatLogCache{
		UserId:  userId,
		Type:    FIAT_LOG_TYPE_PAY_API_QUOTA,
		Meta:    meta,
		OrderNO: RandomOrderNO(),
	}
	if err := tx.Create(&fl).Error; err != nil {
		return 0, err
	}

	return fl.ID, nil
}
