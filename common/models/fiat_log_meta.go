package models

import (
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/shopspring/decimal"
)

type Quota struct {
	CostType enums.CostType `json:"cost_type"`
	Count    int            `json:"count"`
}
type FiatMetaDeposit struct {
	DepositOrderId uint `json:"deposit_order_id"`
}

type FiatMetaBuySponsor struct {
	Address string          `json:"address"`
	TxId    uint            `json:"tx_id"`
	Price   decimal.Decimal `json:"price"`
}

type FiatMetaBuyBillplan struct {
	PlanId         uint `json:"plan_id"`
	UserBillPlanId uint `json:"user_bill_plan_id"`
}

type FiatMetaBuyDatabundle struct {
	DataBundleId     uint `json:"data_bundle_id"`
	Count            uint `json:"count"`
	UserDataBundleId uint `json:"user_data_bundle_id"`
}

type FiatMetaDepositDataBundle struct {
	FiatMetaBuyDatabundle
	Quota
}

type FiatMetaPayApiFee Quota
type FiatMetaPayApiQuota Quota

type FiatMetaRefundSponsor struct {
	RefundForFiatlogId   uint        `json:"refund_for_fiatlog_id"`
	RefundForFiatlogType FiatLogType `json:"refund_for_fiatlog_type"`
	TxId                 uint        `json:"tx_id"`
	Reason               string      `json:"reason"`
}

type FiatMetaRefundApiFee Quota
type FiatMetaRefundApiQuota Quota

type FiatMetaResetQuota Quota
