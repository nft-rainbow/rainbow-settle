package models

import (
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/shopspring/decimal"
)

type FiatMetaDeposit struct {
	DepositOrderId uint `json:"deposit_order_id"`
}

type FiatMetaBuySponsor struct {
	Address string          `json:"address"`
	TxId    uint            `json:"tx_id"`
	Price   decimal.Decimal `json:"price"`
}

type FiatMetaRefundSponsor struct {
	RefundForFiatlogId   uint        `json:"refund_for_fiatlog_id"`
	RefundForFiatlogType FiatLogType `json:"refund_for_fiatlog_type"`
	TxId                 uint        `json:"tx_id"`
	Reason               string      `json:"reason"`
}

type FiatMetaRefundApiFee struct {
	CostType enums.CostType `json:"cost_type"`
	Count    uint           `json:"count"`
}

type FiatMetaPayApiFee FiatMetaRefundApiFee
