package services

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/samber/lo"
)

func LoadAllApikeys() {
	type ApiKey struct {
		UserId uint64 `json:"user_id"`
		ApiKey string `json:"api_key"`
	}
	var apikeys []*ApiKey
	if err := models.GetDB().Table("applications").Where("api_key is not null").Find(&apikeys).Error; err != nil {
		panic(err)
	}

	if len(apikeys) == 0 {
		return
	}

	var vals []string
	lo.ForEach(apikeys, func(a *ApiKey, i int) {
		vals = append(vals, fmt.Sprintf("apikey-%s", crypto.Keccak256Hash([]byte(a.ApiKey)).Hex()))
		vals = append(vals, fmt.Sprintf("%d", a.UserId))
	})

	if _, err := redis.DB().MSet(context.Background(), vals).Result(); err != nil {
		panic(err)
	}
}
