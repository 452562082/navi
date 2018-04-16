/*
Package navi_grpc is a generated protocol buffer package.

It is generated from these files:
	navi.proto

It has these top-level messages:
	PingRequest
	PingResponse
	ServiceNameRequest
	ServiceNameResponse
	ServiceModeRequest
	ServiceModeResponse
*/
package navi_grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PingRequest struct {
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type PingResponse struct {
	Pong string `protobuf:"bytes,1,opt,name=pong" json:"pong,omitempty"`
}

func (m *PingResponse) Reset()                    { *m = PingResponse{} }
func (m *PingResponse) String() string            { return proto.CompactTextString(m) }
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
func (m *ServiceNameRequest) String() string            { return proto.CompactTextString(m) }
func (*ServiceNameRequest) ProtoMessage()               {}
func (*ServiceNameRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type ServiceNameResponse struct {
	ServiceName string `protobuf:"bytes,1,opt,name=service_name,json=serviceName" json:"service_name,omitempty"`
}

func (m *ServiceNameResponse) Reset()                    { *m = ServiceNameResponse{} }
func (m *ServiceNameResponse) String() string            { return proto.CompactTextString(m) }
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
func (m *ServiceModeRequest) String() string            { return proto.CompactTextString(m) }
func (*ServiceModeRequest) ProtoMessage()               {}
func (*ServiceModeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ServiceModeResponse struct {
	ServiceMode string `protobuf:"bytes,1,opt,name=service_mode,json=serviceMode" json:"service_mode,omitempty"`
}

func (m *ServiceModeResponse) Reset()                    { *m = ServiceModeResponse{} }
func (m *ServiceModeResponse) String() string            { return proto.CompactTextString(m) }
func (*ServiceModeResponse) ProtoMessage()               {}
func (*ServiceModeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ServiceModeResponse) GetServiceMode() string {
	if m != nil {
		return m.ServiceMode
	}
	return ""
}

type SayHelloRequest struct {
	YourName string `protobuf:"bytes,1,opt,name=yourName" json:"yourName,omitempty"`
}

func (m *SayHelloRequest) Reset()                    { *m = SayHelloRequest{} }
func (m *SayHelloRequest) String() string            { return proto.CompactTextString(m) }
func (*SayHelloRequest) ProtoMessage()               {}
func (*SayHelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *SayHelloRequest) GetYourName() string {
	if m != nil {
		return m.YourName
	}
	return ""
}

type SayHelloResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *SayHelloResponse) Reset()                    { *m = SayHelloResponse{} }
func (m *SayHelloResponse) String() string            { return proto.CompactTextString(m) }
func (*SayHelloResponse) ProtoMessage()               {}
func (*SayHelloResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *SayHelloResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*PingRequest)(nil), "navi_grpc.PingRequest")
	proto.RegisterType((*PingResponse)(nil), "navi_grpc.PingResponse")
	proto.RegisterType((*ServiceNameRequest)(nil), "navi_grpc.ServiceNameRequest")
	proto.RegisterType((*ServiceNameResponse)(nil), "navi_grpc.ServiceNameResponse")
	proto.RegisterType((*ServiceModeRequest)(nil), "navi_grpc.ServiceModeRequest")
	proto.RegisterType((*ServiceModeResponse)(nil), "navi_grpc.ServiceModeResponse")
	proto.RegisterType((*SayHelloRequest)(nil), "navi_grpc.SayHelloRequest")
	proto.RegisterType((*SayHelloResponse)(nil), "navi_grpc.SayHelloResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Navi service

type NaviClient interface {
	// rpc server必须实现的接口，返回字符串 "pong" 即可
	Ping(ctx context.Context, in *PingRequest, service string, opts ...grpc.CallOption) (*PingResponse, error)
	// rpc server必须实现的接口，返回服务名称，为首字母大写的驼峰格式，例如 "AsvService"
	ServiceName(ctx context.Context, in *ServiceNameRequest, service string, opts ...grpc.CallOption) (*ServiceNameResponse, error)
	// rpc server必须实现的接口，说明该server是以什么模式运行，分为dev和prod；dev为开发版本，prod为生产版本
	ServiceMode(ctx context.Context, in *ServiceModeRequest, service string, opts ...grpc.CallOption) (*ServiceModeResponse, error)
}

type naviClient struct {
	cc *grpc.ClientConn
}

func NewNaviClient(cc *grpc.ClientConn) NaviClient {
	return &naviClient{cc}
}

func (c *naviClient) Ping(ctx context.Context, in *PingRequest, service string, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := grpc.Invoke(ctx, "/navi_grpc."+service+"/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *naviClient) ServiceName(ctx context.Context, in *ServiceNameRequest, service string, opts ...grpc.CallOption) (*ServiceNameResponse, error) {
	out := new(ServiceNameResponse)
	err := grpc.Invoke(ctx, "/navi_grpc."+service+"/ServiceName", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *naviClient) ServiceMode(ctx context.Context, in *ServiceModeRequest, service string, opts ...grpc.CallOption) (*ServiceModeResponse, error) {
	out := new(ServiceModeResponse)
	err := grpc.Invoke(ctx, "/navi_grpc."+service+"/ServiceMode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func init() { proto.RegisterFile("navi.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 280 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0xd1, 0x4a, 0xc3, 0x30,
	0x14, 0x86, 0x37, 0x29, 0xba, 0x9d, 0x4e, 0x94, 0xa3, 0xe8, 0x88, 0x28, 0x9a, 0x2b, 0x2f, 0xb4,
	0x17, 0x7a, 0xa3, 0x0f, 0x20, 0x78, 0x63, 0x91, 0xed, 0x01, 0x46, 0xdc, 0x0e, 0xa5, 0xb0, 0x26,
	0xb5, 0xd9, 0x06, 0x7b, 0x23, 0x1f, 0x53, 0xb2, 0x26, 0x6b, 0x6a, 0xe3, 0x5d, 0xce, 0x7f, 0xfe,
	0x7e, 0x6d, 0x3f, 0x02, 0x20, 0xc5, 0x26, 0x4f, 0xca, 0x4a, 0xad, 0x14, 0x0e, 0xcd, 0x79, 0x96,
	0x55, 0xe5, 0x9c, 0x1f, 0x43, 0xfc, 0x99, 0xcb, 0x6c, 0x42, 0xdf, 0x6b, 0xd2, 0x2b, 0xce, 0x61,
	0x54, 0x8f, 0xba, 0x54, 0x52, 0x13, 0x22, 0x44, 0xa5, 0x92, 0xd9, 0xb8, 0x7f, 0xdb, 0xbf, 0x1f,
	0x4e, 0x76, 0x67, 0x7e, 0x0e, 0x38, 0xa5, 0x6a, 0x93, 0xcf, 0x29, 0x15, 0x05, 0xb9, 0x27, 0x5f,
	0xe0, 0xac, 0x95, 0x5a, 0xc0, 0x1d, 0x8c, 0x74, 0x1d, 0xcf, 0xa4, 0x28, 0xc8, 0x82, 0x62, 0xdd,
	0x54, 0x3d, 0xde, 0x87, 0x5a, 0x04, 0x78, 0x75, 0xda, 0xe5, 0x15, 0x6a, 0xf1, 0x97, 0x67, 0xaa,
	0xfc, 0x11, 0x4e, 0xa6, 0x62, 0xfb, 0x4e, 0xcb, 0xa5, 0xb2, 0x30, 0x64, 0x30, 0xd8, 0xaa, 0x75,
	0x95, 0x36, 0x5f, 0xb0, 0x9f, 0xf9, 0x03, 0x9c, 0x36, 0x75, 0xfb, 0x96, 0x31, 0x1c, 0x15, 0xa4,
	0xb5, 0xc8, 0x5c, 0xdd, 0x8d, 0x4f, 0x3f, 0x07, 0x10, 0x19, 0x7b, 0xf8, 0x0a, 0x91, 0x31, 0x85,
	0x17, 0xc9, 0x5e, 0x66, 0xe2, 0x99, 0x64, 0x97, 0x9d, 0xbc, 0x66, 0xf3, 0x1e, 0xa6, 0x10, 0x7b,
	0xaa, 0xf0, 0xda, 0x6b, 0x76, 0xc5, 0xb2, 0x9b, 0xff, 0xd6, 0x01, 0x9e, 0xf9, 0xff, 0x10, 0xcf,
	0x13, 0x1b, 0xe2, 0xf9, 0x86, 0x79, 0x0f, 0xdf, 0x60, 0xe0, 0x8c, 0x20, 0xf3, 0xdb, 0x6d, 0xab,
	0xec, 0x2a, 0xb8, 0x73, 0x98, 0xaf, 0xc3, 0xdd, 0x65, 0x7b, 0xfe, 0x0d, 0x00, 0x00, 0xff, 0xff,
	0x71, 0x06, 0x58, 0x1e, 0x7a, 0x02, 0x00, 0x00,
}
