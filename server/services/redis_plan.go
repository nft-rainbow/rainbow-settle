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
	setUserPlansToRedis()
	setPlansToRedis()
	refreshOnPlanChanged()
	logrus.Info("set plans related to redis")
}

func setUserPlansToRedis() {
	logrus.Info("refresh all user plans to redis")
	userPlansMap, err := models.GetUserBillPlanOperator().FindAllUsersEffectivePlan()
	if err != nil {
		panic(err)
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
		panic(err)
	}
}

func setPlansToRedis() {
	plans, err := models.GetAllPlansMap()
	if err != nil {
		panic(err)
	}

	var kvs []string
	for planId, plan := range plans {
		key := redis.PlanKey(planId)
		val, err := json.Marshal(plan)
		if err != nil {
			panic(err)
		}
		kvs = append(kvs, key, string(val))
	}

	if _, err := redis.DB().MSet(context.Background(), kvs).Result(); err != nil {
		panic(err)
	}
}

func refreshOnPlanChanged() {
	models.GetUserBillPlanOperator().RegisterOnChangedEvent(func(old, new *models.UserBillPlan) {
		setUserPlansToRedis()
	})
}
