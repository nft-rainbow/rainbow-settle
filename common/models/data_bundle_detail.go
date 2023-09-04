package models

import "github.com/nft-rainbow/rainbow-settle/common/models/enums"

type DataBundleDetail struct {
	BaseModel
	DataBundleId uint           `json:"data_bundle_id"`
	CostType     enums.CostType `json:"cost_type"`
	Count        int            `json:"count"`
}
