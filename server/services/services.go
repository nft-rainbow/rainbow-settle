package services

import (
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models"
)

var (
	userQuotaOperater    *UserQuotaOperator
	userBillPlanOperater *models.UserBillPlanOperator
	cmbDepositNoOperator *models.CmbDepositNoOperator
)

func Init() {
	userQuotaOperater = GetUserQuotaOperator()

	userBillPlanOperater = models.GetUserBillPlanOperator()
	userBillPlanOperater.RegisterOnChangedEvent(ResetQuotaOnPlanUpdated)

	cmbDepositNoOperator = &models.CmbDepositNoOperator{}

	InitUserApiQuota()
}

func Run() {
	SetPlanToRedis()
	SetApiprofilesToRedis()
	go LoopSettle(time.Second * 2)
	go LoopSetRichFlagToRedis()
	go StartWxOrderPolling()
	go StartCmbOrderPolling()
	LoopRunPlan()
	LoopMergeFiatlog()
	LoadAllApikeys()
}
