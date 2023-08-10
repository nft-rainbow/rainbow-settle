package models

import (
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ApiProfile struct {
	BaseModel
	Method    string          `gorm:"varchar(16);index" json:"method"`
	Path      string          `gorm:"type:varchar(64);index" json:"path"`
	IsTestnet bool            `gorm:"type:boolean;index" json:"is_testnet"`
	Price     decimal.Decimal `gorm:"type:decimal(20,5)" json:"price"`
}

var (
	getApiPrice func(method string, path string, isTestnet bool) decimal.Decimal
)

func InitApiProfile() {
	createIfUnexist := func(isTestnet bool) error {
		err := GetDB().Where(&ApiProfile{Method: "default", Path: "default"}).Where("is_testnet=?", isTestnet).First(&ApiProfile{}).Error
		if err == nil {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			if err = GetDB().Create(&ApiProfile{Method: "default", Path: "default", IsTestnet: isTestnet}).Error; err != nil {
				panic(err)
			}
			return nil
		}
		panic(err)
	}

	createIfUnexist(false)
	createIfUnexist(true)

	if err := initGetPriceFunc(); err != nil {
		panic(err)
	}
}

func initGetPriceFunc() error {
	f, err := ApiPriceObtainer()
	if err != nil {
		return err
	}
	getApiPrice = f
	return nil
}

func GetApiProfiles() (map[string]*ApiProfile, error) {
	profiles := []*ApiProfile{}
	if err := GetDB().Find(&profiles).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*ApiProfile)
	for _, p := range profiles {
		result[fmt.Sprintf("%v%v%v", p.Method, p.Path, p.IsTestnet)] = p
	}
	return result, nil
}

// 价格获取优先级如下：
// 1. fmt.Sprintf("%v%v%v", method, path, isTestnet)
// 2. fmt.Sprintf("%v%v%v", method, "default", isTestnet)
// 3. fmt.Sprintf("%v%v%v", "default", "default", isTestnet)
func ApiPriceObtainer() (func(method string, path string, isTestnet bool) decimal.Decimal, error) {
	// 获取价格
	apiProfiles, err := GetApiProfiles()
	if err != nil {
		return nil, err
	}

	getPrice := func(method string, path string, isTestnet bool) decimal.Decimal {
		if val, ok := apiProfiles[fmt.Sprintf("%v%v%v", method, path, isTestnet)]; ok {
			return val.Price
		}

		if val, ok := apiProfiles[fmt.Sprintf("%v%v%v", method, "default", isTestnet)]; ok {
			return val.Price
		}

		return apiProfiles[fmt.Sprintf("%v%v%v", "default", "default", isTestnet)].Price
	}
	return getPrice, nil
}

func ListApiFees() ([]ApiProfile, error) {
	var aps []ApiProfile
	if err := GetDB().Find(&aps).Error; err != nil {
		return nil, err
	}
	return aps, nil
}

// func GetApiPrice(method string, path string) (uint, error) {
// 	var price uint
// 	err := GetDB().Where(&ApiProfile{Method: method, Path: path}).Select("price").First(&price).Error
// 	if err == gorm.ErrRecordNotFound {
// 		return GetApiPrice("default", "default")
// 	}
// 	return price, err
// }
