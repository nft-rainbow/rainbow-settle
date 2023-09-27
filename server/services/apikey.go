package services

import (
	"context"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/samber/lo"
)

func LoadAllApikeys() {
	type AppInfo struct {
		UserId uint64 `json:"user_id"`
		ID     uint64 `json:"id"`
		ApiKey string `json:"api_key"`
	}
	var appInfos []*AppInfo
	if err := models.GetDB().Table("applications").Where("api_key is not null").Find(&appInfos).Error; err != nil {
		panic(err)
	}

	if len(appInfos) == 0 {
		return
	}

	var vals []string
	lo.ForEach(appInfos, func(a *AppInfo, i int) {
		vals = append(vals,
			redis.ApikeyKey(a.ApiKey), //key
			redis.ApikeyValue(uint(a.UserId), uint(a.ID)),
		)
	})

	if _, err := redis.DB().MSet(context.Background(), vals).Result(); err != nil {
		panic(err)
	}
}

func RefreshApikeyToRedis(oldApikey, newApikey string, userId uint64, appId uint64) error {
	if oldApikey != "" {
		if _, err := redis.DB().Del(context.Background(), redis.ApikeyKey(oldApikey)).Result(); err != nil {
			return err
		}
	}
	// key := fmt.Sprintf("apikey-%s", crypto.Keccak256Hash([]byte(newApikey)).Hex()) //key
	// val := fmt.Sprintf("%d-%d", userId, appId)
	key := redis.ApikeyKey(newApikey)                   //key
	val := redis.ApikeyValue(uint(userId), uint(appId)) //value
	if _, err := redis.DB().Set(context.Background(), key, val, 0).Result(); err != nil {
		return err
	}
	return nil
}
