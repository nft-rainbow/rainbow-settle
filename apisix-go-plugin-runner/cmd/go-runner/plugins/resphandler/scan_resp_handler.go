package resphandler

import (
	"encoding/json"
	"net/http"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/count"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
)

func init() {
	err := plugin.RegisterPlugin(&ScanRespHandler{})
	if err != nil {
		log.Fatalf("failed to register plugin scan-resp-handler: %s", err)
	}
}

// ScanRespHandler is a demo to show how to return data directly instead of proxying
// it to the upstream.
type ScanRespHandler struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type ScanRespHandlerConf struct {
}

func (p *ScanRespHandler) Name() string {
	return "scan-resp-handler"
}

func (p *ScanRespHandler) ParseConf(in []byte) (interface{}, error) {
	conf := ScanRespHandlerConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (c *ScanRespHandler) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	// NOTE: 这里置count是由于 1.apisix ext-plugin-post-resp 不支持多个 2.status ok，而scan返回错误的不扣费
	c.determineCount(w)
}

type ScanBody struct {
	Code int `json:"code"`
}

func (c *ScanRespHandler) determineCount(w pkgHTTP.Response) {
	count.DeterminCount(w, func(w pkgHTTP.Response) int {
		if w.StatusCode() != http.StatusOK {
			return 0
		}

		body, err := w.ReadBody()
		if err != nil {
			log.Errorf("failed to read body: %v", err)
			return 0
		}

		var resp ScanBody
		if err = json.Unmarshal(body, &resp); err != nil {
			log.Infof("failed unmarshal rpc response")
			return 0
		}
		if resp.Code == 0 || resp.Code == 1 {
			return 1
		}
		return 0
	})
}
