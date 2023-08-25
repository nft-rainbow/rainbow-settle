package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/testutils"
	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/rainbow-api/services"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestGinContextFullpath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	url, _ := url.Parse(fmt.Sprintf("%v%v", "https://www.baidu.com/mints/1", "?version=1"))
	c.Request = &http.Request{
		Method: "POST",
		URL:    url,
	}

	engine := gin.Default()
	engine.POST("mints", gin.Default().Handlers.Last())
	engine.POST("mints/:id", gin.Default().Handlers.Last())
	engine.HandleContext(c)

	fmt.Println(len(gin.Default().Handlers))
	fmt.Println(c.FullPath())
}

func TestUrlParse(t *testing.T) {
	url, _ := url.Parse(fmt.Sprintf("%v?%v", "http://www.baidu.com", ""))
	fmt.Println(url)
}

func TestParseRainbowApiRequest(t *testing.T) {
	body, _ := json.Marshal(services.CustomMintDto{
		ContractInfoDtoWithoutType: services.ContractInfoDtoWithoutType{
			Chain:           "conflux",
			ContractAddress: "cfx:aamjy3abae3j0ud8ys0npt38ggnunk5r4ps2pg8vcc",
		},
		MintItemDto: services.MintItemDto{
			MintToAddress: "cfx:aamjy3abae3j0ud8ys0npt38ggnunk5r4ps2pg8vcc",
		},
	})
	req := testutils.HttpRequest{
		Method_: http.MethodPost,
		Path_:   []byte("http://localhost:8080/v1/mints/"),
		Header_: testutils.NewHttpHeader(),
		Body_:   body,
	}

	result, err := ParseRainbowApiRequest(&req)
	assert.NoError(t, err)

	assert.Equal(t, true, result.IsMainnet)
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, enums.COST_TYPE_RAINBOW_MINT, result.CostType)

	// fmt.Println(result.CostType.String())

}
