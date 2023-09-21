package plugins

import (
	"context"

	urate "github.com/Conflux-Chain/go-conflux-util/rate"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
)

/*
限流器设计

	库：time/rate

	限流插件
	- 限流Conf：request 或 costtype

	## 流程
	限流器Map:
	- costTypeLimiter 用户-costtype-limiter
	- serverCostTypeLimiter 用户-servertype-limiter
	- serverRequestLimiter 用户-servertype-limiter

	限流规则：
	rainbow：按请求限流
	- 按server限流
	confura：按costtype限流
	- 先按server限流再按costtype限流

	Redis获取限流配置:
	- userplan-userid-servertype: planid
	- defaultplan-servertype: planid
	- plan-id: plan

	getQpsByServer(server,user)
	getQpsByCosttype(costtype,user)
*/

type QpsObtainer interface {
	getQps(serverOrCosttype, userid string) (qps, burst int, err error)
}

type serverQpsObtainer struct{}

func (r *serverQpsObtainer) getQps(server, userid string) (qps, burst int, err error) {
	_userid, err := redis.ParseUserId(userid)
	if err != nil {
		return 0, 0, err
	}

	serverType, err := enums.ParseServerType(server)
	if err != nil {
		return 0, 0, err
	}

	_, qps, err = redis.GetUserServerQps(_userid, *serverType)
	if err != nil {
		return 0, 0, err
	}
	return 5, 10, nil
	// return qps, qps * 2, nil
}

type costtypeQpsObtainer struct{}

func (r *costtypeQpsObtainer) getQps(costtype, userid string) (qps, burst int, err error) {
	_userid, err := redis.ParseUserId(userid)
	if err != nil {
		return 0, 0, err
	}

	costType, err := enums.ParseCostType(costtype)
	if err != nil {
		return 0, 0, err
	}

	qps, err = redis.GetUserCostQps(_userid, *costType)
	if err != nil {
		return 0, 0, err
	}
	return 5, 10, nil
}

type RainbowLimiterFactory struct {
	qpsObtainer QpsObtainer
}

// GetGroupAndKey generates group and key from context
func (r *RainbowLimiterFactory) GetGroupAndKey(ctx context.Context, serverOrCosttype string) (userid, key string, err error) {
	userid = ctx.Value(constants.RAINBOW_USER_ID_HEADER_KEY).(string)
	key = "default"
	return
}

// Create creates limiter by resource and group
func (r *RainbowLimiterFactory) Create(ctx context.Context, serverOrCosttype, userid string) (urate.Limiter, error) {
	qps, burst, err := r.qpsObtainer.getQps(serverOrCosttype, userid)
	if err != nil {
		return nil, err
	}
	log.Infof("create limiter for user %s, server or cost type %v, qps %d, burst %d", userid, serverOrCosttype, qps, burst)
	return urate.NewTokenBucket(qps, burst), nil
}

// type CosttypeLimiterFactory struct {
// }

// func (r *CosttypeLimiterFactory) getQps(costtype, userid string) (qps, burst int, err error) {
// 	return 0, 0, nil
// }

// // GetGroupAndKey generates group and key from context
// func (r *CosttypeLimiterFactory) GetGroupAndKey(ctx context.Context, costtype string) (userid, key string, err error) {
// 	userid = ctx.Value(constants.RAINBOW_USER_ID_HEADER_KEY).(string)
// 	key = "default"
// 	return
// }

// // Create creates limiter by resource and group
// func (r *CosttypeLimiterFactory) Create(ctx context.Context, costtype, userid string) (urate.Limiter, error) {
// 	qps, burst, err := r.getQps(costtype, userid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return urate.NewTokenBucket(qps, burst), nil
// }

// type ServerLimiterFactory struct {
// }

// func (r *ServerLimiterFactory) getQps(server, userid string) (qps, burst int, err error) {
// 	return 0, 0, nil
// }

// // GetGroupAndKey generates group and key from context
// func (r *ServerLimiterFactory) GetGroupAndKey(ctx context.Context, server string) (userid, key string, err error) {
// 	userid = ctx.Value(constants.RAINBOW_USER_ID_HEADER_KEY).(string)
// 	key = "default"
// 	return
// }

// // Create creates limiter by resource and group
// func (r *ServerLimiterFactory) Create(ctx context.Context, server, userid string) (urate.Limiter, error) {
// 	qps, burst, err := r.getQps(server, userid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return urate.NewTokenBucket(qps, burst), nil
// }

// type MixLimiter struct {
// 	costTypeLimiter map[enums.CostType]urate.Limiter
// 	serverLimiter   urate.Limiter
// }

// func NewMixLimiter() *MixLimiter {

// }

// func (c *MixLimiter) Limit() error {
// 	return c.LimitN(1)
// }

// func (c *MixLimiter) LimitN(n int) error {
// 	return c.LimitAt(time.Now(), n)
// }

// func (c *MixLimiter) LimitAt(now time.Time, n int) error {
// 	if err := c.costTypeLimiter.LimitAt(now, n); err != nil {
// 		return err
// 	}
// 	return c.serverLimiter.LimitAt(now, n)
// }

// // Expired indicates whether limiter not updated for a long time.
// // Generally, it is used for garbage collection.
// func (c *MixLimiter) Expired() bool {
// 	return c.costTypeLimiter.Expired() && c.serverLimiter.Expired()
// }
