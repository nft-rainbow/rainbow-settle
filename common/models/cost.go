package models

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Cost struct {
	BaseModel
	UserId    uint           `gorm:"index" json:"user_id"`
	Method    string         `gorm:"varchar(16);index" json:"method"`
	Path      string         `gorm:"type:varchar(64);index" json:"path"`
	IsTestnet bool           `gorm:"type:boolean;index" json:"is_testnet"`
	Count     int            `gorm:"default:0" json:"count"`
	Date      string         `gorm:"type:varchar(10);index" json:"date"`
	Settled   bool           `gorm:"type:bool" json:"settled"`
	IsRefund  bool           `gorm:"type:bool" json:"is_refund"`
	CostType  enums.CostType `gorm:"type:string" json:"cost_type"`
}

type userRemain struct {
	// initial      UserBalance
	balanceQuota decimal.Decimal
	freeApiQuota *FreeApiQuota
}

var (
	userRemainQuotas sync.Map
	getCostLock      func(userId uint) *sync.RWMutex = costLockObtainer()
)

func NewCost(userId uint, method string, path string, isTestnet bool, count uint, date string, isRefund bool, settled bool) *Cost {
	c := &Cost{
		UserId:    userId,
		Method:    method,
		Path:      path,
		IsTestnet: isTestnet,
		Count:     int(count),
		Date:      date,
		IsRefund:  isRefund,
		Settled:   settled,
		CostType:  enums.GetCostType(isTestnet, method, path),
	}
	return c
}

func (c *Cost) IsMainnetMint() bool {
	return enums.GetCostType(c.IsTestnet, c.Method, c.Path) == enums.COST_TYPE_RAINBOW_MINT
}

func (c *Cost) IsMainnetDeploy() bool {
	return enums.GetCostType(c.IsTestnet, c.Method, c.Path) == enums.COST_TYPE_RAINBOW_DEPLOY
}

func FindAllUserCostUnSettledBeforeToday() (map[uint][]*Cost, error) {
	userIds := []uint{}
	if err := GetDB().Model(&Cost{}).Where("date < ? and settled = ?", utils.TodayDateStr(), false).Distinct("user_id").Find(&userIds).Error; err != nil {
		return nil, err
	}

	result := make(map[uint][]*Cost)
	for _, userId := range userIds {
		records, err := FindUserCostUnSettled(GetDB(), userId)
		if err != nil {
			return nil, err
		}
		result[userId] = records
	}

	return result, nil
}

func FindUserCostUnSettled(tx *gorm.DB, userId uint) ([]*Cost, error) {
	records := []*Cost{}
	if err := tx.Where("user_id = ? and date <= ? and settled = ?", userId, utils.TodayDateStr(), false).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func initUsersRemainBalanceQuota() error {
	userRemainQuotas = sync.Map{}

	defer func() {
		logrus.Info("init user remains balance quota")
	}()

	userIds := []uint{}
	if err := GetDB().Model(&User{}).Distinct("id").Find(&userIds).Error; err != nil {
		return err
	}

	for _, userId := range userIds {
		ub, err := GetUserBalance(userId)
		if err != nil {
			return err
		}
		if err := refreshUserRemainBalanceQuota(GetDB(), ub); err != nil {
			return err
		}
	}
	return nil
}

func RefreshUserRemainBalanceQuota(tx *gorm.DB, userId uint) error {
	ub, err := GetUserBalance(userId)
	if err != nil {
		return err
	}
	return refreshUserRemainBalanceQuota(tx, ub)
}

// userbalance变化后调用以刷新内存数据
func refreshUserRemainBalanceQuota(tx *gorm.DB, ub *UserBalance) error {
	userId := ub.UserId
	getCostLock(userId).Lock()
	defer getCostLock(userId).Unlock()

	costs, err := FindUserCostUnSettled(tx, userId)
	if err != nil {
		return err
	}

	// 先减次数，再减剩余额度
	ur := &userRemain{
		// initial:      *ub,
		balanceQuota: ub.Balance.Add(ub.ArrearsQuota),
		freeApiQuota: &ub.FreeApiQuota,
	}

	for _, c := range costs {
		calcUserRemain(ur, c, false)
	}

	userRemainQuotas.Store(userId, ur)

	logrus.WithField("userId", userId).WithFields(logrus.Fields{"remain balanceQuota": ur.balanceQuota, "remain apiFree": ur.freeApiQuota}).Debug("refresh user remains balance quota")

	return nil
}

func calcUserRemain(ur *userRemain, cost *Cost, verfiy bool) error {
	free, unfree := CalcFreeApiCnt(cost, ur.freeApiQuota)
	costFee := CalcApiFee(cost, &unfree)

	if verfiy && ur.balanceQuota.Cmp(costFee) < 0 {
		return fmt.Errorf("balance quota is not enough, need %v got %v", costFee, ur.balanceQuota)
	}
	ur.balanceQuota = ur.balanceQuota.Sub(costFee)
	ur.freeApiQuota.Decrease(free)
	return nil
}

func CalcApiFee(c *Cost, unfree *FreeApiUsed) decimal.Decimal {
	unitPrice := decimal.NewFromInt(0) //getApiPrice(c.Method, c.Path, c.IsTestnet)

	switch c.CostType {
	case enums.COST_TYPE_RAINBOW_MINT:
		return unitPrice.Mul(decimal.NewFromInt(int64(unfree.FreeMint)))
	case enums.COST_TYPE_RAINBOW_DEPLOY:
		return unitPrice.Mul(decimal.NewFromInt(int64(unfree.FreeDeploy)))
	default:
		return unitPrice.Mul(decimal.NewFromInt(int64(unfree.FreeOtherApi)))
	}
}

// 1st returns free api used; 2nd returns unfree api used.
func CalcFreeApiCnt(c *Cost, remainFreeQuota *FreeApiQuota) (FreeApiUsed, FreeApiUsed) {
	var free, unfree FreeApiUsed

	count := c.Count
	if c.IsRefund {
		count = 0 - count
	}

	// 如果是mint 或 deploy，减mint或deploy
	if c.IsMainnetMint() {
		// if !c.IsRefund {
		if remainFreeQuota.FreeMintQuota >= count {
			free.FreeMint = count
			return free, unfree
		} else {
			free.FreeMint = remainFreeQuota.FreeMintQuota
			unfree.FreeMint = count - remainFreeQuota.FreeMintQuota
			return free, unfree
		}
		// }

	}

	if c.IsMainnetDeploy() {
		if remainFreeQuota.FreeDeployQuota >= count {
			free.FreeDeploy = count
			return free, unfree
		} else {
			free.FreeDeploy = remainFreeQuota.FreeDeployQuota
			unfree.FreeDeploy = count - remainFreeQuota.FreeDeployQuota
			return free, unfree
		}
	}

	// 如果不是，减other
	if remainFreeQuota.FreeOtherApiQuota >= count {
		free.FreeOtherApi = count
		return free, unfree
	} else {
		free.FreeOtherApi = remainFreeQuota.FreeOtherApiQuota
		unfree.FreeOtherApi = count - remainFreeQuota.FreeOtherApiQuota
		return free, unfree
	}
}

func costLockObtainer() func(userId uint) *sync.RWMutex {
	costLocks := make(map[uint]*sync.RWMutex)
	var initLock sync.Mutex

	f := func(userId uint) *sync.RWMutex {

		if v, ok := costLocks[userId]; ok {
			return v
		}

		initLock.Lock()
		if _, ok := costLocks[userId]; !ok {
			costLocks[userId] = &sync.RWMutex{}
		}
		initLock.Unlock()

		return costLocks[userId]
	}
	return f
}

// 每有请求导致cost时更新内存
func updateUserRemainsQuota(userId uint, method string, path string, isTestnet bool, count uint, isRefund bool) error {
	getCostLock(userId).Lock()
	defer getCostLock(userId).Unlock()

	current, ok := userRemainQuotas.Load(userId)
	if !ok {
		return fmt.Errorf("unknown users %v", userId)
	}
	ur := current.(*userRemain)

	cost := NewCost(userId, method, path, isTestnet, count, "", isRefund, false)
	if err := calcUserRemain(ur, cost, true); err != nil {
		return err
	}

	userRemainQuotas.Store(userId, ur)
	logrus.WithFields(logrus.Fields{"userId": userId, "method": method, "path": path, "isTestnet": isTestnet, "count": count, "remain balanceQuota": ur.balanceQuota, "remain apiFreeQuota": ur.freeApiQuota}).Trace("update user remains quota in memory")

	return nil
}

func GetRuntimeUserBalance(userId uint) (*UserBalance, error) {
	ub, err := GetUserBalance(userId)
	if err != nil {
		return nil, err
	}

	current, ok := userRemainQuotas.Load(userId)
	if !ok {
		return nil, fmt.Errorf("unknown users %v", userId)
	}
	ur := current.(*userRemain)

	ub.Balance = ur.balanceQuota.Sub(ub.ArrearsQuota)
	ub.FreeApiQuota = *ur.freeApiQuota
	return ub, nil
}

func RecordCost(userId uint, method string, path string, isTestnet bool, count uint) error {
	if method == http.MethodOptions {
		return nil
	}

	if err := updateUserRemainsQuota(userId, method, path, isTestnet, count, false); err != nil {
		return err
	}

	if err := func() error {
		getCostLock(userId).Lock()
		defer getCostLock(userId).Unlock()

		date := utils.TodayDateStr()
		res := db.Debug().Model(&Cost{}).Where("user_id = ? AND method = ? AND path = ? And is_testnet = ? AND date = ? AND settled = 0", userId, method, path, isTestnet, date).
			Update("count", gorm.Expr(fmt.Sprintf("count + %d", count)))
		if res.Error == nil && res.RowsAffected == 0 {
			return db.Create(NewCost(userId, method, path, isTestnet, count, date, false, false)).Error
		}
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func RefundCost(tx *gorm.DB, userId uint, method string, path string, isTestnet bool, count uint) error {
	logrus.Info("fall back cost")
	if err := func() error {
		getCostLock(userId).RLock()
		defer getCostLock(userId).RUnlock()

		date := utils.TodayDateStr()
		res := tx.Model(&Cost{}).Where("user_id = ? AND method = ? AND path = ? And is_testnet = ? AND date = ? AND settled = 0 And is_refund = 1", userId, method, path, isTestnet, date).
			Update("count", gorm.Expr(fmt.Sprintf("count - %d", count)))
		if res.Error == nil && res.RowsAffected == 0 {
			return tx.Create(NewCost(userId, method, path, isTestnet, count, date, true, false)).Error
		}
		return res.Error
	}(); err != nil {
		return err
	}

	ub, err := GetUserBalance(userId)
	if err != nil {
		return err
	}
	return refreshUserRemainBalanceQuota(db, ub)
}
