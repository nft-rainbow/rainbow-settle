package redis

import (
	"context"
	"testing"

	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestCheckIsRich(t *testing.T) {
	assert.True(t, isRich(1055, enums.COST_TYPE_RAINBOW_NORMAL))
}

func TestConnectDB(t *testing.T) {
	_, err := DB().Get(context.Background(), "rich-1").Result()
	assert.NoError(t, err)
}
