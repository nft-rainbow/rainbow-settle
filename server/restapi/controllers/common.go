package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
)

func getApiProfiles(c *gin.Context) {
	aps, err := models.GetApiProfiles()
	ginutils.RenderResp(c, aps, err)
}

func getAllBillPlans(c *gin.Context) {
	ps, err := models.GetAllPlans()
	ginutils.RenderResp(c, ps, err)
}

func getAllDataBundles(c *gin.Context) {
	ps, err := models.GetAllDataBundles()
	ginutils.RenderResp(c, ps, err)
}
