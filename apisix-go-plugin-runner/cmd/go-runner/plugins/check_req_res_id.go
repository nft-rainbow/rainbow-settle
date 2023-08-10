package plugins

import (
	"fmt"
	"net/http"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
)

func init() {
	err := plugin.RegisterPlugin(&CheckId{})
	if err != nil {
		log.Fatalf("failed to register plugin check_id: %s", err)
	}
}

// Say is a demo to show how to return data directly instead of proxying
// it to the upstream.
type CheckId struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type CheckIdConf struct {
}

func (p *CheckId) Name() string {
	return "check_id"
}

func (p *CheckId) ParseConf(in []byte) (interface{}, error) {
	return CheckIdConf{}, nil
}

func (p *CheckId) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	// log.Infof("run request filter, r.ID() %d", r.ID())
	// r.Header().Set("ID", fmt.Sprintf("%d", r.ID()))
	xrIDInReq := r.Header().Get("X-Request-Id")
	fmt.Printf("xrIDInResp %v", xrIDInReq)
}

func (p *CheckId) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	log.Infof("run response filter, w.ID() %d", w.ID())
	// rID := w.Header().Get("rID")
	// xrID := w.Header().Get("xRID")
	xrIDInResp := w.Header().Get("X-Request-Id")
	// w.Write([]byte(fmt.Sprintf("rID %s, wID %d, xrID %v, xrIDInResp %v", rID, w.ID(), xrID, xrIDInResp)))
	w.Write([]byte(fmt.Sprintf("xrIDInResp %v", xrIDInResp)))

	// rsp, ok := w.(*thisHttp.Response)
	// if !ok {
	// 	w.Write([]byte("not internal_http.Response"))
	// 	return
	// }

}
