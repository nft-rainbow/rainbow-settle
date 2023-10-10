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
	UserDataBundleId uint `json:"user_data_bundle_id"`
	Quota
}

type FiatMetaPayApiFeeForCache Quota
type FiatMetaPayApiFee struct {
	Quota
	RefundedCount  int
	RefundedAmount decimal.Decimal
}
type FiatMetaPayApiQuota struct {
	CostType      enums.CostType `json:"cost_type"`
	CountReset    int            `json:"count_reset"`
	CountRollover int            `json:"count_rollover"`
}

type FiatMetaRefundSponsor struct {
	RefundForFiatlogId   uint        `json:"refund_for_fiatlog_id"`
	RefundForFiatlogType FiatLogType `json:"refund_for_fiatlog_type"`
	TxId                 uint        `json:"tx_id"`
	Reason               string      `json:"reason"`
}

type FiatMetaRefundApiFeeForCache Quota
type FiatMetaRefundApiFee struct {
	Quota
	IsPart bool
}
type FiatMetaRefundApiQuota FiatMetaPayApiQuota

type FiatMetaResetQuota Quota
