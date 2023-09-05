package config

import (
	"github.com/nft-rainbow/conflux-gin-helper/logger"
	cfg "github.com/nft-rainbow/rainbow-settle/common/config"
)

type ConfigBase struct {
	Environment string `yaml:"environment"`
	Port        struct {
		Grpc    int `yaml:"grpc"`
		RestApi int `yaml:"rest_api"`
	} `yaml:"port"`
	Log       logger.LogConfig `yaml:"log"`
	Mysql     cfg.Mysql        `yaml:"mysql"`
	Redis     cfg.Redis        `yaml:"redis"`
	WechatPay cfg.WechatPay    `yaml:"wechatPay"`
	Fee       cfg.Fee          `yaml:"fee"`
	CfxPrice  float64          `yaml:"cfxPrice"`
	Schedules struct {
		MergeFiatlog string `yaml:"mergeFiatlog"`
	} `yaml:"schedules"`
}

// type Config struct {
// 	ConfigBase
// 	// QuotaRules []*QuotaRule `yaml:"quotaRules"`
// }

var (
	_config ConfigBase
)

func InitByFile(file string) {
	_config = *cfg.InitByFile[ConfigBase](file)

	// type tmpConfig struct {
	// 	QuotaRules []quotaRuleRaw `yaml:"quotaRules"`
	// }
	// c := *cfg.InitByFile[ConfigBase](file)
	// qrs := *cfg.InitByFile[tmpConfig](file)
	// for _, v := range qrs.QuotaRules {
	// 	v.verify()
	// }

	// _config = Config{ConfigBase: c}
	// _config.QuotaRules = utils.MustMapSlice(qrs.QuotaRules, func(r quotaRuleRaw) *QuotaRule {
	// 	q, err := r.ToQuotaRule()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return q
	// })
}

func Get() *ConfigBase {
	return &_config
}
