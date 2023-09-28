package respformat

import (
	"encoding/json"
	"net/http"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/count"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/openweb3/go-rpc-provider"
	"github.com/samber/lo"
)

func init() {
	err := plugin.RegisterPlugin(&RpcRespFormat{})
	if err != nil {
		log.Fatalf("failed to register plugin rpc-resp-format: %s", err)
	}
}

// RpcRespFormat is a demo to show how to return data directly instead of proxying
// it to the upstream.
type RpcRespFormat struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type RpcRespFormatConf struct {
}

func (p *RpcRespFormat) Name() string {
	return "rpc-resp-format"
}

func (p *RpcRespFormat) ParseConf(in []byte) (interface{}, error) {
	conf := RpcRespFormatConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (c *RpcRespFormat) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	// NOTE: 这里置count是由于 1.apisix ext-plugin-post-resp 不支持多个 2.status ok，而rpc返回错误的不扣费
	determineCount(w)

	log.Infof("in rpc-resp-format response filter, status code: %d", w.StatusCode())
	if w.StatusCode() < http.StatusBadRequest {
		return
	}
	// log.Infof("aaa")
	reqId := w.Header().Get(constants.RAINBOW_REQUEST_ID_HEADER_KEY)
	rpcIdsInfo, err := redis.GetRpcIdsInfo(reqId)
	if err != nil {
		log.Errorf("failed to get rpc ids from redis: %v", err)
		return
	}
	// log.Infof("bbb")
	body, err := w.ReadBody()
	if err != nil {
		log.Errorf("failed to read body: %v", err)
		return
	}
	// log.Infof("ccc")

	rpcResps := lo.Map(rpcIdsInfo.Items, func(item *redis.RpcInfoItem, index int) rpc.JsonRpcMessage {
		return rpc.JsonRpcMessage{
			Version: item.RpcVersion,
			ID:      json.RawMessage(item.RpcId),
			Error: &rpc.JsonError{
				Code:    -32001,
				Message: string(body),
			},
		}
	})

	w.WriteHeader(http.StatusOK)
	if rpcIdsInfo.IsBatchRpc {
		body, _ = json.Marshal(rpcResps)
		w.Write(body)
		return
	} else {
		body, _ = json.Marshal(rpcResps[0])
		w.Write(body)
	}
	// log.Infof("ddd")
}

func determineCount(w pkgHTTP.Response) {
	count.DeterminCount(w, func(w pkgHTTP.Response) int {
		if w.StatusCode() != http.StatusOK {
			return 0
		}

		body, err := w.ReadBody()
		if err != nil {
			log.Errorf("failed to read body: %v", err)
			return 0
		}

		var successCount int
		var resps []*rpc.JsonRpcMessage
		if err := json.Unmarshal(body, &resps); err == nil {
			successCount = len(lo.Filter(resps, func(item *rpc.JsonRpcMessage, index int) bool {
				return item.Error == nil
			}))
			return successCount
		}

		var resp rpc.JsonRpcMessage
		if err = json.Unmarshal(body, &resp); err != nil {
			log.Infof("failed unmarshal rpc response")
			return 0
		}
		return 1
	})
}
