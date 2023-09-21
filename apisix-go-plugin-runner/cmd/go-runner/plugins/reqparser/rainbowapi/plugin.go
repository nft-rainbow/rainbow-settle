package rainbowapi

import (
	"net/http"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

var (
	o RainbowApiRequestOp
)

func init() {
	err := plugin.RegisterPlugin(&RainbowApiParser{})
	if err != nil {
		log.Fatalf("failed to register plugin rainbow_api_parser: %s", err)
	}
}

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
	return "rainbow-api-parser"
}

func (p *RainbowApiParser) ParseConf(in []byte) (interface{}, error) {
	return RainbowApiParserConf{}, nil
}

func (p *RainbowApiParser) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	types.DefaultRequestFilter(&o, w, r)
	r.Header().Set(constants.RAINBOW_SERVER_TYPE_HEADER_KEY, enums.SERVER_TYPE_RAINBOW.String())
}
