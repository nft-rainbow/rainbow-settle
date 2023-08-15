package models

import (
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"gorm.io/datatypes"
)

type SettleStackItem struct {
	SettleType enums.SettleType
	Count      uint
}

type UserSettled struct {
	BaseModel
	UserId   uint                                   `gorm:"user_id"`
	CostType enums.CostType                         `gorm:"cost_type"`
	Stack    datatypes.JSONType[[]*SettleStackItem] `gorm:"stack"`
}

func GetUserSettled(userId uint) (map[enums.CostType]*UserSettled, error) {
	var quotas []*UserSettled
	if err := GetDB().Where("user_id = ?", userId).Find(&quotas).Error; err != nil {
		return nil, err
	}

	result := make(map[enums.CostType]*UserSettled)
	for _, q := range quotas {
		result[q.CostType] = q
	}
	return result, nil
}
