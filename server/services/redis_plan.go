package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/sirupsen/logrus"
)

// Init 时加载所有Plan到redis
// Init 时加载所有用户PlanId到redis
// 当用户PlanId变化时更新redis

func SetPlanToRedis() {
	if err := setAllUserPlansToRedis(); err != nil {
		panic(err)
	}
	if err := setPlansToRedis(); err != nil {
		panic(err)
	}
	logrus.Info("set plans related to redis")
}

func setUserPlansToRedis(userIds []uint) error {
	logrus.WithField("users", userIds).Info("refresh user plans to redis")
	userPlansMap, err := models.GetUserBillPlanOperator().FindUsersEffectivePlans(userIds)
	if err != nil {
		return err
	}

	var kvs []string
	for userId, userPlansMap := range userPlansMap {
		for serverType, userPlan := range userPlansMap {
			key := redis.UserPlanKey(userId, serverType)
			val := fmt.Sprintf("%d", userPlan.PlanId)
			kvs = append(kvs, key, val)
		}
	}

	if _, err := redis.DB().MSet(context.Background(), kvs).Result(); err != nil {
		return err
	}
	return nil
}

func setAllUserPlansToRedis() error {
	logrus.Info("refresh all user plans to redis")
	allUserIds, err := models.GetAllUserIds()
	if err != nil {
		return err
	}
	return setUserPlansToRedis(allUserIds)
}

func setPlansToRedis() error {
	plans, err := models.GetAllPlansMap()
	if err != nil {
		return err
	}

	var kvs []string
	for planId, plan := range plans {
		key := redis.PlanKey(planId)
		val, err := json.Marshal(plan)
		if err != nil {
			return err
		}
		kvs = append(kvs, key, string(val))
	}

	if _, err := redis.DB().MSet(context.Background(), kvs).Result(); err != nil {
		return err
	}
	return nil
}
