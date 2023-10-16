package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/mathutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/pkg/errors"
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
	mergeFiatLogLock.Lock()
	logrus.Debug("lock merge fiat log mutex")
}

func unlockMergeFiatLogMutex() {
	mergeFiatLogLock.Unlock()
	logrus.Debug("unlock merge fiat log mutex")
}

type FiatLogCache struct {
	BaseModel
	FiatLogCore
	UnsettleAmount decimal.Decimal `gorm:"type:decimal(20,10)" json:"unsettle_amount"`
	IsMerged       bool            `gorm:"default:0" json:"isMerged,omitempty"`
}

func (f *FiatLogCache) AfterCreate(tx *gorm.DB) (err error) {
	logrus.Debug("debug fiatlog cache create: start hook after create fiat_log_cache")
	lockMergeFiatLogMutex()
	defer unlockMergeFiatLogMutex()

	logrus.Debug("debug fiatlog cache create: locked")
	if f.IsMerged || lo.Contains(apiRelatedFiatLogTypes, f.Type) {
		return nil
	}
	f.IsMerged = true

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
	err = tx.Save(fl).Error
	logrus.Debug("debug fiatlog cache create: save fiat log and update users.fiat_log_balance")
	return err
}

func FindSponsorFiatlogByTxid(txId uint) (*FiatLogCache, error) {
	var fl FiatLogCache
	if err := db.Model(&FiatLogCache{}).Where("meta->'$.tx_id'=?", txId).
		Where("type =? or type=?", FIAT_LOG_TYPE_BUY_GAS, FIAT_LOG_TYPE_BUY_STORAGE).
		First(&fl).Error; err != nil {
		return nil, err
	}
	return &fl, nil
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
	type TmpFiatLog struct {
		FiatLog
		Meta     string `json:"meta"`
		CacheIds string `json:"cache_ids"`
	}

	var apiQuotaTmpFls []*TmpFiatLog
	err := tx.Debug().Model(&FiatLogCache{}).Group("user_id,type").
		Where("created_at>=? and created_at<?", start, end).
		Where("is_merged=0").
		Where("type in ?", apiQuotaRelatedFiatLogTypes).
		Select("user_id, sum(amount) as amount, type, GROUP_CONCAT(meta) as meta, GROUP_CONCAT(id) as cache_ids").
		Scan(&apiQuotaTmpFls).Error
	if err != nil {
		return errors.WithStack(err)
	}

	logrus.WithField("fls", apiQuotaTmpFls).Trace("get temp api fee fiat logs")

	var apiFeeFls []*FiatLog
	for _, tmpFl := range apiQuotaTmpFls {
		fl := tmpFl.FiatLog

		logrus.WithField("metas", fmt.Sprintf("[%s]", tmpFl.Meta)).Debug("rat metas string")

		// TODO: 这里由于meta是截断后的结果，暂时不存meta，需要处理
		metas, err := summaryMetas(fl.Type, fmt.Sprintf("[%s]", "")) //tmpFl.Meta))
		if err != nil {
			return errors.WithStack(err)
		}
		meta, _ := json.Marshal(metas)
		fl.Meta = meta

		ids, err := unmarshalType[[]uint](fmt.Sprintf("[%s]", tmpFl.CacheIds))
		if err != nil {
			return errors.WithStack(err)
		}
		fl.CacheIds = datatypes.JSONSlice[uint](*ids)
		apiFeeFls = append(apiFeeFls, &fl)
	}

	if len(apiFeeFls) == 0 {
		return nil
	}

	// note: avoid there is unmerged fiat log not in apiRelatedFiatLogTypes
	// var otherFls []*FiatLog
	// err = tx.Model(&FiatLogCache{}).
	// 	Where("created_at>=? and created_at<?", start, end).
	// 	Where("is_merged=0").
	// 	Where("type not in ?", apiRelatedFiatLogTypes).
	// 	Select("user_id, amount, type, meta, CONCAT('[' , id, ']') as cache_ids").
	// 	Scan(&otherFls).Error
	// if err != nil {
	// 	return errors.WithStack(err)
	// }

	// allFls := append(otherFls, apiFeeFls...)
	// if len(allFls) == 0 {
	// 	return nil
	// }

	// allFls := apiFeeFls
	lastBalances := make(map[uint]decimal.Decimal)
	for _, fl := range apiFeeFls {
		if _, ok := lastBalances[fl.UserId]; !ok {
			lastBalance, err := GetLastBlanceByFiatlog(tx, fl.UserId)
			if err != nil {
				return errors.WithStack(err)
			}
			lastBalances[fl.UserId] = lastBalance
		}
		fl.OrderNO = RandomOrderNO()
		fl.Balance = lastBalances[fl.UserId].Add(fl.Amount)
		lastBalances[fl.UserId] = fl.Balance

		logrus.WithField("user", fl.UserId).WithField("val", lastBalances[fl.UserId]).WithField("amount", fl.Amount).Debug("updated last balance")
	}
	logrus.WithField("all fls", apiFeeFls).Debug("save fiat logs")
	if err := tx.Debug().Save(&apiFeeFls).Error; err != nil {
		return errors.WithStack(err)
	}

	// update is_merged flag
	err = tx.Model(&FiatLogCache{}).
		Where("created_at>=? and created_at<?", start, end).
		Where("type in ?", apiQuotaRelatedFiatLogTypes).
		Where("is_merged=0").Update("is_merged", true).Error
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func summaryMetas(fiatLogType FiatLogType, metasJson string) (interface{}, error) {
	switch fiatLogType {
	case FIAT_LOG_TYPE_PAY_API_QUOTA:
		fallthrough
	case FIAT_LOG_TYPE_REFUND_API_QUOTA:
		fallthrough
	case FIAT_LOG_TYPE_REFUND_API_FEE:
		fallthrough
	case FIAT_LOG_TYPE_RESET_API_QUOTA:
		fallthrough
	case FIAT_LOG_TYPE_PAY_API_FEE:
		fms, err := unmarshalType[[]*FiatMetaPayApiFeeForCache](metasJson)
		if err != nil {
			logrus.WithField("input", metasJson).Debug("failed unmarshal metas to []*FiatMetaPayApiFee")
			return nil, err
		}
		fmsByAddr := lo.GroupBy(*fms, func(fm *FiatMetaPayApiFeeForCache) enums.CostType {
			return fm.CostType
		})
		_fms := lo.MapToSlice(fmsByAddr, func(costType enums.CostType, fms []*FiatMetaPayApiFeeForCache) *FiatMetaPayApiFeeForCache {
			return lo.Reduce(fms, func(aggr *FiatMetaPayApiFeeForCache, item *FiatMetaPayApiFeeForCache, index int) *FiatMetaPayApiFeeForCache {
				aggr.Count = aggr.Count + item.Count
				return aggr
			}, &FiatMetaPayApiFeeForCache{CostType: fms[0].CostType})
		})
		return _fms, nil
	default:
		return nil, errors.Errorf("not supported %v", fiatLogType)
	}
}

type ApiFeeAggregated struct {
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

	err := tx.Debug().Model(&FiatLogCache{}).
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

func mergePayApiFeeFiatlogs(tx *gorm.DB, start, end time.Time) error {
	var payFiatlogCaches []*FiatLogCache
	err := tx.Debug().Model(&FiatLogCache{}).
		Where("created_at>=? and created_at<?", start, end).
		Where("type=?", FIAT_LOG_TYPE_PAY_API_FEE).
		Where("is_merged=?", false).
		Find(&payFiatlogCaches).Error
	if err != nil {
		return err
	}

	fmt.Println(payFiatlogCaches)

	if len(payFiatlogCaches) == 0 {
		return nil
	}

	userPayFees, err := groupFlcByUserAndCosttype(payFiatlogCaches, FIAT_LOG_TYPE_PAY_API_FEE)
	if err != nil {
		return err
	}
	// insert db
	userPayFiatlogs, err := convertGroupedFlcToFiatlogs(tx, userPayFees, FIAT_LOG_TYPE_PAY_API_FEE)
	if err != nil {
		return err
	}

	if err = tx.Save(&userPayFiatlogs).Error; err != nil {
		return err
	}

	for _, v := range payFiatlogCaches {
		v.IsMerged = true
	}

	if err = tx.Save(&payFiatlogCaches).Error; err != nil {
		return err
	}

	return nil
}

func groupFlcByUserAndCosttype(source []*FiatLogCache, fiatLogType FiatLogType) (map[uint]map[enums.CostType](*ApiFeeAggregated), error) {
	userPayFees := make(map[uint]map[enums.CostType](*ApiFeeAggregated))
	for _, item := range source {
		if _, ok := userPayFees[item.UserId]; !ok {
			userPayFees[item.UserId] = make(map[enums.CostType]*ApiFeeAggregated)
		}

		var meta FiatMetaPayApiFeeForCache
		err := json.Unmarshal(item.Meta, &meta)
		if err != nil {
			return nil, err
		}

		if _, ok := userPayFees[item.UserId][meta.CostType]; !ok {
			userPayFees[item.UserId][meta.CostType] = &ApiFeeAggregated{}
		}

		userPayFees[item.UserId][meta.CostType].Count += meta.Count
		userPayFees[item.UserId][meta.CostType].Amount = userPayFees[item.UserId][meta.CostType].Amount.Add(item.Amount)
		userPayFees[item.UserId][meta.CostType].CacheIds = append(userPayFees[item.UserId][meta.CostType].CacheIds, item.ID)
	}
	return userPayFees, nil
}

func convertGroupedFlcToFiatlogs(tx *gorm.DB, groupedUserApiCosts map[uint]map[enums.CostType](*ApiFeeAggregated), fiatLogType FiatLogType) ([]FiatLog, error) {
	var userPayFiatlogs []FiatLog
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
			userPayFiatlogs = append(userPayFiatlogs, FiatLog{
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
