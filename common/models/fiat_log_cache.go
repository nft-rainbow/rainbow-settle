package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	needGroupFiatLogTypes = []FiatLogType{FIAT_LOG_TYPE_PAY_API_FEE, FIAT_LOG_TYPE_REFUND_API_FEE, FIAT_LOG_TYPE_PAY_API_QUOTA, FIAT_LOG_TYPE_REFUND_API_QUOTA, FIAT_LOG_TYPE_RESET_API_QUOTA}
	mergeFiatLogLock      sync.Mutex
)

type FiatLogCache struct {
	BaseModel
	FiatLogCore
	UnsettleAmount decimal.Decimal `gorm:"type:decimal(20,10)" json:"unsettle_amount"`
	IsMerged       bool            `gorm:"default:0" json:"isMerged,omitempty"`
}

func (f *FiatLogCache) AfterCreate(tx *gorm.DB) (err error) {
	mergeFiatLogLock.Lock()
	defer mergeFiatLogLock.Unlock()

	if f.IsMerged || lo.Contains(needGroupFiatLogTypes, f.Type) {
		return nil
	}

	// get user last fiat log and calc balance
	lastBalance, err := GetLastBlanceByFiatlog(tx, f.UserId)
	if err != nil {
		return err
	}

	fl := &FiatLog{
		FiatLogCore: f.FiatLogCore,
		CacheIds:    datatypes.JSONSlice[uint]{f.ID},
	}
	fl.Balance = lastBalance.Add(f.Amount)
	f.IsMerged = true

	if err := tx.Save(f).Error; err != nil {
		return err
	}

	return tx.Save(fl).Error
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

func FindLastFiatLogCache(userId uint, logType FiatLogType) (*FiatLogCache, error) {
	var fl FiatLogCache
	if err := db.Model(&FiatLogCache{}).
		Where("user_id=? and type =?", userId, logType).
		Order("id desc").
		First(&fl).Error; err != nil {
		return nil, err
	}
	return &fl, nil
}

func MergeToFiatlog(start, end time.Time) error {
	mergeFiatLogLock.Lock()
	defer mergeFiatLogLock.Unlock()

	type TmpFiatLog struct {
		FiatLog
		Meta     string `json:"meta"`
		CacheIds string `json:"cache_ids"`
	}

	err := GetDB().Transaction(func(tx *gorm.DB) error {
		var apiFeeTmpFls []*TmpFiatLog
		err := tx.Debug().Model(&FiatLogCache{}).Group("user_id,type").
			Where("created_at>=? and created_at<?", start, end).
			Where("is_merged=0").
			Where("type in ?", needGroupFiatLogTypes).
			Select("user_id, sum(amount) as amount, type, GROUP_CONCAT(meta) as meta, GROUP_CONCAT(id) as cache_ids").
			Scan(&apiFeeTmpFls).Error
		if err != nil {
			return errors.WithStack(err)
		}

		logrus.WithField("fls", apiFeeTmpFls).Trace("get temp api fee fiat logs")

		var apiFeeFls []*FiatLog
		for _, tmpFl := range apiFeeTmpFls {
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

		// note: avoid there is unmerged fiat log not in needGroupFiatLogTypes
		var otherFls []*FiatLog
		err = tx.Model(&FiatLogCache{}).
			Where("created_at>=? and created_at<?", start, end).
			Where("is_merged=0").
			Where("type not in ?", needGroupFiatLogTypes).
			Select("user_id, amount, type, meta, CONCAT('[' , id, ']') as cache_ids").
			Scan(&otherFls).Error
		if err != nil {
			return errors.WithStack(err)
		}

		allFls := append(otherFls, apiFeeFls...)
		if len(allFls) == 0 {
			return nil
		}

		lastBalances := make(map[uint]decimal.Decimal)
		for _, fl := range allFls {
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
		logrus.WithField("all fls", allFls).Debug("save fiat logs")
		if err := tx.Debug().Save(&allFls).Error; err != nil {
			return errors.WithStack(err)
		}

		// update is_merged flag
		err = tx.Model(&FiatLogCache{}).
			Where("created_at>=? and created_at<?", start, end).
			Where("is_merged=0").Update("is_merged", true).Error
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
	logrus.WithField("error", fmt.Sprintf("%+v", err)).WithField("start", start).WithField("end", end).Info("merged fiatlog")
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
	case FIAT_LOG_TYPE_RESET_API_QUOTA:
		fallthrough
	case FIAT_LOG_TYPE_PAY_API_FEE:
		fms, err := unmarshalType[[]*FiatMetaPayApiFee](metasJson)
		if err != nil {
			logrus.WithField("input", metasJson).Debug("failed unmarshal metas to []*FiatMetaPayApiFee")
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
		return nil, errors.Errorf("not supported %v", fiatLogType)
	}
}

func unmarshalType[T any](jsonStr string) (*T, error) {
	var fm *T
	if err := json.Unmarshal([]byte(jsonStr), &fm); err != nil {
		return nil, err
	}
	return fm, nil
}
