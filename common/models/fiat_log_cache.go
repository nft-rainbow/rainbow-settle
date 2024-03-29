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
	DeletedAt      gorm.DeletedAt  `gorm:"index;index:idx_type_merge_deleted_user" json:"deleted_at"`
	UnsettleAmount decimal.Decimal `gorm:"type:decimal(20,10)" json:"unsettle_amount"`
	IsMerged       bool            `gorm:"index:idx_type_merge_deleted_user;default:0" json:"isMerged,omitempty"`
}

func (f *FiatLogCache) AfterCreate(tx *gorm.DB) (err error) {
	if f.IsMerged || lo.Contains(apiRelatedFiatLogTypes, f.Type) {
		return nil
	}
	f.IsMerged = true
	if err := tx.Save(f).Error; err != nil {
		return err
	}
	if err := f.copyToFiatLog(tx); err != nil {
		return err
	}

	logrus.Debug("debug fiatlog cache create: save fiat log and update users.fiat_log_balance")
	return nil
}

func (f *FiatLogCache) copyToFiatLog(tx *gorm.DB) error {
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
	if err := tx.Save(fl).Error; err != nil {
		return err
	}

	if err := RelateBuySponsorFiatlog(tx, fl); err != nil {
		return err
	}
	return nil
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
	logrus.WithField("start", start).WithField("end", end).Info("merge to fiat log")
	lockMergeFiatLogMutex()
	defer unlockMergeFiatLogMutex()

	err := func() error {
		if err := mergeApiQuotaFiatlogs(start, end); err != nil {
			return err
		}

		if err := mergePayApiFeeFiatlogs(start, end); err != nil {
			return err
		}

		if err := mergeRefundApiFeeFiatlogs(start, end); err != nil {
			return err
		}
		return nil
	}()

	logrus.WithField("error", fmt.Sprintf("%+v", err)).WithField("start", start).WithField("end", end).Info("merged fiatlog")
	return err
}

func mergeApiQuotaFiatlogs(start, end time.Time) error {

	if err := mergeApiFiatlogs(FIAT_LOG_TYPE_PAY_API_QUOTA, start, end); err != nil {
		return err
	}

	if err := mergeApiFiatlogs(FIAT_LOG_TYPE_REFUND_API_QUOTA, start, end); err != nil {
		return err
	}

	if err := mergeApiFiatlogs(FIAT_LOG_TYPE_RESET_API_QUOTA, start, end); err != nil {
		return err
	}
	return nil
}

func mergePayApiFeeFiatlogs(start, end time.Time) error {
	return mergeApiFiatlogs(FIAT_LOG_TYPE_PAY_API_FEE, start, end)
}

type ApiInfoAggregated struct {
	// Quota count
	CountReset    int
	CountRollover int
	// Balance count
	Count    int
	Amount   decimal.Decimal
	CacheIds datatypes.JSONSlice[uint]
}

type FiatLogWithCount struct {
	FiatLog
	Count int
}

func mergeRefundApiFeeFiatlogs(start, end time.Time) error {
	return GetDB().Transaction(func(tx *gorm.DB) error {
		var refundFiatlogCaches []*FiatLogCache

		err := tx.Model(&FiatLogCache{}).
			Where("created_at>=? and created_at<?", start, end).
			Where("type=?", FIAT_LOG_TYPE_REFUND_API_FEE).
			Where("is_merged=?", false).
			Find(&refundFiatlogCaches).Error
		if err != nil {
			return errors.WithMessage(err, "failed to query FIAT_LOG_TYPE_REFUND_API_FEE fiat log caches")
		}
		fmt.Println(refundFiatlogCaches)

		userRefundFeeFlcs, err := groupFlcByUserAndCosttype(refundFiatlogCaches)
		if err != nil {
			return errors.WithMessage(err, "failed to group fiat_log_caches")
		}

		// 根据refund api fee的数量往回找 pay api fee fiatlog并设置refund logids，如果fiatlog数量不足refund log则拆分refund log
		for userId, refundFees := range userRefundFeeFlcs {
			lastBalance, err := GetLastBlanceByFiatlog(tx, userId)
			if err != nil {
				return errors.WithMessage(err, "failed to get last balance by fiat log")
			}

			for costtype, apiAggre := range refundFees {

				payFls, err := getPayApiFeeFlsForMapRefund(tx, userId, costtype, apiAggre.Count)
				if err != nil {
					return errors.WithMessage(err, "failed to get pay api fee fiat logs")
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

					amount := apiAggre.Amount.Div(decimal.NewFromInt(int64(apiAggre.Count))).Mul(decimal.NewFromInt(int64(count)))
					lastBalance = lastBalance.Add(amount)

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
							Amount:  amount,
							Type:    FIAT_LOG_TYPE_REFUND_API_FEE,
							Balance: lastBalance,
							OrderNO: RandomOrderNO(),
							Meta:    refundMeta,
						},
						CacheIds: apiAggre.CacheIds,
					}
					err = tx.Save(&refundFl).Error
					if err != nil {
						return errors.WithMessage(err, "failed to save refund fiat logs")
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
					return errors.WithMessage(err, "failed to save pay fait logs")
				}

				if err := tx.Model(&FiatLogCache{}).Where("id in ?", []uint(apiAggre.CacheIds)).Update("is_merged", true).Error; err != nil {
					return errors.WithMessage(err, "failed to update fiat log caches")
				}
			}
		}
		return nil
	})
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
func mergeApiFiatlogs(fiatLogType FiatLogType, start, end time.Time) error {
	// split by user and length
	if !lo.Contains([]FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_PAY_API_QUOTA, FIAT_LOG_TYPE_RESET_API_QUOTA, FIAT_LOG_TYPE_REFUND_API_QUOTA}, fiatLogType) {
		return fmt.Errorf("not support %v", fiatLogType)
	}

	// get user_ids
	var user_ids []uint
	if err := GetDB().Model(&FiatLogCache{}).
		Select("user_id").Group("user_id").
		Where("created_at>=? and created_at<?", start, end).
		Where("type=?", fiatLogType).
		Where("is_merged=?", false).
		Find(&user_ids).Error; err != nil {
		return errors.WithMessage(err, "failed to find need merge users")
	}
	logrus.WithField("users", user_ids).WithField("type", fiatLogType).WithField("start", start).WithField("end", end).Info("find need merge fiat logs")

	for _, user_id := range user_ids {
		logrus.WithField("user_id", user_id).Info("merge user fiat log")
		isAllMeged := false
		for {
			if isAllMeged {
				break
			}

			if err := GetDB().Transaction(func(tx *gorm.DB) error {
				var fiatlogCaches []*FiatLogCache
				err := tx.Model(&FiatLogCache{}).
					Where("created_at>=? and created_at<?", start, end).
					Where("user_id=?", user_id).
					Where("type=?", fiatLogType).
					Where("is_merged=?", false).
					Limit(1000).
					Find(&fiatlogCaches).Error
				if err != nil {
					return errors.WithMessage(err, "failed to query unmerged fiat_log_caches")
				}
				fmt.Println("aaa")

				if len(fiatlogCaches) == 0 {
					isAllMeged = true
					return nil
				}
				// fmt.Println("bbb")

				userApiAggregates, err := groupFlcByUserAndCosttype(fiatlogCaches)
				if err != nil {
					return errors.WithMessage(err, "failed to group fiat_log_caches")
				}
				// fmt.Println("ccc")
				// insert db
				userApiFiatlogs, err := convertGroupedFlcToFiatlogs(tx, userApiAggregates, fiatLogType)
				if err != nil {
					return errors.WithMessage(err, "failed to convert to fiatlogs")
				}
				// fmt.Println("ddd")

				j, _ := json.Marshal(userApiFiatlogs)
				fmt.Printf("convertGroupedFlcToFiatlogs %s\n", j)

				if err = tx.Save(&userApiFiatlogs).Error; err != nil {
					return errors.WithMessage(err, "failed to save fiat logs")
				}
				// fmt.Println("eee")

				for _, v := range fiatlogCaches {
					v.IsMerged = true
				}

				if err = tx.Save(&fiatlogCaches).Error; err != nil {
					return errors.WithMessage(err, "failed to save fait log caches")
				}
				// fmt.Println("fff")
				logrus.WithField("user_id", user_id).WithField("length", len(fiatlogCaches)).Info("success merge to fiat log")
				return nil
			}); err != nil {
				logrus.WithField("user_id", user_id).WithError(err).Error("merge to fiat log")
				return err
			}
		}
	}
	return nil
}

func groupFlcByUserAndCosttype(source []*FiatLogCache) (map[uint]map[enums.CostType](*ApiInfoAggregated), error) {
	userPayFees := make(map[uint]map[enums.CostType](*ApiInfoAggregated))
	for _, item := range source {
		if _, ok := userPayFees[item.UserId]; !ok {
			userPayFees[item.UserId] = make(map[enums.CostType]*ApiInfoAggregated)
		}

		if lo.Contains(apiFeeRelatedFiatLogTypes, item.Type) {
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
		} else {
			var meta FiatMetaPayApiQuota
			err := json.Unmarshal(item.Meta, &meta)
			if err != nil {
				return nil, err
			}

			if _, ok := userPayFees[item.UserId][meta.CostType]; !ok {
				userPayFees[item.UserId][meta.CostType] = &ApiInfoAggregated{}
			}

			userPayFees[item.UserId][meta.CostType].CountReset += meta.CountReset
			userPayFees[item.UserId][meta.CostType].CountRollover += meta.CountRollover
			userPayFees[item.UserId][meta.CostType].Amount = userPayFees[item.UserId][meta.CostType].Amount.Add(item.Amount)
			userPayFees[item.UserId][meta.CostType].CacheIds = append(userPayFees[item.UserId][meta.CostType].CacheIds, item.ID)
		}
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

			var meta []byte
			if lo.Contains(apiFeeRelatedFiatLogTypes, fiatLogType) {
				meta, _ = json.Marshal(FiatMetaPayApiFeeForCache{
					CostType: costType,
					Count:    apiAggre.Count,
				})
			} else {
				meta, _ = json.Marshal(FiatMetaPayApiQuota{
					CostType:      costType,
					CountReset:    apiAggre.CountReset,
					CountRollover: apiAggre.CountRollover,
				})
			}

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
