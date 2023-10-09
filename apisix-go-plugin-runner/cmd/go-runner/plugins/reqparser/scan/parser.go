package rainbowapi

import (
	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

type ScanParserConf struct {
	IsMainnet bool `json:"is_mainnet,omitempty"`
	IsCspace  bool `json:"is_cspace,omitempty"`
}

func (o *ScanParserConf) GetCostType() enums.CostType {
	if o.IsMainnet && o.IsCspace {
		return enums.COST_TYPE_SCAN_MAIN_CSPACE_NORMAL
	}

	if o.IsMainnet && !o.IsCspace {
		return enums.COST_TYPE_SCAN_MAIN_ESPACE_NORMAL
	}

	if !o.IsMainnet && o.IsCspace {
		return enums.COST_TYPE_SCAN_TEST_CSPACE_NORMAL
	}

	return enums.COST_TYPE_SCAN_TEST_ESPACE_NORMAL
}

func (o *ScanParserConf) GetServerType() enums.ServerType {
	if o.IsCspace {
		return enums.SERVER_TYPE_SCAN_CSPACE
	}
	return enums.SERVER_TYPE_SCAN_ESPACE
}

type ScanParseResult struct {
	types.DefaultReqParseResult
}

func (o *ScanParserConf) ParseRequest(r pkgHTTP.Request) (types.ReqParseResult, error) {
	result := ScanParseResult{
		DefaultReqParseResult: types.DefaultReqParseResult{
			CostType: o.GetCostType(),
			Count:    1,
		},
	}
	return &result, nil
}
