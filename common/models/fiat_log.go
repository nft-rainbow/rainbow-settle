package models

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PayType int

const (
	PAY_TYPE_WX PayType = iota + 1
	PAY_TYPE_CMB
	PAY_TYPE_BALANCE_REFUND_OTHER
	PAY_TYPE_BALANCE_REFUND_SPONSOR
	PAY_TYPE_BALANCE = 10
)

type FiatLogType int

const (
	FIAT_LOG_TYPE_DEPOSIT             FiatLogType = iota + 1
	FIAT_LOG_TYPE_WITHDRAW                        //2
	FIAT_LOG_TYPE_BUY_GAS                         //3
	FIAT_LOG_TYPE_BUY_STORAGE                     //4
	FIAT_LOG_TYPE_PAY_API_FEE                     //5
	FIAT_LOG_TYPE_CMB_CHARGE                      // 6招行对公充值
	FIAT_LOG_TYPE_REFUND_API_FEE                  //7
	FIAT_LOG_TYPE_REFUND_SPONSOR                  //8
	FIAT_LOG_TYPE_REFUND_API_QUOTA                //9
	FIAT_LOG_TYPE_REFUND_RESERV3_2                //10
	FIAT_LOG_TYPE_REFUND_RESERV3_3                //11
	FIAT_LOG_TYPE_PAY_API_QUOTA                   //12
	FIAT_LOG_TYPE_RESET_API_QUOTA                 //13
	FIAT_LOG_TYPE_DEPOSITE_DATABUNDLE             //14
	FIAT_LOG_TYPE_BUY_BILLPLAN
	FIAT_LOG_TYPE_BUY_DATABUNDLE
)

func (f FiatLogType) PayType() PayType {
	switch f {
	case FIAT_LOG_TYPE_DEPOSIT:
		return PAY_TYPE_WX
	case FIAT_LOG_TYPE_CMB_CHARGE:
		return PAY_TYPE_CMB
	case FIAT_LOG_TYPE_REFUND_API_FEE:
		return PAY_TYPE_BALANCE_REFUND_OTHER
	case FIAT_LOG_TYPE_REFUND_SPONSOR:
		return PAY_TYPE_BALANCE_REFUND_SPONSOR
	}
	return PAY_TYPE_BALANCE
}

type FiatLogCore struct {
	UserId  uint            `gorm:"type:int;index;index:idx_type_merge_deleted_user,priority:20" json:"user_id"`
	Amount  decimal.Decimal `gorm:"type:decimal(20,2)" json:"amount"`                                 // 单位分
	Type    FiatLogType     `gorm:"type:int;index:idx_type_merge_deleted_user;default:0" json:"type"` // 1-deposit
	Meta    datatypes.JSON  `gorm:"type:json" json:"meta"`                                            // metadata
	OrderNO string          `gorm:"type:varchar(255);unique" json:"order_no"`                         // order NO in rainbow platform
	Balance decimal.Decimal `gorm:"type:decimal(20,2)" json:"balance"`                                // apply log balance
}
type FiatLog struct {
	BaseModel
	FiatLogCore
	CacheIds     datatypes.JSONSlice[uint] `json:"cache_ids"`
	InvoiceId    *uint                     `gorm:"type:int;index" json:"invoice_id"` // 发票id, 如果某条消费 log 已开发票, 此字段会有值
	RefundLogIds datatypes.JSONSlice[uint] `json:"refund_log_ids"`                   // 对应的退款 fiatLog。说明当前 fiatLog 中有退款，meta中的 refunded_amount 为该fiatLog中被退款金额
}

func (f *FiatLog) AfterCreate(tx *gorm.DB) (err error) {
	return UpdateUserBalanceOnFiatlog(tx, f.UserId, f.Balance)
}

func FindFiatLogs(userId uint, offset int, limit int) (*[]FiatLog, error) {
	var logs []FiatLog
	res := db.Model(&FiatLog{}).
		Where("user_id = ? AND amount != 0", userId).Order("id desc").Limit(limit).Offset(offset).Find(&logs)
	return &logs, res.Error
}

func FindUserFiatLogsByIds(userId uint, logIds []uint) (*[]FiatLog, error) {
	var logs []FiatLog
	res := db.Model(&FiatLog{}).
		Where("user_id = ? AND id IN ?", userId, logIds).Find(&logs)
	return &logs, res.Error
}

func FindSponsorFiatlogByTxid(txId uint) (*FiatLog, error) {
	var fl FiatLog
	if err := db.Model(&FiatLog{}).Where("meta->'$.tx_id'=?", txId).
		Where("type =? or type=?", FIAT_LOG_TYPE_BUY_GAS, FIAT_LOG_TYPE_BUY_STORAGE).
		First(&fl).Error; err != nil {
		return nil, err
	}
	return &fl, nil
}

// If not found, return default FiatLog
func GetLastFiatLog(tx *gorm.DB, userId uint) (*FiatLog, error) {
	var lastFiatLog FiatLog
	if err := tx.Model(&FiatLog{}).Where("user_id=?", userId).Order("created_at desc,id desc").First(&lastFiatLog).Error; err != nil {
		return nil, err
	}
	return &lastFiatLog, nil
}

func GetLastBlanceByFiatlog(tx *gorm.DB, userId uint) (decimal.Decimal, error) {
	lastFiatLog, err := GetLastFiatLog(tx, userId)
	if err != nil {
		if !gormutils.IsRecordNotFoundError(err) {
			return decimal.Zero, err
		}
		ub, err := GetUserBalance(tx, userId)
		if err != nil {
			return decimal.Zero, err
		}
		return ub.Balance, nil
	}
	return lastFiatLog.Balance, nil
}

func RandomOrderNO() string {
	return fmt.Sprintf("NR%s%d", utils.CompactFormatTime(time.Now()), utils.RandomNumber(1000, 9999))
}

func UserFiatLogCount(userId uint) int64 {
	var count int64
	GetDB().Model(&FiatLog{}).Where("user_id = ? AND amount != 0", userId).Count(&count)
	return count
}

func RelateBuySponsorFiatlog(tx *gorm.DB, refundSponsorFl *FiatLog) error {
	switch refundSponsorFl.Type {
	case FIAT_LOG_TYPE_REFUND_SPONSOR:
		// find related fiat log and set refundMeta
		var refundMeta FiatMetaRefundSponsor
		if err := json.Unmarshal(refundSponsorFl.Meta, &refundMeta); err != nil {
			return err
		}

		var relatedFiatlog FiatLog
		if err := tx.Model(&FiatLog{}).Where("id=?", refundMeta.RefundForFiatlogId).First(&relatedFiatlog).Error; err != nil {
			return err
		}

		if lo.Contains(relatedFiatlog.RefundLogIds, refundSponsorFl.ID) {
			return nil
		}

		relatedFiatlog.RefundLogIds = append(relatedFiatlog.RefundLogIds, refundSponsorFl.ID)

		var sponsorMeta FiatMetaBuySponsor
		if err := json.Unmarshal(relatedFiatlog.Meta, &sponsorMeta); err != nil {
			return err
		}

		sponsorMeta.RefundedAmount = refundSponsorFl.Amount
		sponsorMetaStr, _ := json.Marshal(sponsorMeta)
		relatedFiatlog.Meta = sponsorMetaStr
		err := tx.Save(&relatedFiatlog).Error
		logrus.WithError(err).WithField("pay_sponsor_fiatlog", relatedFiatlog.ID).WithField("refund_sponsor_fiatlog", refundSponsorFl.ID).WithField("pay fiatlog meta", string(sponsorMetaStr)).Info("relate refund with pay sponsor fiatlog")
		if err != nil {
			return err
		}
	}
	return nil
}

type FiatLogWithDetails struct {
	FiatLog
	Email   string `json:"email"`
	TradeNo string `json:"trade_no"`
	Note    string `json:"note"`
}

type FiatLogFilter struct {
	UserId    uint        `form:"user_id" json:"user_id"`
	OrderNO   string      `form:"order_no" json:"order_no"`
	StartedAt *string     `form:"started_at" json:"started_at"`
	EndedAt   *string     `form:"ended_at" json:"ended_at"`
	Type      FiatLogType `form:"type" json:"type"`
	Address   *string     `form:"address" json:"address"`
	InvoiceId *uint       `form:"invoice_id" json:"invoice_id"`
}

func (filter *FiatLogFilter) Where() *gorm.DB {
	where := db.Where("fiat_logs.amount != 0")
	if filter.StartedAt != nil && len(*filter.StartedAt) > 0 {
		where = where.Where("fiat_logs.created_at >= ?", filter.StartedAt)
	}
	if filter.EndedAt != nil && len(*filter.EndedAt) > 0 {
		where = where.Where("fiat_logs.created_at < ?", filter.EndedAt)
	}
	if filter.Address != nil && len(*filter.Address) > 0 {
		where = where.Where("fiat_logs.meta->'$.address' = ?", filter.Address)
	}
	if filter.UserId != 0 {
		where = where.Where("fiat_logs.user_id=?", filter.UserId)
	}
	if filter.OrderNO != "" {
		where = where.Where("fiat_logs.order_no=?", filter.OrderNO)
	}
	if filter.Type != 0 {
		where = where.Where("fiat_logs.type=?", filter.Type)
	}
	if filter.InvoiceId != nil && *filter.InvoiceId != 0 {
		where = where.Where("fiat_logs.invoice_id=?", filter.InvoiceId)
	}
	return where
}

func FindAndCountFiatLogs(filter FiatLogFilter, offset, limit int) (logs *[]FiatLog, count int64, err error) {
	err = db.Debug().Model(&FiatLog{}).Where(filter.Where()).Count(&count).Order("created_at desc").Order("id desc").Offset(offset).Limit(limit).Find(&logs).Error
	return
}

type FiatlogWithDetailFilter struct {
	FiatLogFilter
	UserPayTypes []enums.UserPayType `form:"user_pay_types" json:"user_pay_types"`
}

func (filter *FiatlogWithDetailFilter) Where() *gorm.DB {
	where := filter.FiatLogFilter.Where()
	if len(filter.UserPayTypes) > 0 {
		where = where.Where("users.user_pay_type in ?", filter.UserPayTypes)
	}
	return where
}

func FindAndCountFiatLogWithDetails(filter FiatlogWithDetailFilter, offset, limit int) (logs []*FiatLogWithDetails, count int64, err error) {
	table := db.Debug().Model(&FiatLog{}).
		Where(filter.Where()).
		Joins("LEFT JOIN deposit_orders on fiat_logs.meta->'$.deposit_order_id'=deposit_orders.id").
		Where("deposit_orders.status=1 or deposit_orders.status is null").
		Joins("LEFT JOIN users on fiat_logs.user_id = users.id").
		Select("fiat_logs.*, users.email,  deposit_orders.trade_no")

	err = table.Count(&count).Order("fiat_logs.created_at desc").Offset(offset).Limit(limit).Find(&logs).Error
	return
}

type TimeWindowFilter struct {
	StartedAt time.Time `form:"started_at" json:"started_at"`
	EndedAt   time.Time `form:"ended_at" json:"ended_at"`
}

func (b *TimeWindowFilter) SetDefaults() {
	defaultTime := time.Time{}
	now := time.Now()
	if b.StartedAt == defaultTime {
		b.StartedAt = utils.BeginningOfMonth(now)
	}
	if b.EndedAt == defaultTime {
		b.EndedAt = utils.BeginnigOfNextMonth(now)
	}
}

type FiatlogSummaryFilter struct {
	TimeWindowFilter
	UserPayTypes []enums.UserPayType `form:"user_pay_types" json:"user_pay_types"`
	Types        []FiatLogType       `form:"type" json:"type"`
	UserIds      []uint              `form:"user_ids" json:"user_ids"`
}

func (f *FiatlogSummaryFilter) Where() *gorm.DB {
	where := GetDB()
	if len(f.UserPayTypes) > 0 {
		where = where.Where("users.user_pay_type in ?", f.UserPayTypes)
	}

	if len(f.Types) > 0 {
		where = where.Where("fiat_logs.type in (?)", f.Types)
	}

	if len(f.UserIds) > 0 {
		where = where.Where("fiat_logs.user_id in (?)", f.UserIds)
	}
	return where
}

type FiatlogSummayItem struct {
	UserId  uint            `gorm:"type:int;index" json:"user_id"`
	Email   string          `json:"email"`
	Type    FiatLogType     `form:"fiat_log_type" json:"type"`
	Amount  decimal.Decimal `json:"amount"` // 单位分
	PayType PayType         `form:"pay_type" json:"pay_type"`
}

func FiatLogsMonthSummary(cond FiatlogSummaryFilter, offset, limit int) (items []*FiatlogSummayItem, count int64, err error) {
	cond.SetDefaults()
	logrus.WithField("filter", cond).Info("get FiatLogsMonthSummary")

	sql := db.Debug().Model(&FiatLog{}).
		Joins("LEFT JOIN users on fiat_logs.user_id = users.id").
		Select("fiat_logs.user_id,fiat_logs.type,sum(fiat_logs.amount) as amount,users.email").
		Where("fiat_logs.amount != 0").
		Where("fiat_logs.created_at >= ?", cond.StartedAt).
		Where("fiat_logs.created_at < ?", cond.EndedAt).
		Where(cond.Where()).
		Group("fiat_logs.type").
		Group("fiat_logs.user_id").
		Order("fiat_logs.user_id asc").
		Order("fiat_logs.type asc")

	if err := sql.Count(&count).Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	for _, item := range items {
		item.PayType = item.Type.PayType()
	}

	return items, count, nil
}

type FiatlogChangesItem struct {
	UserId           uint            `json:"user_id"`
	Email            string          `json:"email"`
	Deposit          decimal.Decimal `json:"deposit"`
	Withdraw         decimal.Decimal `json:"withdraw"`
	Gas              decimal.Decimal `json:"gas"`
	Storage          decimal.Decimal `json:"storage"`
	BillPlan         decimal.Decimal `json:"bill_plan"`
	DataBundle       decimal.Decimal `json:"data_bundle"`
	ApiFee           decimal.Decimal `json:"api_fee"`
	CmbDeposit       decimal.Decimal `json:"cmb_deposit"`
	RefundForSponsor decimal.Decimal `json:"refund_for_sponsor"`
	RefundForOther   decimal.Decimal `json:"refund_for_other"`
	StartBalance     decimal.Decimal `json:"start_balance"`
	EndBalance       decimal.Decimal `json:"end_balance"`
}

func FiatlogOfBalanceChanges(cond FiatlogSummaryFilter, offset, limit int) ([]*FiatlogChangesItem, int64, error) {
	cond.SetDefaults()
	logrus.WithField("filter", cond).Info("get FiatlogOfBalanceChanges")

	userIds, count, err := getHasChangesUserIds(cond, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// get month summary of users
	cond.UserIds = userIds
	summaryItems, _, err := FiatLogsMonthSummary(cond, 0, math.MaxInt)
	if err != nil {
		return nil, 0, err
	}

	// convert to changes
	userChanges := map[uint]*FiatlogChangesItem{}
	for _, summaryItem := range summaryItems {
		userId := summaryItem.UserId
		if _, ok := userChanges[userId]; !ok {
			userChanges[userId] = &FiatlogChangesItem{
				UserId: userId,
				Email:  summaryItem.Email,
			}
		}
		switch summaryItem.Type {
		case FIAT_LOG_TYPE_DEPOSIT:
			userChanges[userId].Deposit = summaryItem.Amount
		case FIAT_LOG_TYPE_WITHDRAW:
			userChanges[userId].Withdraw = summaryItem.Amount
		case FIAT_LOG_TYPE_BUY_GAS:
			userChanges[userId].Gas = summaryItem.Amount
		case FIAT_LOG_TYPE_BUY_STORAGE:
			userChanges[userId].Storage = summaryItem.Amount
		case FIAT_LOG_TYPE_PAY_API_FEE:
			userChanges[userId].ApiFee = summaryItem.Amount
		case FIAT_LOG_TYPE_BUY_BILLPLAN:
			userChanges[userId].BillPlan = summaryItem.Amount
		case FIAT_LOG_TYPE_BUY_DATABUNDLE:
			userChanges[userId].DataBundle = summaryItem.Amount
		case FIAT_LOG_TYPE_CMB_CHARGE: // 招行对公充值
			userChanges[userId].CmbDeposit = summaryItem.Amount
		case FIAT_LOG_TYPE_REFUND_API_FEE:
			userChanges[userId].RefundForOther = summaryItem.Amount
		case FIAT_LOG_TYPE_REFUND_SPONSOR:
			userChanges[userId].RefundForSponsor = summaryItem.Amount
		}
	}

	startBalances, err := GetUserBalanceAtDate(userIds, cond.StartedAt)
	if err != nil {
		return nil, 0, err
	}

	endBalances, err := GetUserBalanceAtDate(userIds, cond.EndedAt)
	if err != nil {
		return nil, 0, err
	}

	users, err := FindUserByIds(userIds)
	if err != nil {
		return nil, 0, err
	}

	changeItems := []*FiatlogChangesItem{}
	for i, userId := range userIds {
		if userChanges[userId] == nil {
			userChanges[userId] = &FiatlogChangesItem{
				UserId: userId,
			}
			if users[userId] != nil {
				userChanges[userId].Email = users[userId].Email
			}
		}
		uc := userChanges[userId]
		uc.StartBalance = startBalances[i]
		uc.EndBalance = endBalances[i]
		changeItems = append(changeItems, uc)
	}

	return changeItems, count, nil
}

func FindFiatlogsWithoutSponsorlog(start, end time.Time) ([]*FiatLog, error) {
	fiatLogIdsRefunded, err := FindSponsorFiatlogsRefunded(start, end)
	if err != nil {
		return nil, err
	}

	var fls []*FiatLog
	err = GetDB().
		// Debug().
		Table("fiat_logs").Where("fiat_logs.deleted_at is null").
		Where("fiat_logs.created_at>=?", start).Where("fiat_logs.created_at<?", end).
		Where("fiat_logs.type in (?)", []FiatLogType{FIAT_LOG_TYPE_BUY_GAS, FIAT_LOG_TYPE_BUY_STORAGE}).
		Joins("left join transactions on transactions.id=fiat_logs.meta->\"$.tx_id\"").
		Joins("left join sponsor_logs on sponsor_logs.hash=transactions.hash").
		Where("sponsor_logs.id is null").
		Where("fiat_logs.id not in (?)", fiatLogIdsRefunded).
		Select("fiat_logs.*").
		Scan(&fls).Error
	if err != nil {
		return nil, err
	}

	// find if has refund fiatlog match them
	flIds := GetIds(fls)
	// select a.id,b.id from fiat_logs as a left join fiat_logs as b on a.id=b.meta->'$.refund_for_fiatlog_id' where a.id in (flIds)  and b.id is null;

	fls = nil
	err = GetDB().
		// Debug().
		Table("fiat_logs as a").
		Joins("left join fiat_logs as b on a.id=b.meta->'$.refund_for_fiatlog_id'").
		Where("a.id in ?", flIds).
		Where("b.id is null").
		Select("a.*").
		Scan(&fls).Error
	if err != nil {
		return nil, err
	}

	return fls, nil
}

func getHasChangesUserIds(cond FiatlogSummaryFilter, offset, limit int) (userIds []uint, count int64, err error) {
	// count is user count
	cond.SetDefaults()

	sql := db.Debug().Model(&FiatLog{}).
		Joins("left join users on fiat_logs.user_id=users.id").
		Select("fiat_logs.user_id").
		Where(cond.Where()).
		Where("fiat_logs.amount != 0").
		Where("fiat_logs.created_at < ?", cond.EndedAt).
		Group("fiat_logs.user_id").
		Order("fiat_logs.user_id asc")

	if err := sql.Count(&count).Offset(offset).Limit(limit).Find(&userIds).Error; err != nil {
		return nil, 0, err
	}
	return userIds, count, nil
}

// group by user_id and select the max id of max creted_at desc order by id,
// that's becasue manual added records exists with large id but small created_at.
func GetUserBalanceAtDate(userIds []uint, date time.Time) ([]decimal.Decimal, error) {

	type userMaxCreatedat struct {
		UserId    uint
		CreatedAt time.Time
	}

	// select user_id,max(created_at) from fiat_logs group by user_id where created_at<date;
	var userMaxCreatedats []*userMaxCreatedat
	err := db.Debug().Model(&FiatLog{}).Where("created_at<=?", date).Group("user_id").Select("user_id,max(created_at) as created_at").Scan(&userMaxCreatedats).Error
	if err != nil {
		return nil, err
	}

	var w sync.WaitGroup

	balances := make([]decimal.Decimal, len(userIds))
	for i, userId := range userIds {
		var err error
		for _, v := range userMaxCreatedats {
			if v.UserId == userId {
				w.Add(1)
				go func(_i int) {
					defer w.Done()
					// select * from fiat_logs where created_at="2023-10-27 18:14:48.582" and user_id=1 order by id desc limit 1;
					var fl FiatLog
					_err := db.Debug().Model(&FiatLog{}).Where("created_at=? and user_id=?", v.CreatedAt, v.UserId).Order("id desc").First(&fl).Error
					if _err != nil {
						err = _err
						return
					}
					balances[_i] = fl.Balance
				}(i)

				break
			}
		}
		if i%5 == 0 || i == len(userIds)-1 {
			w.Wait()
			if err != nil {
				return nil, err
			}
		}
	}
	return balances, nil
}

func FindSponsorFiatlogsRefunded(start, end time.Time) ([]uint, error) {
	var fiatLogs []*FiatLog
	err := GetDB().Model(&FiatLog{}).Where("type in (?)", []FiatLogType{FIAT_LOG_TYPE_BUY_GAS, FIAT_LOG_TYPE_BUY_STORAGE}).Where("created_at>=?", start).Where("created_at<?", end).Where("refund_log_ids!=CAST('null' AS JSON)").Find(&fiatLogs).Error
	if err != nil {
		return nil, err
	}

	return GetIds(fiatLogs), nil
}
