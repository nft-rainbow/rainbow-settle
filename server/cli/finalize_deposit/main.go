package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func init() {
	config.InitByFile("../../config.yaml")
	models.ConnectDB(config.Get().Mysql)

	logger.Init(config.Get().Log, "====== CLI: Deposit for user manually ========")
}

func finalizeDeposit(orderId uint, status int) error {
	// create order
	order, err := models.FindDepositOrderById(orderId)
	if err != nil {
		return err
	}

	// fiat_log_caches/fiat_logs 中已存在，则警告并返回
	var flc models.FiatLogCache
	err = models.GetDB().Model(&models.FiatLogCache{}).Where("meta->'$.deposit_order_id'=?", orderId).First(&flc).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// 未找到，继续后续逻辑
	case err != nil:
		return err
	default:
		logrus.WithField("value", flc).Warn("fiat_log_caches already exist")
		return errors.New("fiat_log_caches already exist")
	}

	var fl *models.FiatLog
	err = models.GetDB().Model(&models.FiatLog{}).Where("meta->'$.deposit_order_id'=?", orderId).First(&fl).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// 未找到，继续后续逻辑
	case err != nil:
		return err
	default:
		logrus.WithField("value", fl).Warn("fiat_logs already exist")
		return errors.New("fiat_logs already exist")
	}

	// update order
	err = services.UpdateDepositOrder(orderId, status)
	if err != nil {
		return err
	}

	fmt.Printf("deposit order updated: %d\n", order.ID)
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: finalize-deposit <order_id> <status>\nstatus: 1-success,-1-failed")
		os.Exit(1)
	}

	logrus.WithField("args", os.Args).Info("finalize deposit input")

	orderId := utils.Must(strconv.ParseUint(os.Args[1], 0, 0))
	status := utils.Must(strconv.ParseInt(os.Args[2], 0, 0))
	err := finalizeDeposit(uint(orderId), int(status))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
