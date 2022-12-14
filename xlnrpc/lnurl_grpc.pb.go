// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: lnurl.proto

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

// LNURLClient is the client API for LNURL service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LNURLClient interface {
	// LUD-04: auth
	Auth(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*LNURLResponse, error)
	RequestWithdraw(ctx context.Context, in *RequestWithdrawRequest, opts ...grpc.CallOption) (*RequestWithdrawResponse, error)
	Withdraw(ctx context.Context, in *WithdrawRequest, opts ...grpc.CallOption) (*LNURLResponse, error)
}

type lNURLClient struct {
	cc grpc.ClientConnInterface
}

func NewLNURLClient(cc grpc.ClientConnInterface) LNURLClient {
	return &lNURLClient{cc}
}

func (c *lNURLClient) Auth(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*LNURLResponse, error) {
	out := new(LNURLResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.LNURL/Auth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lNURLClient) RequestWithdraw(ctx context.Context, in *RequestWithdrawRequest, opts ...grpc.CallOption) (*RequestWithdrawResponse, error) {
	out := new(RequestWithdrawResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.LNURL/RequestWithdraw", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lNURLClient) Withdraw(ctx context.Context, in *WithdrawRequest, opts ...grpc.CallOption) (*LNURLResponse, error) {
	out := new(LNURLResponse)
	err := c.cc.Invoke(ctx, "/xlnrpc.LNURL/Withdraw", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LNURLServer is the server API for LNURL service.
// All implementations must embed UnimplementedLNURLServer
// for forward compatibility
type LNURLServer interface {
	// LUD-04: auth
	Auth(context.Context, *AuthRequest) (*LNURLResponse, error)
	RequestWithdraw(context.Context, *RequestWithdrawRequest) (*RequestWithdrawResponse, error)
	Withdraw(context.Context, *WithdrawRequest) (*LNURLResponse, error)
	mustEmbedUnimplementedLNURLServer()
}

// UnimplementedLNURLServer must be embedded to have forward compatible implementations.
type UnimplementedLNURLServer struct {
}

func (UnimplementedLNURLServer) Auth(context.Context, *AuthRequest) (*LNURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Auth not implemented")
}
func (UnimplementedLNURLServer) RequestWithdraw(context.Context, *RequestWithdrawRequest) (*RequestWithdrawResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestWithdraw not implemented")
}
func (UnimplementedLNURLServer) Withdraw(context.Context, *WithdrawRequest) (*LNURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Withdraw not implemented")
}
func (UnimplementedLNURLServer) mustEmbedUnimplementedLNURLServer() {}

// UnsafeLNURLServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LNURLServer will
// result in compilation errors.
type UnsafeLNURLServer interface {
	mustEmbedUnimplementedLNURLServer()
}

func RegisterLNURLServer(s grpc.ServiceRegistrar, srv LNURLServer) {
	s.RegisterService(&LNURL_ServiceDesc, srv)
}

func _LNURL_Auth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LNURLServer).Auth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.LNURL/Auth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LNURLServer).Auth(ctx, req.(*AuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LNURL_RequestWithdraw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestWithdrawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LNURLServer).RequestWithdraw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.LNURL/RequestWithdraw",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LNURLServer).RequestWithdraw(ctx, req.(*RequestWithdrawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LNURL_Withdraw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WithdrawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LNURLServer).Withdraw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlnrpc.LNURL/Withdraw",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LNURLServer).Withdraw(ctx, req.(*WithdrawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LNURL_ServiceDesc is the grpc.ServiceDesc for LNURL service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LNURL_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "xlnrpc.LNURL",
	HandlerType: (*LNURLServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Auth",
			Handler:    _LNURL_Auth_Handler,
		},
		{
			MethodName: "RequestWithdraw",
			Handler:    _LNURL_RequestWithdraw_Handler,
		},
		{
			MethodName: "Withdraw",
			Handler:    _LNURL_Withdraw_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lnurl.proto",
}
