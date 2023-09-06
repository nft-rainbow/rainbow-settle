package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
)

func getApiProfiles(c *gin.Context) {
	aps, err := models.GetApiProfiles()
	ginutils.RenderResp(c, aps, err)
}

func queryApiProfile(c *gin.Context) {
	var filter models.ApiProfileFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}
	aps, err := models.QueryApiProfile(&filter, c.GetInt("offset"), c.GetInt("limit"))
	ginutils.RenderResp(c, aps, err)
}

func getAllBillPlans(c *gin.Context) {
	ps, err := models.GetAllPlans()
	ginutils.RenderResp(c, ps, err)
}

func queryBillPlan(c *gin.Context) {
	var filter models.BillPlanFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}
	aps, err := models.QueryBillPlan(&filter, c.GetInt("offset"), c.GetInt("limit"))
	ginutils.RenderResp(c, aps, err)
}

func getAllDataBundles(c *gin.Context) {
	ps, err := models.GetAllDataBundles()
	ginutils.RenderResp(c, ps, err)
}

func queryDataBundle(c *gin.Context) {
	var filter models.DataBundleFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}
	aps, err := models.QueryDataBundle(&filter, c.GetInt("offset"), c.GetInt("limit"))
	ginutils.RenderResp(c, aps, err)
}
