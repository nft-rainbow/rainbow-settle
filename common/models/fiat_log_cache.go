package models

import (
	"encoding/json"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FiatLogCache struct {
	BaseModel
	UserId  uint            `gorm:"type:int;index" json:"user_id"`
	Amount  decimal.Decimal `gorm:"type:decimal(20,2)" json:"amount"`         // 单位分
	Type    FiatLogType     `gorm:"type:int;default:0" json:"type"`           // 1-deposit
	Meta    datatypes.JSON  `gorm:"type:json" json:"meta"`                    // metadata
	OrderNO string          `gorm:"type:varchar(255);unique" json:"order_no"` // order NO in rainbow platform
	Balance decimal.Decimal `gorm:"type:decimal(20,2)" json:"balance"`        // apply log balance
}

func MergeToFiatlog(start, end time.Time) error {

	type TmpFiatLog struct {
		FiatLog
		Meta     string `json:"meta"`
		CacheIds string `json:"cache_ids"`
	}

	return GetDB().Transaction(func(tx *gorm.DB) error {
		var apiFeeTmpFls []*TmpFiatLog
		err := tx.Debug().Model(&FiatLogCache{}).Group("user_id,type").
			Where("created_at>=? and created_at<?", start, end).
			Where("type in ?", []FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_REFUND_API_FEE}).
			Select("user_id, sum(amount) as amount, type, GROUP_CONCAT(meta) as meta, GROUP_CONCAT(id) as cache_ids").
			Scan(&apiFeeTmpFls).Error
		if err != nil {
			return err
		}

		var apiFeeFls []*FiatLog
		for _, tmpFl := range apiFeeTmpFls {
			fl := tmpFl.FiatLog

			metas, _ := summaryMetas(fl.Type, tmpFl.Meta)
			meta, _ := json.Marshal(metas)
			fl.Meta = meta

			ids, _ := unmarshalType[[]uint]("[" + tmpFl.CacheIds + "]")
			fl.CacheIds = datatypes.JSONSlice[uint](*ids)

			fl.OrderNO = RandomOrderNO()
			apiFeeFls = append(apiFeeFls, &fl)
		}

		var otherFls []*FiatLog
		err = tx.Debug().Model(&FiatLogCache{}).
			Where("created_at>=? and created_at<?", start, end).
			Where("type not in ?", []FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_REFUND_API_FEE}).
			Find(&otherFls).Error
		if err != nil {
			return err
		}

		allFls := append(otherFls, apiFeeFls...)
		for _, fl := range allFls {
			var lastFiatLog FiatLog
			if err := GetDB().Debug().Model(&FiatLog{}).Where("user_id=?", fl.UserId).Order("id desc").First(&lastFiatLog).Error; err != nil {
				if !gormutils.IsRecordNotFoundError(err) {
					return err
				}
				lastFiatLog.Balance = decimal.Zero
			}

			fl.Balance = lastFiatLog.Balance.Add(fl.Amount)
			if err := tx.Save(&fl).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func summaryMetas(fiatLogType FiatLogType, metasJson string) (interface{}, error) {
	switch fiatLogType {
	case FIAT_LOG_TYPE_REFUND_API_FEE:
		fallthrough
	case FIAT_LOG_TYPE_PAY_API_FEE:
		fms, err := unmarshalType[[]*FiatMetaPayApiFee](metasJson)
		if err != nil {
			return nil, err
		}
		fmsByAddr := lo.GroupBy(*fms, func(fm *FiatMetaPayApiFee) enums.CostType {
			return fm.CostType
		})
		_fms := lo.MapToSlice(fmsByAddr, func(costType enums.CostType, fms []*FiatMetaPayApiFee) *FiatMetaPayApiFee {
			return lo.Reduce(fms, func(aggr *FiatMetaPayApiFee, item *FiatMetaPayApiFee, index int) *FiatMetaPayApiFee {
				aggr.Count = aggr.Count + item.Count
				return aggr
			}, &FiatMetaPayApiFee{})
		})
		return _fms, nil
	default:
		return nil, errors.New("not supported")
	}
}

func unmarshalType[T any](jsonStr string) (*T, error) {
	var fm *T
	if err := json.Unmarshal([]byte(jsonStr), &fm); err != nil {
		return nil, err
	}
	return fm, nil
}
