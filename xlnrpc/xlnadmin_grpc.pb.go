// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: xlnadmin.proto

package xlnrpc

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

// XlnAdminClient is the client API for XlnAdmin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type XlnAdminClient interface {
	//
	//Get general info about XLN
	GetInfo(ctx context.Context, in *GetAdminInfoRequest, opts ...grpc.CallOption) (*GetAdminInfoResponse, error)
	CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserResponse, error)
	DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error)
	UpdateWallet(ctx context.Context, in *UpdateWalletRequest, opts ...grpc.CallOption) (*UpdateWalletResponse, error)
	ListUsers(ctx context.Context, in *ListUsersRequest, opts ...grpc.CallOption) (*ListUsersResponse, error)
	AdminDeleteWallet(ctx context.Context, in *AdminDeleteWalletRequest, opts ...grpc.CallOption) (*AdminDeleteWalletResponse, error)
	GetInvoice(ctx context.Context, in *GetInvoiceRequest, opts ...grpc.CallOption) (*GetInvoiceResponse, error)
	ListPendingInvoices(ctx context.Context, in *ListPendingInvoicesRequest, opts ...grpc.CallOption) (*ListPendingInvoicesResponse, error)
	ListPendingPayments(ctx context.Context, in *ListPendingPaymentsRequest, opts ...grpc.CallOption) (*ListPendingPaymentsResponse, error)
}

type xlnAdminClient struct {
	cc grpc.ClientConnInterface
}

func NewXlnAdminClient(cc grpc.ClientConnInterface) XlnAdminClient {
	return &xlnAdminClient{cc}
}

func (c *xlnAdminClient) GetInfo(ctx context.Context, in *GetAdminInfoRequest, opts ...grpc.CallOption) (*GetAdminInfoResponse, error) {
	out := new(GetAdminInfoResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/GetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) CreateUser(ctx context.Context, in *CreateUserRequest, opts ...grpc.CallOption) (*CreateUserResponse, error) {
	out := new(CreateUserResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/CreateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error) {
	out := new(DeleteUserResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/DeleteUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) UpdateWallet(ctx context.Context, in *UpdateWalletRequest, opts ...grpc.CallOption) (*UpdateWalletResponse, error) {
	out := new(UpdateWalletResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/UpdateWallet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) ListUsers(ctx context.Context, in *ListUsersRequest, opts ...grpc.CallOption) (*ListUsersResponse, error) {
	out := new(ListUsersResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/ListUsers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) AdminDeleteWallet(ctx context.Context, in *AdminDeleteWalletRequest, opts ...grpc.CallOption) (*AdminDeleteWalletResponse, error) {
	out := new(AdminDeleteWalletResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/AdminDeleteWallet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) GetInvoice(ctx context.Context, in *GetInvoiceRequest, opts ...grpc.CallOption) (*GetInvoiceResponse, error) {
	out := new(GetInvoiceResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/GetInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) ListPendingInvoices(ctx context.Context, in *ListPendingInvoicesRequest, opts ...grpc.CallOption) (*ListPendingInvoicesResponse, error) {
	out := new(ListPendingInvoicesResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/ListPendingInvoices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *xlnAdminClient) ListPendingPayments(ctx context.Context, in *ListPendingPaymentsRequest, opts ...grpc.CallOption) (*ListPendingPaymentsResponse, error) {
	out := new(ListPendingPaymentsResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.XlnAdmin/ListPendingPayments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// XlnAdminServer is the server API for XlnAdmin service.
// All implementations must embed UnimplementedXlnAdminServer
// for forward compatibility
type XlnAdminServer interface {
	//
	//Get general info about XLN
	GetInfo(context.Context, *GetAdminInfoRequest) (*GetAdminInfoResponse, error)
	CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error)
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error)
	UpdateWallet(context.Context, *UpdateWalletRequest) (*UpdateWalletResponse, error)
	ListUsers(context.Context, *ListUsersRequest) (*ListUsersResponse, error)
	AdminDeleteWallet(context.Context, *AdminDeleteWalletRequest) (*AdminDeleteWalletResponse, error)
	GetInvoice(context.Context, *GetInvoiceRequest) (*GetInvoiceResponse, error)
	ListPendingInvoices(context.Context, *ListPendingInvoicesRequest) (*ListPendingInvoicesResponse, error)
	ListPendingPayments(context.Context, *ListPendingPaymentsRequest) (*ListPendingPaymentsResponse, error)
	mustEmbedUnimplementedXlnAdminServer()
}

// UnimplementedXlnAdminServer must be embedded to have forward compatible implementations.
type UnimplementedXlnAdminServer struct {
}

func (UnimplementedXlnAdminServer) GetInfo(context.Context, *GetAdminInfoRequest) (*GetAdminInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedXlnAdminServer) CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}
func (UnimplementedXlnAdminServer) DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedXlnAdminServer) UpdateWallet(context.Context, *UpdateWalletRequest) (*UpdateWalletResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWallet not implemented")
}
func (UnimplementedXlnAdminServer) ListUsers(context.Context, *ListUsersRequest) (*ListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}
func (UnimplementedXlnAdminServer) AdminDeleteWallet(context.Context, *AdminDeleteWalletRequest) (*AdminDeleteWalletResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminDeleteWallet not implemented")
}
func (UnimplementedXlnAdminServer) GetInvoice(context.Context, *GetInvoiceRequest) (*GetInvoiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInvoice not implemented")
}
func (UnimplementedXlnAdminServer) ListPendingInvoices(context.Context, *ListPendingInvoicesRequest) (*ListPendingInvoicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPendingInvoices not implemented")
}
func (UnimplementedXlnAdminServer) ListPendingPayments(context.Context, *ListPendingPaymentsRequest) (*ListPendingPaymentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPendingPayments not implemented")
}
func (UnimplementedXlnAdminServer) mustEmbedUnimplementedXlnAdminServer() {}

// UnsafeXlnAdminServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to XlnAdminServer will
// result in compilation errors.
type UnsafeXlnAdminServer interface {
	mustEmbedUnimplementedXlnAdminServer()
}

func RegisterXlnAdminServer(s grpc.ServiceRegistrar, srv XlnAdminServer) {
	s.RegisterService(&XlnAdmin_ServiceDesc, srv)
}

func _XlnAdmin_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAdminInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).GetInfo(ctx, req.(*GetAdminInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_CreateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).CreateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/CreateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).CreateUser(ctx, req.(*CreateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/DeleteUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).DeleteUser(ctx, req.(*DeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_UpdateWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateWalletRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).UpdateWallet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/UpdateWallet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).UpdateWallet(ctx, req.(*UpdateWalletRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_ListUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUsersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).ListUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/ListUsers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).ListUsers(ctx, req.(*ListUsersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_AdminDeleteWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminDeleteWalletRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).AdminDeleteWallet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/AdminDeleteWallet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).AdminDeleteWallet(ctx, req.(*AdminDeleteWalletRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_GetInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInvoiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).GetInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/GetInvoice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).GetInvoice(ctx, req.(*GetInvoiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_ListPendingInvoices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPendingInvoicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).ListPendingInvoices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/ListPendingInvoices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).ListPendingInvoices(ctx, req.(*ListPendingInvoicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _XlnAdmin_ListPendingPayments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPendingPaymentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(XlnAdminServer).ListPendingPayments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.XlnAdmin/ListPendingPayments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(XlnAdminServer).ListPendingPayments(ctx, req.(*ListPendingPaymentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// XlnAdmin_ServiceDesc is the grpc.ServiceDesc for XlnAdmin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var XlnAdmin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "xlnrpc.XlnAdmin",
	HandlerType: (*XlnAdminServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetInfo",
			Handler:    _XlnAdmin_GetInfo_Handler,
		},
		{
			MethodName: "CreateUser",
			Handler:    _XlnAdmin_CreateUser_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _XlnAdmin_DeleteUser_Handler,
		},
		{
			MethodName: "UpdateWallet",
			Handler:    _XlnAdmin_UpdateWallet_Handler,
		},
		{
			MethodName: "ListUsers",
			Handler:    _XlnAdmin_ListUsers_Handler,
		},
		{
			MethodName: "AdminDeleteWallet",
			Handler:    _XlnAdmin_AdminDeleteWallet_Handler,
		},
		{
			MethodName: "GetInvoice",
			Handler:    _XlnAdmin_GetInvoice_Handler,
		},
		{
			MethodName: "ListPendingInvoices",
			Handler:    _XlnAdmin_ListPendingInvoices_Handler,
		},
		{
			MethodName: "ListPendingPayments",
			Handler:    _XlnAdmin_ListPendingPayments_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "xlnadmin.proto",
}
