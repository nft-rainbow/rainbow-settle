package models

import (
	"time"
)

type UserDataBundle struct {
	BaseModel
	UserId       uint      `json:"user"`
	DataBundleId uint      `json:"data_bundle_id"`
	Count        uint      `json:"count"`
	BoughtTime   time.Time `json:"bought_time"`
	IsConsumed   bool      `json:"is_consumed"`
}

// func SetOnDataBundlerCreateHandler(handler OnDataBundleCreatHandler) {
// 	onDataBundlerCreateHandler = handler
// }
