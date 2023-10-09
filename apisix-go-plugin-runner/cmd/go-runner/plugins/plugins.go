package plugins

import (
	"os"

	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/auth"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/count"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/ratelimit"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/confura"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/rainbowapi"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/reqparser/scan"
	_ "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/resphandler"
	"github.com/sirupsen/logrus"

	pconfig "github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/config"
	"github.com/nft-rainbow/rainbow-settle/common/redis"
)

func init() {
	wd, err := os.Getwd()
	logrus.WithError(err).WithField("wd", wd).Info("current working directory")
	pconfig.InitByFile("../apisix-go-plugin-runner/config.yaml")
	redis.Init(pconfig.Get().Redis)
}
