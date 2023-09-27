package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
)

var (
	rdb *ExtendClient
)

const (
	PREFIX_COUNT_KEY         = "count-"
	PREFIX_COUNT_PENDING_KEY = "count-pending-"
	PREFIX_REQ_KEY           = "req-"
	PREFIX_RICH_KEY          = "rich-"
	PREFIX_APIKEY_KEY        = "apikey-"
	PREFIX_USER_PLAN_KEY     = "userplan-"
	PREFIX_PLAN_KEY          = "plan-"
	PREFIX_APIPROFILE_KEY    = "apiprofile-"
	PREFIX_RPC_IDS_KEY       = "rpcids-"
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

func Init(cfg config.Redis) {
	c := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password, // 密码
		DB:       0,            // 数据库
		PoolSize: 20,           // 连接池大小
	})
	rdb = &ExtendClient{c}
}

func DB() *ExtendClient {
	return rdb
}

func UserCountKey(userId, costType string) string {
	return fmt.Sprintf("%s%s-%s", PREFIX_COUNT_KEY, userId, costType)
}

func UserPendingCountKey(userId, costType string) string {
	return fmt.Sprintf("%s%s-%s", PREFIX_COUNT_PENDING_KEY, userId, costType)
}

func RequestKey(reqId string) string {
	return fmt.Sprintf("%s%s", PREFIX_REQ_KEY, reqId)
}

func RequestValue(userId, appId, costType, count string) string {
	return fmt.Sprintf("%s-%s-%s-%s", userId, appId, costType, count)
}

func RichKey(userId uint) string {
	return fmt.Sprintf("%s%d", PREFIX_RICH_KEY, userId)
}

func ApikeyKey(apikey string) string {
	return fmt.Sprintf("%s%s", PREFIX_APIKEY_KEY, crypto.Keccak256Hash([]byte(apikey)).Hex())
}

// value of user related apikey
func ApikeyValue(userId, appId uint) string {
	return fmt.Sprintf("%d%d", userId, appId)
}

func UserPlanKey(userId uint, serverType enums.ServerType) string {
	return fmt.Sprintf("%s%d-%s", PREFIX_USER_PLAN_KEY, userId, serverType)
}

func PlanKey(planId uint) string {
	return fmt.Sprintf("%s%d", PREFIX_PLAN_KEY, planId)
}

func ApiProfilesKey(costtype enums.CostType) string {
	return fmt.Sprintf("%s%d", PREFIX_APIPROFILE_KEY, costtype)
}

func RpcIdsInfoKey(requestId string) string {
	return fmt.Sprintf("%s%s", PREFIX_RPC_IDS_KEY, requestId)
}

type RpcInfo struct {
	Items      []*RpcInfoItem
	IsBatchRpc bool `json:"is_batch_rpc"`
}

type RpcInfoItem struct {
	RpcId      string `json:"rpc_id"`
	RpcVersion string `json:"rpc_version"`
}

func RpcIdsValue(info RpcInfo) string {
	v, _ := json.Marshal(info)
	return string(v)
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
			return errors.Errorf("invalid count key: %s, expect prefix %s", key, prefix)
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

func ParseRequestValue(val string) (userId uint, appId uint, costType enums.CostType, count int, err error) {
	err = func() error {
		items := strings.Split(val, "-")
		if len(items) != 4 {
			return errors.Errorf("expect 4 items, got %d", len(items))
		}

		userId, err = ParseUserId(items[0])
		if err != nil {
			return err
		}

		appId, err = ParseAppId(items[1])
		if err != nil {
			return err
		}

		costType, err = ParseCostType(items[2])
		if err != nil {
			return err
		}

		count, err = ParseCount(items[3])
		if err != nil {
			return err
		}

		return nil

	}()
	return
}

func ParseApikeyValue(userInfo string) (userId uint, appId uint, err error) {
	items := strings.Split(userInfo, "-")
	if len(items) != 2 {
		return 0, 0, errors.Errorf("user info format error")
	}
	userId, err = ParseUserId(items[0])
	if err != nil {
		return 0, 0, err
	}

	appId, err = ParseAppId(items[1])
	if err != nil {
		return 0, 0, err
	}
	return userId, appId, nil
}

func ParseUserId(userId string) (uint, error) {
	_userId, err := strconv.Atoi(userId)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse user id %s", userId)
	}
	return uint(_userId), nil
}

func ParseAppId(appId string) (uint, error) {
	_appId, err := strconv.Atoi(appId)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse app id %s", appId)
	}
	return uint(_appId), nil
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

func ParsePlanId(planId string) (int, error) {
	_planId, err := strconv.Atoi(planId)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse plan_id %s", planId)
	}
	return _planId, nil
}

func ParsePlan(planstr string) (*models.BillPlan, error) {
	var up models.BillPlan
	if err := json.Unmarshal([]byte(planstr), &up); err != nil {
		return nil, err
	}
	return &up, nil
}

func ParseRpcIdsInfo(rpcInfoStr string) (*RpcInfo, error) {
	var rpcInfo RpcInfo
	if err := json.Unmarshal([]byte(rpcInfoStr), &rpcInfo); err != nil {
		return nil, err
	}
	return &rpcInfo, nil
}

func GetUserCounts() (map[string]string, error) {
	return GetValuesByRegexKey(fmt.Sprintf("%s[0-9]*-*", PREFIX_COUNT_KEY))
}

func GetValuesByRegexKey(pattern string) (map[string]string, error) {
	// 使用 SCAN 命令获取匹配的 key
	var cursor uint64
	var keys []string

	for {
		var partialKeys []string
		var err error

		partialKeys, cursor, err = DB().Scan(context.Background(), cursor, pattern, 10).Result()
		if err != nil {
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
		value, err := DB().Get(context.Background(), key).Result()
		if err != nil {
			// fmt.Printf("Error getting value for key %s: %v\n", key, err)
			return nil, errors.Wrapf(err, "failed to get value for key %s", key)
		}

		if c, _ := ParseCount(value); c > 0 {
			result[key] = value
		}
	}
	return result, nil
}

func GetUserInfoByApikey(apikey string) (uint, uint, error) {
	key := ApikeyKey(apikey)
	val, err := DB().Get(context.Background(), key).Result()
	if err != nil {
		return 0, 0, errors.WithMessage(err, "failed to access db")
	}

	return ParseApikeyValue(val)
}

func CheckIsRich(userId uint, costType enums.CostType) (bool, error) {
	flagStr, err := DB().Get(context.Background(), RichKey(userId)).Result()
	if err != nil {
		return false, err
	}
	flag, err := strconv.Atoi(flagStr)
	if err != nil {
		return false, err
	}
	return isRich(flag, costType), nil
}

func isRich(flag int, costType enums.CostType) bool {
	result := (1 << costType & flag) > 0
	return result
}

func GetUserPlan(userid uint, server enums.ServerType) (*models.BillPlan, error) {
	userPlanKey := UserPlanKey(userid, server)
	planIdStr, err := DB().Get(context.Background(), userPlanKey).Result()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	planId, err := ParsePlanId(planIdStr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	planKey := PlanKey(uint(planId))
	planStr, err := DB().Get(context.Background(), planKey).Result()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	plan, err := ParsePlan(planStr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return plan, nil
}

func GetUserServerQps(userid uint, server enums.ServerType) (bool, int, error) {
	plan, err := GetUserPlan(userid, server)
	if err != nil {
		return false, 0, errors.WithStack(err)
	}
	return plan.IsQpsByRequset, plan.Qps, nil
}

func GetUserCostQps(userid uint, costType enums.CostType) (int, error) {
	apiProfile, err := GetApiProfile(costType)
	if err != nil {
		return 0, err
	}

	plan, err := GetUserPlan(userid, apiProfile.ServerType)
	if err != nil {
		return 0, err
	}

	detail, ok := lo.Find(plan.BillPlanDetails, func(d *models.BillPlanDetail) bool {
		return d.CostType == costType
	})

	if ok {
		return detail.Qps, nil
	}

	return plan.Qps, nil
}

func GetApiProfile(costType enums.CostType) (*models.ApiProfile, error) {
	apiProfileStr, err := DB().Get(context.Background(), ApiProfilesKey(costType)).Result()
	if err != nil {
		return nil, err
	}

	var apiProfile *models.ApiProfile

	if err := json.Unmarshal([]byte(apiProfileStr), &apiProfile); err != nil {
		return nil, err
	}

	return apiProfile, nil
}

func GetRequest(reqId string) (userId uint, appId uint, costType enums.CostType, count int, err error) {
	reqKey := RequestKey(reqId)
	val, err := DB().Get(context.Background(), reqKey).Result()
	if err != nil {
		// log.Errorf("failed to get req %d val: %s", w.ID(), err)
		return 0, 0, 0, 0, errors.Wrapf(err, "failed to get req %s", reqId)
	}

	userId, appId, costType, count, err = ParseRequestValue(val)
	if err != nil {
		// log.Errorf("failed to parse req %d val %s: %s", w.ID(), val, err)
		return 0, 0, 0, 0, errors.Wrapf(err, "failed to parse req %s", reqId)
	}
	return
}

func GetRpcIdsInfo(requestId string) (*RpcInfo, error) {
	rpcIdsStr, err := DB().Get(context.Background(), RpcIdsInfoKey(requestId)).Result()
	if err != nil {
		return nil, err
	}

	return ParseRpcIdsInfo(rpcIdsStr)
}
