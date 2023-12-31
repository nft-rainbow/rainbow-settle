// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.23.4
// source: settle.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SettleClient is the client API for Settle service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SettleClient interface {
	Deposite(ctx context.Context, in *DepositRequest, opts ...grpc.CallOption) (*WxOrder, error)
	Withdraw(ctx context.Context, in *WithdrawRequest, opts ...grpc.CallOption) (*UserBalance, error)
	GetDepositeOrder(ctx context.Context, in *ID, opts ...grpc.CallOption) (*DepositOrder, error)
	BuyGas(ctx context.Context, in *BuySponsorRequest, opts ...grpc.CallOption) (*Empty, error)
	BuyStorage(ctx context.Context, in *BuySponsorRequest, opts ...grpc.CallOption) (*Empty, error)
	BuyDataBundle(ctx context.Context, in *BuyDataBundleRequest, opts ...grpc.CallOption) (*UserDataBundle, error)
	BuyBillPlan(ctx context.Context, in *BuyBillPlanRequest, opts ...grpc.CallOption) (*UerBillPlan, error)
	UpdateBillPlanRenew(ctx context.Context, in *UpdateUpdateBillPlanRenewRequest, opts ...grpc.CallOption) (*UerBillPlan, error)
	RefundSponsor(ctx context.Context, in *RefundSponsorRequest, opts ...grpc.CallOption) (*Empty, error)
	RefundApiFee(ctx context.Context, in *RefundApiFeeRequest, opts ...grpc.CallOption) (*Empty, error)
	GetUserBalance(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*UserBalance, error)
	// rpc GetUserApiQuota(UserID) returns (UserApiQuotas);
	UserCreated(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*Empty, error)
	ApikeyUpdated(ctx context.Context, in *ApiKeyUpdated, opts ...grpc.CallOption) (*Empty, error)
	CreateCmcDepositNo(ctx context.Context, in *CreateCmcDepositNoReqeust, opts ...grpc.CallOption) (*Empty, error)
	GetCmcDepositNo(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*CmbDepositNo, error)
	QueryRecentCmbHistory(ctx context.Context, in *Pagenation, opts ...grpc.CallOption) (*QueryRecentCmbHistoryResponse, error)
	UpdateCmcDepositNoRelation(ctx context.Context, in *UpdateCmcDepositNoRelationRequest, opts ...grpc.CallOption) (*Empty, error)
}

type settleClient struct {
	cc grpc.ClientConnInterface
}

func NewSettleClient(cc grpc.ClientConnInterface) SettleClient {
	return &settleClient{cc}
}

func (c *settleClient) Deposite(ctx context.Context, in *DepositRequest, opts ...grpc.CallOption) (*WxOrder, error) {
	out := new(WxOrder)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/Deposite", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) Withdraw(ctx context.Context, in *WithdrawRequest, opts ...grpc.CallOption) (*UserBalance, error) {
	out := new(UserBalance)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/Withdraw", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) GetDepositeOrder(ctx context.Context, in *ID, opts ...grpc.CallOption) (*DepositOrder, error) {
	out := new(DepositOrder)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/GetDepositeOrder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) BuyGas(ctx context.Context, in *BuySponsorRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/BuyGas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) BuyStorage(ctx context.Context, in *BuySponsorRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/BuyStorage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) BuyDataBundle(ctx context.Context, in *BuyDataBundleRequest, opts ...grpc.CallOption) (*UserDataBundle, error) {
	out := new(UserDataBundle)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/BuyDataBundle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) BuyBillPlan(ctx context.Context, in *BuyBillPlanRequest, opts ...grpc.CallOption) (*UerBillPlan, error) {
	out := new(UerBillPlan)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/BuyBillPlan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) UpdateBillPlanRenew(ctx context.Context, in *UpdateUpdateBillPlanRenewRequest, opts ...grpc.CallOption) (*UerBillPlan, error) {
	out := new(UerBillPlan)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/UpdateBillPlanRenew", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) RefundSponsor(ctx context.Context, in *RefundSponsorRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/RefundSponsor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) RefundApiFee(ctx context.Context, in *RefundApiFeeRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/RefundApiFee", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) GetUserBalance(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*UserBalance, error) {
	out := new(UserBalance)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/GetUserBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) UserCreated(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/UserCreated", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) ApikeyUpdated(ctx context.Context, in *ApiKeyUpdated, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/ApikeyUpdated", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) CreateCmcDepositNo(ctx context.Context, in *CreateCmcDepositNoReqeust, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/CreateCmcDepositNo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) GetCmcDepositNo(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*CmbDepositNo, error) {
	out := new(CmbDepositNo)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/GetCmcDepositNo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) QueryRecentCmbHistory(ctx context.Context, in *Pagenation, opts ...grpc.CallOption) (*QueryRecentCmbHistoryResponse, error) {
	out := new(QueryRecentCmbHistoryResponse)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/QueryRecentCmbHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settleClient) UpdateCmcDepositNoRelation(ctx context.Context, in *UpdateCmcDepositNoRelationRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/rainbowsettle.Settle/UpdateCmcDepositNoRelation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SettleServer is the server API for Settle service.
// All implementations must embed UnimplementedSettleServer
// for forward compatibility
type SettleServer interface {
	Deposite(context.Context, *DepositRequest) (*WxOrder, error)
	Withdraw(context.Context, *WithdrawRequest) (*UserBalance, error)
	GetDepositeOrder(context.Context, *ID) (*DepositOrder, error)
	BuyGas(context.Context, *BuySponsorRequest) (*Empty, error)
	BuyStorage(context.Context, *BuySponsorRequest) (*Empty, error)
	BuyDataBundle(context.Context, *BuyDataBundleRequest) (*UserDataBundle, error)
	BuyBillPlan(context.Context, *BuyBillPlanRequest) (*UerBillPlan, error)
	UpdateBillPlanRenew(context.Context, *UpdateUpdateBillPlanRenewRequest) (*UerBillPlan, error)
	RefundSponsor(context.Context, *RefundSponsorRequest) (*Empty, error)
	RefundApiFee(context.Context, *RefundApiFeeRequest) (*Empty, error)
	GetUserBalance(context.Context, *UserID) (*UserBalance, error)
	// rpc GetUserApiQuota(UserID) returns (UserApiQuotas);
	UserCreated(context.Context, *UserID) (*Empty, error)
	ApikeyUpdated(context.Context, *ApiKeyUpdated) (*Empty, error)
	CreateCmcDepositNo(context.Context, *CreateCmcDepositNoReqeust) (*Empty, error)
	GetCmcDepositNo(context.Context, *UserID) (*CmbDepositNo, error)
	QueryRecentCmbHistory(context.Context, *Pagenation) (*QueryRecentCmbHistoryResponse, error)
	UpdateCmcDepositNoRelation(context.Context, *UpdateCmcDepositNoRelationRequest) (*Empty, error)
	mustEmbedUnimplementedSettleServer()
}

// UnimplementedSettleServer must be embedded to have forward compatible implementations.
type UnimplementedSettleServer struct {
}

func (UnimplementedSettleServer) Deposite(context.Context, *DepositRequest) (*WxOrder, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Deposite not implemented")
}
func (UnimplementedSettleServer) Withdraw(context.Context, *WithdrawRequest) (*UserBalance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Withdraw not implemented")
}
func (UnimplementedSettleServer) GetDepositeOrder(context.Context, *ID) (*DepositOrder, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDepositeOrder not implemented")
}
func (UnimplementedSettleServer) BuyGas(context.Context, *BuySponsorRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuyGas not implemented")
}
func (UnimplementedSettleServer) BuyStorage(context.Context, *BuySponsorRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuyStorage not implemented")
}
func (UnimplementedSettleServer) BuyDataBundle(context.Context, *BuyDataBundleRequest) (*UserDataBundle, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuyDataBundle not implemented")
}
func (UnimplementedSettleServer) BuyBillPlan(context.Context, *BuyBillPlanRequest) (*UerBillPlan, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuyBillPlan not implemented")
}
func (UnimplementedSettleServer) UpdateBillPlanRenew(context.Context, *UpdateUpdateBillPlanRenewRequest) (*UerBillPlan, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateBillPlanRenew not implemented")
}
func (UnimplementedSettleServer) RefundSponsor(context.Context, *RefundSponsorRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefundSponsor not implemented")
}
func (UnimplementedSettleServer) RefundApiFee(context.Context, *RefundApiFeeRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefundApiFee not implemented")
}
func (UnimplementedSettleServer) GetUserBalance(context.Context, *UserID) (*UserBalance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserBalance not implemented")
}
func (UnimplementedSettleServer) UserCreated(context.Context, *UserID) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserCreated not implemented")
}
func (UnimplementedSettleServer) ApikeyUpdated(context.Context, *ApiKeyUpdated) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApikeyUpdated not implemented")
}
func (UnimplementedSettleServer) CreateCmcDepositNo(context.Context, *CreateCmcDepositNoReqeust) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCmcDepositNo not implemented")
}
func (UnimplementedSettleServer) GetCmcDepositNo(context.Context, *UserID) (*CmbDepositNo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCmcDepositNo not implemented")
}
func (UnimplementedSettleServer) QueryRecentCmbHistory(context.Context, *Pagenation) (*QueryRecentCmbHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryRecentCmbHistory not implemented")
}
func (UnimplementedSettleServer) UpdateCmcDepositNoRelation(context.Context, *UpdateCmcDepositNoRelationRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCmcDepositNoRelation not implemented")
}
func (UnimplementedSettleServer) mustEmbedUnimplementedSettleServer() {}

// UnsafeSettleServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SettleServer will
// result in compilation errors.
type UnsafeSettleServer interface {
	mustEmbedUnimplementedSettleServer()
}

func RegisterSettleServer(s grpc.ServiceRegistrar, srv SettleServer) {
	s.RegisterService(&Settle_ServiceDesc, srv)
}

func _Settle_Deposite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DepositRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).Deposite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/Deposite",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).Deposite(ctx, req.(*DepositRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_Withdraw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WithdrawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).Withdraw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/Withdraw",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).Withdraw(ctx, req.(*WithdrawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_GetDepositeOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).GetDepositeOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/GetDepositeOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).GetDepositeOrder(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_BuyGas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuySponsorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).BuyGas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/BuyGas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).BuyGas(ctx, req.(*BuySponsorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_BuyStorage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuySponsorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).BuyStorage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/BuyStorage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).BuyStorage(ctx, req.(*BuySponsorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_BuyDataBundle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuyDataBundleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).BuyDataBundle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/BuyDataBundle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).BuyDataBundle(ctx, req.(*BuyDataBundleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_BuyBillPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuyBillPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).BuyBillPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/BuyBillPlan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).BuyBillPlan(ctx, req.(*BuyBillPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_UpdateBillPlanRenew_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUpdateBillPlanRenewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).UpdateBillPlanRenew(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/UpdateBillPlanRenew",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).UpdateBillPlanRenew(ctx, req.(*UpdateUpdateBillPlanRenewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_RefundSponsor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefundSponsorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).RefundSponsor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/RefundSponsor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).RefundSponsor(ctx, req.(*RefundSponsorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_RefundApiFee_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefundApiFeeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).RefundApiFee(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/RefundApiFee",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).RefundApiFee(ctx, req.(*RefundApiFeeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_GetUserBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).GetUserBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/GetUserBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).GetUserBalance(ctx, req.(*UserID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_UserCreated_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).UserCreated(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/UserCreated",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).UserCreated(ctx, req.(*UserID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_ApikeyUpdated_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApiKeyUpdated)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).ApikeyUpdated(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/ApikeyUpdated",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).ApikeyUpdated(ctx, req.(*ApiKeyUpdated))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_CreateCmcDepositNo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCmcDepositNoReqeust)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).CreateCmcDepositNo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/CreateCmcDepositNo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).CreateCmcDepositNo(ctx, req.(*CreateCmcDepositNoReqeust))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_GetCmcDepositNo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).GetCmcDepositNo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/GetCmcDepositNo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).GetCmcDepositNo(ctx, req.(*UserID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_QueryRecentCmbHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Pagenation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).QueryRecentCmbHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/QueryRecentCmbHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).QueryRecentCmbHistory(ctx, req.(*Pagenation))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settle_UpdateCmcDepositNoRelation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCmcDepositNoRelationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettleServer).UpdateCmcDepositNoRelation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rainbowsettle.Settle/UpdateCmcDepositNoRelation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettleServer).UpdateCmcDepositNoRelation(ctx, req.(*UpdateCmcDepositNoRelationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Settle_ServiceDesc is the grpc.ServiceDesc for Settle service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Settle_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rainbowsettle.Settle",
	HandlerType: (*SettleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Deposite",
			Handler:    _Settle_Deposite_Handler,
		},
		{
			MethodName: "Withdraw",
			Handler:    _Settle_Withdraw_Handler,
		},
		{
			MethodName: "GetDepositeOrder",
			Handler:    _Settle_GetDepositeOrder_Handler,
		},
		{
			MethodName: "BuyGas",
			Handler:    _Settle_BuyGas_Handler,
		},
		{
			MethodName: "BuyStorage",
			Handler:    _Settle_BuyStorage_Handler,
		},
		{
			MethodName: "BuyDataBundle",
			Handler:    _Settle_BuyDataBundle_Handler,
		},
		{
			MethodName: "BuyBillPlan",
			Handler:    _Settle_BuyBillPlan_Handler,
		},
		{
			MethodName: "UpdateBillPlanRenew",
			Handler:    _Settle_UpdateBillPlanRenew_Handler,
		},
		{
			MethodName: "RefundSponsor",
			Handler:    _Settle_RefundSponsor_Handler,
		},
		{
			MethodName: "RefundApiFee",
			Handler:    _Settle_RefundApiFee_Handler,
		},
		{
			MethodName: "GetUserBalance",
			Handler:    _Settle_GetUserBalance_Handler,
		},
		{
			MethodName: "UserCreated",
			Handler:    _Settle_UserCreated_Handler,
		},
		{
			MethodName: "ApikeyUpdated",
			Handler:    _Settle_ApikeyUpdated_Handler,
		},
		{
			MethodName: "CreateCmcDepositNo",
			Handler:    _Settle_CreateCmcDepositNo_Handler,
		},
		{
			MethodName: "GetCmcDepositNo",
			Handler:    _Settle_GetCmcDepositNo_Handler,
		},
		{
			MethodName: "QueryRecentCmbHistory",
			Handler:    _Settle_QueryRecentCmbHistory_Handler,
		},
		{
			MethodName: "UpdateCmcDepositNoRelation",
			Handler:    _Settle_UpdateCmcDepositNoRelation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "settle.proto",
}
