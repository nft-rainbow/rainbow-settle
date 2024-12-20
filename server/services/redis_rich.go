package services

import (
	"context"
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func LoopSetRichFlagToRedis() {
	for {
		refreshRichFlag()
		time.Sleep(time.Second * 10)
	}
}

type userCostState struct {
	UserId        uint
	UserPayType   enums.UserPayType
	CostType      enums.CostType
	CountReset    int
	CountRollover int
	Balance       decimal.Decimal
	ArrearsQuota  decimal.Decimal
}

// balance 为 0 且 quota 为 0 时置 false；否则置 true
// select user_id, user_api_quota.cost_type, user_api_quota.count_reset, user_api_quota.count_rollover, user_balances.balance, user_balances.arrears_quota
// from users left join user_balances on users.id=user_balances.user_id left join user_api_quota on users.id=user_api_quota.user_id
func refreshRichFlag() error {
	var tmps []*userCostState
	if err := models.GetDB().Model(&models.User{}).
		Joins("left join user_balances on users.id=user_balances.user_id").
		Joins("left join user_api_quota on users.id=user_api_quota.user_id").
		Select("users.id as user_id, users.user_pay_type, user_api_quota.cost_type, user_api_quota.count_reset, user_api_quota.count_rollover, user_balances.balance, user_balances.arrears_quota").
		Scan(&tmps).Error; err != nil {
		return err
	}

	userCostStates := lo.GroupBy(tmps, func(v *userCostState) uint {
		return v.UserId
	})

	// logrus.WithField("user cost states", userCostStates).Trace("debug for refresh rich flag")
	for userId, costStates := range userCostStates {
		flag, err := calcRichFlag(costStates)
		if err != nil {
			return err
		}
		if _, err := redis.DB().Set(context.Background(), redis.RichKey(userId), flag, 0).Result(); err != nil {
			return err
		}
	}
	return nil
}

func calcRichFlag(states []*userCostState) (int, error) {
	flag := 0
	for _, cs := range states {
		apiPrice, err := models.GetApiPrice(cs.UserId, cs.CostType)
		if err != nil {
			return 0, err
		}

		isRich := cs.CountReset+cs.CountRollover > 0 || cs.Balance.Add(cs.ArrearsQuota).GreaterThanOrEqual(apiPrice)
		// only mint need pay for USER_PAY_TYPE_POST users
		if cs.UserPayType == enums.USER_PAY_TYPE_POST && cs.CostType != enums.COST_TYPE_RAINBOW_MINT {
			isRich = true
		}
		if isRich {
			flag = flag | 1<<int(cs.CostType)
		}
	}
	// logrus.WithField("states", states).WithField("flag", flag).Info("calc rich flag")
	return flag, nil
}
