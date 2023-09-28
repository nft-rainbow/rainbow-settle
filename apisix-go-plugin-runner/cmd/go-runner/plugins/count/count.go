package count

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/pkg/errors"
)

func init() {
	err := plugin.RegisterPlugin(&Count{})
	if err != nil {
		log.Fatalf("failed to register plugin count: %s", err)
	}
	InitQuotaLimit()
}

type Count struct {
	plugin.DefaultPlugin
}

func (c *Count) Name() string {
	return "count"
}

type CountConf struct {
}

func (c *Count) ParseConf(in []byte) (conf interface{}, err error) {
	return CountConf{}, nil
}

func (c *Count) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	fn := func() error {
		userIdStr := r.Header().Get(constants.RAINBOW_USER_ID_HEADER_KEY)
		appIdStr := r.Header().Get(constants.RAINBOW_APP_ID_HEADER_KEY)
		costTypeStr := r.Header().Get(constants.RAINBOW_COST_TYPE_HEADER_KEY)
		costCountStr := r.Header().Get(constants.RAINBOW_COST_COUNT_HEADER_KEY)
		reqId := r.Header().Get(constants.RAINBOW_REQUEST_ID_HEADER_KEY)

		log.Infof("userId: %v, costType: %v, costCount %v", userIdStr, costTypeStr, costCountStr)

		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			return errors.Wrapf(err, "failed to parse user id %s", userIdStr)
		}

		costCount, err := strconv.Atoi(costCountStr)
		if err != nil {
			return errors.Wrapf(err, "failed to parse cost count %s", costCountStr)
		}

		costType, err := enums.ParseCostType(costTypeStr)
		if err != nil {
			return errors.Wrapf(err, "failed to parse cost type")
		}

		// 如果rich标记为0，返回失败
		isRich, err := redis.CheckIsRich(uint(userId), *costType)
		if err != nil {
			return errors.Wrapf(err, "failed to check rich")
		}

		if !isRich {
			log.Infof("balance not enough when rich flag check, user %d cost type %s  ", userId, costType)
			return errors.New("balance not enough")
		}

		// 如果超过 quotalimit，返回失败
		pengdingCountKey := redis.UserPendingCountKey(userIdStr, costTypeStr)
		countKey := redis.UserCountKey(userIdStr, costTypeStr)

		currentPendingCount, err := redis.DB().GetIntOrDefault(context.Background(), pengdingCountKey).Result()
		if err != nil {
			return errors.Wrapf(err, "failed to get pending cost count")
		}
		currentCount, err := redis.DB().GetIntOrDefault(context.Background(), countKey).Result()
		if err != nil {
			return errors.Wrapf(err, "failed to get cost count")
		}
		log.Infof("currentCount %d, currentPendingCount %d, costCount %d", currentCount, currentPendingCount, costCount)

		// 不加pending，pending的表示未响应或在pre 插件执行中就失败的，比如被限流的
		if int(currentCount)+costCount > getQuotaLimit(*costType) {
			log.Infof("balance not enough when clac by un-settled count, user %d cost type %s  ", userId, costType)
			return errors.Errorf("balance not enough")
		}

		_, err = redis.DB().IncrBy(context.Background(), pengdingCountKey, int64(costCount)).Result()
		if err != nil {
			return errors.Wrapf(err, "failed to increase pending cost count")
		}

		reqKey, reqVal := redis.RequestKey(reqId), redis.RequestValue(userIdStr, appIdStr, costTypeStr, costCountStr)
		if _, err = redis.DB().Set(context.Background(), reqKey, reqVal, time.Minute*10).Result(); err != nil {
			return errors.Wrapf(err, "failed to cache request")
		}

		return nil
	}

	if err := fn(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed count for request: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}

func (c *Count) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	// log.Infof("get content-type %s", w.Header().Get("Content-Type"))
	// w.Header().Set("Content-Type", w.Header().Get("Content-Type"))
	log.Infof("in responsoe filter")
	DeterminCount(w, nil)
}

type GetSuccessCountHandler func(w pkgHTTP.Response) int

func DeterminCount(w pkgHTTP.Response, successCountHandler GetSuccessCountHandler) {

	reqId := w.Header().Get(constants.RAINBOW_REQUEST_ID_HEADER_KEY)
	// log.Infof("get x-rainbow-request-id %s", reqId)
	if reqId == "" {
		return
	}

	reqKey := redis.RequestKey(reqId)
	defer func() {
		_, err := redis.DB().Del(context.Background(), reqKey).Result()
		if err != nil {
			log.Errorf("failed to del key %d: %s", reqKey, err)
		}
	}()

	userId, _, costType, count, err := redis.GetRequest(reqId)
	if err != nil {
		log.Errorf("failed to get req %d val: %s", reqId, err)
		return
	}

	pengdingCountKey := redis.UserPendingCountKey(fmt.Sprintf("%d", userId), costType.String())
	countKey := redis.UserCountKey(fmt.Sprintf("%d", userId), costType.String())

	// 无论成功失败都减去pending count
	_, err = redis.DB().DecrBy(context.Background(), pengdingCountKey, int64(count)).Result()
	if err != nil {
		log.Errorf("failed to decrease pending cost count of req %d: %s", reqId, err)
		return
	}

	successCount := 0
	if successCountHandler != nil {
		successCount = successCountHandler(w)
	} else {
		if w.StatusCode() >= http.StatusOK && w.StatusCode() < http.StatusMultipleChoices {
			successCount = count
		}
	}

	if successCount <= 0 {
		return
	}

	// 请求成功，改变pending count 为 count
	// if w.StatusCode() >= http.StatusOK && w.StatusCode() < http.StatusMultipleChoices {
	_, err = redis.DB().IncrBy(context.Background(), countKey, int64(successCount)).Result()
	if err != nil {
		log.Errorf("failed to increase cost count of req %d: %s", reqId, err)
		return
	}
	// }
}
