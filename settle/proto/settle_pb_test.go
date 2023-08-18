package proto

import (
	"context"
	"log"
	"testing"

	"github.com/influxdata/influxdb/pkg/testing/assert"
	"github.com/nft-rainbow/rainbow-fiat/common/models/enums"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRefund(t *testing.T) {
	addr := "localhost:8090"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewSettleClient(conn)
	_, err = c.RefundCost(context.Background(), &RefundCostRequest{
		UserId:   1,
		CostType: enums.COST_TYPE_RAINBOW_MINT.String(),
		Count:    1,
	})
	assert.NoError(t, err)
}
