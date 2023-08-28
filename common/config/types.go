package config

type Mysql struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Db       string `yaml:"db"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type WechatPay struct {
	URL string `yaml:"url"`
}

type Fee struct {
	UserDefaultArrearsQuota      int64 `yaml:"userDefaultArrearsQuota"`
	UserDefaultFreeOtherAPIQuota int   `yaml:"userDefaultFreeOtherAPIQuota"`
	UserDefaultFreeMintQuota     int   `yaml:"userDefaultFreeMintQuota"`
	UserDefaultFreeDeployQuota   int   `yaml:"userDefaultFreeDeployQuota"`
}
