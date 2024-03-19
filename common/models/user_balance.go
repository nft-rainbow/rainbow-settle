package models

import (
	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserBalance struct {
	BaseModel
	UserId           uint            `gorm:"type:int;unique" json:"user_id"`
	Balance          decimal.Decimal `gorm:"type:decimal(20,2)" json:"balance"`                   // 实时余额,单位元
	BalanceOnFiatlog decimal.Decimal `gorm:"type:decimal(20,2)" json:"balance_on_fiatlog"`        // 与Fiatlog同步的余额,单位元
	ArrearsQuota     decimal.Decimal `gorm:"type:decimal(20,2)" json:"arrears_quota"`             // 单位元
	CfxPrice         decimal.Decimal `gorm:"type:decimal(20,2);default:0.8" json:"storage_price"` // 存储售卖单价单位元
}

func (u *UserBalance) BalanceWithArrears() decimal.Decimal {
	return u.ArrearsQuota.Add(u.Balance)
}

func NewUserBalance(userId uint) *UserBalance {
	return &UserBalance{
		UserId:       userId,
		ArrearsQuota: decimal.NewFromInt(fee.UserDefaultArrearsQuota),
	}
}

func InitUserBalances() {
	var exists []uint
	if err := GetDB().Model(&UserBalance{}).Distinct("user_id").Find(&exists).Error; err != nil {
		panic(err)
	}

	var unExists []uint
	if err := GetDB().Model(&User{}).Not(exists).Select("id").Find(&unExists).Error; err != nil {
		panic(err)
	}

	for _, u := range unExists {
		if err := GetDB().Debug().Create(NewUserBalance(u)).Error; err != nil {
			panic(err)
		}
	}

	// init balance_on_fiatlog
	rawQuery := "update user_balances set balance_on_fiatlog=(select balance from fiat_logs where fiat_logs.user_id=user_balances.user_id order by created_at desc,id desc limit 1);"
	if err := GetDB().Debug().Exec(rawQuery).Error; err != nil {
		panic(err)
	}
}

func GetUserBalance(tx *gorm.DB, userId uint) (*UserBalance, error) {
	var ub UserBalance
	if err := tx.Where("user_id = ?", userId).First(&ub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ub = UserBalance{UserId: userId, Balance: decimal.Zero}
			if err = tx.Create(&ub).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &ub, nil
}

func GetUserCfxPrice(userId uint) (decimal.Decimal, error) {
	normalCfxPrice := decimal.NewFromFloat(cfxPrice)

	ub, err := GetUserBalance(GetDB(), userId)
	if err != nil {
		if gormutils.IsRecordNotFoundError(err) {
			return normalCfxPrice, nil
		}
		return decimal.Zero, err
	}
	// if ub.CfxPrice.IsZero() {
	// 	return normalCfxPrice, nil
	// }
	return ub.CfxPrice, nil
}

func UpdateUserBalanceOnFiatlog(tx *gorm.DB, userId uint, balanceOnFiatlog decimal.Decimal) error {
	err := tx.Model(&UserBalance{}).Where("user_id=?", userId).Update("balance_on_fiatlog", balanceOnFiatlog).Error
	logrus.WithError(err).WithField("user_id", userId).WithField("balance_on_fiatlog", balanceOnFiatlog).Debug("update user balance_on_fiatlog")
	if err != nil {
		return err
	}
	ub, err := GetUserBalance(tx, userId)
	if err != nil {
		return err
	}
	logrus.WithField("ub", ub).Debug("after update user balance_on_fiatlog")
	return err
}

func UpdateUserBalance(tx *gorm.DB, userId uint, balance decimal.Decimal) error {
	err := tx.Model(&UserBalance{}).Where("user_id=?", userId).Update("balance", balance).Error
	logrus.WithError(err).WithField("user_id", userId).WithField("balance", balance).Debug("update user balance")
	return err
}

func UpdateUserCfxPrice(userId uint, price decimal.Decimal) error {
	if price.LessThan(decimal.Zero) {
		return errors.New("could not negative price")
	}
	ub, err := GetUserBalance(GetDB(), userId)
	if err != nil {
		return err
	}
	ub.CfxPrice = price
	return GetDB().Save(ub).Error
}
