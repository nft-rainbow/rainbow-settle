package models

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var (
	userBillPlanOperator          UserBillPlanOperator
	onUserBillPlanChangedHandlers []OnUserBillPlanChangedHandler
)

type OnUserBillPlanChangedHandler func(old, new *UserBillPlan)

// 用户套餐：用户ID，套餐ID，购买时间（生效时间），是否自动续费
type UserBillPlan struct {
	BaseModel
	UserId        uint             `gorm:"index:idx_userid_servertype" json:"user_id"`
	ServerType    enums.ServerType `gorm:"index:idx_userid_servertype" json:"server_type"`
	PlanId        uint             `json:"plan_id"`
	BoughtTime    time.Time        `json:"bought_time"`
	ExpireTime    time.Time        `json:"expire_time"`
	IsAutoRenewal bool             `json:"is_auto_renewal"`
	Plan          *BillPlan        `gorm:"-" json:"plan"`
}

func (p *UserBillPlan) PopulatePlan() error {
	plan, err := GetBillPlanById(p.PlanId)
	if err != nil {
		return err
	}
	p.Plan = plan
	return nil
}

func (u *UserBillPlan) AfterFind(tx *gorm.DB) (err error) {
	if u.Plan == nil {
		return u.PopulatePlan()
	}
	return
}

// func (u *UserBillPlan) AfterSave(tx *gorm.DB) (err error) {
// 	for _, h := range onUserBillPlanChangedHandlers {
// 		h(u)
// 	}
// 	return nil
// }

type UserBillPlanMap map[uint]map[enums.ServerType]*UserBillPlan

func (u *UserBillPlanMap) PopulatePlans() error {
	for _, server2UserPlan := range *u {
		for _, userPlan := range server2UserPlan {
			if err := userPlan.PopulatePlan(); err != nil {
				return err
			}
		}
	}
	return nil
}

type UserBillPlanOperator struct {
	// onChanged []OnUserBillPlanChangedHandler
}

func GetUserBillPlanOperator() *UserBillPlanOperator {
	return &userBillPlanOperator
}

func (u *UserBillPlanOperator) RegisterOnChangedEvent(handler OnUserBillPlanChangedHandler) {
	onUserBillPlanChangedHandlers = append(onUserBillPlanChangedHandlers, handler)
}

type UserBillPlanFilter struct {
	UserId     uint             `json:"user_id"`
	ServerType enums.ServerType `json:"server_type"`
	PlanId     uint             `json:"plan_id"`
}

func (u *UserBillPlanOperator) First(filter *UserBillPlanFilter) (*UserBillPlan, error) {
	var userPlan UserBillPlan
	if err := GetDB().Where(&filter).First(&userPlan).Error; err != nil {
		return nil, err
	}
	return &userPlan, nil
}

// 购买新套餐时，直接覆盖现有套餐
func (u *UserBillPlanOperator) UpdateUserBillPlan(tx *gorm.DB, userId uint, planId uint, isAutoRenew bool) (*UserBillPlan, error) {

	var oldUserPlan, newUserPlan *UserBillPlan
	coreFn := func(_tx *gorm.DB) error {
		newPlan, err := GetBillPlanById(planId)
		if err != nil {
			return err
		}

		if err := _tx.Where("user_id=? and server_type=?", userId, newPlan.Server).First(&oldUserPlan).Error; err != nil {
			oldUserPlan = nil
			if !gormutils.IsRecordNotFoundError(err) {
				return err
			}
		}

		if oldUserPlan != nil {
			if oldUserPlan.Plan.Priority == newPlan.Priority && oldUserPlan.ExpireTime.After(time.Now()) {
				return errors.New("the plan is using currently")
			}
		}

		now := time.Now()
		newUserPlan = &UserBillPlan{
			UserId:        userId,
			ServerType:    newPlan.Server,
			PlanId:        planId,
			BoughtTime:    now,
			ExpireTime:    newPlan.EffectivePeroid.EndTime(now),
			IsAutoRenewal: isAutoRenew,
		}
		if oldUserPlan != nil {
			newUserPlan.BaseModel = oldUserPlan.BaseModel
		}

		if err := _tx.Save(&newUserPlan).Error; err != nil {
			return err
		}

		return nil
	}

	var err error
	if tx == db {
		err = GetDB().Transaction(func(tx *gorm.DB) error {
			return coreFn(tx)
		})
	} else {
		err = coreFn(tx)
	}

	if err != nil {
		return nil, err
	}

	// TODO: 如果事务失败需要重新触发事件
	for _, h := range onUserBillPlanChangedHandlers {
		h(oldUserPlan, newUserPlan)
	}

	return newUserPlan, nil
}

func (u *UserBillPlanOperator) UpdateRenew(userId uint, serverType enums.ServerType, isAutoRenew bool) error {
	var userPlan UserBillPlan
	if err := GetDB().Where("user_id=? and server_type=?", userId, serverType).First(&userPlan).Error; err != nil {
		return err
	}
	userPlan.IsAutoRenewal = isAutoRenew
	return GetDB().Save(&userPlan).Error
}

// 获取 用户 priority 最高的 plan, 如果没有，设置default
func (u *UserBillPlanOperator) FindAllUsersEffectivePlan() (UserBillPlanMap, error) {
	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}
	return u.FindUsersEffectivePlans(allUserIds)
}

func (u *UserBillPlanOperator) GetUserEffectivePlans(userId uint, isContainRainbow bool) (map[enums.ServerType]*UserBillPlan, error) {
	allUserPlans, err := u.FindUsersEffectivePlans([]uint{userId})
	if err != nil {
		return nil, err
	}

	plans := allUserPlans[userId]

	if !isContainRainbow {
		delete(plans, enums.SERVER_TYPE_RAINBOW)
	}

	return plans, nil
}

func (u *UserBillPlanOperator) FindAllUserNeedRenewPlans() (UserBillPlanMap, error) {
	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}
	allUserPlans, err := u.getUserPlansNeedRenew(allUserIds)
	if err != nil {
		return nil, err
	}
	return allUserPlans, nil
}

func (u *UserBillPlanOperator) FindUsersEffectivePlans(userIds []uint) (UserBillPlanMap, error) {
	userUnexpirePlans, err := u.getUserUnExpiredPlans(userIds, true)
	if err != nil {
		return nil, err
	}

	return userUnexpirePlans, nil
}

func (u *UserBillPlanOperator) getUserPlansNeedRenew(userIds []uint) (UserBillPlanMap, error) {
	var _ubps []*UserBillPlan
	err := GetDB().Model(&UserBillPlan{}).
		Where("user_id in ?", userIds).
		Where("expire_time<=?", time.Now()).
		Where("is_auto_renewal=?", true).
		Find(&_ubps).Error
	if err != nil {
		return nil, err
	}

	grouped := groupUserPlanByUserId(_ubps)
	return grouped, nil
}

func (u *UserBillPlanOperator) getUserUnExpiredPlans(userIds []uint, isUseDefaultIfEmpty bool) (UserBillPlanMap, error) {
	var _ubps []*UserBillPlan
	err := GetDB().Model(&UserBillPlan{}).
		Where("user_id in ?", userIds).
		Where("expire_time>?", time.Now()).
		Find(&_ubps).Error
	if err != nil {
		return nil, err
	}

	defaultPlans, err := GetDefaultPlans()
	if err != nil {
		return nil, err
	}

	grouped := groupUserPlanByUserId(_ubps)
	if !isUseDefaultIfEmpty {
		return grouped, nil
	}

	serverTypes := enums.GetAllServerTypes()
	for _, userId := range userIds {
		if grouped[userId] == nil {
			grouped[userId] = make(map[enums.ServerType]*UserBillPlan)
		}
		for _, serverType := range serverTypes {
			if grouped[userId][serverType] == nil {
				grouped[userId][serverType] = &UserBillPlan{
					UserId:     userId,
					ServerType: serverType,
					PlanId:     defaultPlans[serverType].ID,
					Plan:       defaultPlans[serverType],
					ExpireTime: time.Now().AddDate(10, 0, 0),
				}
			}

		}
	}
	return grouped, nil
}

func groupUserPlanByUserId(input []*UserBillPlan) UserBillPlanMap {
	user2Bplans := lo.GroupBy(input, func(v *UserBillPlan) uint {
		return v.UserId
	})

	result := make(map[uint]map[enums.ServerType]*UserBillPlan)
	for userId, billPlans := range user2Bplans {
		server2Bplan := lo.SliceToMap(billPlans, func(p *UserBillPlan) (enums.ServerType, *UserBillPlan) {
			return p.ServerType, p
		})
		result[userId] = server2Bplan
	}
	return result
}
