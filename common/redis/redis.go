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

const (
	PREFIX_COUNT_KEY         = "count-"
	PREFIX_COUNT_PENDING_KEY = "count-pending-"
	PREFIX_REQ_KEY           = "req-"
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
	return fmt.Sprintf("%s%s-%s", PREFIX_COUNT_KEY, userId, costType)
}

func UserPendingCountKey(userId, costType string) string {
	return fmt.Sprintf("%s%s-%s", PREFIX_COUNT_PENDING_KEY, userId, costType)
}

func RequestKey(reqId uint) string {
	return fmt.Sprintf("%s%d", PREFIX_REQ_KEY, reqId)
}

func RequestValue(userId, costType, count string) string {
	return fmt.Sprintf("%s-%s-%s", userId, costType, count)
}

func ParseCountKey(key string) (userId uint, costType enums.CostType, err error) {
	return parseCountKeyByPrefix(key, PREFIX_COUNT_KEY)
}

func ParsePendingCountKey(key string) (userId uint, costType enums.CostType, err error) {
	return parseCountKeyByPrefix(key, PREFIX_COUNT_PENDING_KEY)
}

func parseCountKeyByPrefix(key string, prefix string) (userId uint, costType enums.CostType, err error) {
	err = func() error {
		core, ok := strings.CutPrefix(key, prefix)
		if !ok {
			return errors.Errorf("invalid count key: %s, expect prefix %s", prefix)
		}

		items := strings.Split(core, "-")
		if len(items) != 2 {
			return errors.Errorf("expect 2 items, got %d", len(items))
		}

		userId, err = ParseUserId(items[0])
		if err != nil {
			return err
		}

		costType, err = ParseCostType(items[1])
		if err != nil {
			return err
		}

		return nil
	}()
	return
}

func ParseRequestValue(val string) (userId uint, costType enums.CostType, count int, err error) {
	err = func() error {
		items := strings.Split(val, "-")
		if len(items) != 3 {
			return errors.Errorf("expect 3 items, got %d", len(items))
		}

		userId, err = ParseUserId(items[0])
		if err != nil {
			return err
		}

		costType, err = ParseCostType(items[1])
		if err != nil {
			return err
		}

		count, err = ParseCount(items[2])
		if err != nil {
			return err
		}

		return nil

	}()
	return
}

func ParseUserId(userId string) (uint, error) {
	_userId, err := strconv.Atoi(userId)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse user id %s", userId)
	}
	return uint(_userId), nil
}

func ParseCostType(costType string) (enums.CostType, error) {
	_costType, err := enums.ParseCostType(costType)
	if err != nil {
		return enums.CostType(0), errors.Wrapf(err, "failed parse costtype %s", costType)
	}
	return *_costType, nil
}

func ParseCount(count string) (int, error) {
	_count, err := strconv.Atoi(count)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse count %s", count)
	}
	return _count, nil
}

func GetUserCounts() (map[string]string, error) {
	return GetValuesByRegexKey(fmt.Sprintf("^%s-\\d*-.*$", PREFIX_COUNT_KEY))
}

func GetValuesByRegexKey(pattern string) (map[string]string, error) {
	// 使用 SCAN 命令获取匹配的 key
	var cursor uint64
	var keys []string

	for {
		var partialKeys []string
		var err error

		partialKeys, cursor, err = rdb.Scan(context.Background(), cursor, pattern, 10).Result()
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}

		keys = append(keys, partialKeys...)

		if cursor == 0 {
			break
		}
	}

	result := make(map[string]string)
	// 获取匹配的 key 对应的 value
	for _, key := range keys {
		value, err := rdb.Get(context.Background(), key).Result()
		if err != nil {
			// fmt.Printf("Error getting value for key %s: %v\n", key, err)
			return nil, errors.Wrapf(err, "failed to get value for key %s", key)
		}

		result[key] = value

		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	return result, nil
}
