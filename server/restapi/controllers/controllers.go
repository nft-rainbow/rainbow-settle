package controllers

import "github.com/gin-gonic/gin"

func SetupRouter(c *gin.Engine) {
	v0 := c.Group("v0")
	common := v0.Group("common")
	common.GET("api-profile", getApiProfiles)
	common.GET("bill-plan", getAllBillPlans)
	common.GET("data-bundler", getAllDataBundles)

	user := v0.Group("user")
	user.GET("quota", getUserApiQuotas)
	user.GET("bill-plan", getUserWorkingBillPlans)
}
