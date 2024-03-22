package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/server/config"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 将 fl，flc 除了 mint cost type 的 amount 都设置为0
// 重新计算 flc balacne 和 fl balance

var (
	payApiTypes    = []models.FiatLogType{models.FIAT_LOG_TYPE_PAY_API_FEE, models.FIAT_LOG_TYPE_PAY_API_QUOTA}
	refundApiTypes = []models.FiatLogType{models.FIAT_LOG_TYPE_REFUND_API_FEE, models.FIAT_LOG_TYPE_REFUND_API_QUOTA}
)

func main() {
	err := tidyPostUserFls(time.Date(2024, 01, 10, 15, 0, 0, 0, time.Local), 289)
	logrus.WithError(err).Info("Tidy post users fiat logs done")
}

func tidyPostUserFls(startTime time.Time, userId uint) error {
	var needFixInvoiceIds []uint
	if err := models.GetDB().Debug().Model(&models.FiatLog{}).
		Group("invoice_id").
		Where("created_at>? and invoice_id is not null", startTime).
		Select("invoice_id").Scan(&needFixInvoiceIds).Error; err != nil {
		return errors.WithMessage(err, "Failed to get invoice ids")
	}
	logrus.WithField("needFixInvoiceIds", needFixInvoiceIds).Info("find need fix invoice ids")

	rawNeedFixFlsOfInvoice := make(map[uint][]models.FiatLog)
	for _, invoiceId := range needFixInvoiceIds {
		fls, err := findNeedUpdateFlsOfInvoice(userId, startTime, invoiceId)
		if err != nil {
			return err
		}
		rawNeedFixFlsOfInvoice[invoiceId] = fls
	}

	if err := updateFlcAmountAndBalance(userId, startTime); err != nil {
		return err
	}
	logrus.Info("update flc amount and balance done")

	if err := updateFlAmountAndBalance(userId, startTime); err != nil {
		return err
	}
	logrus.Info("update fl amount and balance done")

	// fix invoices
	for _, invoiceId := range needFixInvoiceIds {
		if err := fixInvoice(userId, startTime, invoiceId, rawNeedFixFlsOfInvoice[invoiceId]); err != nil {
			return err
		}
		logrus.WithField("invoiceId", invoiceId).Info("fix fls of invoice done")
	}

	return nil
}

func updateFlcAmountAndBalance(userId uint, startTime time.Time) error {
	err := models.GetDB().Debug().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.FiatLogCache{}).Where("user_id=? and created_at>?", userId, startTime).Update("amount", 0).Error; err != nil {
			return err
		}
		// update pay api fls to -0.7*count
		if err := tx.Model(&models.FiatLogCache{}).
			Where(`meta->'$.cost_type'='rainbow_mint' and user_id=? and created_at>?`, userId, startTime).
			Where("type in ?", payApiTypes).
			Update("amount", gorm.Expr("((JSON_UNQUOTE(JSON_EXTRACT(meta, '$.count_reset')) + JSON_UNQUOTE(JSON_EXTRACT(meta, '$.count_rollover'))) * -0.7)")).
			Error; err != nil {
			return err
		}
		// update refund api fls to 0.7*count
		if err := tx.Model(&models.FiatLogCache{}).
			Where(`meta->'$.cost_type'='rainbow_mint' and user_id=? and created_at>?`, userId, startTime).
			Where("type in ?", refundApiTypes).
			Update("amount", gorm.Expr("((JSON_UNQUOTE(JSON_EXTRACT(meta, '$.count_reset')) + JSON_UNQUOTE(JSON_EXTRACT(meta, '$.count_rollover'))) * 0.7)")).
			Error; err != nil {
			return err
		}

		// update flc balance
		// select * from fiat_log_caches where user_id=289 and created_at>startTime
		var flcs []*models.FiatLogCache
		if err := tx.Model(&models.FiatLogCache{}).Where("user_id=? and created_at>?", userId, startTime).Find(&flcs).Error; err != nil {
			return err
		}
		for i, flc := range flcs {
			if i == 0 {
				// select previous record before start time
				var previousFl *models.FiatLogCache
				if err := tx.Model(&models.FiatLogCache{}).Where("user_id=? and created_at<=?", userId, startTime).Order("id desc").First(&previousFl).Error; err != nil {
					return errors.WithMessage(err, "Failed to find previous fl before start time")
				}

				flc.Balance = flc.Amount.Add(previousFl.Balance)
				continue
			}
			flc.Balance = flc.Amount.Add(flcs[i-1].Balance)
		}
		if err := tx.Save(flcs).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func updateFlAmountAndBalance(userId uint, startTime time.Time) error {
	err := models.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.FiatLog{}).Where("user_id=? and created_at>?", userId, startTime).Update("amount", 0).Error; err != nil {
			return errors.WithMessage(err, "Failed to update amount of all of matched fiatlogs")
		}

		// update fl amount and balance according to previous's fiatLogCaches
		var fls []*models.FiatLog
		if err := tx.Model(&models.FiatLog{}).Where("user_id=? and created_at>?", userId, startTime).Find(&fls).Error; err != nil {
			return errors.WithMessage(err, "Failed to find")
		}
		for i, fl := range fls {
			var flcs []*models.FiatLogCache
			if err := tx.Model(&models.FiatLogCache{}).
				Where("created_at>?", startTime).
				Where("id in ?", []uint(fl.CacheIds)).Find(&flcs).Error; err != nil {
				return errors.WithMessage(err, "Failed to find")
			}

			// recalculate totalAmount
			totalAmount := decimal.Zero
			for _, flc := range flcs {

				// reset amount if fiat log type is "pay api fee" or "pay api quota"
				if flc.Type == models.FIAT_LOG_TYPE_PAY_API_FEE {
					var meta models.FiatMetaPayApiFee
					err := json.Unmarshal(flc.Meta, &meta)
					if err != nil {
						return errors.WithMessage(err, "Failed to unmarshal flc meta")
					}
					if meta.CostType == enums.COST_TYPE_RAINBOW_MINT {
						totalAmount = totalAmount.Add(decimal.NewFromFloat32(-0.7 * float32(meta.Count)))
					}
					continue
				}
				if flc.Type == models.FIAT_LOG_TYPE_PAY_API_QUOTA {
					var meta models.FiatMetaPayApiQuota
					err := json.Unmarshal(flc.Meta, &meta)
					if err != nil {
						return errors.WithMessage(err, "Failed to unmarshal flc meta")
					}
					if meta.CostType == enums.COST_TYPE_RAINBOW_MINT {
						totalAmount = totalAmount.Add(decimal.NewFromFloat32(-0.7 * float32(meta.CountReset+meta.CountRollover)))
					}
					continue
				}
				if flc.Type == models.FIAT_LOG_TYPE_REFUND_API_FEE {
					var meta models.FiatMetaRefundApiFee
					err := json.Unmarshal(flc.Meta, &meta)
					if err != nil {
						return errors.WithMessage(err, "Failed to unmarshal flc meta")
					}
					if meta.CostType == enums.COST_TYPE_RAINBOW_MINT {
						totalAmount = totalAmount.Add(decimal.NewFromFloat32(0.7 * float32(meta.Count)))
					}
					continue
				}
				if flc.Type == models.FIAT_LOG_TYPE_REFUND_API_QUOTA {
					var meta models.FiatMetaRefundApiQuota
					err := json.Unmarshal(flc.Meta, &meta)
					if err != nil {
						return errors.WithMessage(err, "Failed to unmarshal flc meta")
					}
					if meta.CostType == enums.COST_TYPE_RAINBOW_MINT {
						totalAmount = totalAmount.Add(decimal.NewFromFloat32(0.7 * float32(meta.CountReset+meta.CountRollover)))
					}
					continue
				}
			}
			fl.Amount = totalAmount

			// update balance
			if i == 0 {
				// select previous record before start time
				var previousFl *models.FiatLog
				if err := tx.Model(&models.FiatLog{}).Where("user_id=? and created_at<=?", userId, startTime).Order("id desc").First(&previousFl).Error; err != nil {
					return errors.WithMessage(err, "Failed to find previous fl before start time")
				}

				fl.Balance = fl.Amount.Add(previousFl.Balance)
				continue
			}
			fl.Balance = fl.Amount.Add(fls[i-1].Balance)
			logrus.WithField("id", fl.ID).WithField("amount", fl.Amount).WithField("balance", fl.Balance).Info("Updated fiatlog balance")
		}

		if err := tx.Save(fls).Error; err != nil {
			return errors.WithMessage(err, "Failed to save fls")
		}

		return nil
	})
	return err
}

func findNeedUpdateFlsOfInvoice(userId uint, startTime time.Time, invoicedId uint) ([]models.FiatLog, error) {
	var needClearFiatlogs []models.FiatLog
	err := models.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := models.GetDB().Model(&models.FiatLog{}).
			Where("user_id=? and invoice_id=? and created_at>? and balance<0", userId, invoicedId, startTime).
			Find(&needClearFiatlogs).
			Error; err != nil {
			return errors.WithMessage(err, "Failed to find fiatlogs")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return needClearFiatlogs, nil
}

// rawNeedClearFiatlogs is FiatLogs before update amount and balance, it's used to calculate need fixed amount
func fixInvoice(userId uint, startTime time.Time, invoicedId uint, rawNeedClearFiatlogs []models.FiatLog) error {
	err := models.GetDB().Transaction(func(tx *gorm.DB) error {
		for _, fl := range rawNeedClearFiatlogs {
			if err := tx.Model(&models.FiatLog{}).Where("id=?", fl.InvoiceId).Update("invoice_id", nil).Error; err != nil {
				return errors.WithMessage(err, "Failed to clear invoice id")
			}
		}

		sum := lo.Reduce(rawNeedClearFiatlogs, func(sum decimal.Decimal, fl models.FiatLog, index int) decimal.Decimal {
			return sum.Add(fl.Amount)
		}, decimal.Zero)

		cfxNum := sum.Div(decimal.NewFromFloat32(0.8))
		isInt := cfxNum.Equal(cfxNum.Ceil())
		if !isInt {
			return errors.WithMessage(fmt.Errorf("sum of fiatlogs is not integer %v, div result %v", sum, sum.Div(decimal.NewFromFloat32(0.8))), "Failed to fix invoices")
		}

		num40 := sum.Div(decimal.NewFromInt(-40)).Floor()
		amount40 := decimal.NewFromInt(-40).Mul(num40)
		num8 := sum.Sub(amount40).Div(decimal.NewFromFloat32(-8)).Floor()
		amount8 := decimal.NewFromInt(-8).Mul(num8)
		num0Dot8 := sum.Sub(amount40).Sub(amount8).Div(decimal.NewFromFloat32(-0.8))
		logrus.WithField("-40", num40).WithField("-8", num8).WithField("-0.8", num0Dot8).Info("calculate need fixed amount")

		var fiatlogsOf40 []*models.FiatLog
		if err := tx.Model(&models.FiatLog{}).
			Where("refund_log_ids=CAST('null' AS JSON) and balance>=0").
			Where("user_id=? and invoice_id is null and amount=?", userId, -40).
			Limit(int(num40.IntPart())).
			Find(&fiatlogsOf40).Error; err != nil {
			return errors.WithMessage(err, "Failed to find amount is 40 fiatlogs")
		}
		if len(fiatlogsOf40) != int(num40.IntPart()) {
			return errors.WithMessage(fmt.Errorf("amount of 40 fiatlogs is not enough %v", num40), "Failed to fix invoices")
		}
		for _, fl := range fiatlogsOf40 {
			fl.InvoiceId = &invoicedId
			logrus.WithField("fiat_log_id", fl.ID).WithField("amount", fl.Amount).WithField("invoice id", invoicedId).Info("update fiatlog invoice id")
		}

		var fiatlogsOf8 []*models.FiatLog
		if err := tx.Model(&models.FiatLog{}).
			Where("refund_log_id is null and balance>=0").
			Where("user_id=? and invoice_id is null and amount=?", userId, -8).
			Limit(int(num8.IntPart())).
			Find(&fiatlogsOf8).Error; err != nil {
			return errors.WithMessage(err, "Failed to find amount is 8 fiatlogs")
		}
		if len(fiatlogsOf8) != int(num8.IntPart()) {
			return errors.WithMessage(fmt.Errorf("amount of 8 fiatlogs is not enough %v", num8), "Failed to fix invoices")
		}
		for _, fl := range fiatlogsOf8 {
			fl.InvoiceId = &invoicedId
			logrus.WithField("fiat_log_id", fl.ID).WithField("amount", fl.Amount).WithField("invoice id", invoicedId).Info("update fiatlog invoice id")
		}

		var fiatlogsOf0Dot8 []*models.FiatLog
		if err := models.GetDB().Model(&models.FiatLog{}).
			Where("refund_log_id is null and balance>=0").
			Where("user_id=? and invoice_id is null and amount=?", userId, -0.8).
			Limit(int(num0Dot8.IntPart())).
			Find(&fiatlogsOf0Dot8).Error; err != nil {
			return errors.WithMessage(err, "Failed to find amount is -0.8 fiatlogs")
		}
		if len(fiatlogsOf0Dot8) != int(num0Dot8.IntPart()) {
			return errors.WithMessage(fmt.Errorf("amount of 0.8 fiatlogs is not enough %v", num0Dot8), "Failed to fix invoices")
		}
		for _, fl := range fiatlogsOf0Dot8 {
			fl.InvoiceId = &invoicedId
			logrus.WithField("fiat_log_id", fl.ID).WithField("amount", fl.Amount).WithField("invoice id", invoicedId).Info("update fiatlog invoice id")
		}

		if err := tx.Save(fiatlogsOf40).Error; err != nil {
			return err
		}
		if err := tx.Save(fiatlogsOf8).Error; err != nil {
			return err
		}
		if err := tx.Save(fiatlogsOf0Dot8).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func init() {
	config.InitByFile("../../config.yaml")
	models.ConnectDB(config.Get().Mysql)

	logger.Init(config.Get().Log, "====== CLI: Tidy Post Users Fiatlogs ========")
}
