package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestJsonMarshalFiatMetaPayApiFee(t *testing.T) {
	j, err := json.Marshal(FiatMetaPayApiFee{
		RefundedAmount: decimal.NewFromInt(10),
	})
	assert.NoError(t, err)

	fmt.Printf("%s", j)
}

func TestJsonMarshalFiatMetaBuySponsor(t *testing.T) {
	j, err := json.Marshal(FiatMetaBuySponsor{
		Price:          decimal.NewFromFloat32(0.8),
		RefundedAmount: decimal.NewFromInt(10),
	})
	assert.NoError(t, err)

	fmt.Printf("%s", j)
}
