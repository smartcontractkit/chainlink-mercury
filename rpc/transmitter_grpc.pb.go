// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: transmitter.proto

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Transmitter_Transmit_FullMethodName     = "/rpc.Transmitter/Transmit"
	Transmitter_LatestReport_FullMethodName = "/rpc.Transmitter/LatestReport"
)

// TransmitterClient is the client API for Transmitter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TransmitterClient interface {
	Transmit(ctx context.Context, in *TransmitRequest, opts ...grpc.CallOption) (*TransmitResponse, error)
	LatestReport(ctx context.Context, in *LatestReportRequest, opts ...grpc.CallOption) (*LatestReportResponse, error)
}

type transmitterClient struct {
	cc grpc.ClientConnInterface
}

func NewTransmitterClient(cc grpc.ClientConnInterface) TransmitterClient {
	return &transmitterClient{cc}
}

func (c *transmitterClient) Transmit(ctx context.Context, in *TransmitRequest, opts ...grpc.CallOption) (*TransmitResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TransmitResponse)
	err := c.cc.Invoke(ctx, Transmitter_Transmit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transmitterClient) LatestReport(ctx context.Context, in *LatestReportRequest, opts ...grpc.CallOption) (*LatestReportResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LatestReportResponse)
	err := c.cc.Invoke(ctx, Transmitter_LatestReport_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TransmitterServer is the server API for Transmitter service.
// All implementations must embed UnimplementedTransmitterServer
// for forward compatibility.
type TransmitterServer interface {
	Transmit(context.Context, *TransmitRequest) (*TransmitResponse, error)
	LatestReport(context.Context, *LatestReportRequest) (*LatestReportResponse, error)
	mustEmbedUnimplementedTransmitterServer()
}

// UnimplementedTransmitterServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTransmitterServer struct{}

func (UnimplementedTransmitterServer) Transmit(context.Context, *TransmitRequest) (*TransmitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transmit not implemented")
}
func (UnimplementedTransmitterServer) LatestReport(context.Context, *LatestReportRequest) (*LatestReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LatestReport not implemented")
}
func (UnimplementedTransmitterServer) mustEmbedUnimplementedTransmitterServer() {}
func (UnimplementedTransmitterServer) testEmbeddedByValue()                     {}

// UnsafeTransmitterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TransmitterServer will
// result in compilation errors.
type UnsafeTransmitterServer interface {
	mustEmbedUnimplementedTransmitterServer()
}

func RegisterTransmitterServer(s grpc.ServiceRegistrar, srv TransmitterServer) {
	// If the following call pancis, it indicates UnimplementedTransmitterServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Transmitter_ServiceDesc, srv)
}

func _Transmitter_Transmit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransmitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransmitterServer).Transmit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Transmitter_Transmit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransmitterServer).Transmit(ctx, req.(*TransmitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Transmitter_LatestReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LatestReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransmitterServer).LatestReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Transmitter_LatestReport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransmitterServer).LatestReport(ctx, req.(*LatestReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Transmitter_ServiceDesc is the grpc.ServiceDesc for Transmitter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Transmitter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Transmitter",
	HandlerType: (*TransmitterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Transmit",
			Handler:    _Transmitter_Transmit_Handler,
		},
		{
			MethodName: "LatestReport",
			Handler:    _Transmitter_LatestReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "transmitter.proto",
}
