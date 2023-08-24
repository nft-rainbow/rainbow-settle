package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FiatLogCache struct {
	BaseModel
	FiatLogCore
	IsMerged bool `gorm:"default:0" json:"isMerged,omitempty"`
}

func MergeToFiatlog(start, end time.Time) error {

	type TmpFiatLog struct {
		FiatLog
		Meta     string `json:"meta"`
		CacheIds string `json:"cache_ids"`
	}

	err := GetDB().Transaction(func(tx *gorm.DB) error {
		// tx = tx.Debug()
		needGroupFiatLogTypes := []FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_REFUND_API_FEE, FIAT_LOG_TYPE_PAY_API_QUOTA, FIAT_LOG_TYPE_REFUND_API_QUOTA}

		var apiFeeTmpFls []*TmpFiatLog
		err := tx.Model(&FiatLogCache{}).Group("user_id,type").
			Where("created_at>=? and created_at<?", start, end).
			Where("is_merged=0").
			Where("type in ?", needGroupFiatLogTypes).
			Select("user_id, sum(amount) as amount, type, GROUP_CONCAT(meta) as meta, GROUP_CONCAT(id) as cache_ids").
			Scan(&apiFeeTmpFls).Error
		if err != nil {
			return err
		}

		var apiFeeFls []*FiatLog
		for _, tmpFl := range apiFeeTmpFls {
			fl := tmpFl.FiatLog

			metas, err := summaryMetas(fl.Type, fmt.Sprintf("[%s]", tmpFl.Meta))
			if err != nil {
				return err
			}
			meta, _ := json.Marshal(metas)
			fl.Meta = meta

			ids, err := unmarshalType[[]uint](fmt.Sprintf("[%s]", tmpFl.CacheIds))
			if err != nil {
				return err
			}
			fl.CacheIds = datatypes.JSONSlice[uint](*ids)

			fl.OrderNO = RandomOrderNO()
			apiFeeFls = append(apiFeeFls, &fl)
		}

		var otherFls []*FiatLog
		err = tx.Model(&FiatLogCache{}).
			Where("created_at>=? and created_at<?", start, end).
			Where("is_merged=0").
			Where("type not in ?", needGroupFiatLogTypes).
			Select("user_id, amount, type, meta, order_no, CONCAT('[' , id, ']') as cache_ids").
			Scan(&otherFls).Error
		if err != nil {
			return err
		}

		allFls := append(otherFls, apiFeeFls...)
		if len(allFls) == 0 {
			return nil
		}

		for _, fl := range allFls {
			var lastFiatLog FiatLog
			if err := tx.Model(&FiatLog{}).Where("user_id=?", fl.UserId).Order("id desc").First(&lastFiatLog).Error; err != nil {
				if !gormutils.IsRecordNotFoundError(err) {
					return err
				}
				lastFiatLog.Balance = decimal.Zero
			}

			fl.Balance = lastFiatLog.Balance.Add(fl.Amount)

		}
		if err := tx.Save(&allFls).Error; err != nil {
			return err
		}

		// update is_merged flag
		err = tx.Model(&FiatLogCache{}).
			Where("created_at>=? and created_at<?", start, end).
			Where("is_merged=0").Update("is_merged", true).Error
		if err != nil {
			return err
		}

		return nil
	})
	logrus.WithError(err).WithField("start", start).WithField("end", end).Info("merged fiatlog")
	return err
}

func summaryMetas(fiatLogType FiatLogType, metasJson string) (interface{}, error) {
	switch fiatLogType {
	case FIAT_LOG_TYPE_PAY_API_QUOTA:
		fallthrough
	case FIAT_LOG_TYPE_REFUND_API_QUOTA:
		fallthrough
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
			}, &FiatMetaPayApiFee{CostType: fms[0].CostType})
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
