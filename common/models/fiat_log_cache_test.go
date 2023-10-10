package models

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/influxdata/influxdb/pkg/testing/assert"
	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/mathutils"
	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	ConnectDB(config.Mysql{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "new-password",
		Db:       "rainbowtest",
	})
}

func TestGroupPayAndRefundApiFeeFiatlog(t *testing.T) {
	err := GetDB().Transaction(func(tx *gorm.DB) error {

		if err := MergePayApiFeeFiatlogs(tx); err != nil {
			return err
		}

		if err := MergeRefundApiFeeFiatlogs(tx); err != nil {
			return err
		}

		return nil

	})
	assert.NoError(t, err)
}

func getPayApiFeeMeta(payApiFeefiatlog FiatLog) *FiatMetaPayApiFee {
	var payFlMeta FiatMetaPayApiFee
	json.Unmarshal(payApiFeefiatlog.Meta, &payFlMeta)
	return &payFlMeta
}

type ApiFeeAggregate struct {
	Count    int
	Amount   decimal.Decimal
	CacheIds datatypes.JSONSlice[uint]
}

type FiatLogWithCount struct {
	FiatLog
	Count int
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

func MergeRefundApiFeeFiatlogs(tx *gorm.DB) error {

	var refundFiatlogCaches []*FiatLogCache

	err := tx.Debug().Model(&FiatLogCache{}).
		Where("type=?", FIAT_LOG_TYPE_REFUND_API_FEE).
		Where("is_merged=?", false).
		Find(&refundFiatlogCaches).Error
	if err != nil {
		return err
	}
	fmt.Println(refundFiatlogCaches)

	userRefundFeeFlcs, err := GroupFlcByUserAndCosttype(refundFiatlogCaches, FIAT_LOG_TYPE_REFUND_API_FEE)
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

func MergePayApiFeeFiatlogs(tx *gorm.DB) error {
	var payFiatlogCaches []*FiatLogCache
	err := tx.Debug().Model(&FiatLogCache{}).
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

	userPayFees, err := GroupFlcByUserAndCosttype(payFiatlogCaches, FIAT_LOG_TYPE_PAY_API_FEE)
	if err != nil {
		return err
	}
	// insert db
	userPayFiatlogs, err := ConvertGroupedFlcToFiatlogs(tx, userPayFees, FIAT_LOG_TYPE_PAY_API_FEE)
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

func GroupFlcByUserAndCosttype(source []*FiatLogCache, fiatLogType FiatLogType) (map[uint]map[enums.CostType](*ApiFeeAggregate), error) {
	userPayFees := make(map[uint]map[enums.CostType](*ApiFeeAggregate))
	for _, item := range source {
		if _, ok := userPayFees[item.UserId]; !ok {
			userPayFees[item.UserId] = make(map[enums.CostType]*ApiFeeAggregate)
		}

		var meta FiatMetaPayApiFeeForCache
		err := json.Unmarshal(item.Meta, &meta)
		if err != nil {
			return nil, err
		}

		if _, ok := userPayFees[item.UserId][meta.CostType]; !ok {
			userPayFees[item.UserId][meta.CostType] = &ApiFeeAggregate{}
		}

		userPayFees[item.UserId][meta.CostType].Count += meta.Count
		userPayFees[item.UserId][meta.CostType].Amount = userPayFees[item.UserId][meta.CostType].Amount.Add(item.Amount)
		userPayFees[item.UserId][meta.CostType].CacheIds = append(userPayFees[item.UserId][meta.CostType].CacheIds, item.ID)
	}
	return userPayFees, nil
}

func ConvertGroupedFlcToFiatlogs(tx *gorm.DB, groupedUserApiCosts map[uint]map[enums.CostType](*ApiFeeAggregate), fiatLogType FiatLogType) ([]FiatLog, error) {
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
