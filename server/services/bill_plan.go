package services

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
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
		return ResetAllUsersQuotas()
	})
	logrus.WithError(err).Info("run plan completed")
}

func RenewPlans() error {
	needRenews, err := models.GetUserBillPlanOperator().FindAllUserNeedRenewPlans()
	if err != nil {
		return err
	}
	logrus.WithField("need renews", needRenews).Debug("ready to renew plans")

	for userId, server2Userplan := range needRenews {
		for _, userPlan := range server2Userplan {
			utils.Retry(10, time.Second, func() error {
				fl, _, err := BuyBillPlan(userId, userPlan.PlanId, true)
				if err != nil {
					return err
				}
				logrus.WithField("user id", userId).WithField("plan", userPlan.PlanId).WithField("fiat log id", fl).Info("renewing plan completed")
				return nil
			})
		}
	}
	return nil
}

func ResetUsersQuotas(userIds []uint) error {
	logrus.WithField("users", userIds).Info("reset users quotas")
	userPlans, err := models.GetUserBillPlanOperator().FindUsersEffectivePlans(userIds)
	logrus.WithField("result", userPlans).WithField("user ids", userIds).WithError(err).Debug("debug on reset user quotas: find user effective plans")
	if err != nil {
		return err
	}

	// to plan => userids
	planId2UserIds := make(map[uint][]uint)
	for userId, v := range userPlans {
		for _, userPlan := range v {
			planId2UserIds[userPlan.PlanId] = append(planId2UserIds[userPlan.PlanId], userId)
		}
	}

	for planId, userIds := range planId2UserIds {
		plan, err := models.GetBillPlanById(planId)
		if err != nil {
			return err
		}
		err = resetQuotaByPlan(plan, userIds)
		logrus.Debug("debug on reset user quotas: reset quota by plan done")
		if err != nil {
			return err
		}
	}

	return nil
}

// 1. find users highest priority plan
// 2. set users without plan for default plan
func ResetAllUsersQuotas() error {
	logrus.Info("reset all users quotas")
	allUserIds, err := models.GetAllUserIds()
	if err != nil {
		return err
	}
	return ResetUsersQuotas(allUserIds)
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

func ResetQuotaOnUserCreated(userId uint) error {
	return ResetUsersQuotas([]uint{userId})
}

func ResetQuotaOnPlanUpdated(old, new *models.UserBillPlan) {
	err := utils.Retry(10, time.Second, func() error {
		allPlansMap, err := models.GetAllPlansMap()
		if err != nil {
			return err
		}

		// update if current quota small than new plan
		newPlan := allPlansMap[new.PlanId]
		newQuotas := newPlan.GetQuotas()
		userQuotasMap, err := GetUserQuotaOperator().GetUserQuotasMap(new.UserId)
		if err != nil {
			return err
		}

		needUpdates := make(map[enums.CostType]int)
		for costType, newCount := range newQuotas {
			if userQuotasMap[costType].CountReset < newCount {
				needUpdates[costType] = newCount
			}
		}

		if len(needUpdates) == 0 {
			return nil
		}

		nextRefreshTime, err := newPlan.NextRefreshQuotaTime()
		if err != nil {
			return err
		}
		return GetUserQuotaOperator().Reset(models.GetDB(), []uint{new.UserId}, needUpdates, nextRefreshTime, true)
	})
	logrus.WithError(err).WithField("old", old).WithField("new", new).Info("reset quota on plan updated")
}
