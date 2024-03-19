package main

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/sirupsen/logrus"
)

func updateFlCreatedByCacheIds(start time.Time, end time.Time) error {
	var fls []*models.FiatLog
	if err := models.GetDB().Where("created_at>? and created_at<? and type in (7,8)", start, end).Find(&fls).Error; err != nil {
		return err
	}

	for _, fl := range fls {
		flcId := fl.CacheIds[len(fl.CacheIds)-1]
		var flc models.FiatLogCache
		if err := models.GetDB().First(&flc, flcId).Error; err != nil {
			return err
		}

		if err := models.GetDB().Debug().
			Model(&models.FiatLog{}).
			Where("id=?", fl.ID).
			Update("created_at", flc.CreatedAt).Error; err != nil {
			return err
		}
	}
	return nil
}

func init() {
	config.InitByFile("../../config.yaml")
	models.ConnectDB(config.Get().Mysql)

	logger.Init(config.Get().Log, "====== CLI: Update Fiatlog Created_at By Cache_Ids ========")
}

func main() {
	err := updateFlCreatedByCacheIds(time.Date(2024, 1, 8, 0, 0, 0, 0, time.Local), time.Date(2024, 1, 9, 0, 0, 0, 0, time.Local))
	logrus.WithError(err).Info("update fiatlog created_at by cache_ids completed")
}
