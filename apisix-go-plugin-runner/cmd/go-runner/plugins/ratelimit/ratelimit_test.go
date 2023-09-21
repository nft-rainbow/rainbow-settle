package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	rhttp "github.com/Conflux-Chain/go-conflux-util/rate/http"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	// 定义要执行的Shell语句
	command := `curl -k -X POST --data '{"jsonrpc":"2.0","method":"cfx_getNextNonce","params":["cfx:aajj1b1gm7k51mhzm80czcx31kwxrm2f6jxvy30mvk"],"id":1}' -H "Content-Type: application/json" http://dev-rpc-cspace-main.nftrainbow.me:9080/rvbDstNuuN`

	// 创建用于执行Shell语句的命令
	cmd := exec.Command("sh", "-c", command)

	// 将命令的输出连接到当前进程的标准输出
	// 如果您需要将输出重定向到缓冲区，可以创建一个字节缓冲区，并将cmd.Stdout设置为缓冲区
	cmd.Stdout = os.Stderr
	// 运行命令，并检查错误
	err := cmd.Run()
	if err != nil {
		fmt.Printf("运行Shell语句时出错：%s\n", err)
		return
	}
}

func BenchmarkXxx(b *testing.B) {

	fmt.Printf("n: %d", b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 定义要执行的Shell语句
		command := `curl -k -X POST --data '{"jsonrpc":"2.0","method":"cfx_getNextNonce","params":["cfx:aajj1b1gm7k51mhzm80czcx31kwxrm2f6jxvy30mvk"],"id":1}' -H "Content-Type: application/json" http://dev-rpc-cspace-main.nftrainbow.me:9080/rvbDstNuuN`

		// 创建用于执行Shell语句的命令
		cmd := exec.Command("sh", "-c", command)

		// 将命令的输出连接到当前进程的标准输出
		// 如果您需要将输出重定向到缓冲区，可以创建一个字节缓冲区，并将cmd.Stdout设置为缓冲区
		cmd.Stdout = os.Stderr
		// 运行命令，并检查错误
		err := cmd.Run()
		if err != nil {
			fmt.Printf("运行Shell语句时出错：%s\n", err)
			return
		}
	}
}

type qpsObtainerForTest struct{}

func (q *qpsObtainerForTest) getQps(serverOrCosttype, userid string) (qps, burst int, err error) {
	return 5, 10, nil
}

func TestRateRegistry(t *testing.T) {
	lm := RainbowLimiterFactory{new(qpsObtainerForTest)}
	serverReqRegistry = rhttp.NewRegistry(&lm)

	ctx := context.WithValue(context.Background(), constants.RAINBOW_USER_ID_HEADER_KEY, "1")
	err := serverReqRegistry.LimitN(ctx, "test-server-type", 100)
	assert.Error(t, err)
}
