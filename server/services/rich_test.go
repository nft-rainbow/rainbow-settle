package services

import (
	"testing"

	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalcRichFlag(t *testing.T) {
	costStates := []*userCostState{
		{1, enums.USER_PAY_TYPE_PRE, enums.COST_TYPE_CONFURA_MAIN_CSPACE_NORMAL, 100, 100, decimal.NewFromInt(1), decimal.NewFromInt(1)},
	}
	flag, err := calcRichFlag(costStates)
	assert.NoError(t, err)
	assert.Equal(t, 0b10000000000, flag)
}
