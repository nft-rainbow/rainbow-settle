package services

import (
	"math"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// 1. 根据规则reset user quota
// 1.1 rainbow 每月reset
func LoopResetQuota() {
	c := cron.New()
	for _, r := range config.Get().QuotaRules {
		fn := func() {
			utils.Retry(10, time.Second*5, func() error { return ResetDefaultQuotas(r) })
		}
		fn()

		c.AddFunc(r.Schedule, fn)
	}
	c.Start()
}

func ResetDefaultQuotas(r *config.QuotaRule) error {
	logrus.WithField("rule", r).Info("reset default quotas")
	users, err := models.GetAllUser(nil, 0, math.MaxInt32)
	if err != nil {
		return err
	}

	userIds, _ := utils.MapSlice(users, func(u *models.User) (uint, error) { return u.ID, nil })
	if err := userQuotaOperater.CreateIfNotExists(models.GetDB(), userIds, r.GetRelatedCostTypes()); err != nil {
		return err
	}

	nextSchedule, err := utils.NextScheduleTime(r.Schedule)
	if err != nil {
		return err
	}

	if err := userQuotaOperater.Reset(models.GetDB(), userIds, r.Quotas, nextSchedule); err != nil {
		return err
	}
	return nil
}
