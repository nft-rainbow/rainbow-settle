package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/shopspring/decimal"
)

type DepositForUserReq struct {
	UserId uint            `json:"user_id"`
	Amount decimal.Decimal `json:"amount"`
}

type WithdrawForUserReq struct {
	UserId uint            `json:"user_id"`
	Amount decimal.Decimal `json:"amount"`
}

type RefundForUserReq struct {
	UserId   uint           `json:"user_id" binding:"required"`
	CostType enums.CostType `json:"cost_type" binding:"required"`
	Count    uint           `json:"count" binding:"required"`
}

func depositForUser(c *gin.Context) {
	if config.Get().Environment == "production" || config.Get().Environment == "prod" {
		ginutils.RenderRespError(c, errors.New("not support on prod environment"), http.StatusMethodNotAllowed)
		return
	}

	var req DepositForUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	fl, err := services.DepositBalance(req.UserId, req.Amount, 0, models.FIAT_LOG_TYPE_DEPOSIT)
	ginutils.RenderResp(c, gin.H{"fiat_log_id": fl}, err)
}

func withdrawForUser(c *gin.Context) {
	if config.Get().Environment == "production" || config.Get().Environment == "prod" {
		ginutils.RenderRespError(c, errors.New("not support on prod environment"), http.StatusMethodNotAllowed)
		return
	}

	var req WithdrawForUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	fl, err := services.WithdrawBalance(req.UserId, req.Amount, "user apply")
	ginutils.RenderResp(c, gin.H{"fiat_log_id": fl}, err)
}

func refundForUser(c *gin.Context) {
	if config.Get().Environment == "production" || config.Get().Environment == "prod" {
		ginutils.RenderRespError(c, errors.New("not support on prod environment"), http.StatusMethodNotAllowed)
		return
	}

	var req RefundForUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutils.RenderRespError(c, err, http.StatusInternalServerError)
		return
	}

	fl, err := services.RefundApiFee(models.GetDB(), req.UserId, req.CostType, req.Count)
	ginutils.RenderResp(c, gin.H{"fiat_log_id": fl}, err)
}
