package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

var (
	rdb    *ExtendClient
	dbSync sync.Once
)

type ExtendClient struct {
	*redis.Client
}

func (c *ExtendClient) GetIntOrDefault(ctx context.Context, key string) *redis.IntCmd {
	result := redis.NewIntCmd(ctx)

	str, err := c.Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			result.SetVal(int64(0))
			return result
		}
		result.SetErr(err)
		return result
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		result.SetErr(errors.Wrap(err, "failed to convert to int"))
		return result
	}

	result.SetVal(int64(i))
	return result
}

func DB() *ExtendClient {
	dbSync.Do(
		func() {
			c := redis.NewClient(&redis.Options{
				Addr:     "redis:6379",
				Password: "", // 密码
				DB:       0,  // 数据库
				PoolSize: 20, // 连接池大小
			})
			rdb = &ExtendClient{c}
		},
	)
	return rdb
}

func UserCountKey(userId, costType string) string {
	return fmt.Sprintf("count-%s-%s", userId, costType)
}

func UserPendingCountKey(userId, costType string) string {
	return fmt.Sprintf("count-pending-%s-%s", userId, costType)
}

func RequestCountKey(reqId uint) string {
	return fmt.Sprintf("req-%d", reqId)
}

func RequestCountValue(userId, costType, count string) string {
	return fmt.Sprintf("%s-%s-%s", userId, costType, count)
}

func ParseRequestValue(val string) (userId uint, costType enums.CostType, count int, err error) {
	items := strings.Split(val, "-")
	if len(items) != 3 {
		return 0, enums.CostType(0), 0, errors.Errorf("expect 3 items, got %d", len(items))
	}

	_userId, err := strconv.Atoi(items[0])
	if err != nil {
		return 0, enums.CostType(0), 0, errors.Wrapf(err, "failed to parse user id %s", items[0])
	}

	_costType, err := enums.ParseCostType(items[1])
	if err != nil {
		return 0, enums.CostType(0), 0, err
	}

	_count, err := strconv.Atoi(items[2])
	if err != nil {
		return 0, enums.CostType(0), 0, errors.Wrapf(err, "failed to parse count %s", items[0])
	}

	return uint(_userId), *_costType, _count, nil
}
