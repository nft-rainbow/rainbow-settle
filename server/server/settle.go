package server

import (
	"context"

	"github.com/nft-rainbow/rainbow-settle/common/models"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/nft-rainbow/rainbow-settle/proto"
	"github.com/nft-rainbow/rainbow-settle/server/services"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	confluxpay "github.com/web3-identity/conflux-pay-sdk-go"
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

func (s *SettleServer) GetDepositeOrder(ctx context.Context, in *proto.ID) (*proto.DepositOrder, error) {
	o, err := models.FindDepositOrderById(uint(in.Id))
	if err != nil {
		return nil, err
	}
	return &proto.DepositOrder{
		ID:          uint32(o.ID),
		UserId:      uint32(o.UserId),
		Amount:      float32(o.Amount.InexactFloat64()),
		Type:        uint32(o.Type),
		Status:      uint32(o.Status),
		Description: o.Description,
		TradeNo:     o.TradeNo,
		Meta:        string(o.Meta),
	}, nil
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
	fiatlog, err := models.FindSponsorFiatlogByTxid(uint(in.TxId))
	if err != nil {
		return nil, err
	}

	fl, err := services.RefundSponsor(uint(in.UserId), fiatlog.Amount, fiatlog.ID, fiatlog.Type, uint(in.TxId))
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

func (s *SettleServer) CreateCmcDepositNo(ctx context.Context, in *proto.CreateCmcDepositNoReqeust) (*proto.Empty, error) {
	if err := services.CreateCmcDepositNo(uint(in.UserId), parseCmbDepositNo(in.Info)); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

func (s *SettleServer) GetCmcDepositNo(ctx context.Context, in *proto.UserID) (*proto.CmbDepositNo, error) {
	result, err := (&models.CmbDepositNoOperator{}).FindByUserId(uint(in.UserId))
	if err != nil {
		return nil, err
	}
	return convertCmbDepositNo(result), nil
}

func (s *SettleServer) QueryRecentCmbHistory(ctx context.Context, in *proto.Pagenation) (*proto.QueryRecentCmbHistoryResponse, error) {
	resp, err := services.QueryRecentCmbHistory(int32(in.Limit), int32(in.Offset))
	if err != nil {
		return nil, err
	}
	return convertCmbHistory(resp), nil
}

func (s *SettleServer) UpdateCmcDepositNoRelation(ctx context.Context, in *proto.UpdateCmcDepositNoRelationRequest) (*proto.Empty, error) {
	if err := services.UpdateCmcDepositNoRelation(uint(in.UserId), parseCmbDepositNo(in.Info)); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

func convertCmbDepositNo(in *models.CmbDepositNo) *proto.CmbDepositNo {
	return &proto.CmbDepositNo{
		ID:           uint32(in.ID),
		UserId:       uint32(in.UserId),
		UserName:     in.UserName,
		UserBankNo:   in.UserBankNo,
		UserBankName: in.UserBankName,
		CmbNo:        in.CmbNo,
	}
}

func convertCmbHistory(records []confluxpay.ModelsCmbRecord) *proto.QueryRecentCmbHistoryResponse {
	var result proto.QueryRecentCmbHistoryResponse

	result.List = lo.Map(records, func(r confluxpay.ModelsCmbRecord, index int) *proto.ModelsCmbRecord {
		return &proto.ModelsCmbRecord{
			AccNbr:    r.AccNbr,
			AutFlg:    r.AutFlg,
			CcyNbr:    r.CcyNbr,
			CreatedAt: r.CreatedAt,
			DmaNam:    r.DmaNam,
			DmaNbr:    r.DmaNbr,
			Id:        r.Id,
			NarInn:    r.NarInn,
			RpyAcc:    r.RpyAcc,
			RpyNam:    r.RpyNam,
			TrxAmt:    r.TrxAmt,
			TrxDat:    r.TrxDat,
			TrxDir:    r.TrxDir,
			TrxNbr:    r.TrxNbr,
			TrxTim:    r.TrxTim,
			TrxTxt:    r.TrxTxt,
		}
	})
	return &result
}

func parseCmbDepositNo(info *proto.CmbDepositNoDto) services.CmbDepositNoDto {
	return services.CmbDepositNoDto{
		Name:   info.Name,
		Bank:   info.Bank,
		CardNo: info.CardNo,
	}
}
