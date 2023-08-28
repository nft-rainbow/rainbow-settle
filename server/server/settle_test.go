package server

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/influxdata/influxdb/pkg/testing/assert"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSettleImpleteInterface(t *testing.T) {
	var _ proto.SettleServer = &SettleServer{}
}

func _TestSettle(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", "localhost", 8090) //"localhost:8090"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	settleClient := proto.NewSettleClient(conn)

	balance, err := settleClient.GetUserBalance(context.Background(), &proto.UserID{UserId: 1})
	assert.NoError(t, err)
	fmt.Println(balance)

	_, err = settleClient.BuyGas(context.Background(), &proto.BuySponsorRequest{UserId: 1, Amount: 1, TxId: 10})
	assert.NoError(t, err)

	_, err = settleClient.BuyStorage(context.Background(), &proto.BuySponsorRequest{UserId: 1, Amount: 1, TxId: 0})
	assert.NoError(t, err)

	_, err = settleClient.Deposite(context.Background(), &proto.DepositRequest{UserId: 1, Amount: 1})
	assert.NoError(t, err)

	_, err = settleClient.RefundApiFee(context.Background(), &proto.RefundApiFeeRequest{UserId: 1, CostType: enums.COST_TYPE_RAINBOW_NORMAL.String(), Count: 1})
	assert.NoError(t, err)

	_, err = settleClient.RefundSponsor(context.Background(), &proto.RefundSponsorRequest{UserId: 1, TxId: 10})
	assert.NoError(t, err)

}
