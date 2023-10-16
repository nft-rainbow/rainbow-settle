package services

import (
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"gorm.io/gorm"
)

var (
	userQuotaOperater    *UserQuotaOperator
	userBillPlanOperater *models.UserBillPlanOperator
	cmbDepositNoOperator *models.CmbDepositNoOperator
)

func Init() {
	userQuotaOperater = GetUserQuotaOperator()
	userBillPlanOperater = models.GetUserBillPlanOperator()
	cmbDepositNoOperator = &models.CmbDepositNoOperator{}

	InitUserApiQuota()
	RegisterEvents()
}

func RegisterEvents() {
	userBillPlanOperater.RegisterOnChangedEvent(ResetQuotaOnPlanUpdated)

	h := models.UserCreatedHandler(func(tx *gorm.DB, user *models.User) error {
		costTypes, err := models.GetAllCostTypes()
		if err != nil {
			return err
		}
		if err := userQuotaOperater.CreateIfNotExists(tx, []uint{user.ID}, costTypes); err != nil {
			return err
		}
		if err := models.GetUserSettledOperator().CreateIfNotExists(tx, []uint{user.ID}, costTypes); err != nil {
			return err
		}
		return nil
	})
	models.RegisterUserCreatedEvent(h)
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
