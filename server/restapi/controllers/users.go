package controllers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func getUserApiQuotas(c *gin.Context) {
	userId, err := getUserIdByQuery(c)
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}
	uaqs, err := services.GetUserQuotaOperator().GetUserQuotas(userId, c.GetInt("offset"), c.GetInt("limit"))
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	ginutils.RenderRespOK(c, uaqs)
}

type UserWorkingBillPlanReq struct {
	UserId           uint `form:"user_id"`
	IsContainRainbow bool `form:"is_contain_rainbow"`
}

func getUserWorkingBillPlans(c *gin.Context) {
	var req UserWorkingBillPlanReq
	if err := c.ShouldBindQuery(&req); err != nil {
		ginutils.RenderRespError(c, err, http.StatusBadRequest)
		return
	}

	ueps, err := models.GetUserBillPlanOperator().GetUserEffectivePlans(req.UserId, req.IsContainRainbow)
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	values := lo.Values(ueps)
	sort.Slice(values, func(i, j int) bool {
		return values[i].ServerType < values[j].ServerType
	})

	ginutils.RenderRespOK(c, values)
}

func getUserIdByQuery(c *gin.Context) (uint, error) {
	userIdStr := c.Query("user_id")
	if userIdStr == "" {
		return 0, errors.New("missing user_id in query")
	}
	userId, err := strconv.Atoi(userIdStr)
	return uint(userId), err
}

func getUserCostTypePrice(c *gin.Context) {
	userIdStr := c.Query("user_id")
	if userIdStr == "" {
		ginutils.RenderRespError(c, errors.New("missing user_id in query"), http.StatusInternalServerError)
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	prices := make(map[enums.CostType]decimal.Decimal)
	for v := range enums.CostTypeValue2StrMap {
		price := models.GetApiPrice(uint(userId), v)
		prices[v] = price
	}
	ginutils.RenderRespOK(c, prices)
}
