package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/rand"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/settle/config"
	"github.com/shopspring/decimal"

	confluxpay "github.com/web3-identity/conflux-pay-sdk-go"
	"golang.org/x/net/context"
)

const DEFAULT_TRADE_TYPE = "native"
const DEFAULT_TRADE_PROVIDER = "wechat"
const DEFAULT_TRADE_APP_NAME = "rainbow"

var apiClient *confluxpay.APIClient

func initConfluxPayCli() {
	cfg := confluxpay.NewConfiguration()
	cfg.Servers = []confluxpay.ServerConfiguration{
		{
			URL: config.Get().WechatPay.URL,
		},
	}
	//configuration.Debug = true 	// enable debug mode
	apiClient = confluxpay.NewAPIClient(cfg)
}

func makeWechatOrder(amount int32, desc string) (*confluxpay.ModelsOrder, *http.Response, error) {
	expire := time.Now().Unix() + int64(60*30) // default 30 minutes
	makeOrdReq := *confluxpay.NewServicesMakeOrderReq(amount, DEFAULT_TRADE_APP_NAME, "Rainbow-"+desc, int32(expire), DEFAULT_TRADE_PROVIDER, DEFAULT_TRADE_TYPE)
	return apiClient.OrdersApi.MakeOrder(context.Background()).MakeOrdReq(makeOrdReq).Execute()
}

func queryWechatOrderSummary(tradeNo string) (*confluxpay.ModelsOrder, *http.Response, error) {
	return apiClient.OrdersApi.QueryOrder(context.Background(), tradeNo).Execute()
}

// cmb methods
func queryCmbHistory(no *string, limit int32, offset int32) ([]confluxpay.ModelsCmbRecord, error) {
	// transactionDate := "transactionDate_example"           // string | specified date, e.g. 20230523
	// transactionDirection := "transactionDirection_example" // string | transaction direction, C for recieve and D for out

	configuration := confluxpay.NewConfiguration()
	apiClient := confluxpay.NewAPIClient(configuration)
	r := apiClient.CmbApi.QueryHistoryCmbRecords(context.Background()).
		Limit(limit).
		Offset(offset)
		// TransactionDate(transactionDate).
		// TransactionDirection(transactionDirection).

	if no != nil {
		r.UnitAccountNbr(*no)
	}

	resp, _, err := r.Execute()

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func queryRecentCmbHistory(limit, offset int32) ([]confluxpay.ModelsCmbRecord, error) {
	resp, _, err := apiClient.CmbApi.QueryRecentCmbRecords(context.Background()).Limit(limit).Offset(offset).Execute()
	return resp, err
}

func QueryRecentCmbHistory(limit, offset int32) ([]confluxpay.ModelsCmbRecord, error) {
	return queryRecentCmbHistory(limit, offset)
}

func addCmbUnitAccount(name, no string) error {
	addUnitAccountReq := *confluxpay.NewControllersAddUnitAccountReq(name, no)
	_, err := apiClient.CmbApi.AddUnitAccount(context.Background()).AddUnitAccountReq(addUnitAccountReq).Execute()
	return err
}

func setCmbUnitAccountRelation(chargeNo, unitNo string) error {
	setUnitAccountRelationReq := *confluxpay.NewControllersSetUnitAccountRelationReq(chargeNo, unitNo)
	_, err := apiClient.CmbApi.SetUnitAccountRelation(context.Background()).SetUnitAccountRelationReq(setUnitAccountRelationReq).Execute()
	return err
}

func CreateWechatOrder(userId uint, amount int32, desc string) (*confluxpay.ModelsOrder, error) {
	order, _, err := makeWechatOrder(amount, desc)
	if err != nil {
		return nil, err
	}
	orderJsonStr, _ := json.Marshal(order)
	models.GetDB().Create(&models.DepositOrder{
		UserId:      userId,
		Amount:      decimal.NewFromFloat(float64(amount) / 100),
		TradeNo:     *order.TradeNo,
		Type:        models.DEPOSIT_TYPE_WECHAT,
		Status:      models.DEPOSIT_INIT,
		Description: desc,
		Meta:        orderJsonStr,
	})
	return order, nil
}

func UpdateDepositOrder(orderId uint, status int) error {
	err := models.GetDB().Model(&models.DepositOrder{}).Where("id = ?", orderId).Update("status", status).Error
	if err != nil {
		return err
	}
	if status == models.DEPOSIT_SUCCESS {
		order := models.DepositOrder{}
		if err := models.GetDB().First(&order, orderId).Error; err != nil {
			return err
		}
		if _, err = DepositBalance(order.UserId, order.Amount, orderId, models.FIAT_LOG_TYPE_DEPOSIT); err != nil {
			utils.DingWarnf("deposit balance failed: %d %s", orderId, err.Error())
			return err
		}
	}
	return nil
}

type CmbDepositNoDto struct {
	Name   string `form:"name" json:"name" binding:"required"`
	Bank   string `form:"bank" json:"bank" binding:"required"`
	CardNo string `form:"card_no" json:"card_no" binding:"required"`
}

func CreateCmcDepositNo(userId uint, info CmbDepositNoDto) error {
	// create order no for user: prefix num is 12
	cmbNo := fmt.Sprintf("1%s%s", rand.NumString(3), utils.LeftPadZero(strconv.Itoa(int(userId)), 6))
	err := addCmbUnitAccount(info.Name, cmbNo)
	if err != nil {
		return err
	}
	// bind info for user
	err = setCmbUnitAccountRelation(info.CardNo, cmbNo)
	if err != nil {
		return err
	}
	// save in db
	item := models.CmbDepositNo{
		UserId:       userId,
		UserName:     info.Name,
		UserBankNo:   info.CardNo,
		UserBankName: info.Bank,
		CmbNo:        cmbNo,
	}
	err = models.GetDB().Save(&item).Error
	return err
}

// 更新 cmb 关联充值卡信息
func UpdateCmcDepositNoRelation(userId uint, info CmbDepositNoDto) error {
	var cmbDepositNo *models.CmbDepositNo
	item, err := cmbDepositNo.FindByUserId(userId)
	if err != nil {
		return err
	}
	// update cmb relation
	if item.UserBankNo != info.CardNo {
		err = setCmbUnitAccountRelation(info.CardNo, item.CmbNo)
		if err != nil {
			return err
		}
	}
	item.UserName = info.Name
	item.UserBankNo = info.CardNo
	item.UserBankName = info.Bank
	err = models.GetDB().Save(&item).Error
	return err
}

const WECHAT_PAY_SUCCESS = "SUCCESS"
const WECHAT_APY_ERROR = "PAYERROR" // 支付失败
const WECHAT_PAY_REVOKED = "REVOKED"
const WECHAT_PAY_CLOSED = "CLOSED"
const WECHAT_PAY_NOTPAY = "NOTPAY" // 未支付
// 其他
// USERPAYING
// REFUND

func StartWxOrderPolling() {
	initConfluxPayCli()
	for {
		time.Sleep(time.Second * 10)

		orders := []models.DepositOrder{}
		err := models.GetDB().Where("status = ?", models.DEPOSIT_INIT).Find(&orders).Limit(10).Error
		if err != nil {
			continue
		}

		for _, order := range orders {
			wxOrder, resp, err := queryWechatOrderSummary(order.TradeNo)
			if err != nil {
				continue
			}
			if resp.StatusCode == 200 {
				if *wxOrder.TradeState == WECHAT_PAY_SUCCESS { // check the status
					_ = UpdateDepositOrder(order.ID, models.DEPOSIT_SUCCESS)
				} else if *wxOrder.TradeState == WECHAT_PAY_CLOSED || *wxOrder.TradeState == WECHAT_PAY_REVOKED || *wxOrder.TradeState == WECHAT_APY_ERROR {
					_ = UpdateDepositOrder(order.ID, models.DEPOSIT_FAILED)
				}
			}
		}
		// TODO deal unpay and outdated orders
	}
}

func StartCmbOrderPolling() {
	for {
		time.Sleep(time.Second * 10)

		offset := 0
		limit := 50
		for {
			orders, err := queryRecentCmbHistory(int32(limit), int32(offset))
			if len(orders) == 0 || err != nil {
				break
			}

			allInserted := true
			for _, order := range orders {
				inserted, err := saveCmbDepositOrder(&order)
				if err != nil {
					allInserted = false
					continue
				}
				if !inserted {
					allInserted = false
					break
				}
			}

			if !allInserted {
				break
			}
			offset += limit
		}
	}
}

const UNIT float32 = 100

func saveCmbDepositOrder(order *confluxpay.ModelsCmbRecord) (bool, error) {
	// check exist
	exist, err := models.FindDepositOrderByTradeNo(*order.TrxNbr)
	if exist != nil && err == nil {
		return false, nil
	}
	var cmb *models.CmbDepositNo
	cmbInfo, err := cmb.FindByCmbNo(*order.DmaNbr)
	if err != nil {
		return false, err
	}

	// create record in DepositOrder
	orderJsonStr, _ := json.Marshal(order)
	amount := decimal.NewFromFloat32(*order.TrxAmt)
	item := models.DepositOrder{
		UserId:      cmbInfo.UserId,
		Amount:      amount,
		TradeNo:     *order.TrxNbr,
		Type:        models.DEPOSIT_TYPE_CMB,
		Status:      models.DEPOSIT_SUCCESS,
		Description: "招行对公",
		Meta:        orderJsonStr,
	}
	err = models.GetDB().Create(&item).Error
	if err != nil {
		return false, err
	}

	// deposit balance
	_, err = DepositBalance(cmbInfo.UserId, amount, item.ID, models.FIAT_LOG_TYPE_CMB_CHARGE)
	if err != nil {
		return false, err
	}
	return true, nil
}
