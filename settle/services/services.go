package services

import (
	"time"

	"github.com/nft-rainbow/rainbow-fiat/common/models"
)

var (
	userQuotaOperater *models.UserQuotaOperator
)

func Init() {
	userQuotaOperater = &models.UserQuotaOperator{}
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
