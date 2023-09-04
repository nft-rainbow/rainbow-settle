package models

import "github.com/nft-rainbow/rainbow-settle/common/models/enums"

// 套餐详情：重置周期,CostType,Count,qps(优先)
type BillPlanDetail struct {
	BaseModel
	BillPlanId uint           `json:"bill_plan_id"`
	CostType   enums.CostType `json:"cost_type"`
	Count      int            `json:"count"`
	Qps        int            `json:"qps"`
}
