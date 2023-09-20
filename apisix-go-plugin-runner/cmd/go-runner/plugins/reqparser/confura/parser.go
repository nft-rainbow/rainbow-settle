package rainbowapi

import (
	"encoding/json"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/openweb3/go-rpc-provider"
)

type ConfuraParserConf struct {
	IsMainnet bool `json:"is_mainnet,omitempty"`
	IsCspace  bool `json:"is_cspace,omitempty"`
}

// type ConfuraParserConf struct {
// 	IsMainnet bool `json:"is_mainnet,omitempty"`
// 	IsCspace  bool `json:"is_cspace,omitempty"`
// }

func (o *ConfuraParserConf) GetCostType() enums.CostType {
	if o.IsMainnet {
		if o.IsCspace {
			return enums.COST_TYPE_CONFURA_MAIN_CSPACE_NOMRAL
		}
		return enums.COST_TYPE_CONFURA_TEST_CSPACE_NOMRAL
	} else {
		if o.IsCspace {
			return enums.COST_TYPE_CONFURA_MAIN_ESPACE_NOMRAL
		}
		return enums.COST_TYPE_CONFURA_TEST_ESPACE_NOMRAL
	}
}

func (o *ConfuraParserConf) GetServerType() enums.ServerType {
	if o.IsCspace {
		return enums.SERVER_TYPE_CONFURA_CSPACE
	}
	return enums.SERVER_TYPE_CONFURA_ESPACE
}

func (o *ConfuraParserConf) ParseRequest(r pkgHTTP.Request) (*types.ReqParseResult, error) {
	body, err := r.Body()
	if err != nil {
		return nil, err
	}

	result := &types.ReqParseResult{
		CostType: o.GetCostType(),
		Count:    1,
	}

	var ms []rpc.JsonRpcMessage
	err = json.Unmarshal(body, &ms)
	if err == nil {
		result.Count = len(ms)
		return result, nil
	}

	var m rpc.JsonRpcMessage
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	return result, nil
}
