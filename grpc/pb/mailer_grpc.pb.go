//grpc-интерфейс сервиса рассылки почты

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: grpc/proto/mailer.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MailerService_SendMail_FullMethodName = "/orgdem.services.mailer.MailerService/SendMail"
)

// MailerServiceClient is the client API for MailerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MailerServiceClient interface {
	// отправка письма
	SendMail(ctx context.Context, in *MailerMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type mailerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMailerServiceClient(cc grpc.ClientConnInterface) MailerServiceClient {
	return &mailerServiceClient{cc}
}

func (c *mailerServiceClient) SendMail(ctx context.Context, in *MailerMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, MailerService_SendMail_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MailerServiceServer is the server API for MailerService service.
// All implementations must embed UnimplementedMailerServiceServer
// for forward compatibility.
type MailerServiceServer interface {
	// отправка письма
	SendMail(context.Context, *MailerMessage) (*emptypb.Empty, error)
	mustEmbedUnimplementedMailerServiceServer()
}

// UnimplementedMailerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMailerServiceServer struct{}

func (UnimplementedMailerServiceServer) SendMail(context.Context, *MailerMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMail not implemented")
}
func (UnimplementedMailerServiceServer) mustEmbedUnimplementedMailerServiceServer() {}
func (UnimplementedMailerServiceServer) testEmbeddedByValue()                       {}

// UnsafeMailerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MailerServiceServer will
// result in compilation errors.
type UnsafeMailerServiceServer interface {
	mustEmbedUnimplementedMailerServiceServer()
}

func RegisterMailerServiceServer(s grpc.ServiceRegistrar, srv MailerServiceServer) {
	// If the following call pancis, it indicates UnimplementedMailerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MailerService_ServiceDesc, srv)
}

func _MailerService_SendMail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MailerMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailerServiceServer).SendMail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MailerService_SendMail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailerServiceServer).SendMail(ctx, req.(*MailerMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// MailerService_ServiceDesc is the grpc.ServiceDesc for MailerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MailerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "orgdem.services.mailer.MailerService",
	HandlerType: (*MailerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMail",
			Handler:    _MailerService_SendMail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/proto/mailer.proto",
}
