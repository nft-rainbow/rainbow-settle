package rainbowapi

import (
	"encoding/json"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/openweb3/go-rpc-provider"
)

type ConfuraRequestOp struct {
	IsMainnet bool
}

func (o *ConfuraRequestOp) ParseRequest(r pkgHTTP.Request) (*types.ReqParseResult, error) {
	body, err := r.Body()
	if err != nil {
		return nil, err
	}

	result := &types.ReqParseResult{
		CostType: enums.COST_TYPE_CONFURA_MAIN_CSPACE_NOMRAL,
		Count:    1,
	}

	if !o.IsMainnet {
		result.CostType = enums.COST_TYPE_CONFURA_TEST_CSPACE_NOMRAL
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
