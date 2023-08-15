package services

import (
	"context"
	"errors"
	"time"

	"github.com/nft-rainbow/rainbow-api/utils/mathutils"
	"github.com/nft-rainbow/rainbow-fiat/common/models"
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/nft-rainbow/rainbow-fiat/common/redis"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func SettleLooping(interval time.Duration) {
	for {
		go settle()
		time.Sleep(time.Second)
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
	var userBalances map[uint]*models.UserBalance
	var userFreeQuotas map[uint]map[enums.CostType]*models.UserApiQuota
	var userSettleds map[uint]map[enums.CostType]*models.UserSettled

	userCounts, err := redis.GetUserCounts()
	if err != nil {
		return err
	}

	for k, v := range userCounts {
		userId, costType, err := redis.ParseCountKey(k)
		if err != nil {
			continue
		}

		// load userbalance and free quota
		if userBalances[userId] == nil {
			ub, err := models.GetUserBalance(userId)
			if err != nil {
				continue
			}
			userBalances[userId] = ub
		}

		if userFreeQuotas[userId] == nil {
			uq, err := userQuotaOperater.GetUserQuotas(userId)
			if err != nil {
				continue
			}
			userFreeQuotas[userId] = uq
		}

		if userSettleds[userId] == nil {
			us, err := models.GetUserSettled(userId)
			if err != nil {
				continue
			}
			userSettleds[userId] = us
		}

		// calc cost
		count, err := redis.ParseCount(v)
		if err != nil {
			continue
		}

		// if has quota, cost quota
		countInQuota := mathutils.Min[int](count, userFreeQuotas[userId][costType].Total())
		countInResetQuota := mathutils.Min[int](countInQuota, userFreeQuotas[userId][costType].CountReset)
		countInRolloverQuota := countInQuota - countInResetQuota

		// calc count in balance
		price := models.GetApiPrice(costType)
		needCost := price.Mul(decimal.NewFromInt(int64(count - countInQuota)))
		actualCost := decimal.Min(needCost, userBalances[userId].Balance.Add(userBalances[userId].ArrearsQuota))
		countInBalance := actualCost.Div(price).BigInt().Int64()

		// update mysql
		err = models.GetDB().Transaction(func(tx *gorm.DB) error {
			// update free quota
			if countInQuota > 0 {
				fl, _err := userQuotaOperater.Pay(tx, userId, costType, countInResetQuota, countInRolloverQuota)
				if _err != nil {
					return _err
				}
				logrus.WithField("user id", userId).WithField("fl", fl).Info("pay quota")
			}

			// update user balance
			if countInBalance > 0 {
				if _, _err := models.PayAPIFee(tx, userId, costType, uint(countInBalance)); _err != nil {
					return _err
				}
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
			return tx.Save(us).Error
		})
		if err != nil {
			continue
		}

		// update state on memory
		userFreeQuotas[userId][costType].CountReset -= countInResetQuota
		userFreeQuotas[userId][costType].CountRollover -= countInRolloverQuota
		userBalances[userId].Balance = userBalances[userId].Balance.Sub(actualCost)

		// update redis
		if _, err = redis.DB().DecrBy(context.Background(), k, int64(countInQuota+int(countInBalance))).Result(); err != nil {
			logrus.WithError(err).WithField("key", k).Error("redis: failed decr count")
		}
	}
	return nil
}

func RefundApiCost(userId uint, costType enums.CostType, count int) error {
	us, err := models.GetUserSettled(uint(userId))
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
			fl, err = models.RefundApiFee(models.GetDB(), userId, costType, uint(matchCount))
		case enums.SETTLE_TYPE_QUOTA_RESET:
			fl, err = userQuotaOperater.Refund(models.GetDB(), userId, costType, matchCount, 0)
		case enums.SETTLE_TYPE_QUOTA_ROLLOVER:
			fl, err = userQuotaOperater.Refund(models.GetDB(), userId, costType, 0, matchCount)
		}
		if err != nil {
			return err
		}
		logrus.WithField("fiatlog", fl).Info("refund api fee")

		if data[i].Count == 0 {
			data = data[:i]
		}
		handledCount += matchCount
	}

	return nil
}
