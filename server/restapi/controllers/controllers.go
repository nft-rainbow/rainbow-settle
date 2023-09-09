package controllers

import "github.com/gin-gonic/gin"

func SetupRouter(c *gin.Engine) {
	v0 := c.Group("v0")
	common := v0.Group("common")
	{
		// common.GET("api-profile/list", getApiProfiles)
		common.GET("api-profile", queryApiProfile)
		// common.GET("bill-plan/list", getAllBillPlans)
		common.GET("bill-plan", queryBillPlan)
		// common.GET("data-bundler/list", getAllDataBundles)
		common.GET("data-bundle", queryDataBundle)
	}

	user := v0.Group("user")
	{
		user.GET("quota", getUserApiQuotas)
		user.GET("bill-plan", getUserWorkingBillPlans)
	}
}
