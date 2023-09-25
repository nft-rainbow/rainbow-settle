package rainbowapi

import (
	"encoding/json"
	"net/http"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/sirupsen/logrus"
)

func init() {
	err := plugin.RegisterPlugin(&ScanParser{})
	if err != nil {
		log.Fatalf("failed to register plugin scan-parser: %s", err)
	}
}

// Say is a demo to show how to return data directly instead of proxying
// it to the upstream.
type ScanParser struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

func (p *ScanParser) Name() string {
	return "scan-parser"
}

func (p *ScanParser) ParseConf(in []byte) (interface{}, error) {
	conf := ScanParserConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *ScanParser) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	// log.Infof("in scan-parser request filter")
	c := conf.(ScanParserConf)
	if _, err := types.DefaultRequestFilter(&c, w, r); err != nil {
		logrus.WithError(err).Error("failed to parse scan-parser request")
	}
	r.Header().Set(constants.RAINBOW_SERVER_TYPE_HEADER_KEY, c.GetServerType().String())
}
