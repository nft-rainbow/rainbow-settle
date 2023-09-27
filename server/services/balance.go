package services

import (
	"encoding/json"

	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func DepositBalance(userId uint, amount decimal.Decimal, depositOrderId uint, logType models.FiatLogType) (uint, error) {
	return updateUserBalance(userId, amount, logType, models.FiatMetaDeposit{depositOrderId})
}

func WithdrawBalance(userId uint, amount decimal.Decimal) (uint, error) {
	return updateUserBalance(userId, decimal.Zero.Sub(amount), models.FIAT_LOG_TYPE_WITHDRAW, nil)
}

func BuyGas(userId uint, amount decimal.Decimal, txId uint, address string, price decimal.Decimal) (uint, error) {
	return updateUserBalance(userId, decimal.Zero.Sub(amount), models.FIAT_LOG_TYPE_BUY_GAS, models.FiatMetaBuySponsor{address, txId, price})
}

func BuyStorage(userId uint, amount decimal.Decimal, txId uint, address string, price decimal.Decimal) (uint, error) {
	return updateUserBalance(userId, decimal.Zero.Sub(amount), models.FIAT_LOG_TYPE_BUY_STORAGE, models.FiatMetaBuySponsor{address, txId, price})
}

func BuyBillPlan(userId uint, planId uint, isAutoRenewal bool) (fiatlogId uint, userBillPlan *models.UserBillPlan, err error) {
	plan, err := models.GetBillPlanById(planId)
	if err != nil {
		return 0, nil, err
	}

	var up *models.UserBillPlan
	var fl uint

	err = models.GetDB().Transaction(func(tx *gorm.DB) error {
		up, err = models.GetUserBillPlanOperator().UpdateUserBillPlan(tx, userId, planId, isAutoRenewal)
		if err != nil {
			return err
		}
		fl, err = updateUserBalanceWithTx(tx, userId, decimal.Zero.Sub(plan.Price), models.FIAT_LOG_TYPE_BUY_BILLPLAN, models.FiatMetaBuyBillplan{up.PlanId, up.ID})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return fl, up, nil
}

func BuyDataBundle(userId uint, dataBundleId uint, count uint) (fiatlogId uint, userDataBundle *models.UserDataBundle, err error) {
	plan, err := models.GetDataBundleById(dataBundleId)
	if err != nil {
		return 0, nil, err
	}

	var udb *models.UserDataBundle
	var fl uint
	err = models.GetDB().Transaction(func(tx *gorm.DB) error {
		udb, err = models.CreateUserDataBundleAndConsume(tx, userId, dataBundleId, count)
		if err != nil {
			return err
		}
		fl, err = updateUserBalanceWithTx(tx, userId, decimal.Zero.Sub(plan.Price), models.FIAT_LOG_TYPE_BUY_DATABUNDLE, models.FiatMetaBuyDatabundle{udb.DataBundleId, udb.Count, udb.ID})
		if err != nil {
			return err
		}
		return nil
	})

	return fl, udb, nil
}

func RefundSponsor(userId uint, amount decimal.Decimal, sponsorFiatlogId uint, sponsorFiatlogType models.FiatLogType, txId uint) (uint, error) {
	return RefundSponsorWithTx(models.GetDB(), userId, amount, sponsorFiatlogId, sponsorFiatlogType, txId)
}

func RefundSponsorWithTx(tx *gorm.DB, userId uint, amount decimal.Decimal, sponsorFiatlogId uint, sponsorFiatlogType models.FiatLogType, txId uint) (uint, error) {
	return updateUserBalanceWithTx(tx, userId, amount, models.FIAT_LOG_TYPE_REFUND_SPONSOR, models.FiatMetaRefundSponsor{sponsorFiatlogId, sponsorFiatlogType, txId, "tx failed"})
}

func RefundApiFee(tx *gorm.DB, userId uint, costType enums.CostType, count uint) (uint, error) {
	amount := models.GetApiPrice(costType).Mul(decimal.NewFromInt(int64(count)))
	return updateUserBalanceWithTx(tx, userId, amount, models.FIAT_LOG_TYPE_REFUND_API_FEE, models.FiatMetaRefundApiFee{costType, int(count)}, false)
}

func PayAPIFee(tx *gorm.DB, userId uint, costType enums.CostType, count uint) (uint, error) {
	amount := models.GetApiPrice(costType).Mul(decimal.NewFromInt(int64(count)))
	return updateUserBalanceWithTx(tx, userId, decimal.Zero.Sub(amount), models.FIAT_LOG_TYPE_PAY_API_FEE, models.FiatMetaPayApiFee{costType, int(count)}, false)
}

func updateUserBalance(userId uint, amount decimal.Decimal, logType models.FiatLogType, meta interface{}, checkBalance ...bool) (uint, error) {
	var fiatLogId uint
	err := models.GetDB().Transaction(func(tx *gorm.DB) error {
		l, err := updateUserBalanceWithTx(tx, userId, amount, logType, meta, checkBalance...)
		fiatLogId = l
		return err
	})
	return fiatLogId, err
}

func updateUserBalanceWithTx(tx *gorm.DB, userId uint, amount decimal.Decimal, logType models.FiatLogType, meta interface{}, checkBalance ...bool) (uint, error) {
	// 找 logtype 上一条记录的unsettle
	flc, err := models.FindLastFiatLogCache(userId, logType)
	if err != nil {
		if gormutils.IsRecordNotFoundError(err) {
			flc = &models.FiatLogCache{}
		} else {
			return 0, err
		}
	}

	// 小于1分的零头只记录，等凑齐1分以上才结算
	_amount := amount.Add(flc.UnsettleAmount)
	amount, leftover := calcLeftover(_amount)

	fl, err := func() (uint, error) {
		if err := checkDecimalQualified(amount); err != nil {
			return 0, err
		}

		userBalance := models.UserBalance{
			UserId: userId,
		}

		if err := tx.Where(&userBalance).Find(&userBalance).Error; err != nil {
			return 0, err
		}

		if (len(checkBalance) == 0 || checkBalance[0]) && userBalance.Balance.Add(userBalance.ArrearsQuota).Add(amount).Cmp(decimal.Zero) < 0 {
			return 0, errors.New("insufficient balance")
		}

		metaJson, _ := json.Marshal(meta)
		l := models.FiatLogCache{
			FiatLogCore: models.FiatLogCore{
				UserId:  userId,
				Amount:  amount,
				Type:    logType,
				Meta:    metaJson,
				OrderNO: models.RandomOrderNO(),
				Balance: userBalance.Balance.Add(amount),
			},
			UnsettleAmount: leftover,
		}
		if err := tx.Create(&l).Error; err != nil {
			return 0, err
		}

		userBalance.Balance = userBalance.Balance.Add(amount)
		return l.ID, tx.Save(&userBalance).Error
	}()

	logrus.WithFields(logrus.Fields{
		"userId":    userId,
		"amount":    amount,
		"logType":   logType,
		"fiatLogId": fl,
	}).WithError(err).Info("update user balance")

	return fl, err
}

func UpdateUserArrearsQuota(userId uint, amount decimal.Decimal) error {
	var ub models.UserBalance
	if err := models.GetDB().Where(&models.UserBalance{UserId: userId}).First(&ub).Error; err != nil {
		return err
	}
	ub.ArrearsQuota = amount
	return models.GetDB().Save(&ub).Error
}

func UpdateUserCfxPrice(userId uint, price decimal.Decimal) error {
	if err := checkDecimalQualified(price); err != nil {
		return err
	}
	return models.GetDB().Model(&models.UserBalance{}).Where(&models.UserBalance{UserId: userId}).Update("cfx_price", price).Error
}

func checkDecimalQualified(amount decimal.Decimal) error {
	amountX100 := amount.Mul(decimal.NewFromInt(100))
	if amountX100.Cmp(amountX100.Floor()) != 0 {
		return errors.Errorf("the decimal place of value cannot be greater than 2, but got %v", amount)
	}
	return nil
}

// returns amount in fen and leftover small than 1 fen
func calcLeftover(rawAmount decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	if !rawAmount.Round(2).Equal(rawAmount) {
		fen, _ := decimal.NewFromString(".01")
		leftover := rawAmount.Mod(fen)
		amount := rawAmount.Sub(leftover)
		return amount, leftover
	}
	return rawAmount, decimal.Zero
}
