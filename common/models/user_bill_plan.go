package models

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils/gormutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// 用户套餐：用户ID，套餐ID，购买时间（生效时间），是否自动续费
type UserBillPlan struct {
	BaseModel
	UserId        uint             `gorm:"index:idx_userid_servertype" json:"user_id"`
	ServerType    enums.ServerType `gorm:"index:idx_userid_servertype" json:"server_type"`
	PlanId        uint             `json:"plan_id"`
	BoughtTime    time.Time        `json:"bought_time"`
	ExpireTime    time.Time        `json:"expire_time"`
	IsAutoRenewal bool             `json:"is_auto_renewal"`
}

// func InitUserUserBillPlan() {
// 	userIds := MustGetAllUserIds()
// 	serverTypes := enums.GetAllServerTypes()

// 	if err := GetUserBillPlanOperator().CreateIfNotExists(GetDB(), userIds, serverTypes); err != nil {
// 		panic(err)
// 	}
// }

var (
	userBillPlanOperator UserBillPlanOperator
)

func GetUserBillPlanOperator() *UserBillPlanOperator {
	return &userBillPlanOperator
}

type OnUserBillPlanChangedHandler func(old, new *UserBillPlan)
type UserBillPlanOperator struct {
	onChanged []OnUserBillPlanChangedHandler
}

func (u *UserBillPlanOperator) RegisterOnChangedEvent(handler OnUserBillPlanChangedHandler) {
	u.onChanged = append(u.onChanged, handler)
}

// // @deprecateds
// func (u *UserBillPlanOperator) createIfNotExists(tx *gorm.DB, userIds []uint, serverTypes []enums.ServerType) error {
// 	var userPlans []*UserBillPlan
// 	if err := tx.Where("user_id in ?", userIds).Where("server_type in ?", serverTypes).Find(&userPlans).Error; err != nil {
// 		return err
// 	}

// 	exists := make(map[uint]map[enums.ServerType]*UserBillPlan)
// 	for _, q := range userPlans {
// 		if _, ok := exists[q.UserId]; !ok {
// 			exists[q.UserId] = make(map[enums.ServerType]*UserBillPlan)
// 		}
// 		exists[q.UserId][q.ServerType] = q
// 	}

// 	defaultPlans, err := GetDefaultPlans()
// 	if err != nil {
// 		return err
// 	}

// 	var unexists []*UserBillPlan
// 	for _, userId := range userIds {
// 		for _, serverType := range serverTypes {
// 			if _, ok := exists[userId][serverType]; !ok {
// 				unexists = append(unexists, &UserBillPlan{
// 					UserId:        userId,
// 					ServerType:    serverType,
// 					PlanId:        defaultPlans[serverType].ID,
// 					BoughtTime:    time.Now(),
// 					ExpireTime:    defaultPlans[serverType].EffectivePeroid.EndTime(time.Now()),
// 					IsAutoRenewal: true,
// 				})
// 			}
// 		}
// 	}
// 	if len(unexists) == 0 {
// 		return nil
// 	}

// 	for _, h := range u.onChanged {
// 		h(unexists)
// 	}

// 	return tx.Save(&unexists).Error
// }

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
func (u *UserBillPlanOperator) UpdateUserBillPlan(userId uint, planId uint, isAutoRenew bool) (*UserBillPlan, error) {
	var userPlan UserBillPlan
	var old *UserBillPlan
	err := GetDB().Transaction(func(tx *gorm.DB) error {
		plan, err := GetBillPlanById(planId)
		if err != nil {
			return err
		}

		if err := tx.Where("user_id=? and server_type=?", userId, plan.Server).First(&userPlan).Error; err != nil {
			if gormutils.IsRecordNotFoundError(err) {
				userPlan.UserId = userId
				userPlan.ServerType = plan.Server
			} else {
				return err
			}
		} else {
			_old := userPlan
			old = &_old
		}

		now := time.Now()
		userPlan.PlanId = planId
		userPlan.BoughtTime = now
		userPlan.ExpireTime = plan.EffectivePeroid.EndTime(now)
		userPlan.IsAutoRenewal = isAutoRenew

		if err := tx.Save(&userPlan).Error; err != nil {
			return err
		}

		for _, h := range u.onChanged {
			h(old, &userPlan)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &userPlan, nil
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
func (u *UserBillPlanOperator) FindAllUsersEffectivePlan() (map[uint]map[enums.ServerType]*BillPlan, error) {
	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}
	return u.FindUsersEffectivePlans(allUserIds)
}

func (u *UserBillPlanOperator) GetUserEffectivePlans(userId uint) (map[enums.ServerType]*BillPlan, error) {
	plans, err := u.FindUsersEffectivePlans([]uint{userId})
	if err != nil {
		return nil, err
	}
	return plans[userId], nil
}

func (u *UserBillPlanOperator) FindAllUserNeedRenewPlans() (map[uint]map[enums.ServerType]*BillPlan, error) {
	allPlansMap, err := GetAllPlansMap()
	if err != nil {
		return nil, err
	}

	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}
	allUserPlans, err := u.getUserPlansNeedRenew(allUserIds)
	if err != nil {
		return nil, err
	}

	result := make(map[uint]map[enums.ServerType]*BillPlan)
	for userId, userPlans := range allUserPlans {
		result[userId] = make(map[enums.ServerType]*BillPlan)
		for serverType, userPlan := range userPlans {
			result[userId][serverType] = allPlansMap[userPlan.PlanId]
		}
	}
	return result, nil
}

func (u *UserBillPlanOperator) FindUsersEffectivePlans(userIds []uint) (map[uint]map[enums.ServerType]*BillPlan, error) {
	allPlansMap, err := GetAllPlansMap()
	if err != nil {
		return nil, err
	}

	defaultPlans, err := GetDefaultPlans()
	if err != nil {
		return nil, err
	}

	userUnexpirePlans, err := u.getUserUnExpiredPlans(userIds)
	if err != nil {
		return nil, err
	}

	allServerTypes := enums.GetAllServerTypes()

	result := make(map[uint]map[enums.ServerType]*BillPlan)
	for _, userId := range userIds {
		result[userId] = make(map[enums.ServerType]*BillPlan)
		if userUnexpirePlans[userId] == nil {
			userUnexpirePlans[userId] = make(map[enums.ServerType]*UserBillPlan)
		}

		for _, servetType := range allServerTypes {
			userPlan := userUnexpirePlans[userId][servetType]
			if userPlan != nil {
				result[userId][servetType] = allPlansMap[userPlan.ID]
				continue
			}
			result[userId][servetType] = defaultPlans[servetType]
		}
	}
	return result, nil
}

// func findUsersEffectivePlan(userIds []uint) (map[uint]map[enums.ServerType]*BillPlan, error) {
// 	// select user_bill_plans.user_id, max(plans.priority) from user_bill_plans left join plans on user_bill_plans.plan_id=plans.id where user_bill_plans.expire_time>now() group by user_bill_plans.user_id

// 	// userid => plan.priority
// 	// plan.priority => plan

// 	allPlansMap, err := GetAllPlansMap()
// 	if err != nil {
// 		return nil, err
// 	}

// 	defaultPlans, err := GetDefaultPlans()
// 	if err != nil {
// 		return nil, err
// 	}

// 	ups, err := getUserUnExpiredPlans()
// 	if err != nil {
// 		return nil, err
// 	}

// 	userPlans := make(map[uint]map[enums.ServerType]*BillPlan)
// 	for _, userId := range userIds {
// 		userPlans[userId] = make(map[enums.ServerType]*BillPlan)
// 		// set default plans
// 		for _, plan := range defaultPlans {
// 			userPlans[userId][plan.Server] = plan
// 		}

// 		// set max priority plans

// 		for _, userPlan := range ups[userId] {
// 			plan := allPlansMap[userPlan.PlanId]
// 			if userPlans[userId][plan.Server].Priority < plan.Priority {
// 				userPlans[userId][plan.Server] = plan
// 			}
// 		}
// 	}
// 	return userPlans, nil
// }

// // 找用户所有plan，标记renew的优先级最高的且优先级大于当前生效的
// func findUserNeedRenewPlans(userIds []uint) (map[uint]map[enums.ServerType]*BillPlan, error) {
// 	allUps, err := getUserAllPlans()
// 	if err != nil {
// 		return nil, err
// 	}

// 	allPlansMap, err := GetAllPlansMap()
// 	if err != nil {
// 		return nil, err
// 	}

// 	allUeps, err := FindAllUsersEffectivePlan()
// 	if err != nil {
// 		return nil, err
// 	}

// 	userNeedRenews := make(map[uint]map[enums.ServerType]*BillPlan)
// 	for _, userId := range userIds {
// 		userNeedRenews[userId] = make(map[enums.ServerType]*BillPlan)
// 		// set max priority plans
// 		for _, userPlan := range allUps[userId] {
// 			plan := allPlansMap[userPlan.PlanId]
// 			// 不续费的跳过，优先级低于正生效的跳过
// 			if !userPlan.IsAutoRenewal || plan.Priority < allUeps[userId][plan.Server].Priority {
// 				continue
// 			}

// 			curPlan := userNeedRenews[userId][plan.Server]
// 			if curPlan == nil || curPlan.Priority < plan.Priority {
// 				userNeedRenews[userId][plan.Server] = plan
// 			}
// 		}
// 	}
// 	return userNeedRenews, nil

// }

func (u *UserBillPlanOperator) getUserPlansNeedRenew(userIds []uint) (map[uint]map[enums.ServerType]*UserBillPlan, error) {
	var _ubps []*UserBillPlan
	err := GetDB().Model(&UserBillPlan{}).
		Where("user_id in ?", userIds).
		Where("expire_time<=?", time.Now()).
		Where("is_auto_renewal=?", true).
		Find(&_ubps).Error
	if err != nil {
		return nil, err
	}

	return groupUserPlanByUserId(_ubps), nil
}

// func getUserAllPlans(userIds []uint) (map[uint]map[enums.ServerType]*UserBillPlan, error) {
// 	var _ubps []*UserBillPlan
// 	err := GetDB().Model(&UserBillPlan{}).Where("user_id in ?", userIds).Find(&_ubps).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	return groupUserPlanByUserId(_ubps), nil
// }

func (u *UserBillPlanOperator) getUserUnExpiredPlans(userIds []uint) (map[uint]map[enums.ServerType]*UserBillPlan, error) {
	var _ubps []*UserBillPlan
	err := GetDB().Model(&UserBillPlan{}).
		Where("user_id in ?", userIds).
		Where("expire_time>?", time.Now()).
		Find(&_ubps).Error
	if err != nil {
		return nil, err
	}

	return groupUserPlanByUserId(_ubps), nil
}

func groupUserPlanByUserId(input []*UserBillPlan) map[uint]map[enums.ServerType]*UserBillPlan {
	user2Bplans := lo.GroupBy(input, func(v *UserBillPlan) uint {
		return v.PlanId
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
