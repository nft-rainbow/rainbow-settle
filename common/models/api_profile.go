package models

import (
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/shopspring/decimal"
)

type ApiProfile struct {
	BaseModel
	CostType enums.CostType  `json:"cost_type"`
	Price    decimal.Decimal `gorm:"type:decimal(20,5)" json:"price"`
}

var (
	GetApiPrice func(costType enums.CostType) decimal.Decimal
)

func InitApiProfile() {
	apiProfiles, err := GetApiProfiles()
	if err != nil {
		panic(err)
	}
	GetApiPrice = func(costType enums.CostType) decimal.Decimal {
		return apiProfiles[costType].Price
	}
}

func GetApiProfiles() (map[enums.CostType]*ApiProfile, error) {
	profiles := []*ApiProfile{}
	if err := GetDB().Find(&profiles).Error; err != nil {
		return nil, err
	}

	result := make(map[enums.CostType]*ApiProfile)
	for _, p := range profiles {
		result[p.CostType] = p
	}
	return result, nil
}

func ListApiFees() ([]ApiProfile, error) {
	var aps []ApiProfile
	if err := GetDB().Find(&aps).Error; err != nil {
		return nil, err
	}
	return aps, nil
}
