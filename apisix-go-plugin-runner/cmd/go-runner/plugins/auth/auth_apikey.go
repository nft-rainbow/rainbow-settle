package auth

import (
	"encoding/json"
	"strings"

	"fmt"
	"net/http"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// 备注：
// 1. 路由设计与现rainbow-api 使用cost middleware的一致，如 login/refresh 不计数，不验证token

func init() {
	err := plugin.RegisterPlugin(&ApikeyAuth{})
	if err != nil {
		log.Fatalf("failed to register plugin apikey-auth: %s", err)
	}
}

type ApikeyAuthConf struct {
	Lookup string `json:"lookup"`
}

func (a *ApikeyAuthConf) ExtractApiKey(r pkgHTTP.Request) (string, error) {
	path, args := r.Path(), r.Args()
	switch a.Lookup {
	case "path":
		items := strings.Split(string(path), "/")
		if len(items) < 1 {
			return "", fmt.Errorf("missing apikey")
		}
		return items[len(items)-1], nil

	case "query":
		if !args.Has(constants.RAINBOW_APIKEY_KEY) {
			return "", fmt.Errorf("missing apikey")
		}
		return args.Get(constants.RAINBOW_APIKEY_KEY), nil

	case "header":
		apikey := r.Header().Get(constants.RAINBOW_APIKEY_KEY)
		if len(apikey) == 0 {
			return "", fmt.Errorf("missing apikey")
		}
		return apikey, nil
	default:
		return "", errors.Errorf("not support lookup: %s", a.Lookup)
	}
}

type ApikeyAuth struct {
	plugin.DefaultPlugin
}

func (c *ApikeyAuth) Name() string {
	return "apikey-auth"
}

func (c *ApikeyAuth) ParseConf(in []byte) (interface{}, error) {
	// logrus.WithField("stack", string(debug.Stack())).WithField("in", string(in)).Info("parse conf")
	conf := ApikeyAuthConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (c *ApikeyAuth) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	logrus.WithField("conf", conf).Info("request filter")
	// parse jwt
	fn := func() error {
		ApikeyAuthConf := conf.(ApikeyAuthConf)
		if len(ApikeyAuthConf.Lookup) == 0 {
			return errors.New("must specity auth place, must be one of header/query/path")
		}

		apikey, err := ApikeyAuthConf.ExtractApiKey(r)
		if err != nil {
			return err
		}

		if apikey == "" {
			return errors.New("missing api key")
		}

		userId, appId, err := redis.GetUserInfoByApikey(apikey)
		if err != nil {
			return err
		}
		log.Infof("get user info from redis: %d,%d,%v", userId, appId, err)
		r.Header().Set(constants.RAINBOW_USER_ID_HEADER_KEY, fmt.Sprintf("%d", userId))
		r.Header().Set(constants.RAINBOW_APP_ID_HEADER_KEY, fmt.Sprintf("%d", appId))
		return nil
	}

	if err := fn(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed check auth: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}
