package services

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/shopspring/decimal"
)

func TestCalcRichFlag(t *testing.T) {
	costStates := []*userCostState{
		{1, enums.USER_PAY_TYPE_PRE, enums.COST_TYPE_CONFURA_MAIN_CSPACE_NORMAL, 100, 100, decimal.NewFromInt(1), decimal.NewFromInt(1)},
	}
	flag := calcRichFlag(costStates)
	assert.Equal(t, 0b10000000000, flag)
}
