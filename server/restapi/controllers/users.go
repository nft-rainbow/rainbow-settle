package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/samber/lo"
)

func getUserApiQuotas(c *gin.Context) {
	userId, err := getUserIdByQuery(c)
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}
	uaqs, err := models.GetUserQuotaOperator().GetUserQuotas(userId, c.GetInt("offset"), c.GetInt("limit"))
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	ginutils.RenderRespOK(c, uaqs)
}

func getUserWorkingBillPlans(c *gin.Context) {
	userId, err := getUserIdByQuery(c)
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}
	ueps, err := models.GetUserEffectivePlans(userId)
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	ginutils.RenderRespOK(c, lo.Values(ueps))
}

func getUserIdByQuery(c *gin.Context) (uint, error) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.Atoi(userIdStr)
	return uint(userId), err
}
