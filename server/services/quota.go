package services

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/cronutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// 1. find users highest priority plan
// 2. set users without plan for default plan (default_rainbow,default_confura,default_scan)
// 3. run cron schedules daily
func LoopResetQuota() {
	c := cron.New()
	// for _, r := range config.Get().QuotaRules {
	// 	fn := func() {
	// 		utils.Retry(10, time.Second*5, func() error { return ResetDefaultQuotas(r) })
	// 	}
	// 	fn()

	// 	c.AddFunc(r.Schedule, fn)
	// }

	fn := func() {
		utils.Retry(10, time.Second*5, ResetQuotas)
	}
	fn()
	c.AddFunc("@daily", fn)

	c.Start()
}

// func ResetDefaultQuotas(r *config.QuotaRule) error {
// 	logrus.WithField("rule", r).Info("reset default quotas")
// 	users, err := models.GetAllUser(nil, 0, math.MaxInt32)
// 	if err != nil {
// 		return err
// 	}

// 	userIds, _ := utils.MapSlice(users, func(u *models.User) (uint, error) { return u.ID, nil })
// 	if err := userQuotaOperater.CreateIfNotExists(models.GetDB(), userIds, r.GetRelatedCostTypes()); err != nil {
// 		return err
// 	}

// 	nextSchedule, err := utils.NextScheduleTime(r.Schedule)
// 	if err != nil {
// 		return err
// 	}

// 	if err := userQuotaOperater.Reset(models.GetDB(), userIds, r.Quotas, nextSchedule); err != nil {
// 		return err
// 	}
// 	return nil
// }

func ResetQuotas() error {
	logrus.Info("reset default quotas")
	userPlans, err := models.FindAllUsersEffectivePlan()
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
		quotas := plan.GetQuotas()
		nextSchedule, err := cronutils.NextScheduleTime(plan.RefreshQuotaSchedule.ToCronSchedule())
		if err != nil {
			return err
		}
		if err := userQuotaOperater.Reset(models.GetDB(), userIds, quotas, nextSchedule); err != nil {
			return err
		}
	}

	return nil
}

// func ConsumeDataBundle(tx *gorm.DB, udb *models.UserDataBundle) error {
// 	return userQuotaOperater.DepositDataBundle(tx, udb)
// }
