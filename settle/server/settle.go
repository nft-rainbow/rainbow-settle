package server

import (
	"context"

	"github.com/nft-rainbow/rainbow-fiat/settle/proto"
	"github.com/nft-rainbow/rainbow-fiat/settle/services"
	"google.golang.org/grpc"
)

// type SettleClient interface {
// 	Deposite(ctx context.Context, in *proto.DepositRequest, opts ...grpc.CallOption) (*proto.WxOrder, error)
// 	GetWxOrder(ctx context.Context, in *proto.WxOrderRequest, opts ...grpc.CallOption) (*proto.WxOrder, error)
// 	RefundQuota(ctx context.Context, in *proto.RefundQuotaRequest, opts ...grpc.CallOption) (*proto.UserBalance, error)
// 	GetUserBalance(ctx context.Context, in *proto.UserID, opts ...grpc.CallOption) (*proto.UserBalance, error)
// }

type SettleServer struct {
}

func (s *SettleServer) Deposite(ctx context.Context, in *proto.DepositRequest, opts ...grpc.CallOption) (*proto.WxOrder, error) {
	order, err := services.CreateWechatOrder(uint(in.UserId), int32(in.Amount), in.Description)
	if err != nil {
		return nil, err
	}
	return &proto.WxOrder{
		CodeUrl:       order.CodeUrl,
		H5Url:         order.H5Url,
		TradeNo:       order.TradeNo,
		TradeProvider: order.TradeProvider,
		TradeState:    order.TradeState,
	}, nil
}

func (s *SettleServer) GetWxOrder(ctx context.Context, in *proto.WxOrderRequest, opts ...grpc.CallOption) (*proto.WxOrder, error) {
	return nil, nil
}

func (s *SettleServer) RefundQuota(ctx context.Context, in *proto.RefundQuotaRequest, opts ...grpc.CallOption) (*proto.UserBalance, error) {

}

func (s *SettleServer) GetUserBalance(ctx context.Context, in *proto.UserID, opts ...grpc.CallOption) (*proto.UserBalance, error) {

}
