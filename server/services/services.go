package services

import (
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/sirupsen/logrus"
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
		logrus.Debug("get all cost types")
		if err := userQuotaOperater.CreateIfNotExists(tx, []uint{user.ID}, costTypes); err != nil {
			return err
		}
		logrus.Debug("CreateIfNotExists user_api_quota")
		if err := models.GetUserSettledOperator().CreateIfNotExists(tx, []uint{user.ID}, costTypes); err != nil {
			return err
		}
		logrus.Debug("CreateIfNotExists user_settleds")
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
