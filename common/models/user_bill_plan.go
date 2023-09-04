package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/samber/lo"
)

// 用户套餐：用户ID，套餐ID，购买时间（生效时间），是否自动续费
type UserBillPlan struct {
	BaseModel
	UserId        uint      `json:"user"`
	PlanId        uint      `json:"plan_id"`
	BoughtTime    time.Time `json:"bought_time"`
	ExpireTime    time.Time `json:"expire_time"`
	IsAutoRenewal bool      `json:"is_auto_renewal"`
}

// 获取 用户 priority 最高的 plan, 如果没有，设置default
func FindAllUsersEffectivePlan() (map[uint]map[PlanServer]*BillPlan, error) {
	// select user_bill_plans.user_id, max(plans.priority) from user_bill_plans left join plans on user_bill_plans.plan_id=plans.id where user_bill_plans.expire_time>now() group by user_bill_plans.user_id

	// userid => plan.priority
	// plan.priority => plan

	allPlans, err := GetAllPlans()
	if err != nil {
		return nil, err
	}

	allUserIds, err := GetAllUserIds()
	if err != nil {
		return nil, err
	}

	defaultPlans, err := GetDefaultPlans()
	if err != nil {
		return nil, err
	}

	ups, err := getUserPlanIds()
	if err != nil {
		return nil, err
	}

	allPlansMap := lo.SliceToMap(allPlans, func(p *BillPlan) (uint, *BillPlan) { return p.ID, p })
	userPlans := make(map[uint]map[PlanServer]*BillPlan)
	for _, userId := range allUserIds {
		userPlans[userId] = make(map[PlanServer]*BillPlan)
		// set default plans
		for _, plan := range defaultPlans {
			userPlans[userId][plan.Server] = plan
		}

		// set max priority plans

		for _, planId := range ups[userId] {
			plan := allPlansMap[planId]
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

func getUserPlanIds() (map[uint][]uint, error) {
	type UserPlans struct {
		UserId  uint   `json:"user_id"`
		PlanIds string `json:"plan_ids"`
	}

	var _ups []*UserPlans
	err := GetDB().Model(&UserBillPlan{}).Joins("left join bill_plans on user_bill_plans.plan_id = bill_plans.id").
		Where("user_bill_plans.expire_time>?", time.Now()).
		Group("user_bill_plans.user_id").
		Select("user_bill_plans.user_id,GROUP_CONCAT(user_bill_plans.plan_id) as plan_ids").Scan(&_ups).Error
	if err != nil {
		return nil, err
	}
	ups := lo.SliceToMap(_ups, func(p *UserPlans) (uint, []uint) {
		var planIds []uint
		json.Unmarshal([]byte(fmt.Sprintf("[%s]", p.PlanIds)), &planIds)
		return p.UserId, planIds
	})
	return ups, nil
}
