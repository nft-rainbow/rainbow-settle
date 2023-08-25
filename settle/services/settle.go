package services

import (
	"context"

	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils/mathutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func LoopSettle(interval time.Duration) {
	logrus.Info("start settle looping")
	for {
		if err := settle(); err != nil {
			logrus.WithError(err).Info("failed to settle")
		}
		time.Sleep(interval)
	}
}

//  1. 结算redis中的quota，减user_balance, 加redis quota
//  2. 结算时记录多少是使用的免费quota，用于退款的时候区分退quota还是balance，需要仔细斟酌！
//     记录扣费堆栈：退费时根据堆栈退还
//     begin    [balance5, quota100]
//     mint5    [balance10,quota100]
//     refund15 [quota95]
//  3. TODO: 发现余额<=0且free quota为0后置标记 USER-COSTTYPE-RICH 为false
//  4. 写fiat log cache
func settle() error {
	userBalances := make(map[uint]*models.UserBalance)
	userApiQuotas := make(map[uint]map[enums.CostType]*models.UserApiQuota)
	userSettleds := make(map[uint]map[enums.CostType]*models.UserSettled)

	userCounts, err := redis.GetUserCounts()
	if err != nil {
		return errors.WithMessage(err, "failed to get user count")
	}

	if len(userCounts) == 0 {
		return nil
	}

	logrus.WithField("val", userCounts).Info("found need settle counts")

	for k, v := range userCounts {
		userId, costType, err := redis.ParseCountKey(k)
		if err != nil {
			logrus.WithError(err).WithField("key", k).Info("failed to parse count key")
			continue
		}

		// load userbalance and free quota
		if userBalances[userId] == nil {
			ub, err := models.GetUserBalance(userId)
			if err != nil {
				logrus.WithError(err).WithField("user id", userId).Info("failed to get user balance")
				continue
			}
			userBalances[userId] = ub
		}

		if userApiQuotas[userId] == nil {
			uq, err := userQuotaOperater.GetUserQuotas(userId)
			if err != nil {
				logrus.WithError(err).WithField("user id", userId).Info("failed to get user quotas")
				continue
			}
			userApiQuotas[userId] = uq
		}

		if userSettleds[userId] == nil {
			us, err := models.GetUserSettledOperator().GetUserSettled(userId)
			if err != nil {
				logrus.WithError(err).WithField("user id", userId).Info("failed to get user settled")
				continue
			}
			userSettleds[userId] = us
		}

		// calc cost
		count, err := redis.ParseCount(v)
		if err != nil {
			logrus.WithError(err).WithField("val", v).Info("failed to parse count")
			continue
		}

		// if has quota, cost quota
		var countInQuota, countInResetQuota, countInRolloverQuota int
		if q, ok := userApiQuotas[userId][costType]; ok {
			countInQuota = mathutils.Min(count, q.Total())
			countInResetQuota = mathutils.Min(countInQuota, q.CountReset)
			countInRolloverQuota = countInQuota - countInResetQuota
		}

		// calc count in balance
		price := models.GetApiPrice(costType)
		needCost := price.Mul(decimal.NewFromInt(int64(count - countInQuota)))
		actualCost := decimal.Min(needCost, userBalances[userId].Balance.Add(userBalances[userId].ArrearsQuota))
		countInBalance := actualCost.Div(price).BigInt().Int64()

		if countInBalance == 0 && countInQuota == 0 {
			continue
		}
		logrus.WithField("cost type", costType).WithField("in reset", countInResetQuota).WithField("in rollover", countInRolloverQuota).WithField("in balance", countInBalance).Info("calculated cost quota")

		// update mysql
		err = models.GetDB().Transaction(func(tx *gorm.DB) error {
			// update free quota
			if countInQuota > 0 {
				fl, _err := userQuotaOperater.Pay(tx, userId, costType, countInResetQuota, countInRolloverQuota)
				if _err != nil {
					return errors.WithMessage(_err, "failed to pay api quota")
				}
				logrus.WithField("user id", userId).WithField("fl", fl).Info("pay quota")
			}

			// update user balance
			if countInBalance > 0 {
				fl, _err := PayAPIFee(tx, userId, costType, uint(countInBalance))
				if _err != nil {
					return errors.WithMessage(_err, "failed to pay api fee")
				}
				logrus.WithField("user id", userId).WithField("fl", fl).Info("pay api fee")
			}

			// update settle
			us := userSettleds[userId][costType]
			data := us.Stack.Data()

			if countInResetQuota > 0 {
				if len(data) > 0 && data[len(data)-1].SettleType == enums.SETTLE_TYPE_QUOTA_RESET {
					data[len(data)-1].Count += uint(countInResetQuota)
				} else {
					data = append(us.Stack.Data(), &models.SettleStackItem{SettleType: enums.SETTLE_TYPE_QUOTA_RESET, Count: uint(countInResetQuota)})
				}
			}

			if countInRolloverQuota > 0 {
				if len(data) > 0 && data[len(data)-1].SettleType == enums.SETTLE_TYPE_QUOTA_ROLLOVER {
					data[len(data)-1].Count += uint(countInRolloverQuota)
				} else {
					data = append(us.Stack.Data(), &models.SettleStackItem{SettleType: enums.SETTLE_TYPE_QUOTA_ROLLOVER, Count: uint(countInRolloverQuota)})
				}
			}

			if countInBalance > 0 {
				if len(data) > 0 && data[len(data)-1].SettleType == enums.SETTLE_TYPE_BALANCE {
					data[len(data)-1].Count += uint(countInBalance)
				} else {
					data = append(us.Stack.Data(), &models.SettleStackItem{SettleType: enums.SETTLE_TYPE_BALANCE, Count: uint(countInBalance)})
				}
			}
			us.Stack = datatypes.NewJSONType(data)
			logrus.WithField("stack", us.Stack).Info("update user settle stack")
			return tx.Save(us).Error
		})
		if err != nil {
			logrus.WithError(err).Info("failed to update mysql user_balance and user_api_quota")
			continue
		}

		// update state on memory
		userApiQuotas[userId][costType].CountReset -= countInResetQuota
		userApiQuotas[userId][costType].CountRollover -= countInRolloverQuota
		userBalances[userId].Balance = userBalances[userId].Balance.Sub(actualCost)

		// update redis
		if _, err = redis.DB().DecrBy(context.Background(), k, int64(countInQuota+int(countInBalance))).Result(); err != nil {
			logrus.WithError(err).WithField("key", k).Error("redis: failed decr count")
		}
	}
	return nil
}

func RefundApiCost(userId uint, costType enums.CostType, count int) error {
	us, err := models.GetUserSettledOperator().GetUserSettled(uint(userId))
	if err != nil {
		return err
	}

	data := us[costType].Stack.Data()
	if len(data) == 0 {
		return errors.New("failed refund due to no cost stack")
	}

	handledCount := 0
	// 循环 pop 最后一个元素，直到偿还完毕
	l := len(data)
	for i := l - 1; i >= 0; i-- {
		remains := count - handledCount
		if remains == 0 {
			return nil
		}

		matchCount := mathutils.Min[int](int(data[i].Count), remains)
		data[i].Count -= uint(matchCount)

		var fl uint
		var err error
		switch data[i].SettleType {
		case enums.SETTLE_TYPE_BALANCE:
			fl, err = RefundApiFee(models.GetDB(), userId, costType, uint(matchCount))
		case enums.SETTLE_TYPE_QUOTA_RESET:
			fl, err = userQuotaOperater.Refund(models.GetDB(), userId, costType, matchCount, 0)
		case enums.SETTLE_TYPE_QUOTA_ROLLOVER:
			fl, err = userQuotaOperater.Refund(models.GetDB(), userId, costType, 0, matchCount)
		}
		if err != nil {
			return err
		}
		logrus.WithField("fiatlog", fl).Info("refund api cost")

		if data[i].Count == 0 {
			data = data[:i]
		}
		handledCount += matchCount
	}

	return nil
}
