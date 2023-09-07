package models

import (
	"time"

	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type PeroidType int

func (p PeroidType) ToCronSchedule() string {
	coll := map[PeroidType]string{
		PEROID_TYPE_DAY:   "@daily",
		PEROID_TYPE_MONTH: "@monthly",
		PEROID_TYPE_YEAR:  "@yearly",
	}
	return coll[p]
}

func (p PeroidType) EndTime(beginTime time.Time) time.Time {
	// coll := map[PeroidType]time.Time{
	// 	PEROID_TYPE_DAY:   time.Date(0, 0, 1, 0, 0, 0, 0, time.Local),
	// 	PEROID_TYPE_MONTH: time.Date(0, 1, 0, 0, 0, 0, 0, time.Local),
	// 	PEROID_TYPE_YEAR:  time.Date(1, 0, 0, 0, 0, 0, 0, time.Local),
	// }
	// return beginTime.AddDate(coll[p].Year(), int(coll[p].Month()), coll[p].Day())

	y, m, d := beginTime.Date()
	switch p {
	case PEROID_TYPE_DAY:
		return beginTime.AddDate(0, 0, 1)
	case PEROID_TYPE_MONTH:
		return time.Date(y, m+1, d, 0, 0, 0, 0, time.Local)
	case PEROID_TYPE_YEAR:
		return time.Date(y+1, 1, d, 0, 0, 0, 0, time.Local)
	}
	return beginTime
}

const (
	PEROID_TYPE_DAY PeroidType = iota + 1
	PEROID_TYPE_MONTH
	PEROID_TYPE_YEAR
)

type PlanServer int

const (
	PLAN_SERVER_RAINBOW = iota + 1
	PLAN_SERVER_CONFURA
	PLAN_SERVER_SCAN
)

// 包月/年套餐：名称，生效时长，qps，价格
type BillPlan struct {
	BaseModel
	Name                 string            `json:"name"`
	EffectivePeroid      PeroidType        `json:"effective_peroid"` // 月，年
	RefreshQuotaSchedule PeroidType        `json:"reset_duration"`   // 日，月，年
	Qps                  int               `json:"qps"`
	Price                decimal.Decimal   `json:"price"`
	Server               PlanServer        `json:"server"`
	Priority             int               `json:"priority"` // 同一plan server下plan都是互斥的，哪个生效由priority决定，如企业版>普通版
	BillPlanDetails      []*BillPlanDetail `json:"bill_plan_details"`
}

func (p *BillPlan) GetQuotas() map[enums.CostType]int {
	return lo.SliceToMap(p.BillPlanDetails, func(d *BillPlanDetail) (enums.CostType, int) {
		return d.CostType, d.Count
	})
}

type BillPlanFilter struct {
	ID     uint       `form:"id" json:"id"`
	Server PlanServer `form:"server" json:"server"`
}

func QueryBillPlan(filter *BillPlanFilter, offset, limit int) (*ginutils.List[*BillPlan], error) {
	var aps []*BillPlan
	var count int64
	if err := GetDB().Model(&BillPlan{}).Where(&filter).Count(&count).Offset(offset).Limit(limit).Find(&aps).Error; err != nil {
		return nil, err
	}
	return &ginutils.List[*BillPlan]{Items: aps, Count: count}, nil
}

func GetBillPlanById(id uint) (*BillPlan, error) {
	var p *BillPlan
	if err := GetDB().Model(&BillPlan{}).Preload("BillPlanDetails").First(id, &p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func GetAllPlans() ([]*BillPlan, error) {
	var plans []*BillPlan
	if err := GetDB().Model(&BillPlan{}).Preload("BillPlanDetails").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func GetAllPlansMap() (map[uint]*BillPlan, error) {
	allPlans, err := GetAllPlans()
	if err != nil {
		return nil, err
	}
	allPlansMap := lo.SliceToMap(allPlans, func(p *BillPlan) (uint, *BillPlan) { return p.ID, p })
	return allPlansMap, nil
}

func GetDefaultPlans() (map[PlanServer]*BillPlan, error) {
	var plans []*BillPlan
	if err := GetDB().Model(&BillPlan{}).Preload("BillPlanDetails").Where("name like \"default_%\"").Find(&plans).Error; err != nil {
		return nil, err
	}

	result := lo.SliceToMap(plans, func(p *BillPlan) (PlanServer, *BillPlan) {
		return p.Server, p
	})
	return result, nil
}
