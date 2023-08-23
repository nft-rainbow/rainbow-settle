package models

import (
	"encoding/json"

	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserBalance struct {
	BaseModel
	UserId       uint            `gorm:"type:int;unique" json:"user_id"`
	Balance      decimal.Decimal `gorm:"type:decimal(20,2)" json:"balance"`       // 单位元
	ArrearsQuota decimal.Decimal `gorm:"type:decimal(20,2)" json:"arrears_quota"` // 单位元
	// FreeApiQuota
	// FreeApiWorkMonth string          `gorm:"type:varchar(10);index" json:"free_api_work_month"`   // 生效的月份
	CfxPrice decimal.Decimal `gorm:"type:decimal(20,2);default:0.8" json:"storage_price"` // 存储售卖单价单位分
}

func NewUserBalance(userId uint) *UserBalance {
	return &UserBalance{
		UserId:       userId,
		ArrearsQuota: decimal.NewFromInt(fee.UserDefaultArrearsQuota),
		// FreeApiQuota: FreeApiQuota{
		// 	FreeOtherApiQuota: fee.UserDefaultFreeOtherAPIQuota,
		// 	FreeMintQuota:     fee.UserDefaultFreeMintQuota,
		// 	FreeDeployQuota:   fee.UserDefaultFreeDeployQuota,
		// },

		// FreeApiWorkMonth: utils.CurrentMonthStr(),
	}
}

// type FreeApiQuota struct {
// 	FreeOtherApiQuota int `gorm:"type:int;default:0" json:"free_other_api_quota" binding:"required"` // 单位次，每月重置
// 	FreeMintQuota     int `gorm:"type:int;default:0" json:"free_mint_quota" binding:"required"`      // 单位次，每月重置
// 	FreeDeployQuota   int `gorm:"type:int;default:0" json:"free_deploy_quota" binding:"required"`    // 单位次，每月重置
// }

// func (f *FreeApiQuota) Decrease(dec FreeApiUsed) {
// 	f.FreeMintQuota -= dec.FreeMint
// 	f.FreeDeployQuota -= dec.FreeDeploy
// 	f.FreeOtherApiQuota -= dec.FreeOtherApi
// }

// type ApiFeeCostItem struct {
// 	Method    string          `json:"method"`
// 	Path      string          `json:"path"`
// 	Count     uint            `json:"count"`
// 	IsTestnet bool            `json:"is_testnet"`
// 	Fee       decimal.Decimal `json:"fee"`
// 	CostType  enums.CostType  `json:"cost_type"`
// 	FreeApiUsed
// }

// func NewApiFeeCostItem(method string, path string, count uint, isTestNet bool, fee decimal.Decimal, free FreeApiUsed) *ApiFeeCostItem {
// 	item := ApiFeeCostItem{
// 		Method:      method,
// 		Path:        path,
// 		Count:       count,
// 		Fee:         fee,
// 		CostType:    enums.GetCostType(isTestNet, method, path),
// 		FreeApiUsed: free,
// 	}
// 	return &item
// }

// type FreeApiUsed struct {
// 	FreeMint     int `json:"free_mint"`      // 单位次，每月重置
// 	FreeDeploy   int `json:"free_deploy"`    // 单位次，每月重置
// 	FreeOtherApi int `json:"free_other_api"` // 单位次，每月重置
// }

// func (f *FreeApiUsed) Increase(inc FreeApiUsed) {
// 	f.FreeMint += inc.FreeMint
// 	f.FreeDeploy += inc.FreeDeploy
// 	f.FreeOtherApi += inc.FreeOtherApi
// }

// var (
// 	FreeApiUsedDefault = FreeApiUsed{}
// )

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

	// if err := initUsersRemainBalanceQuota(); err != nil {
	// 	panic(err)
	// }
}

func GetUserBalance(userId uint) (*UserBalance, error) {
	var ub UserBalance
	if err := GetDB().Where("user_id = ?", userId).First(&ub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ub = UserBalance{UserId: userId, Balance: decimal.Zero}
			if err = GetDB().Create(&ub).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &ub, nil
}

func DepositBalance(userId uint, amount decimal.Decimal, depositOrderId uint, logType FiatLogType) (uint, error) {
	return UpdateUserBalance(userId, amount, logType, FiatMetaDeposit{depositOrderId})
}

func WithdrawBalance(userId uint, amount decimal.Decimal) (uint, error) {
	return UpdateUserBalance(userId, decimal.Zero.Sub(amount), FIAT_LOG_TYPE_WITHDRAW, nil)
}

func BuyGas(userId uint, amount decimal.Decimal, txId uint, address string, price decimal.Decimal) (uint, error) {
	return UpdateUserBalance(userId, decimal.Zero.Sub(amount), FIAT_LOG_TYPE_BUY_GAS, FiatMetaBuySponsor{address, txId, price})
}

func BuyStorage(userId uint, amount decimal.Decimal, txId uint, address string, price decimal.Decimal) (uint, error) {
	return UpdateUserBalance(userId, decimal.Zero.Sub(amount), FIAT_LOG_TYPE_BUY_STORAGE, FiatMetaBuySponsor{address, txId, price})
}

func RefundSponsor(userId uint, amount decimal.Decimal, sponsorFiatlogId uint, sponsorFiatlogType FiatLogType, txId uint) (uint, error) {
	return RefundSponsorWithTx(GetDB(), userId, amount, sponsorFiatlogId, sponsorFiatlogType, txId)
}

func RefundSponsorWithTx(tx *gorm.DB, userId uint, amount decimal.Decimal, sponsorFiatlogId uint, sponsorFiatlogType FiatLogType, txId uint) (uint, error) {
	return UpdateUserBalanceWithTx(tx, userId, amount, FIAT_LOG_TYPE_REFUND_SPONSOR, FiatMetaRefundSponsor{sponsorFiatlogId, sponsorFiatlogType, txId, "tx failed"})
}

func RefundApiFee(tx *gorm.DB, userId uint, costType enums.CostType, count uint) (uint, error) {
	amount := GetApiPrice(costType).Mul(decimal.NewFromInt(int64(count)))
	return UpdateUserBalanceWithTx(tx, userId, amount, FIAT_LOG_TYPE_REFUND_API_FEE, FiatMetaRefundApiFee{costType, count}, false)
}

func PayAPIFee(tx *gorm.DB, userId uint, costType enums.CostType, count uint) (uint, error) {
	amount := GetApiPrice(costType).Mul(decimal.NewFromInt(int64(count)))
	return UpdateUserBalanceWithTx(tx, userId, decimal.Zero.Sub(amount), FIAT_LOG_TYPE_PAY_API_FEE, FiatMetaPayApiFee{costType, count}, false)
}

func UpdateUserBalance(userId uint, amount decimal.Decimal, logType FiatLogType, meta interface{}, checkBalance ...bool) (uint, error) {
	var fiatLogId uint
	err := GetDB().Transaction(func(tx *gorm.DB) error {
		l, err := UpdateUserBalanceWithTx(tx, userId, amount, logType, meta, checkBalance...)
		fiatLogId = l
		return err
	})
	return fiatLogId, err
}

func UpdateUserBalanceWithTx(tx *gorm.DB, userId uint, amount decimal.Decimal, logType FiatLogType, meta interface{}, checkBalance ...bool) (uint, error) {

	fl, err := func() (uint, error) {
		if err := checkDecimalQualified(amount); err != nil {
			return 0, err
		}

		userBalance := UserBalance{
			UserId: userId,
		}

		if err := tx.Where(&userBalance).Find(&userBalance).Error; err != nil {
			return 0, err
		}

		if (len(checkBalance) == 0 || checkBalance[0]) && userBalance.Balance.Add(userBalance.ArrearsQuota).Add(amount).Cmp(decimal.Zero) < 0 {
			return 0, errors.New("insufficient balance")
		}

		metaJson, _ := json.Marshal(meta)
		l := FiatLogCache{
			FiatLogCore: FiatLogCore{
				UserId:  userId,
				Amount:  amount,
				Type:    logType,
				Meta:    metaJson,
				OrderNO: RandomOrderNO(),
				Balance: userBalance.Balance.Add(amount),
			},
		}
		if err := tx.Create(&l).Error; err != nil {
			return 0, err
		}

		userBalance.Balance = userBalance.Balance.Add(amount)
		// if freeApiCost != nil {
		// 	userBalance.FreeOtherApiQuota -= freeApiCost.FreeOtherApi
		// 	userBalance.FreeMintQuota -= freeApiCost.FreeMint
		// 	userBalance.FreeDeployQuota -= freeApiCost.FreeDeploy
		// }

		return l.ID, tx.Save(&userBalance).Error
	}()

	logrus.WithFields(logrus.Fields{
		"userId": userId,
		"amount": amount,
		// "freeApiCost": freeApiCost,
		"logType":   logType,
		"fiatLogId": fl,
	}).WithError(err).Info("update user balance")

	return fl, err
}

func UpdateUserArrearsQuota(userId uint, amount decimal.Decimal) error {
	var ub UserBalance
	if err := GetDB().Where(&UserBalance{UserId: userId}).First(&ub).Error; err != nil {
		return err
	}
	ub.ArrearsQuota = amount
	return GetDB().Save(&ub).Error
}

// func UpdateUserFreeQuotas(userId uint, quotas FreeApiQuota) error {
// 	var ub UserBalance
// 	if err := GetDB().Where(&UserBalance{UserId: userId}).First(&ub).Error; err != nil {
// 		return err
// 	}
// 	ub.FreeApiQuota = quotas
// 	return GetDB().Save(&ub).Error
// }

// func ResetAllUserFreeApiQuota() error {
// 	currntMonth := utils.CurrentMonthStr()

// 	return GetDB().Transaction(func(tx *gorm.DB) error {
// 		if err := GetDB().Model(&UserBalance{}).
// 			Where("free_api_work_month<?", currntMonth).
// 			Or("free_api_work_month is NULL").
// 			Updates(map[string]interface{}{
// 				"free_other_api_quota": fee.UserDefaultFreeOtherAPIQuota,
// 				"free_mint_quota":      fee.UserDefaultFreeMintQuota,
// 				"free_deploy_quota":    fee.UserDefaultFreeDeployQuota,
// 				"free_api_work_month":  utils.CurrentMonthStr(),
// 			}).
// 			Error; err != nil {
// 			return err
// 		}

// 		return initUsersRemainBalanceQuota()
// 	})
// }

// func (u *UserBalance) AfterSave(tx *gorm.DB) error {
// 	if u.ID == 0 {
// 		return nil
// 	}
// 	err := refreshUserRemainBalanceQuota(tx, u)
// 	logrus.WithError(err).WithField("user balance", u).Info("save user balance")
// 	return err
// }

func (u *UserBalance) BalanceWithArrears() decimal.Decimal {
	return u.ArrearsQuota.Add(u.Balance)
}

func GetUserCfxPrice(userId uint) (decimal.Decimal, error) {
	normalCfxPrice := decimal.NewFromFloat(cfxPrice)

	ub, err := GetUserBalance(userId)
	if err != nil {
		if gormutils.IsRecordNotFoundError(err) {
			return normalCfxPrice, nil
		}
		return decimal.Zero, err
	}
	if ub.CfxPrice.IsZero() {
		return normalCfxPrice, nil
	}
	return ub.CfxPrice, nil
}

func UpdateUserCfxPrice(userId uint, price decimal.Decimal) error {
	if err := checkDecimalQualified(price); err != nil {
		return err
	}
	return GetDB().Model(&UserBalance{}).Where(&UserBalance{UserId: userId}).Update("cfx_price", price).Error
}

func checkDecimalQualified(amount decimal.Decimal) error {
	amountX100 := amount.Mul(decimal.NewFromInt(100))
	if amountX100.Cmp(amountX100.Floor()) != 0 {
		return errors.Errorf("the decimal place of value cannot be greater than 2, but got %v", amount)
	}
	return nil
}
