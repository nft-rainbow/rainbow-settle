package services

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// 1. renew plan if on month begin
// 1. run cron schedules daily
func LoopRunPlan() {
	c := cron.New()

	RunPlan()
	c.AddFunc("@daily", RunPlan)

	c.Start()
}

func RunPlan() {
	err := utils.Retry(10, time.Second*5, func() error {
		if err := RenewPlans(); err != nil {
			return err
		}
		return ResetQuotas()
	})
	logrus.WithError(err).Info("run plan completed")
}

func RenewPlans() error {
	needRenews, err := models.GetUserBillPlanOperator().FindAllUserNeedRenewPlans()
	if err != nil {
		return err
	}

	for userId, server2Plan := range needRenews {
		for _, plan := range server2Plan {
			utils.Retry(10, time.Second, func() error {
				fl, _, err := BuyBillPlan(userId, plan.ID, true)
				if err != nil {
					return err
				}
				logrus.WithField("user id", userId).WithField("plan", plan.ID).WithField("fiat log id", fl).Info("renewing plan completed")
				return nil
			})
		}
	}
	return nil
}

// 1. find users highest priority plan
// 2. set users without plan for default plan (default_rainbow,default_confura,default_scan)
func ResetQuotas() error {
	logrus.Info("reset default quotas")
	userPlans, err := models.GetUserBillPlanOperator().FindAllUsersEffectivePlan()
	logrus.WithField("result", userPlans).WithError(err).Info("find all user effective plans")
	if err != nil {
		return err
	}

	// to plan => userids
	plan2UserIds := make(map[*models.BillPlan][]uint)
	for userId, v := range userPlans {
		for _, plan := range v {
			plan2UserIds[plan] = append(plan2UserIds[plan], userId)
		}
	}

	for plan, userIds := range plan2UserIds {
		if err := resetQuotaByPlan(plan, userIds); err != nil {
			return err
		}
	}

	return nil
}

func resetQuotaByPlan(plan *models.BillPlan, userIds []uint) error {
	quotas := plan.GetQuotas()
	nextSchedule, err := plan.NextRefreshQuotaTime()
	if err != nil {
		return err
	}
	if err := userQuotaOperater.Reset(models.GetDB(), userIds, quotas, nextSchedule, false); err != nil {
		return err
	}
	return nil
}

func ResetQuotaOnPlanUpdated(old, new *models.UserBillPlan) {
	err := utils.Retry(10, time.Second, func() error {
		allPlansMap, err := models.GetAllPlansMap()
		if err != nil {
			return err
		}
		newPlan := allPlansMap[new.PlanId]
		if old == nil || newPlan.Priority > allPlansMap[old.PlanId].Priority {
			nextRefreshTime, err := newPlan.NextRefreshQuotaTime()
			if err != nil {
				return err
			}
			return models.GetUserQuotaOperator().Reset(models.GetDB(), []uint{new.UserId}, newPlan.GetQuotas(), nextRefreshTime, true)
		}
		return nil
	})
	logrus.WithError(err).WithField("old", old).WithField("new", new).Info("reset quota on plan updated")
}
