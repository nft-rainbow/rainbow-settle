package rainbowapi

import (
	"encoding/json"
	"net/http"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
)

var (
	o ConfuraRequestOp
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
	IsMainnet bool `json:"is_mainnet,omitempty"`
	IsCspace  bool `json:"is_cspace,omitempty"`
}

func (p *ConfuraParser) Name() string {
	return "confura_parser"
}

func (p *ConfuraParser) ParseConf(in []byte) (interface{}, error) {
	conf := ConfuraParserConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ConfuraParser) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	_conf := conf.(ConfuraParserConf)
	o.IsMainnet = _conf.IsMainnet
	types.DefaultRequestFilter(&o, w, r)
	// convert url
	// costTypeStr := r.Header().Get(constants.RAINBOW_COST_TYPE_HEADER_KEY)
	// costType, _ := enums.ParseCostType(costTypeStr)
	// serverType, _ := costType.ServerType()

	// if _conf.IsMainnet {
	// 	r.SetPath([]byte("https://main.confluxrpc.com/6G5LxkA1P3EXMfpArsPBBRxDL8GJk78ceeVRXDSBSwaxDab3YyyKLLZRE4NF6gQqejPoxfNsmJ4wBBJwdwGS4Vg8T"))
	// } else {
	// 	r.SetPath([]byte("https://test.confluxrpc.com/6G5LxkA1P3EXMfpArsPBBRxDL8GJk78ceeVRXDSBSwaxDab3YyyKLLZRE4NF6gQqejPoxfNsmJ4wBBJwdwGS4Vg8T"))
	// }
}
