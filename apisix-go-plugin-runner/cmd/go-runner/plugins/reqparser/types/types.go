package types

import (
	"fmt"
	"net/http"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/google/uuid"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

type Parser interface {
	ParseRequest(r pkgHTTP.Request) (ReqParseResult, error)
}

type ReqParseResult interface {
	GetCostType() enums.CostType
	GetCount() int
}

type DefaultReqParseResult struct {
	CostType enums.CostType
	Count    int
}

func (d *DefaultReqParseResult) GetCostType() enums.CostType { return d.CostType }
func (d *DefaultReqParseResult) GetCount() int               { return d.Count }

func DefaultRequestFilter(o Parser, w http.ResponseWriter, r pkgHTTP.Request) (ReqParseResult, error) {
	fn := func() (ReqParseResult, error) {
		result, err := o.ParseRequest(r)
		if err != nil {
			return nil, err
		}

		r.Header().Set(constants.RAINBOW_COST_TYPE_HEADER_KEY, result.GetCostType().String())
		r.Header().Set(constants.RAINBOW_COST_COUNT_HEADER_KEY, fmt.Sprintf("%d", result.GetCount()))
		r.Header().Set(constants.RAINBOW_REQUEST_ID_HEADER_KEY, uuid.New().String())

		return result, nil
	}

	result, err := fn()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed parse rainbow request: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
	return result, err
}
