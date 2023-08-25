package models

import (
	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
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
