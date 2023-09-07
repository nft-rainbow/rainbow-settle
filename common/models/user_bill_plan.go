package models

import (
	"time"

	"github.com/samber/lo"
)

// 用户套餐：用户ID，套餐ID，购买时间（生效时间），是否自动续费
type UserBillPlan struct {
	BaseModel
	UserId        uint      `json:"user_id"`
	ServerType    uint      `json:"server_type"`
	PlanId        uint      `json:"plan_id"`
	BoughtTime    time.Time `json:"bought_time"`
	ExpireTime    time.Time `json:"expire_time"`
	IsAutoRenewal bool      `json:"is_auto_renewal"`
}

// 购买新套餐时，直接覆盖现有套餐
func CreateUserBillPlan(userId, planId uint, isAutoRenew bool) (*UserBillPlan, error) {
	plan, err := GetBillPlanById(planId)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	up := &UserBillPlan{
		UserId:        userId,
		PlanId:        planId,
		BoughtTime:    now,
		ExpireTime:    plan.EffectivePeroid.EndTime(now),
		IsAutoRenewal: isAutoRenew,
	}
	if err := GetDB().Create(&up).Error; err != nil {
		return nil, err
	}
	return up, nil
}

func GetUserEffectivePlans(userId uint) (map[PlanServer]*BillPlan, error) {
	plans, err := findUsersEffectivePlan([]uint{userId})
	if err != nil {
		return nil, err
	}
	return plans[userId], nil
}

// 获取 用户 priority 最高的 plan, 如果没有，设置default
func FindAllUsersEffectivePlan() (map[uint]map[PlanServer]*BillPlan, error) {
	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}
	return findUsersEffectivePlan(allUserIds)
}

func findUsersEffectivePlan(userIds []uint) (map[uint]map[PlanServer]*BillPlan, error) {
	// select user_bill_plans.user_id, max(plans.priority) from user_bill_plans left join plans on user_bill_plans.plan_id=plans.id where user_bill_plans.expire_time>now() group by user_bill_plans.user_id

	// userid => plan.priority
	// plan.priority => plan

	allPlansMap, err := GetAllPlansMap()
	if err != nil {
		return nil, err
	}

	defaultPlans, err := GetDefaultPlans()
	if err != nil {
		return nil, err
	}

	ups, err := getUserUnExpiredPlans()
	if err != nil {
		return nil, err
	}

	userPlans := make(map[uint]map[PlanServer]*BillPlan)
	for _, userId := range userIds {
		userPlans[userId] = make(map[PlanServer]*BillPlan)
		// set default plans
		for _, plan := range defaultPlans {
			userPlans[userId][plan.Server] = plan
		}

		// set max priority plans

		for _, userPlan := range ups[userId] {
			plan := allPlansMap[userPlan.PlanId]
			if userPlans[userId][plan.Server].Priority < plan.Priority {
				userPlans[userId][plan.Server] = plan
			}
		}
	}
	return userPlans, nil

	// set max priority plans
	// for _, ump := range _ups {
	// 	for _, plan := range allPlans {
	// 		if plan.Server == ump.Server && plan.Priority == ump.MaxPriority {
	// 			userPlans[ump.UserId][plan.Server] = plan
	// 		}
	// 	}
	// }

	// priorities := lo.Map(ump, func(v *UserMaxPlanPriority, i int) uint { return v.MaxPriority })
	// priorities = lo.Uniq(priorities)

	// var plans []*Plan
	// err = GetDB().Model(&Plan{}).Where("priority in ?", priorities).Find(&plans).Error
	// if err != nil {
	// 	return nil, err
	// }

	// priority2Plan := lo.SliceToMap(plans, func(v *Plan) (int, *Plan) {
	// 	return v.Priority, v
	// })

	// user2Plan := lo.SliceToMap(ump, func(v *UserMaxPlanPriority) (uint, []*Plan) {
	// 	return v.UserId, priority2Plan[int(v.MaxPriority)]
	// })

	// for _, user := range ump {

	// }

	// return user2Plan, nil
}

func FindAllUserNeedRenewPlans() (map[uint]map[PlanServer]*BillPlan, error) {
	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}
	return findUserNeedRenewPlans(allUserIds)
}

// 找用户所有plan，标记renew的优先级最高的且优先级大于当前生效的
func findUserNeedRenewPlans(userIds []uint) (map[uint]map[PlanServer]*BillPlan, error) {
	allUps, err := getUserAllPlans()
	if err != nil {
		return nil, err
	}

	allPlansMap, err := GetAllPlansMap()
	if err != nil {
		return nil, err
	}

	allUeps, err := FindAllUsersEffectivePlan()
	if err != nil {
		return nil, err
	}

	userNeedRenews := make(map[uint]map[PlanServer]*BillPlan)
	for _, userId := range userIds {
		userNeedRenews[userId] = make(map[PlanServer]*BillPlan)
		// set max priority plans
		for _, userPlan := range allUps[userId] {
			plan := allPlansMap[userPlan.PlanId]
			// 不续费的跳过，优先级低于正生效的跳过
			if !userPlan.IsAutoRenewal || plan.Priority < allUeps[userId][plan.Server].Priority {
				continue
			}

			curPlan := userNeedRenews[userId][plan.Server]
			if curPlan == nil || curPlan.Priority < plan.Priority {
				userNeedRenews[userId][plan.Server] = plan
			}
		}
	}
	return userNeedRenews, nil

}

func getUserAllPlans() (map[uint][]*UserBillPlan, error) {
	var _ubps []*UserBillPlan
	err := GetDB().Model(&UserBillPlan{}).Find(&_ubps).Error
	if err != nil {
		return nil, err
	}

	return groupUserPlanByUserId(_ubps), nil
}

func getUserUnExpiredPlans() (map[uint][]*UserBillPlan, error) {
	var _ubps []*UserBillPlan
	err := GetDB().Model(&UserBillPlan{}).
		Where("expire_time>?", time.Now()).
		Find(&_ubps).Error
	if err != nil {
		return nil, err
	}

	return groupUserPlanByUserId(_ubps), nil
}

func groupUserPlanByUserId(input []*UserBillPlan) map[uint][]*UserBillPlan {
	return lo.GroupBy(input, func(v *UserBillPlan) uint {
		return v.PlanId
	})
}
