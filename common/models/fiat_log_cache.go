package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/mathutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	apiQuotaRelatedFiatLogTypes = []FiatLogType{FIAT_LOG_TYPE_PAY_API_QUOTA, FIAT_LOG_TYPE_REFUND_API_QUOTA, FIAT_LOG_TYPE_RESET_API_QUOTA}
	apiFeeRelatedFiatLogTypes   = []FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_REFUND_API_FEE}
	apiRelatedFiatLogTypes      = append(apiQuotaRelatedFiatLogTypes, apiFeeRelatedFiatLogTypes...)
	mergeFiatLogLock            sync.Mutex
)

func lockMergeFiatLogMutex() {
	logrus.Debug("start lock merge fiat log mutex")
	mergeFiatLogLock.Lock()
	logrus.Debug("merge fiat log mutex locked")
}

func unlockMergeFiatLogMutex() {
	logrus.Debug("start unlock merge fiat log mutex")
	mergeFiatLogLock.Unlock()
	logrus.Debug("merge fiat log mutex unlocked")
}

type FiatLogCache struct {
	BaseModel
	FiatLogCore
	UnsettleAmount decimal.Decimal `gorm:"type:decimal(20,10)" json:"unsettle_amount"`
	IsMerged       bool            `gorm:"default:0" json:"isMerged,omitempty"`
}

func (f *FiatLogCache) AfterCreate(tx *gorm.DB) (err error) {
	if f.IsMerged || lo.Contains(apiRelatedFiatLogTypes, f.Type) {
		return nil
	}
	f.IsMerged = true
	if err := tx.Save(f).Error; err != nil {
		return err
	}

	// get user last fiat log and calc balance
	lastBalance, err := GetLastBlanceByFiatlog(tx, f.UserId)
	if err != nil {
		return err
	}
	logrus.Debug("debug fiatlog cache create: get last balance by fiat log")

	fl := &FiatLog{
		FiatLogCore: f.FiatLogCore,
		CacheIds:    datatypes.JSONSlice[uint]{f.ID},
	}
	fl.Balance = lastBalance.Add(f.Amount)
	// switch fl.Type {
	// case FIAT_LOG_TYPE_REFUND_SPONSOR:
	// 	// find related fiat log and set meta
	// 	var meta FiatMetaRefundSponsor
	// 	if err := json.Unmarshal(fl.Meta, &meta); err != nil {
	// 		return err
	// 	}

	// 	FindSponsorFiatlogCacheByTxid(meta.TxId)

	// }

	err = tx.Save(fl).Error
	logrus.Debug("debug fiatlog cache create: save fiat log and update users.fiat_log_balance")
	return err
}

func FindLastFiatLogCache(tx *gorm.DB, userId uint, logType FiatLogType) (*FiatLogCache, error) {
	var fl FiatLogCache
	if err := tx.Model(&FiatLogCache{}).
		Where("user_id=? and type =?", userId, logType).
		Order("id desc").
		First(&fl).Error; err != nil {
		return nil, err
	}
	return &fl, nil
}

func MergeToFiatlog(start, end time.Time) error {
	lockMergeFiatLogMutex()
	defer unlockMergeFiatLogMutex()

	err := GetDB().Transaction(func(tx *gorm.DB) error {
		if err := mergeApiQuotaFiatlogs(tx, start, end); err != nil {
			return err
		}

		if err := mergePayApiFeeFiatlogs(tx, start, end); err != nil {
			return err
		}

		if err := mergeRefundApiFeeFiatlogs(tx, start, end); err != nil {
			return err
		}
		return nil
	})
	logrus.WithField("error", fmt.Sprintf("%+v", err)).WithField("start", start).WithField("end", end).Info("merged fiatlog")
	return err
}

func mergeApiQuotaFiatlogs(tx *gorm.DB, start, end time.Time) error {
	if err := mergeApiFiatlogs(tx, FIAT_LOG_TYPE_PAY_API_QUOTA, start, end); err != nil {
		return err
	}
	if err := mergeApiFiatlogs(tx, FIAT_LOG_TYPE_REFUND_API_QUOTA, start, end); err != nil {
		return err
	}
	if err := mergeApiFiatlogs(tx, FIAT_LOG_TYPE_RESET_API_QUOTA, start, end); err != nil {
		return err
	}
	return nil
}

func mergePayApiFeeFiatlogs(tx *gorm.DB, start, end time.Time) error {
	return mergeApiFiatlogs(tx, FIAT_LOG_TYPE_PAY_API_FEE, start, end)
}

type ApiInfoAggregated struct {
	Count    int
	Amount   decimal.Decimal
	CacheIds datatypes.JSONSlice[uint]
}

type FiatLogWithCount struct {
	FiatLog
	Count int
}

func mergeRefundApiFeeFiatlogs(tx *gorm.DB, start, end time.Time) error {

	var refundFiatlogCaches []*FiatLogCache

	err := tx.Model(&FiatLogCache{}).
		Where("created_at>=? and created_at<?", start, end).
		Where("type=?", FIAT_LOG_TYPE_REFUND_API_FEE).
		Where("is_merged=?", false).
		Find(&refundFiatlogCaches).Error
	if err != nil {
		return err
	}
	fmt.Println(refundFiatlogCaches)

	userRefundFeeFlcs, err := groupFlcByUserAndCosttype(refundFiatlogCaches, FIAT_LOG_TYPE_REFUND_API_FEE)
	if err != nil {
		return err
	}

	// 根据refund api fee的数量往回找 pay api fee fiatlog并设置refund logids，如果fiatlog数量不足refund log则拆分refund log
	for userId, refundFees := range userRefundFeeFlcs {
		lastBalance, err := GetLastBlanceByFiatlog(tx, userId)
		if err != nil {
			return err
		}

		for costtype, apiAggre := range refundFees {

			payFls, err := getPayApiFeeFlsForMapRefund(tx, userId, costtype, apiAggre.Count)
			if err != nil {
				return err
			}

			if len(payFls) == 0 {
				continue
			}

			// 循环匹配refund fees到 pay fiatlogs，不足的分割refund fees并标注为part
			remain := apiAggre.Count
			for _, payFl := range payFls {
				if remain == 0 {
					break
				}

				payFlMeta := getPayApiFeeMeta(payFl.FiatLog)
				aliveCountOfPayFl := payFl.Count - payFlMeta.RefundedCount
				if aliveCountOfPayFl == 0 {
					continue
				}

				count := mathutils.Min(remain, aliveCountOfPayFl)
				remain -= count

				lastBalance = lastBalance.Add(apiAggre.Amount)
				refundMeta, _ := json.Marshal(FiatMetaRefundApiFee{
					Quota: Quota{
						CostType: costtype,
						Count:    count,
					},
					IsPart: count < apiAggre.Count,
				})

				refundFl := FiatLog{
					FiatLogCore: FiatLogCore{
						UserId:  userId,
						Amount:  apiAggre.Amount.Div(decimal.NewFromInt(int64(apiAggre.Count))).Mul(decimal.NewFromInt(int64(count))),
						Type:    FIAT_LOG_TYPE_REFUND_API_FEE,
						Balance: lastBalance,
						OrderNO: RandomOrderNO(),
						Meta:    refundMeta,
					},
					CacheIds: apiAggre.CacheIds,
				}
				err = tx.Save(&refundFl).Error
				if err != nil {
					return err
				}

				payFl.FiatLog.RefundLogIds = append(payFl.FiatLog.RefundLogIds, refundFl.ID)
				// recored refunded count and amount for invoice
				payFlMeta.RefundedCount += count
				payFlMeta.RefundedAmount = payFlMeta.RefundedAmount.Add(refundFl.Amount)
				payFl.Meta = utils.Must(json.Marshal(payFlMeta))
			}

			_payFls := lo.Map(payFls, func(payFl *FiatLogWithCount, index int) FiatLog {
				return payFl.FiatLog
			})
			if err = tx.Save(&_payFls).Error; err != nil {
				return err
			}

			if err := tx.Model(&FiatLogCache{}).Where("id in ?", []uint(apiAggre.CacheIds)).Update("is_merged", true).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// count represents the queried pay_api_fee_fiatlogs should contains the number of costtype which unrefuned
func getPayApiFeeFlsForMapRefund(tx *gorm.DB, userId uint, costtype enums.CostType, count int) ([]*FiatLogWithCount, error) {
	// find last 100 pay api fee
	var payFls []*FiatLogWithCount

	offset := 0
	limit := 100
	for {
		var fragment []*FiatLogWithCount
		if err := tx.Debug().Model(&FiatLog{}).
			Where("user_id=?", userId).
			Where("type=?", FIAT_LOG_TYPE_PAY_API_FEE).
			Where("meta->'$.cost_type'=?", costtype.String()).
			Order("id desc").Offset(offset).Limit(limit).
			Select("*,meta->'$.count' as count").
			Scan(&fragment).Error; err != nil {
			return nil, err
		}

		// means queried all but sum small than refund count, it's impossible but we still handle it
		if len(fragment) == 0 {
			return nil, nil
		}
		payFls = append(payFls, fragment...)

		sum := 0
		for _, fl := range payFls {
			meta := getPayApiFeeMeta(fl.FiatLog)
			sum += fl.Count - meta.RefundedCount
		}

		if sum >= count {
			break
		}
		offset += limit
	}

	return payFls, nil
}

// Note: not support refund api fee fiat log caches
func mergeApiFiatlogs(tx *gorm.DB, fiatLogType FiatLogType, start, end time.Time) error {
	if !lo.Contains([]FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_PAY_API_QUOTA, FIAT_LOG_TYPE_RESET_API_QUOTA, FIAT_LOG_TYPE_REFUND_API_QUOTA}, fiatLogType) {
		return fmt.Errorf("not support %v", fiatLogType)
	}

	var fiatlogCaches []*FiatLogCache
	err := tx.Model(&FiatLogCache{}).
		Where("created_at>=? and created_at<?", start, end).
		Where("type=?", fiatLogType).
		Where("is_merged=?", false).
		Find(&fiatlogCaches).Error
	if err != nil {
		return err
	}

	// fmt.Println(fiatlogCaches)

	if len(fiatlogCaches) == 0 {
		return nil
	}

	userApiAggregates, err := groupFlcByUserAndCosttype(fiatlogCaches, fiatLogType)
	if err != nil {
		return err
	}
	// insert db
	userApiFiatlogs, err := convertGroupedFlcToFiatlogs(tx, userApiAggregates, fiatLogType)
	if err != nil {
		return err
	}

	if err = tx.Save(&userApiFiatlogs).Error; err != nil {
		return err
	}

	for _, v := range fiatlogCaches {
		v.IsMerged = true
	}

	if err = tx.Save(&fiatlogCaches).Error; err != nil {
		return err
	}

	return nil
}

func groupFlcByUserAndCosttype(source []*FiatLogCache, fiatLogType FiatLogType) (map[uint]map[enums.CostType](*ApiInfoAggregated), error) {
	userPayFees := make(map[uint]map[enums.CostType](*ApiInfoAggregated))
	for _, item := range source {
		if _, ok := userPayFees[item.UserId]; !ok {
			userPayFees[item.UserId] = make(map[enums.CostType]*ApiInfoAggregated)
		}

		var meta FiatMetaPayApiFeeForCache
		err := json.Unmarshal(item.Meta, &meta)
		if err != nil {
			return nil, err
		}

		if _, ok := userPayFees[item.UserId][meta.CostType]; !ok {
			userPayFees[item.UserId][meta.CostType] = &ApiInfoAggregated{}
		}

		userPayFees[item.UserId][meta.CostType].Count += meta.Count
		userPayFees[item.UserId][meta.CostType].Amount = userPayFees[item.UserId][meta.CostType].Amount.Add(item.Amount)
		userPayFees[item.UserId][meta.CostType].CacheIds = append(userPayFees[item.UserId][meta.CostType].CacheIds, item.ID)
	}
	return userPayFees, nil
}

func convertGroupedFlcToFiatlogs(tx *gorm.DB, groupedUserApiCosts map[uint]map[enums.CostType](*ApiInfoAggregated), fiatLogType FiatLogType) ([]*FiatLog, error) {
	var userPayFiatlogs []*FiatLog
	for userId, payFees := range groupedUserApiCosts {
		lastBalance, err := GetLastBlanceByFiatlog(tx, userId)
		if err != nil {
			return nil, err
		}

		for costType, apiAggre := range payFees {
			lastBalance = lastBalance.Add(apiAggre.Amount)
			meta, _ := json.Marshal(FiatMetaPayApiFeeForCache{
				CostType: costType,
				Count:    apiAggre.Count,
			})
			userPayFiatlogs = append(userPayFiatlogs, &FiatLog{
				FiatLogCore: FiatLogCore{
					UserId:  userId,
					Amount:  apiAggre.Amount,
					Type:    fiatLogType,
					Balance: lastBalance,
					OrderNO: RandomOrderNO(),
					Meta:    meta,
				},
				CacheIds: apiAggre.CacheIds,
			})
		}
	}
	return userPayFiatlogs, nil
}

func unmarshalType[T any](jsonStr string) (*T, error) {
	var fm *T
	if err := json.Unmarshal([]byte(jsonStr), &fm); err != nil {
		return nil, err
	}
	return fm, nil
}

func getPayApiFeeMeta(payApiFeefiatlog FiatLog) *FiatMetaPayApiFee {
	var payFlMeta FiatMetaPayApiFee
	json.Unmarshal(payApiFeefiatlog.Meta, &payFlMeta)
	return &payFlMeta
}
