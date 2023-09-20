package plugins

import (
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/auth"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/count"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/confura"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/rainbowapi"
	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
)

func init() {
	redis.Init(config.Redis{
		Host: "redis",
		Port: 6379,
	})
}
