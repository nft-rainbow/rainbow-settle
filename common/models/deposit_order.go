package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

const (
	DEPOSIT_FAILED = iota - 1
	DEPOSIT_INIT
	DEPOSIT_SUCCESS
)

const (
	DEPOSIT_TYPE_WECHAT = iota + 1
	DEPOSIT_TYPE_CMB    // 招行对公充值
)

type DepositOrder struct {
	BaseModel
	UserId      uint            `gorm:"type:int;index" json:"user_id"`
	Amount      decimal.Decimal `gorm:"type:decimal(20,2)" json:"amount"` // 单位分
	Type        uint            `gorm:"type:int;default:0" json:"type"`   // 1-wechat
	Status      uint            `gorm:"type:int;default:0" json:"status"` // 0-init,1-success,-1-failed
	Description string          `gorm:"type:varchar(255)" json:"description"`
	TradeNo     string          `gorm:"type:varchar(255);index" json:"trade_no"`
	Meta        datatypes.JSON  `gorm:"type:json" json:"meta"` // metadata
}

func FindDepositOrderById(id uint) (*DepositOrder, error) {
	var item DepositOrder
	err := db.First(&item, id).Error
	return &item, err
}

func FindDepositOrderByTradeNo(no string) (*DepositOrder, error) {
	var item DepositOrder
	err := db.Where("trade_no = ?", no).First(&item).Error
	return &item, err
}
