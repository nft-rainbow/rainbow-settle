package models

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreatFiatLogWillUpdateUserBalance(t *testing.T) {
	err := GetDB().Save([]FiatLog{
		{
			FiatLogCore: FiatLogCore{
				UserId:  1,
				Balance: decimal.NewFromInt(7),
				OrderNO: RandomOrderNO(),
			},
		},
	}).Error
	assert.NoError(t, err)

	var ub UserBalance
	err = GetDB().Model(&UserBalance{}).Where("user_id=?", 1).Find(&ub).Error
	assert.NoError(t, err)

	assert.True(t, decimal.NewFromInt(7).Equal(ub.BalanceOnFiatlog))
}
