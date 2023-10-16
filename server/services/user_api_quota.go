package services

import (
	"math"
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	userQuotaOperator UserQuotaOperator
)

func InitUserApiQuota() {
	userIds := models.MustGetAllUserIds()
	costTypes := models.MustGetAllCostTypes()

	if err := GetUserQuotaOperator().CreateIfNotExists(models.GetDB(), userIds, costTypes); err != nil {
		panic(err)
	}
}

func GetUserQuotaOperator() *UserQuotaOperator {
	return &userQuotaOperator
}

type UserQuotaOperator struct {
}

func (u *UserQuotaOperator) GetUserQuotasMap(userId uint) (map[enums.CostType]*models.UserApiQuota, error) {
	uqs, err := u.GetUserQuotas(userId, 0, math.MaxInt32)
	if err != nil {
		return nil, err
	}

	uqsMap := lo.SliceToMap(uqs.Items, func(v *models.UserApiQuota) (enums.CostType, *models.UserApiQuota) {
		return v.CostType, v
	})
	return uqsMap, nil
}

func (*UserQuotaOperator) GetUserQuotas(userId uint, offset int, limit int) (*ginutils.List[*models.UserApiQuota], error) {
	var quotas []*models.UserApiQuota
	var count int64
	if err := models.GetDB().Model(&models.UserApiQuota{}).Where("user_id = ?", userId).Count(&count).Offset(offset).Limit(limit).Find(&quotas).Error; err != nil {
		return nil, err
	}

	return &ginutils.List[*models.UserApiQuota]{Count: count, Items: quotas}, nil
}

func (u *UserQuotaOperator) CreateIfNotExists(tx *gorm.DB, userIds []uint, costTypes []enums.CostType) error {
	logrus.WithField("user ids", userIds).WithField("cost types", costTypes).Info("create user api quota")
	var quotas []*models.UserApiQuota
	if err := tx.Where("user_id in ?", userIds).Where("cost_type in ?", costTypes).Find(&quotas).Error; err != nil {
		return err
	}

	exists := make(map[uint]map[enums.CostType]*models.UserApiQuota)
	for _, q := range quotas {
		if _, ok := exists[q.UserId]; !ok {
			exists[q.UserId] = make(map[enums.CostType]*models.UserApiQuota)
		}
		exists[q.UserId][q.CostType] = q
	}

	var unexists []*models.UserApiQuota
	for _, userId := range userIds {
		for _, costType := range costTypes {
			if _, ok := exists[userId][costType]; !ok {
				unexists = append(unexists, &models.UserApiQuota{
					UserId:             userId,
					CostType:           costType,
					NextResetCountTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local),
				})
			}
		}
	}
	if len(unexists) == 0 {
		return nil
	}

	return tx.Save(&unexists).Error
}

// force为 true 则即使 此时在nextResetTime之前 也会强制更新（当升级套餐时需要force reset）
func (u *UserQuotaOperator) Reset(tx *gorm.DB, userIds []uint, resetQuotas map[enums.CostType]int, nextResetTime time.Time, force bool) error {
	logrus.WithField("user id", userIds).WithField("force", force).Debug("start reset user quota")
	err := func() error {
		// logrus.Debug("aaa")
		var matched []*models.UserApiQuota

		where := tx.
			Where("cost_type in ?", utils.GetMapKeys(resetQuotas)).
			Where("user_id in ?", userIds)

		if !force {
			where = where.Where("next_reset_count_time<?", nextResetTime)
		}
		// logrus.Debug("bbb")

		if err := tx.Table("user_api_quota").
			Where(where).
			Find(&matched).Error; err != nil {
			return err
		}
		if len(matched) == 0 {
			return nil
		}

		// logrus.Debug("ccc")

		for _, uaq := range matched {
			uaq.CountReset = resetQuotas[uaq.CostType]
			uaq.NextResetCountTime = nextResetTime
		}
		if err := tx.Save(&matched).Error; err != nil {
			return err
		}

		// logrus.Debug("ddd")

		for _, uaq := range matched {
			fl, err := ResetQuota(tx, uaq.UserId, uaq.CostType, uint(resetQuotas[uaq.CostType]))
			if err != nil {
				return err
			}
			logrus.WithField("fiat log cache", fl).WithField("user", uaq.UserId).WithField("cost type", uaq.CostType).WithField("count", resetQuotas[uaq.CostType]).Info("reset quota done")
		}
		return nil
	}()
	logrus.WithField("user id", userIds).WithField("force", force).WithError(err).Debug("reset user quota done")
	return err
}

func (u *UserQuotaOperator) ConsumeDataBundle(tx *gorm.DB, udb *models.UserDataBundle) error {
	if udb.IsConsumed {
		return nil
	}

	dataBundle, err := models.GetDataBundleById(udb.DataBundleId)
	if err != nil {
		return err
	}

	rolloverQuotas := lo.SliceToMap(dataBundle.DataBundleDetails, func(d *models.DataBundleDetail) (enums.CostType, int) {
		return d.CostType, d.Count * int(udb.Count)
	})

	quotas, err := u.depositRollover(tx, udb.UserId, rolloverQuotas)
	if err != nil {
		return err
	}

	udb.IsConsumed = true
	if err := tx.Save(&udb).Error; err != nil {
		return err
	}

	for _, v := range quotas {
		fl, err := DepositeDatabundle(tx, udb.UserId, udb.DataBundleId, *v)
		if err != nil {
			return err
		}
		logrus.WithField("fiat log cache", fl).Info("deposit databundle done")
	}

	return nil
}

func (u *UserQuotaOperator) depositRollover(tx *gorm.DB, userId uint, rolloverQuotas map[enums.CostType]int) ([]*models.Quota, error) {
	var costCounts []*models.Quota
	err := func() error {
		var matched []*models.UserApiQuota
		if err := tx.Table("user_api_quota").
			Where("cost_type in ?", utils.GetMapKeys(rolloverQuotas)).
			Where("user_id = ?", userId).
			Find(&matched).Error; err != nil {
			return err
		}

		for _, uaq := range matched {
			err := tx.Table("user_api_quota").
				Where("cost_type = ? and user_id = ?", uaq.CostType, uaq.UserId).
				Update("count_rollover", gorm.Expr("count_rollover+?", rolloverQuotas[uaq.CostType])).Error

			if err != nil {
				return err
			}
		}

		for _, uaq := range matched {
			costCounts = append(costCounts, &models.Quota{uaq.CostType, rolloverQuotas[uaq.CostType]})
		}
		return nil
	}()

	logrus.WithError(err).WithField("user id", userId).WithField("reset quotas", rolloverQuotas).Info("deposite user rollover quota completed")
	return costCounts, err
}

func (u *UserQuotaOperator) Refund(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {
	if err := tx.Model(&models.UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset+?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover+?", countRollover)).
		Error; err != nil {
		return 0, err
	}

	return RefundQuota(tx, userId, costType, countReset, countRollover)
}

func (u *UserQuotaOperator) Pay(tx *gorm.DB, userId uint, costType enums.CostType, countReset int, countRollover int) (uint, error) {
	if err := tx.Model(&models.UserApiQuota{}).Where("cost_type=?", costType).Where("user_id = ?", userId).
		Update("count_reset", gorm.Expr("count_reset-?", countReset)).
		Update("count_rollover", gorm.Expr("count_rollover-?", countRollover)).
		Error; err != nil {
		return 0, err
	}
	return PayQuota(tx, userId, costType, countReset, countRollover)
}
