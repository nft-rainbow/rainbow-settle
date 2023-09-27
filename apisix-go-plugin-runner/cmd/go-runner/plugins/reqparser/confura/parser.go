package rainbowapi

import (
	"encoding/json"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/openweb3/go-rpc-provider"
	"github.com/samber/lo"
)

type ConfuraParserConf struct {
	IsMainnet bool `json:"is_mainnet,omitempty"`
	IsCspace  bool `json:"is_cspace,omitempty"`
}

func (o *ConfuraParserConf) GetCostType() enums.CostType {
	if o.IsMainnet {
		if o.IsCspace {
			return enums.COST_TYPE_CONFURA_MAIN_CSPACE_NORMAL
		}
		return enums.COST_TYPE_CONFURA_MAIN_ESPACE_NORMAL
	} else {
		if o.IsCspace {
			return enums.COST_TYPE_CONFURA_TEST_CSPACE_NORMAL
		}
		return enums.COST_TYPE_CONFURA_TEST_ESPACE_NORMAL
	}
}

func (o *ConfuraParserConf) GetServerType() enums.ServerType {
	if o.IsCspace {
		return enums.SERVER_TYPE_CONFURA_CSPACE
	}
	return enums.SERVER_TYPE_CONFURA_ESPACE
}

type ConfuraParseResult struct {
	types.DefaultReqParseResult
	redis.RpcInfo
}

func (o *ConfuraParserConf) ParseRequest(r pkgHTTP.Request) (types.ReqParseResult, error) {
	body, err := r.Body()
	if err != nil {
		return nil, err
	}

	result := ConfuraParseResult{
		DefaultReqParseResult: types.DefaultReqParseResult{
			CostType: o.GetCostType(),
			Count:    1,
		},
	}

	var ms []rpc.JsonRpcMessage
	err = json.Unmarshal(body, &ms)
	if err == nil {
		result.Count = len(ms)
		result.Items = lo.Map(ms, func(item rpc.JsonRpcMessage, index int) *redis.RpcInfoItem {
			return &redis.RpcInfoItem{string(item.ID), item.Version}
		})
		result.IsBatchRpc = true
		return &result, nil
	}

	var m rpc.JsonRpcMessage
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	result.Items = []*redis.RpcInfoItem{{string(m.ID), m.Version}}
	return &result, nil
}
