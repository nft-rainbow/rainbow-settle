package models

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"gorm.io/gorm"
)

type UserDataBundle struct {
	BaseModel
	UserId       uint      `json:"user"`
	DataBundleId uint      `json:"data_bundle_id"`
	Count        uint      `json:"count"`
	BoughtTime   time.Time `json:"bought_time"`
	IsConsumed   bool      `json:"is_consumed"`
}

func CreateUserDataBundleAndConsume(userId, dataBundleId, count uint) (*UserDataBundle, error) {
	udb := &UserDataBundle{
		UserId:       userId,
		DataBundleId: dataBundleId,
		Count:        count,
		BoughtTime:   time.Now(),
	}

	err := GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&udb).Error; err != nil {
			return err
		}
		if err := GetUserQuotaOperator().DepositDataBundle(tx, udb); err != nil {
			return err
		}

		utils.Retry(10, time.Second, func() error { return tx.Save(&udb).Error })
		return nil
	})
	if err != nil {
		return nil, err
	}

	return udb, nil
}

// func SetOnDataBundlerCreateHandler(handler OnDataBundleCreatHandler) {
// 	onDataBundlerCreateHandler = handler
// }
