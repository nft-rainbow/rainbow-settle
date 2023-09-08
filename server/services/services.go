package services

import (
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models"
)

var (
	userQuotaOperater    *models.UserQuotaOperator
	userBillPlanOperater *models.UserBillPlanOperator
	cmbDepositNoOperator *models.CmbDepositNoOperator
)

func Init() {
	userQuotaOperater = models.GetUserQuotaOperator()
	userBillPlanOperater = models.GetUserBillPlanOperator()
	userBillPlanOperater.RegisterOnChangedEvent(ResetQuotaOnPlanUpdated)
	cmbDepositNoOperator = &models.CmbDepositNoOperator{}
}

func Run() {
	go LoopSettle(time.Second * 2)
	go LoopRunPlan()
	go LoopMergeFiatlog()
	go LoopSetRichFlag()
	go StartWxOrderPolling()
	go StartCmbOrderPolling()
	LoadAllApikeys()
}
