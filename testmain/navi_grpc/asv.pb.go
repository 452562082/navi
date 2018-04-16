/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	proto/grpc.proto

It has these top-level messages:
	PingRequest
	PingResponse
	ServiceNameRequest
	ServiceNameResponse
	ServiceModeRequest
	ServiceModeResponse
*/
package navi_grpc

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type PingRequest struct {
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto1.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type PingResponse struct {
	Pong string `protobuf:"bytes,1,opt,name=pong" json:"pong,omitempty"`
}

func (m *PingResponse) Reset()                    { *m = PingResponse{} }
func (m *PingResponse) String() string            { return proto1.CompactTextString(m) }
func (*PingResponse) ProtoMessage()               {}
func (*PingResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PingResponse) GetPong() string {
	if m != nil {
		return m.Pong
	}
	return ""
}

type ServiceNameRequest struct {
}

func (m *ServiceNameRequest) Reset()                    { *m = ServiceNameRequest{} }
func (m *ServiceNameRequest) String() string            { return proto1.CompactTextString(m) }
func (*ServiceNameRequest) ProtoMessage()               {}
func (*ServiceNameRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type ServiceNameResponse struct {
	ServiceName string `protobuf:"bytes,1,opt,name=service_name,json=serviceName" json:"service_name,omitempty"`
}

func (m *ServiceNameResponse) Reset()                    { *m = ServiceNameResponse{} }
func (m *ServiceNameResponse) String() string            { return proto1.CompactTextString(m) }
func (*ServiceNameResponse) ProtoMessage()               {}
func (*ServiceNameResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ServiceNameResponse) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

type ServiceModeRequest struct {
}

func (m *ServiceModeRequest) Reset()                    { *m = ServiceModeRequest{} }
func (m *ServiceModeRequest) String() string            { return proto1.CompactTextString(m) }
func (*ServiceModeRequest) ProtoMessage()               {}
func (*ServiceModeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ServiceModeResponse struct {
	ServiceMode string `protobuf:"bytes,1,opt,name=service_mode,json=serviceMode" json:"service_mode,omitempty"`
}

func (m *ServiceModeResponse) Reset()                    { *m = ServiceModeResponse{} }
func (m *ServiceModeResponse) String() string            { return proto1.CompactTextString(m) }
func (*ServiceModeResponse) ProtoMessage()               {}
func (*ServiceModeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ServiceModeResponse) GetServiceMode() string {
	if m != nil {
		return m.ServiceMode
	}
	return ""
}

func init() {
	proto1.RegisterType((*PingRequest)(nil), "proto.PingRequest")
	proto1.RegisterType((*PingResponse)(nil), "proto.PingResponse")
	proto1.RegisterType((*ServiceNameRequest)(nil), "proto.ServiceNameRequest")
	proto1.RegisterType((*ServiceNameResponse)(nil), "proto.ServiceNameResponse")
	proto1.RegisterType((*ServiceModeRequest)(nil), "proto.ServiceModeRequest")
	proto1.RegisterType((*ServiceModeResponse)(nil), "proto.ServiceModeResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Grpc service

type GrpcClient interface {
	// rpc server必须实现的接口，返回字符串 "pong" 即可
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// rpc server必须实现的接口，返回服务名称，为首字母大写的驼峰格式，例如 "GrpcService"
	ServiceName(ctx context.Context, in *ServiceNameRequest, opts ...grpc.CallOption) (*ServiceNameResponse, error)
	// rpc server必须实现的接口，说明该server是以什么模式运行，分为dev和prod；dev为开发版本，prod为生产版本
	ServiceMode(ctx context.Context, in *ServiceModeRequest, opts ...grpc.CallOption) (*ServiceModeResponse, error)
}

type grpcClient struct {
	cc *grpc.ClientConn
}

func NewGrpcClient(cc *grpc.ClientConn) GrpcClient {
	return &grpcClient{cc}
}

func (c *grpcClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := grpc.Invoke(ctx, "/proto.grpc/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcClient) ServiceName(ctx context.Context, in *ServiceNameRequest, opts ...grpc.CallOption) (*ServiceNameResponse, error) {
	out := new(ServiceNameResponse)
	err := grpc.Invoke(ctx, "/proto.grpc/ServiceName", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcClient) ServiceMode(ctx context.Context, in *ServiceModeRequest, opts ...grpc.CallOption) (*ServiceModeResponse, error) {
	out := new(ServiceModeResponse)
	err := grpc.Invoke(ctx, "/proto.grpc/ServiceMode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Grpc service

type GrpcServer interface {
	// rpc server必须实现的接口，返回字符串 "pong" 即可
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// rpc server必须实现的接口，返回服务名称，为首字母大写的驼峰格式，例如 "GrpcService"
	ServiceName(context.Context, *ServiceNameRequest) (*ServiceNameResponse, error)
	// rpc server必须实现的接口，说明该server是以什么模式运行，分为dev和prod；dev为开发版本，prod为生产版本
	ServiceMode(context.Context, *ServiceModeRequest) (*ServiceModeResponse, error)
}

func RegisterGrpcServer(s *grpc.Server, srv GrpcServer) {
	s.RegisterService(&_Grpc_serviceDesc, srv)
}

func _Grpc_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.grpc/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Grpc_ServiceName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServer).ServiceName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.grpc/ServiceName",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServer).ServiceName(ctx, req.(*ServiceNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Grpc_ServiceMode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceModeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcServer).ServiceMode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.grpc/ServiceMode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcServer).ServiceMode(ctx, req.(*ServiceModeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Grpc_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.grpc",
	HandlerType: (*GrpcServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Grpc_Ping_Handler,
		},
		{
			MethodName: "ServiceName",
			Handler:    _Grpc_ServiceName_Handler,
		},
		{
			MethodName: "ServiceMode",
			Handler:    _Grpc_ServiceMode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/grpc.proto",
}

func init() { proto1.RegisterFile("proto/grpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 213 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0x2c, 0x2e, 0xd3, 0x03, 0xb3, 0x84, 0x58, 0xc1, 0x94, 0x12, 0x2f, 0x17, 0x77,
	0x40, 0x66, 0x5e, 0x7a, 0x50, 0x6a, 0x61, 0x69, 0x6a, 0x71, 0x89, 0x92, 0x12, 0x17, 0x0f, 0x84,
	0x5b, 0x5c, 0x90, 0x9f, 0x57, 0x9c, 0x2a, 0x24, 0xc4, 0xc5, 0x52, 0x90, 0x9f, 0x97, 0x2e, 0xc1,
	0xa8, 0xc0, 0xa8, 0xc1, 0x19, 0x04, 0x66, 0x2b, 0x89, 0x70, 0x09, 0x05, 0xa7, 0x16, 0x95, 0x65,
	0x26, 0xa7, 0xfa, 0x25, 0xe6, 0xa6, 0xc2, 0x74, 0x5a, 0x70, 0x09, 0xa3, 0x88, 0x42, 0x0d, 0x50,
	0xe4, 0xe2, 0x29, 0x86, 0x08, 0xc7, 0xe7, 0x25, 0xe6, 0xa6, 0x42, 0x0d, 0xe2, 0x2e, 0x46, 0x28,
	0x45, 0x32, 0xcf, 0x37, 0x3f, 0x05, 0x8b, 0x79, 0x10, 0x51, 0x4c, 0xf3, 0x72, 0xf3, 0x53, 0xd0,
	0xcd, 0x03, 0x29, 0x35, 0x3a, 0xc1, 0xc8, 0xc5, 0x9c, 0x58, 0x5c, 0x26, 0x64, 0xc8, 0xc5, 0x02,
	0xf2, 0x8b, 0x90, 0x10, 0xc4, 0xc7, 0x7a, 0x48, 0xfe, 0x94, 0x12, 0x46, 0x11, 0x83, 0x98, 0xad,
	0xc4, 0x20, 0xe4, 0xc6, 0xc5, 0x8d, 0xe4, 0x09, 0x21, 0x49, 0xa8, 0x2a, 0x4c, 0xef, 0x4a, 0x49,
	0x61, 0x93, 0xc2, 0x62, 0x0e, 0xc8, 0x45, 0xe8, 0xe6, 0x20, 0x79, 0x13, 0xdd, 0x1c, 0x64, 0xbf,
	0x2a, 0x31, 0x24, 0xb1, 0x81, 0x25, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8d, 0x95, 0xbf,
	0x2f, 0xbe, 0x01, 0x00, 0x00,
}
