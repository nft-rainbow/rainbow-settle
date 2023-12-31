package models

import (
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/shopspring/decimal"
)

type Quota struct {
	CostType enums.CostType `json:"cost_type,omitempty"`
	Count    int            `json:"count,omitempty"`
}
type FiatMetaDeposit struct {
	DepositOrderId uint `json:"deposit_order_id"`
}

type FiatMetaWithdraw struct {
	Reason string `json:"reason,omitempty"`
}

type FiatMetaBuySponsor struct {
	Address        string          `json:"address"`
	TxId           uint            `json:"tx_id"`
	Price          decimal.Decimal `json:"price"`
	RefundedAmount decimal.Decimal `json:"refunded_amount"`
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
	RefundedCount  int             `json:"refunded_count"`
	RefundedAmount decimal.Decimal `json:"refunded_amount"`
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
	IsPart bool // 一条Refund可能会拆分成多条，目的是为了每一条refund_api_fee_log都最多对应一条 pay_api_fee_fiat_log
}
type FiatMetaRefundApiQuota FiatMetaPayApiQuota

type FiatMetaResetQuota Quota
