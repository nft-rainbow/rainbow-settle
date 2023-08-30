package rainbowapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/testutils"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestPluginParseRainbowApiRequest(t *testing.T) {
	var p RainbowApiParser

	w := httptest.NewRecorder()
	r := testutils.HttpRequest{
		Method_: http.MethodGet,
		Path_:   []byte("http://localhost:8080/v1/mints/"),
		Header_: testutils.NewHttpHeader(),
	}
	p.RequestFilter(RainbowApiParserConf{}, w, &r)
	assert.Equal(t, "normal", r.Header().Get(constants.RAINBOW_COST_TYPE_HEADER_KEY))
}

func TestCostType(t *testing.T) {
	fmt.Println(enums.COST_TYPE_RAINBOW_NORMAL.String())
	fmt.Println(enums.COST_TYPE_RAINBOW_DEPLOY.String())
	fmt.Println(enums.COST_TYPE_RAINBOW_MINT.String())
}
