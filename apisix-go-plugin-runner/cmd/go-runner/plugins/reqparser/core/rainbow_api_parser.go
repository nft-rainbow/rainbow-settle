package core

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nft-rainbow/rainbow-api/services"
	"github.com/nft-rainbow/rainbow-api/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

var (
	rainbowApiEngine *gin.Engine
)

func init() {
	rainbowApiEngine = gin.New()
	emptyHandler := gin.HandlerFunc(func(c *gin.Context) {})
	for k := range utils.MintPaths {
		rainbowApiEngine.POST(k, emptyHandler)
	}

	for k := range utils.DeployPaths {
		rainbowApiEngine.POST(k, emptyHandler)
	}
}

type RainbowApiReqParseResult struct {
	IsMainnet bool
	CostType  enums.CostType
	Count     int
}

func ParseRainbowApiRequest(r pkgHTTP.Request) (*RainbowApiReqParseResult, error) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	urlStr := string(r.Path())
	if len(r.Args()) > 0 {
		urlStr = fmt.Sprintf("%s?%s", urlStr, r.Args().Encode())
	}
	url, _ := url.Parse(urlStr)

	c.Request = &http.Request{
		Header: r.Header().View(),
		Method: r.Method(),
		URL:    url,
	}
	// 此处主要用于解析fullpath
	rainbowApiEngine.HandleContext(c)

	if getAction(c.Request.Method, c.FullPath()) == enums.COST_TYPE_RAINBOW_NORMAL {
		return &RainbowApiReqParseResult{false, enums.COST_TYPE_RAINBOW_NORMAL, 1}, nil
	}

	body, err := r.Body()
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return parseRainbowApiRequestByGin(c)
}

func parseRainbowApiRequestByGin(c *gin.Context) (*RainbowApiReqParseResult, error) {
	result := RainbowApiReqParseResult{
		CostType: getAction(c.Request.Method, c.FullPath()),
		Count:    1,
	}

	switch c.FullPath() {
	case "/v1/mints/":
		var mintMeta services.CustomMintDto
		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(mintMeta.Chain)
		return &result, err

	case "/v1/mints/customizable":
		var mintMeta services.CustomMintDto
		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(mintMeta.Chain)
		return &result, err

	case "/v1/mints/customizable/batch":
		var mintMeta services.CustomMintBatchDto
		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(mintMeta.Chain)
		return &result, err

	case "/v1/mints/easy/files":
		var mintMeta services.EasyMintFileDto
		err := c.ShouldBind(&mintMeta)
		result.IsMainnet = utils.IsMainnetByName(mintMeta.Chain)
		return &result, err

	case "/v1/mints/easy/urls":
		var mintMeta services.EasyMintMetaPartsDto
		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(mintMeta.Chain)
		return &result, err

	case "/dashboard/apps/:id/contracts":
		fallthrough
	case "/v1/contracts/":
		var contractDeployDto services.ContractDeployDto
		err := c.ShouldBindBodyWith(&contractDeployDto, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(contractDeployDto.Chain)
		return &result, err

	case "/v1/transfers/customizable":
		var transferMeta services.TransferDto
		err := c.ShouldBindBodyWith(&transferMeta, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(transferMeta.Chain)
		return &result, err

	case "/v1/transfers/customizable/batch":
		var transferBatchMeta services.TransferBatchDto
		err := c.ShouldBindBodyWith(&transferBatchMeta, binding.JSON)
		result.IsMainnet = utils.IsMainnetByName(transferBatchMeta.Chain)
		result.Count = len(transferBatchMeta.Items)
		return &result, err
	}

	return &result, nil
}

func getAction(method, fullPath string) enums.CostType {
	isMint := utils.IsMint(method, fullPath)
	isDeploy := utils.IsDeploy(method, fullPath)
	if isMint {
		return enums.COST_TYPE_RAINBOW_MINT
	}
	if isDeploy {
		return enums.COST_TYPE_RAINBOW_DEPLOY
	}
	return enums.COST_TYPE_RAINBOW_NORMAL
}

// func IsMainNet(c *gin.Context) (bool, uint, error) {
// 	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
// 		return false, 1, nil
// 	}

// 	if !utils.IsMint(c.Request.Method, c.FullPath()) && !utils.IsDeploy(c.Request.Method, c.FullPath()) {
// 		return false, 1, nil
// 	}

// 	switch c.FullPath() {
// 	case "/v1/mints/":
// 		var mintMeta services.CustomMintDto
// 		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
// 		return utils.IsMainnetByName(mintMeta.Chain), 1, err

// 	case "/v1/mints/customizable":
// 		var mintMeta services.CustomMintDto
// 		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
// 		return utils.IsMainnetByName(mintMeta.Chain), 1, err

// 	case "/v1/mints/customizable/batch":
// 		var mintMeta services.CustomMintBatchDto
// 		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
// 		return utils.IsMainnetByName(mintMeta.Chain), uint(len(mintMeta.MintItems)), err

// 	case "/v1/mints/easy/files":
// 		var mintMeta services.EasyMintFileDto
// 		err := c.ShouldBind(&mintMeta)
// 		return utils.IsMainnetByName(mintMeta.Chain), 1, err

// 	case "/v1/mints/easy/urls":
// 		var mintMeta services.EasyMintMetaPartsDto
// 		err := c.ShouldBindBodyWith(&mintMeta, binding.JSON)
// 		return utils.IsMainnetByName(mintMeta.Chain), 1, err

// 	case "/dashboard/apps/:id/contracts":
// 		fallthrough
// 	case "/v1/contracts/":
// 		var contractDeployDto services.ContractDeployDto
// 		err := c.ShouldBindBodyWith(&contractDeployDto, binding.JSON)
// 		return utils.IsMainnetByName(contractDeployDto.Chain), 1, err

// 	case "/v1/transfers/customizable":
// 		var transferMeta services.TransferDto
// 		err := c.ShouldBindBodyWith(&transferMeta, binding.JSON)
// 		return utils.IsMainnetByName(transferMeta.Chain), 1, err

// 	case "/v1/transfers/customizable/batch":
// 		var transferBatchMeta services.TransferBatchDto
// 		err := c.ShouldBindBodyWith(&transferBatchMeta, binding.JSON)
// 		return utils.IsMainnetByName(transferBatchMeta.Chain), uint(len(transferBatchMeta.Items)), err
// 	}

// 	return false, 1, nil
// }
