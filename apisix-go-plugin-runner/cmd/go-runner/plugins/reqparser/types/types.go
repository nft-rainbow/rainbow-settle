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
	ParseRequest(r pkgHTTP.Request) (*ReqParseResult, error)
}

type ReqParseResult struct {
	CostType enums.CostType
	Count    int
}

type DefaultParserPlugin struct {
}

func DefaultRequestFilter(o Parser, w http.ResponseWriter, r pkgHTTP.Request) {
	fn := func() error {
		result, err := o.ParseRequest(r)
		if err != nil {
			return err
		}

		r.Header().Set(constants.RAINBOW_COST_TYPE_HEADER_KEY, result.CostType.String())
		r.Header().Set(constants.RAINBOW_COST_COUNT_HEADER_KEY, fmt.Sprintf("%d", result.Count))
		r.Header().Set(constants.RAINBOW_REQUEST_ID_HEADER_KEY, uuid.New().String())

		return nil
	}

	if err := fn(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed parse rainbow request: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}
