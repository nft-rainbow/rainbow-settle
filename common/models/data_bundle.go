package models

import "github.com/shopspring/decimal"

// 流量包
type DataBundle struct {
	BaseModel
	Name              string            `json:"name"`
	Price             decimal.Decimal   `json:"price"`
	DataBundleDetails []*UserDataBundle `json:"data_bundle_details"`
}
