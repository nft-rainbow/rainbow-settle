package services

import (
	"context"
	"time"

	"github.com/nft-rainbow/rainbow-fiat/common/models"
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/nft-rainbow/rainbow-fiat/common/redis"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func LoopSetRichFlag() {
	for {
		refreshRichFlag()
		time.Sleep(time.Second * 10)
	}
}

// balance 为 0 且 quota 为 0 时置 false；否则置 true
// select user_id, user_api_quota.cost_type, user_api_quota.count_reset, user_api_quota.count_rollover, user_balances.balance, user_balances.arrears_quota
// from users left join user_balances on users.id=user_balances.user_id left join user_api_quota on users.id=user_api_quota.user_id
func refreshRichFlag() error {
	type UserCostState struct {
		UserId        uint
		CostType      enums.CostType
		CountReset    int
		CountRollover int
		Balance       decimal.Decimal
		ArrearsQuota  decimal.Decimal
	}

	var tmps []*UserCostState
	if err := models.GetDB().Model(&models.User{}).
		Joins("left join user_balances on users.id=user_balances.user_id").
		Joins("left join user_api_quota on users.id=user_api_quota.user_id").
		Select("users.id as user_id, user_api_quota.cost_type, user_api_quota.count_reset, user_api_quota.count_rollover, user_balances.balance, user_balances.arrears_quota").
		Scan(&tmps).Error; err != nil {
		return err
	}

	userCostStates := lo.GroupBy(tmps, func(v *UserCostState) uint {
		return v.UserId
	})
	for userId, costStates := range userCostStates {
		flag := 0
		for _, cs := range costStates {
			isRich := cs.CountReset+cs.CountRollover > 0 || cs.Balance.Add(cs.ArrearsQuota).GreaterThanOrEqual(models.GetApiPrice(cs.CostType))
			if isRich {
				flag = flag | 1<<int(cs.CostType)
			}
		}
		if _, err := redis.DB().Set(context.Background(), redis.RichKey(userId), flag, 0).Result(); err != nil {
			return err
		}
	}
	return nil
}
