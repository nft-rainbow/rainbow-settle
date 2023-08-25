package services

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// TODO: 每天合并当天的fiat_log_cache，生成fiat_log
func LoopMergeFiatlog() {
	c := cron.New()

	var eid cron.EntryID

	fn := func() {
		// TODO: TomorrowBegin 改为 TodayBegin
		start, end := utils.EarlistDate(), utils.TomorrowBegin()
		utils.Retry(10, time.Second*5, func() error { return models.MergeToFiatlog(start, end) })
		logrus.WithField("val", c.Entry(eid).Next).Info("next merge fiat log time")
	}
	fn()

	eid, _ = c.AddFunc("@every 10s", fn)

	c.Start()
}
