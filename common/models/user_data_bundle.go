package models

import "time"

type UserDataBundle struct {
	BaseModel
	UserId       uint      `json:"user"`
	DataBundleId uint      `json:"data_bundle_id"`
	BoughtTime   time.Time `json:"bought_time"`
}
