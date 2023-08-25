package server

import (
	"testing"

	"github.com/nft-rainbow/rainbow-settle/proto"
)

func TestSettle(t *testing.T) {
	var _ proto.SettleServer = &SettleServer{}
}
