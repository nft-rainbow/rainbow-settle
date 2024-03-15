package models

import (
	"fmt"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/shopspring/decimal"
)

type ApiProfile struct {
	BaseModel
	CostType       enums.CostType   `gorm:"unique" json:"cost_type"`
	CostTypeName   string           `json:"-"`
	ServerType     enums.ServerType `json:"server_type"`
	ServerTypeName string           `json:"-"`
	Price          decimal.Decimal  `gorm:"type:decimal(10,6)" json:"price"`
}

type ApiProfileFilter struct {
	CostType   enums.CostType   `form:"cost_type" json:"cost_type"`
	ServerType enums.ServerType `form:"server_type" json:"server_type"`
}

var (
	GetApiPrice func(userId uint, costType enums.CostType) decimal.Decimal
)

func InitApiProfile() {
	apiProfiles, err := GetApiProfiles()
	if err != nil {
		panic(err)
	}
	// check if has all cost types, and if cost name match
	for k, v := range enums.CostTypeValue2StrMap {
		ap, ok := apiProfiles[k]
		if !ok {
			panic("missing api profile of type: " + k.String())
		}

		if ap.CostTypeName != v {
			panic("cost type name not match for type: " + v)
		}
	}

	um, err := GetAllUsersMap()
	if err != nil {
		panic(err)
	}

	GetApiPrice = func(userId uint, costType enums.CostType) decimal.Decimal {
		if um[userId].UserPayType == enums.USER_PAY_TYPE_POST {
			if costType == enums.COST_TYPE_RAINBOW_MINT {
				return decimal.NewFromFloat32(0.7)
			}
			return decimal.Zero
		}
		if _, ok := apiProfiles[costType]; !ok {
			panic(fmt.Sprintf("not found api price of %d", int(costType)))
		}
		return apiProfiles[costType].Price
	}
}

func QueryApiProfile(filter *ApiProfileFilter, offset, limit int) (*ginutils.List[*ApiProfile], error) {
	var aps []*ApiProfile
	var count int64
	if err := GetDB().Model(&ApiProfile{}).Where(&filter).Count(&count).Offset(offset).Limit(limit).Find(&aps).Error; err != nil {
		return nil, err
	}
	return &ginutils.List[*ApiProfile]{Items: aps, Count: count}, nil
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

func GetAllCostTypes() ([]enums.CostType, error) {
	aps, err := GetApiProfiles()
	if err != nil {
		return nil, err
	}
	return utils.GetMapKeys(aps), nil
}

func ListApiFees() ([]*ApiProfile, error) {
	var aps []*ApiProfile
	if err := GetDB().Find(&aps).Error; err != nil {
		return nil, err
	}
	return aps, nil
}

func MustGetAllCostTypes() []enums.CostType {
	return utils.Must(GetAllCostTypes())
}
