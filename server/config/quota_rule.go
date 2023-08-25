package config

import (
	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

type QuotaRule struct {
	Name     string                 `yaml:"name"`
	Schedule string                 `yaml:"schedule"`
	Quotas   map[enums.CostType]int `yaml:"quotas,omitempty"`
}

func (q *QuotaRule) GetRelatedCostTypes() []enums.CostType {
	return utils.GetMapKeys(q.Quotas)
}

type quotaRuleRaw struct {
	Name     string         `yaml:"name"`
	Schedule string         `yaml:"schedule"`
	Quotas   map[string]int `yaml:"quotas,omitempty"`
}

func (q *quotaRuleRaw) verify() {
	ks := utils.GetMapKeys(q.Quotas)
	_, err := utils.MapSlice(ks, func(k string) (*enums.CostType, error) {
		return enums.ParseCostType(k)
	})
	if err != nil {
		panic(err)
	}
}

func (q *quotaRuleRaw) ToQuotaRule() (*QuotaRule, error) {
	quotas := make(map[enums.CostType]int)
	for k, v := range q.Quotas {
		costType, _ := enums.ParseCostType(k)
		quotas[*costType] = v
	}

	return &QuotaRule{
		Name:     q.Name,
		Schedule: q.Schedule,
		Quotas:   quotas,
	}, nil
}
