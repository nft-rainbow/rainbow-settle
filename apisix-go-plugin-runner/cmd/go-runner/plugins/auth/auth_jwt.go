package auth

import (
	"encoding/json"
	"strings"

	"fmt"
	"net/http"
	"net/url"
	"time"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// 备注：
// 1. 路由设计与现rainbow-api 使用cost middleware的一致，如 login/refresh 不计数，不验证token

func init() {
	err := plugin.RegisterPlugin(&JwtAuth{})
	if err != nil {
		log.Fatalf("failed to register plugin jwt-auth: %s", err)
	}
}

type JwtAuthConf struct {
	TokenLookup string `json:"token_lookup"`
	JwtKey      string `json:"jwt_key"`
	// APP         string `json:"app"` // rainbow-api, rainbow-dashboard, rainbow-admin
	// Env         string `json:"env"` // prod, env, local
}

func (j *JwtAuthConf) getJwtKey() string {
	return j.JwtKey

	// var keys = map[string]string{
	// 	"rainbow-api-prod":  "",
	// 	"rainbow-api-dev":   "jwt-openapi-key",
	// 	"rainbow-api-local": "jwt-openapi-key",

	// 	"rainbow-dashboard-prod":  "",
	// 	"rainbow-dashboard-dev":   "jwt-dashboard-key",
	// 	"rainbow-dashboard-local": "jwt-dashboard-key",

	// 	"rainbow-admin-prod":  "",
	// 	"rainbow-admin-dev":   "jwt-admin-key",
	// 	"rainbow-admin-local": "jwt-ddmin-key",
	// }
	// val, ok := keys[j.appEnvString()]
	// return val, ok
}

// func (j *JwtAuthConf) appEnvString() string {
// 	return fmt.Sprintf("%v-%v", j.APP, j.Env)
// }

type JwtAuth struct {
	plugin.DefaultPlugin
	jwtAuthMiddlewares map[string]*jwt.GinJWTMiddleware
}

func (c *JwtAuth) Name() string {
	return "jwt-auth"
}

func (c *JwtAuth) ParseConf(in []byte) (interface{}, error) {
	// logrus.WithField("stack", string(debug.Stack())).WithField("in", string(in)).Info("parse conf")
	conf := JwtAuthConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (c *JwtAuth) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	logrus.WithField("conf", conf).Info("request filter")
	// parse jwt
	fn := func() error {
		jwtAuthConf := conf.(JwtAuthConf)
		if len(jwtAuthConf.TokenLookup) == 0 {
			// w.WriteHeader(http.StatusBadRequest)
			// _, err := w.Write([]byte("must specity auth place, must be one of header/query/path"))
			// if err != nil {
			// 	log.Errorf("failed to write: %s", err)
			// }
			// return nil
			return errors.New("must specity auth place, must be one of header/query/path")
		}

		jwtMid, err := c.getJwtMiddleware(jwtAuthConf)
		if err != nil {
			return err
		}

		ctx, _ := gin.CreateTestContext(w)
		urlStr := string(r.Path())
		if len(r.Args()) > 0 {
			urlStr = fmt.Sprintf("%s?%s", urlStr, r.Args().Encode())
		}
		url, _ := url.Parse(urlStr)
		ctx.Request = &http.Request{
			Header: r.Header().View(),
			Method: r.Method(),
			URL:    url,
		}

		jwtMid.MiddlewareFunc()(ctx)
		if ctx.Writer.Status() >= http.StatusBadRequest {
			return nil
		}

		userId, appId, err := extractUserInfoFromJwt(jwtMid, ctx)
		if err != nil {
			return err
		}
		r.Header().Set(constants.RAINBOW_USER_ID_HEADER_KEY, userId)
		r.Header().Set(constants.RAINBOW_APP_ID_HEADER_KEY, appId)
		return nil
	}

	if err := fn(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed check auth: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}

func (j *JwtAuth) getJwtMiddleware(conf JwtAuthConf) (*jwt.GinJWTMiddleware, error) {
	if j.jwtAuthMiddlewares == nil {
		j.jwtAuthMiddlewares = make(map[string]*jwt.GinJWTMiddleware)
	}

	jwtKey := conf.getJwtKey()

	if jwtKey == "" {
		return nil, errors.Errorf("missing JWT key")
	}

	if j.jwtAuthMiddlewares[conf.getJwtKey()] == nil {
		timeout := time.Hour
		jwtMid, err := jwt.New(&jwt.GinJWTMiddleware{
			Key:           []byte(jwtKey),
			Timeout:       timeout,
			MaxRefresh:    time.Hour * 5,
			TokenLookup:   conf.TokenLookup,
			TokenHeadName: "Bearer",
			TimeFunc:      time.Now,
		})
		if err != nil {
			return nil, err
		}
		j.jwtAuthMiddlewares[jwtKey] = jwtMid
	}
	return j.jwtAuthMiddlewares[jwtKey], nil
}

// TODO: support rainbow-dashboard
func extractUserInfoFromJwt(jwtMid *jwt.GinJWTMiddleware, ctx *gin.Context) (string, string, error) {
	claims, err := jwtMid.GetClaimsFromJWT(ctx)
	if err != nil {
		return "", "", err
	}

	if strings.HasPrefix(ctx.Request.URL.Path, "/v1") {
		log.Infof("claims: %v", claims)
		return fmt.Sprintf("%v", claims["AppUserId"]), fmt.Sprintf("%v", claims["id"]), nil
	}

	if strings.HasPrefix(ctx.Request.URL.Path, "/dashboard") {
		userId := fmt.Sprintf("%v", claims["id"])
		appId := "0"
		if strings.HasPrefix(ctx.Request.URL.Path, "/dashboard/apps/") {
			s := strings.TrimPrefix(ctx.Request.URL.Path, "/dashboard/apps/")
			len := strings.Index(s, "/")
			var id string
			if len == -1 {
				id = s
			} else {
				id = s[:len]
			}
			appId = id
		}
		return userId, appId, nil
	}

	if strings.HasPrefix(ctx.Request.URL.Path, "/admin") {
		return fmt.Sprintf("%v", claims["id"]), fmt.Sprintf("%d", 0), nil
	}
	return "", "", fmt.Errorf("unsupport get user id of path %s", ctx.Request.URL.Path)
}
