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
	type AppInfo struct {
		UserId uint64 `json:"user_id"`
		AppId  uint64 `json:"app_id"`
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
		vals = append(vals, fmt.Sprintf("apikey-%s", crypto.Keccak256Hash([]byte(a.ApiKey)).Hex()))
		vals = append(vals, fmt.Sprintf("%d-%d", a.UserId, a.AppId))
	})

	if _, err := redis.DB().MSet(context.Background(), vals).Result(); err != nil {
		panic(err)
	}
}
