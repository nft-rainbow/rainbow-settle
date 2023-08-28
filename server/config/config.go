package config

import (
	"github.com/nft-rainbow/conflux-gin-helper/logger"
	"github.com/nft-rainbow/conflux-gin-helper/utils"
	cfg "github.com/nft-rainbow/rainbow-settle/common/config"
)

type ConfigBase struct {
	Environment string           `yaml:"environment"`
	Log         logger.LogConfig `yaml:"log"`
	Port        int              `yaml:"port"`
	Mysql       cfg.Mysql        `yaml:"mysql"`
	Redis       cfg.Redis        `yaml:"redis"`
	WechatPay   cfg.WechatPay    `yaml:"wechatPay"`
	Fee         cfg.Fee          `yaml:"fee"`
	CfxPrice    float64          `yaml:"cfxPrice"`
	Schedules   struct {
		MergeFiatlog string `yaml:"mergeFiatlog"`
	} `yaml:"schedules"`
}

type Config struct {
	ConfigBase
	QuotaRules []*QuotaRule `yaml:"quotaRules"`
}

var (
	_config Config
)

func InitByFile(file string) {
	type tmpConfig struct {
		QuotaRules []quotaRuleRaw `yaml:"quotaRules"`
	}
	c := *cfg.InitByFile[ConfigBase](file)
	qrs := *cfg.InitByFile[tmpConfig](file)
	for _, v := range qrs.QuotaRules {
		v.verify()
	}

	_config = Config{ConfigBase: c}
	_config.QuotaRules = utils.MustMapSlice(qrs.QuotaRules, func(r quotaRuleRaw) *QuotaRule {
		q, err := r.ToQuotaRule()
		if err != nil {
			panic(err)
		}
		return q
	})
}

func Get() *Config {
	return &_config
}
