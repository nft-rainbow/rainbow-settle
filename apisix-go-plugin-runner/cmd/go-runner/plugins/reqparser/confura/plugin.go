package rainbowapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
)

func init() {
	err := plugin.RegisterPlugin(&ConfuraParser{})
	if err != nil {
		log.Fatalf("failed to register plugin rainbow_api_parser: %s", err)
	}
}

// Say is a demo to show how to return data directly instead of proxying
// it to the upstream.
type ConfuraParser struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

func (p *ConfuraParser) Name() string {
	return "confura-parser"
}

func (p *ConfuraParser) ParseConf(in []byte) (interface{}, error) {
	conf := ConfuraParserConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ConfuraParser) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	// log.Infof("in confura-parser request filter")
	c := conf.(ConfuraParserConf)
	result, err := types.DefaultRequestFilter(&c, w, r)
	if err != nil {
		return
	}

	// log.Infof("aaa")
	r.Header().Set(constants.RAINBOW_SERVER_TYPE_HEADER_KEY, c.GetServerType().String())

	// write ids to redis
	rpqId := r.Header().Get(constants.RAINBOW_REQUEST_ID_HEADER_KEY)
	rpcInfo := result.(*ConfuraParseResult).RpcInfo
	rpcInfoJson, _ := json.Marshal(rpcInfo)
	_, err = redis.DB().Set(context.Background(), redis.RpcIdsInfoKey(rpqId), rpcInfoJson, time.Minute).Result()
	if err != nil {
		log.Infof("failed set rpcids to redis: %v", err)
	}
	// log.Infof("bbb")
}
