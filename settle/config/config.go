package config

import cconfig "github.com/nft-rainbow/rainbow-fiat/common/config"

type Config struct {
	Mysql     cconfig.Mysql
	WechatPay cconfig.WechatPay
	Fee       cconfig.Fee
	CfxPrice  float64 `yaml:"cfxPrice"`
}

func Get() Config {
	return Config{}
}
