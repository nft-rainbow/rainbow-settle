package services

import (
	"context"
	"encoding/json"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/sirupsen/logrus"
)

func SetApiprofilesToRedis() {
	profiles, err := models.GetApiProfiles()
	if err != nil {
		panic(err)
	}
	var kvs []string
	for costtype, profile := range profiles {
		key := redis.ApiProfilesKey(costtype)
		val, err := json.Marshal(profile)
		if err != nil {
			panic(err)
		}
		kvs = append(kvs, key, string(val))
	}

	if _, err := redis.DB().MSet(context.Background(), kvs).Result(); err != nil {
		panic(err)
	}
	logrus.Info("set apiprofiles to redis")
}
