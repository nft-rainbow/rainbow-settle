package models

import (
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

func InitUserSettleds() {
	userIds := MustGetAllUserIds()
	costTypes := MustGetAllCostTypes()

	if err := GetUserSettledOperator().CreateIfNotExists(GetDB(), userIds, costTypes); err != nil {
		panic(err)
	}
}

var (
	userSettledOperator UserSettledOperator
)

func GetUserSettledOperator() *UserSettledOperator {
	return &userSettledOperator
}

type UserSettledOperator struct {
}

func (u *UserSettledOperator) GetUserSettled(userId uint) (map[enums.CostType]*UserSettled, error) {
	var us []*UserSettled
	if err := GetDB().Where("user_id = ?", userId).Find(&us).Error; err != nil {
		return nil, err
	}

	result := make(map[enums.CostType]*UserSettled)
	for _, q := range us {
		result[q.CostType] = q
	}
	return result, nil
}

func (u *UserSettledOperator) CreateIfNotExists(tx *gorm.DB, userIds []uint, costTypes []enums.CostType) error {
	var us []*UserSettled
	if err := tx.Where("user_id in ?", userIds).Where("cost_type in ?", costTypes).Find(&us).Error; err != nil {
		return err
	}

	exists := make(map[uint]map[enums.CostType]*UserSettled)
	for _, q := range us {
		if _, ok := exists[q.UserId]; !ok {
			exists[q.UserId] = make(map[enums.CostType]*UserSettled)
		}
		exists[q.UserId][q.CostType] = q
	}

	var unexists []*UserSettled
	for _, userId := range userIds {
		for _, costType := range costTypes {
			if _, ok := exists[userId][costType]; !ok {
				unexists = append(unexists, &UserSettled{
					UserId:   userId,
					CostType: costType,
				})
			}
		}
	}
	if len(unexists) == 0 {
		return nil
	}

	return tx.Save(&unexists).Error
}
