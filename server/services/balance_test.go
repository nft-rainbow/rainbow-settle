package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalcLeftover(t *testing.T) {
	var null decimal.Decimal
	assert.Equal(t, decimal.Zero.String(), null.String())

	rawAmount, err := decimal.NewFromString("0.066")
	assert.NoError(t, err)
	amount, leftover := calcLeftover(rawAmount)
	assert.Equal(t, decimal.NewFromFloat(0.06).String(), amount.String())
	assert.Equal(t, decimal.NewFromFloat(0.006).String(), leftover.String())

	rawAmount, err = decimal.NewFromString("-0.066")
	assert.NoError(t, err)
	amount, leftover = calcLeftover(rawAmount)
	assert.Equal(t, decimal.NewFromFloat(-0.06).String(), amount.String())
	assert.Equal(t, decimal.NewFromFloat(-0.006).String(), leftover.String())
}
