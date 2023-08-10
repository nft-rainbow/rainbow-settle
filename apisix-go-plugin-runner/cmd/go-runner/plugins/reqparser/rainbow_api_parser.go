package reqparser

import (
	"fmt"
	"net/http"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/core"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
)

func init() {
	err := plugin.RegisterPlugin(&RainbowApiParser{})
	if err != nil {
		log.Fatalf("failed to register plugin rainbow_api_parser: %s", err)
	}
}

const (
	RAINBOW_COST_TYPE_HEADER_KEY  = "x-rainbow-cost-type"
	RAINBOW_COST_COUNT_HEADER_KEY = "x-rainbow-cost-count"
)

// Say is a demo to show how to return data directly instead of proxying
// it to the upstream.
type RainbowApiParser struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type RainbowApiParserConf struct {
}

func (p *RainbowApiParser) Name() string {
	return "rainbow_api_parser"
}

func (p *RainbowApiParser) ParseConf(in []byte) (interface{}, error) {
	return RainbowApiParserConf{}, nil
}

func (p *RainbowApiParser) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {

	fn := func() error {
		result, err := core.ParseRainbowApiRequest(r)
		if err != nil {
			return err
		}

		// log.Infof("result %v", result)
		r.Header().Set(RAINBOW_COST_TYPE_HEADER_KEY, result.CostType.String())
		r.Header().Set(RAINBOW_COST_COUNT_HEADER_KEY, fmt.Sprintf("%d", result.Count))

		return nil
	}

	if err := fn(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed parse rainbow api request: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}
