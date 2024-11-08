package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func _TestCreatFiatLogWillUpdateUserBalance(t *testing.T) {
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

func _TestGetUserBalanceAtDate(t *testing.T) {
	userIds := []uint{1, 71, 90}
	balances, err := GetUserBalanceAtDate(userIds, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, len(userIds), len(balances))
	fmt.Println(balances)
}

func TestGetLastFiatLog(t *testing.T) {
	f, err := GetLastFiatLog(GetDB(), 1)
	assert.NoError(t, err)
	fmt.Println(*f)
}

func TestFindFiatlogIdsRefunded(t *testing.T) {
	start := time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 11, 1, 0, 0, 0, 0, time.Local)
	ids, err := FindSponsorFiatlogsRefunded(start, end)
	assert.NoError(t, err)
	fmt.Println(ids)
}
