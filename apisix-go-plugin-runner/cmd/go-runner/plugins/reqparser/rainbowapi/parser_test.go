package rainbowapi

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
	o := RainbowApiRequestOp{}
	mints1Body, _ := json.Marshal(services.CustomMintDto{
		ContractInfoDtoWithoutType: services.ContractInfoDtoWithoutType{
			Chain:           "conflux",
			ContractAddress: "cfx:aamjy3abae3j0ud8ys0npt38ggnunk5r4ps2pg8vcc",
		},
		MintItemDto: services.MintItemDto{
			MintToAddress: "cfx:aamjy3abae3j0ud8ys0npt38ggnunk5r4ps2pg8vcc",
		},
	})
	mints1Req := testutils.HttpRequest{
		Method_: http.MethodPost,
		Path_:   []byte("http://localhost:8080/v1/mints/"),
		Header_: testutils.NewHttpHeader(),
		Body_:   mints1Body,
	}

	mints2Body, _ := json.Marshal(services.CustomMintDto{
		ContractInfoDtoWithoutType: services.ContractInfoDtoWithoutType{
			Chain:           "conflux_test",
			ContractAddress: "cfxtest:acfgbf21bj612uth2xekuj5xh8cmgbj56j3fawd5c2",
		},
		MintItemDto: services.MintItemDto{
			MintToAddress: "cfxtest:acfgbf21bj612uth2xekuj5xh8cmgbj56j3fawd5c2",
		},
	})
	mints2Req := testutils.HttpRequest{
		Method_: http.MethodPost,
		Path_:   []byte("http://localhost:8080/v1/mints/"),
		Header_: testutils.NewHttpHeader(),
		Body_:   mints2Body,
	}

	deploy1Body, _ := json.Marshal(services.ContractDeployDto{
		Chain:  "conflux",
		Name:   "xxx",
		Symbol: "xxx",
		Type:   "erc721",
	})
	deploy1Req := testutils.HttpRequest{
		Method_: http.MethodPost,
		Path_:   []byte("http://localhost:8080/dashboard/apps/7/contracts"),
		Header_: testutils.NewHttpHeader(),
		Body_:   deploy1Body,
	}

	table := []struct {
		Req       testutils.HttpRequest
		IsMainNet bool
		Count     int
		CostType  enums.CostType
	}{
		{
			Req:       mints1Req,
			IsMainNet: true,
			Count:     1,
			CostType:  enums.COST_TYPE_RAINBOW_MINT,
		},
		{
			Req:       mints2Req,
			IsMainNet: false,
			Count:     1,
			CostType:  enums.COST_TYPE_RAINBOW_NORMAL,
		},
		{
			Req:       deploy1Req,
			IsMainNet: false,
			Count:     1,
			CostType:  enums.COST_TYPE_RAINBOW_DEPLOY,
		},
	}

	for i, item := range table {
		result, err := o.ParseRequest(&item.Req)
		assert.NoError(t, err)

		// assert.Equal(t, item.IsMainNet, result.IsMainnet, i)
		assert.Equal(t, item.Count, result.GetCount(), i)
		assert.Equal(t, item.CostType, result.GetCostType(), i)
	}

}
