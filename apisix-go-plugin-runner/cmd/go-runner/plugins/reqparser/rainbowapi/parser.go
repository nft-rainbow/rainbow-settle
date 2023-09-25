package rainbowapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/types"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-api/services"
	"github.com/nft-rainbow/rainbow-api/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

var (
	rainbowApiEngine *gin.Engine
)

func init() {
	rainbowApiEngine = gin.New()
	// rainbowApiEngine.Any("*path", func(c *gin.Context) {})
	emptyHandler := gin.HandlerFunc(func(c *gin.Context) {})
	for k := range utils.MintPaths {
		rainbowApiEngine.POST(k, emptyHandler)
	}

	for k := range utils.DeployPaths {
		rainbowApiEngine.POST(k, emptyHandler)
	}
	ginutils.RegisterValidation()
}

// type RainbowApiReqParseResult struct {
// 	IsMainnet bool
// 	CostType  enums.CostType
// 	Count     int
// }

type RainbowApiRequestOp struct {
}

func (o *RainbowApiRequestOp) ParseRequest(r pkgHTTP.Request) (types.ReqParseResult, error) {
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

	if o.getAction(c.Request.Method, c.FullPath()) == enums.COST_TYPE_RAINBOW_NORMAL {
		return &types.DefaultReqParseResult{enums.COST_TYPE_RAINBOW_NORMAL, 1}, nil
	}

	body, err := r.Body()
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return o.parseRequestByGin(c)
}

func (o *RainbowApiRequestOp) parseRequestByGin(c *gin.Context) (*types.DefaultReqParseResult, error) {
	result := types.DefaultReqParseResult{
		CostType: o.getAction(c.Request.Method, c.FullPath()),
		Count:    1,
	}

	var err error
	var isMainnet bool

	switch c.FullPath() {
	case "/v1/mints/":
		var mintMeta services.CustomMintDto
		err = c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(mintMeta.Chain)

	case "/v1/mints/customizable":
		var mintMeta services.CustomMintDto
		err = c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(mintMeta.Chain)

	case "/v1/mints/customizable/batch":
		var mintMeta services.CustomMintBatchDto
		err = c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(mintMeta.Chain)

	case "/v1/mints/easy/files":
		var mintMeta services.EasyMintFileDto
		err = c.ShouldBind(&mintMeta)
		isMainnet = utils.IsMainnetByName(mintMeta.Chain)

	case "/v1/mints/easy/urls":
		var mintMeta services.EasyMintMetaPartsDto
		err = c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(mintMeta.Chain)

	case "/dashboard/apps/:id/nft":
		var mintMeta services.MintMetaPartsDto
		err = c.ShouldBindBodyWith(&mintMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(mintMeta.Chain)

	case "/dashboard/apps/:id/nft/batch/by-meta-parts":
		var batchMintMeta []*services.AppMintByMetaPartsDto
		err = c.ShouldBindBodyWith(&batchMintMeta, binding.JSON)
		result.Count = len(batchMintMeta)
		if len(batchMintMeta) > 0 {
			isMainnet = utils.IsMainnetByName(batchMintMeta[0].Chain)
		}

	case "/dashboard/apps/:id/nft/batch/by-meta-uri":
		var batchMintMeta services.AppBatchMintByMetaUriDto
		err = c.ShouldBindBodyWith(&batchMintMeta, binding.JSON)
		result.Count = len(batchMintMeta.MintItems)
		isMainnet = utils.IsMainnetByName(batchMintMeta.Chain)

	case "/dashboard/apps/:id/contracts":
		fallthrough
	case "/v1/contracts/":
		var contractDeployDto services.ContractDeployDto
		err = c.ShouldBindBodyWith(&contractDeployDto, binding.JSON)
		isMainnet = utils.IsMainnetByName(contractDeployDto.Chain)

	case "/v1/transfers/customizable":
		var transferMeta services.TransferDto
		err = c.ShouldBindBodyWith(&transferMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(transferMeta.Chain)

	case "/v1/transfers/customizable/batch":
		var transferBatchMeta services.TransferBatchDto
		err = c.ShouldBindBodyWith(&transferBatchMeta, binding.JSON)
		isMainnet = utils.IsMainnetByName(transferBatchMeta.Chain)
		result.Count = len(transferBatchMeta.Items)

	}
	if err != nil {
		return nil, err
	}

	if !isMainnet {
		result.CostType = enums.COST_TYPE_RAINBOW_NORMAL
	}

	return &result, nil
}

func (*RainbowApiRequestOp) getAction(method, fullPath string) enums.CostType {
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
