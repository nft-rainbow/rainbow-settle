package config

import (
	cconfig "github.com/nft-rainbow/rainbow-settle/common/config"
	cfg "github.com/nft-rainbow/rainbow-settle/common/config"
)

type Config struct {
	Environment string        `yaml:"environment"`
	Redis       cconfig.Redis `yaml:"redis"`
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
