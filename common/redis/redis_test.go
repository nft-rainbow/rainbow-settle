package redis

import (
	"context"
	"os"
	"testing"

	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	Init(config.Redis{Host: "172.16.100.252", Port: 6379})
	code := m.Run()
	os.Exit(code)
}

func TestCheckIsRich(t *testing.T) {
	assert.True(t, isRich(1055, enums.COST_TYPE_RAINBOW_NORMAL))
	assert.True(t, isRich(103910431, enums.COST_TYPE_CONFURA_MAIN_CSPACE_NORMAL))
}

func TestConnectDB(t *testing.T) {
	_, err := DB().Get(context.Background(), "rich-1").Result()
	assert.NoError(t, err)
}

func TestGetUserServerQps(t *testing.T) {
	_, _, err := GetUserServerQps(1, enums.SERVER_TYPE_CONFURA_CSPACE)
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func BenchmarkGetUserInfoByApikey(b *testing.B) {
	// Init(config.Redis{Port: 6379})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// _, _, err := GetUserInfoByApikey("rvbDstNuuN")
		_, _, err := GetUserServerQps(1, enums.SERVER_TYPE_CONFURA_CSPACE)
		// _, err := DB().Get(context.Background(), "count-pending-1-rainbow_normal").Result()
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}
