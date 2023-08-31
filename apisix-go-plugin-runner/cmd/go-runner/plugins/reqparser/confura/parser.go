package rainbowapi

import (
	"encoding/json"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/openweb3/go-rpc-provider"
)

type ConfuraOp struct {
}

func (o *ConfuraOp) ParseRequest(r pkgHTTP.Request) (*types.ReqParseResult, error) {
	body, err := r.Body()
	if err != nil {
		return nil, err
	}

	var ms []rpc.JsonRpcMessage
	err = json.Unmarshal(body, &ms)
	if err == nil {
		return &types.ReqParseResult{enums.COST_TYPE_CONFURA_NOMRAL, len(ms)}, nil
	}

	var m rpc.JsonRpcMessage
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}
	return &types.ReqParseResult{enums.COST_TYPE_CONFURA_NOMRAL, 1}, nil
}
