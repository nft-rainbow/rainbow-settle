package services

import (
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models"
)

var (
	userQuotaOperater    *models.UserQuotaOperator
	cmbDepositNoOperator *models.CmbDepositNoOperator
)

func Init() {
	userQuotaOperater = &models.UserQuotaOperator{}
	cmbDepositNoOperator = &models.CmbDepositNoOperator{}
}

func GetUserQuotaOperator() *models.UserQuotaOperator {
	return userQuotaOperater
}

func Run() {
	go LoopSettle(time.Second * 2)
	go LoopResetQuota()
	go LoopMergeFiatlog()
	go LoopSetRichFlag()
	go StartWxOrderPolling()
	go StartCmbOrderPolling()
}
