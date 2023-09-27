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

func CreateUserDataBundleAndConsume(tx *gorm.DB, userId, dataBundleId, count uint) (*UserDataBundle, error) {
	udb := &UserDataBundle{
		UserId:       userId,
		DataBundleId: dataBundleId,
		Count:        count,
		BoughtTime:   time.Now(),
	}

	coreFn := func(_tx *gorm.DB) error {
		if err := _tx.Create(&udb).Error; err != nil {
			return err
		}
		if err := GetUserQuotaOperator().DepositDataBundle(_tx, udb); err != nil {
			return err
		}

		utils.Retry(10, time.Second, func() error { return _tx.Save(&udb).Error })
		return nil
	}

	var err error
	if tx == db {
		err = GetDB().Transaction(func(tx *gorm.DB) error {
			return coreFn(tx)
		})
	} else {
		err = coreFn(tx)
	}

	if err != nil {
		return nil, err
	}

	return udb, nil
}

// func SetOnDataBundlerCreateHandler(handler OnDataBundleCreatHandler) {
// 	onDataBundlerCreateHandler = handler
// }
