// Code generated by protoc-gen-go. DO NOT EDIT.
// source: server.proto

package tucson

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Simple ping message
type PingMessage struct {
	Msg                  string   `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingMessage) Reset()         { *m = PingMessage{} }
func (m *PingMessage) String() string { return proto.CompactTextString(m) }
func (*PingMessage) ProtoMessage()    {}
func (*PingMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad098daeda4239f7, []int{0}
}

func (m *PingMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingMessage.Unmarshal(m, b)
}
func (m *PingMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingMessage.Marshal(b, m, deterministic)
}
func (m *PingMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingMessage.Merge(m, src)
}
func (m *PingMessage) XXX_Size() int {
	return xxx_messageInfo_PingMessage.Size(m)
}
func (m *PingMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_PingMessage.DiscardUnknown(m)
}

var xxx_messageInfo_PingMessage proto.InternalMessageInfo

func (m *PingMessage) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

// Object for asking the model
type ModelRequestMessage struct {
	PublicationPoint     string   `protobuf:"bytes,1,opt,name=publicationPoint,proto3" json:"publicationPoint,omitempty"`
	Campaign             string   `protobuf:"bytes,2,opt,name=campaign,proto3" json:"campaign,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ModelRequestMessage) Reset()         { *m = ModelRequestMessage{} }
func (m *ModelRequestMessage) String() string { return proto.CompactTextString(m) }
func (*ModelRequestMessage) ProtoMessage()    {}
func (*ModelRequestMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad098daeda4239f7, []int{1}
}

func (m *ModelRequestMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ModelRequestMessage.Unmarshal(m, b)
}
func (m *ModelRequestMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ModelRequestMessage.Marshal(b, m, deterministic)
}
func (m *ModelRequestMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ModelRequestMessage.Merge(m, src)
}
func (m *ModelRequestMessage) XXX_Size() int {
	return xxx_messageInfo_ModelRequestMessage.Size(m)
}
func (m *ModelRequestMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ModelRequestMessage.DiscardUnknown(m)
}

var xxx_messageInfo_ModelRequestMessage proto.InternalMessageInfo

func (m *ModelRequestMessage) GetPublicationPoint() string {
	if m != nil {
		return m.PublicationPoint
	}
	return ""
}

func (m *ModelRequestMessage) GetCampaign() string {
	if m != nil {
		return m.Campaign
	}
	return ""
}

// Model name as returned message
type ModelResponseMessage struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ModelResponseMessage) Reset()         { *m = ModelResponseMessage{} }
func (m *ModelResponseMessage) String() string { return proto.CompactTextString(m) }
func (*ModelResponseMessage) ProtoMessage()    {}
func (*ModelResponseMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad098daeda4239f7, []int{2}
}

func (m *ModelResponseMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ModelResponseMessage.Unmarshal(m, b)
}
func (m *ModelResponseMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ModelResponseMessage.Marshal(b, m, deterministic)
}
func (m *ModelResponseMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ModelResponseMessage.Merge(m, src)
}
func (m *ModelResponseMessage) XXX_Size() int {
	return xxx_messageInfo_ModelResponseMessage.Size(m)
}
func (m *ModelResponseMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ModelResponseMessage.DiscardUnknown(m)
}

var xxx_messageInfo_ModelResponseMessage proto.InternalMessageInfo

func (m *ModelResponseMessage) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*PingMessage)(nil), "tucson.PingMessage")
	proto.RegisterType((*ModelRequestMessage)(nil), "tucson.ModelRequestMessage")
	proto.RegisterType((*ModelResponseMessage)(nil), "tucson.ModelResponseMessage")
}

func init() { proto.RegisterFile("server.proto", fileDescriptor_ad098daeda4239f7) }

var fileDescriptor_ad098daeda4239f7 = []byte{
	// 218 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x41, 0x4f, 0x84, 0x30,
	0x10, 0x85, 0x17, 0xdd, 0x90, 0x75, 0xf4, 0xb0, 0x99, 0xf5, 0x40, 0xd0, 0x44, 0xd3, 0x93, 0xe1,
	0xc0, 0x01, 0x7f, 0x04, 0x27, 0x12, 0x82, 0x67, 0x0f, 0x05, 0x27, 0x4d, 0x13, 0x68, 0x2b, 0x53,
	0xfc, 0x07, 0xfe, 0x6f, 0x43, 0x05, 0xa3, 0x2e, 0xb7, 0xe9, 0xeb, 0xeb, 0xeb, 0xfb, 0x06, 0x6e,
	0x98, 0xc6, 0x0f, 0x1a, 0x73, 0x37, 0x5a, 0x6f, 0x31, 0xf6, 0x53, 0xc7, 0xd6, 0x88, 0x07, 0xb8,
	0xae, 0xb5, 0x51, 0x15, 0x31, 0x4b, 0x45, 0x78, 0x84, 0xcb, 0x81, 0x55, 0x12, 0x3d, 0x46, 0x4f,
	0x57, 0xcd, 0x3c, 0x8a, 0x57, 0x38, 0x55, 0xf6, 0x8d, 0xfa, 0x86, 0xde, 0x27, 0x62, 0xbf, 0x1a,
	0x33, 0x38, 0xba, 0xa9, 0xed, 0x75, 0x27, 0xbd, 0xb6, 0xa6, 0xb6, 0xda, 0xf8, 0xe5, 0xd5, 0x99,
	0x8e, 0x29, 0x1c, 0x3a, 0x39, 0x38, 0xa9, 0x95, 0x49, 0x2e, 0x82, 0xe7, 0xe7, 0x2c, 0x32, 0xb8,
	0x5d, 0xe2, 0xd9, 0x59, 0xc3, 0xb4, 0xe6, 0x23, 0xec, 0x8d, 0x1c, 0x68, 0xc9, 0x0c, 0x73, 0xf1,
	0x19, 0x41, 0xfc, 0x12, 0x20, 0xb0, 0x80, 0xfd, 0x5c, 0x1b, 0x4f, 0xf9, 0x37, 0x47, 0xfe, 0x0b,
	0x22, 0xdd, 0x12, 0xc5, 0x0e, 0x4b, 0x38, 0x94, 0xe4, 0xc3, 0x6f, 0x78, 0xb7, 0x5a, 0x36, 0xd8,
	0xd2, 0xfb, 0x7f, 0x97, 0x7f, 0x9a, 0x89, 0x5d, 0x1b, 0x87, 0x15, 0x3e, 0x7f, 0x05, 0x00, 0x00,
	0xff, 0xff, 0x69, 0xb9, 0xd0, 0xb7, 0x52, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ServerClient is the client API for Server service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServerClient interface {
	// Simple Ping call, useful to check gRPC connection
	// liveness / readiness status.
	Ping(ctx context.Context, in *PingMessage, opts ...grpc.CallOption) (*PingMessage, error)
	//Given the message in input, it returns the model name to the client
	GetModel(ctx context.Context, in *ModelRequestMessage, opts ...grpc.CallOption) (*ModelResponseMessage, error)
}

type serverClient struct {
	cc *grpc.ClientConn
}

func NewServerClient(cc *grpc.ClientConn) ServerClient {
	return &serverClient{cc}
}

func (c *serverClient) Ping(ctx context.Context, in *PingMessage, opts ...grpc.CallOption) (*PingMessage, error) {
	out := new(PingMessage)
	err := c.cc.Invoke(ctx, "/tucson.Server/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serverClient) GetModel(ctx context.Context, in *ModelRequestMessage, opts ...grpc.CallOption) (*ModelResponseMessage, error) {
	out := new(ModelResponseMessage)
	err := c.cc.Invoke(ctx, "/tucson.Server/GetModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServerServer is the server API for Server service.
type ServerServer interface {
	// Simple Ping call, useful to check gRPC connection
	// liveness / readiness status.
	Ping(context.Context, *PingMessage) (*PingMessage, error)
	//Given the message in input, it returns the model name to the client
	GetModel(context.Context, *ModelRequestMessage) (*ModelResponseMessage, error)
}

// UnimplementedServerServer can be embedded to have forward compatible implementations.
type UnimplementedServerServer struct {
}

func (*UnimplementedServerServer) Ping(ctx context.Context, req *PingMessage) (*PingMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedServerServer) GetModel(ctx context.Context, req *ModelRequestMessage) (*ModelResponseMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModel not implemented")
}

func RegisterServerServer(s *grpc.Server, srv ServerServer) {
	s.RegisterService(&_Server_serviceDesc, srv)
}

func _Server_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tucson.Server/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).Ping(ctx, req.(*PingMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Server_GetModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModelRequestMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServerServer).GetModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tucson.Server/GetModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServerServer).GetModel(ctx, req.(*ModelRequestMessage))
	}
	return interceptor(ctx, in, info, handler)
}

var _Server_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tucson.Server",
	HandlerType: (*ServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Server_Ping_Handler,
		},
		{
			MethodName: "GetModel",
			Handler:    _Server_GetModel_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}
