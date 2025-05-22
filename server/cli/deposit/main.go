package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/shopspring/decimal"
)

func init() {
	config.InitByFile("../../config.yaml")
	models.ConnectDB(config.Get().Mysql)

	logger.Init(config.Get().Log, "====== CLI: Deposit for user manually ========")
}

func depositForUser(userId uint, amountInFen int32, description string) error {
	// create order
	order := &models.DepositOrder{
		UserId:      userId,
		Amount:      decimal.NewFromFloat(float64(amountInFen) / 100),
		TradeNo:     models.RandomOrderNO(),
		Type:        10,
		Status:      models.DEPOSIT_SUCCESS,
		Description: fmt.Sprintf("manual deposit: %s", description),
	}
	result := models.GetDB().Create(order)
	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("deposit order created: %d\n", order.ID)

	// update order
	err := services.UpdateDepositOrder(order.ID, models.STATUS_SUCCESS)
	if err != nil {
		return err
	}

	fmt.Printf("deposit order updated: %d\n", order.ID)
	return nil
}

func main() {
	fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 4 {
		fmt.Println("usage: deposit <user_id> <amount_in_fen> <description>")
		os.Exit(1)
	}
	userId := utils.Must(strconv.ParseUint(os.Args[1], 0, 0))
	amountInFen := utils.Must(strconv.ParseInt(os.Args[2], 0, 0))
	description := os.Args[3]
	err := depositForUser(uint(userId), int32(amountInFen), description)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
