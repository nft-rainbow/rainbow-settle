package server

import (
	"context"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/proto"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type SettleServer struct {
	proto.UnimplementedSettleServer
}

func (s *SettleServer) Deposite(ctx context.Context, in *proto.DepositRequest) (*proto.WxOrder, error) {
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

func (s *SettleServer) GetWxOrder(ctx context.Context, in *proto.WxOrderRequest) (*proto.WxOrder, error) {
	return nil, nil
}

func (s *SettleServer) BuyGas(ctx context.Context, in *proto.BuySponsorRequest) (*proto.Empty, error) {
	price, err := models.GetUserCfxPrice(uint(in.UserId))
	if err != nil {
		return nil, err
	}
	fl, err := services.BuyGas(uint(in.UserId), decimal.NewFromFloat32(in.Amount), uint(in.TxId), in.Address, price)
	if err != nil {
		return nil, err
	}
	logrus.WithField("fiatlog", fl).Info("buy gas completed")
	return &proto.Empty{}, nil
}

func (s *SettleServer) BuyStorage(ctx context.Context, in *proto.BuySponsorRequest) (*proto.Empty, error) {
	price, err := models.GetUserCfxPrice(uint(in.UserId))
	if err != nil {
		return nil, err
	}
	fl, err := services.BuyStorage(uint(in.UserId), decimal.NewFromFloat32(in.Amount), uint(in.TxId), in.Address, price)
	if err != nil {
		return nil, err
	}
	logrus.WithField("fiatlog", fl).Info("buy storage completed")
	return &proto.Empty{}, nil
}

func (s *SettleServer) RefundSponsor(ctx context.Context, in *proto.RefundSponsorRequest) (*proto.Empty, error) {

	fl, err := services.RefundSponsor(uint(in.UserId), decimal.NewFromFloat32(in.Amount), uint(in.SponsorFiatlogId), models.FiatLogType(in.FiatlogType), uint(in.TxId))
	if err != nil {
		return nil, err
	}
	logrus.WithField("fiatlog", fl).Info("refund sponsor completed")
	return &proto.Empty{}, nil
}

// 根据settle堆栈判断退quota还是balance
func (s *SettleServer) RefundApiFee(ctx context.Context, in *proto.RefundApiFeeRequest) (*proto.Empty, error) {
	costType, err := enums.ParseCostType(in.CostType)
	if err != nil {
		return nil, err
	}

	err = services.RefundApiCost(uint(in.UserId), *costType, int(in.Count))
	if err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

func (s *SettleServer) GetUserBalance(ctx context.Context, in *proto.UserID) (*proto.UserBalance, error) {
	ub, err := models.GetUserBalance(uint(in.UserId))
	if err != nil {
		return nil, err
	}
	return &proto.UserBalance{
		UserId: uint32(ub.UserId),
	}, nil
}

func (s *SettleServer) GetUserApiQuota(ctx context.Context, in *proto.UserID) (*proto.UserApiQuotas, error) {
	_uqs, err := services.GetUserQuotaOperator().GetUserQuotas(uint(in.UserId))
	if err != nil {
		return nil, err
	}

	uqs := &proto.UserApiQuotas{}
	uqs.Items = make(map[string]*proto.UserApiQuota)
	for _, _u := range _uqs {

		u := proto.UserApiQuota{
			UserID:        uint32(_u.UserId),
			CostType:      _u.CostType.String(),
			CountReset:    uint32(_u.CountReset),
			CountRollover: uint32(_u.CountRollover),
		}
		uqs.Items[_u.CostType.String()] = &u
	}
	return uqs, nil
}
