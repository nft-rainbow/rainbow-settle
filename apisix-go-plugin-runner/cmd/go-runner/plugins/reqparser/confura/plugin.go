package rainbowapi

import (
	"fmt"
	"net/http"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/google/uuid"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
)

var (
	o ConfuraOp
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

type ConfuraParserConf struct {
}

func (p *ConfuraParser) Name() string {
	return "confura_parser"
}

func (p *ConfuraParser) ParseConf(in []byte) (interface{}, error) {
	return ConfuraParserConf{}, nil
}

func (p *ConfuraParser) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {

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
		if _, err := w.Write([]byte(fmt.Sprintf("failed parse rainbow api request: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}
