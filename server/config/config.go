package config

import (
	"github.com/nft-rainbow/conflux-gin-helper/logger"
	cfg "github.com/nft-rainbow/rainbow-settle/common/config"
)

type Config struct {
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

var (
	_config Config
)

func InitByFile(file string) {
	_config = *cfg.InitByFile[Config](file)
}

func Get() *Config {
	return &_config
}
