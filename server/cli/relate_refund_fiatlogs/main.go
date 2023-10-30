package main

import (
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func relateAllRefundSponsorFiatlogs() error {
	err := models.GetDB().Transaction(func(tx *gorm.DB) error {
		var refundSponsorFls []*models.FiatLog
		err := tx.Model(&models.FiatLog{}).Where("type=?", models.FIAT_LOG_TYPE_REFUND_SPONSOR).Find(&refundSponsorFls).Error
		if err != nil {
			return err
		}

		logrus.WithField("fls", refundSponsorFls).Info("find refund sponsor fiatlogs")

		for _, v := range refundSponsorFls {
			err = models.RelateBuySponsorFiatlog(tx, v)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func init() {
	config.InitByFile("../../config.yaml")
	models.ConnectDB(config.Get().Mysql)
}

func main() {
	err := relateAllRefundSponsorFiatlogs()
	logrus.WithError(err).Info("relate all refund sponsor fiatlogs completed")
}
