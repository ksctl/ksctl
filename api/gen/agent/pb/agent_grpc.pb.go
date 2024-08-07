// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v5.27.1
// source: agent.proto

package pb

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

// KsctlAgentClient is the client API for KsctlAgent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KsctlAgentClient interface {
	Scale(ctx context.Context, in *ReqScale, opts ...grpc.CallOption) (*ResScale, error)
	LoadBalancer(ctx context.Context, in *ReqLB, opts ...grpc.CallOption) (*ResLB, error)
	Application(ctx context.Context, in *ReqApplication, opts ...grpc.CallOption) (*ResApplication, error)
}

type ksctlAgentClient struct {
	cc grpc.ClientConnInterface
}

func NewKsctlAgentClient(cc grpc.ClientConnInterface) KsctlAgentClient {
	return &ksctlAgentClient{cc}
}

func (c *ksctlAgentClient) Scale(ctx context.Context, in *ReqScale, opts ...grpc.CallOption) (*ResScale, error) {
	out := new(ResScale)
	err := c.cc.Invoke(ctx, "/ksctlAgent.KsctlAgent/Scale", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ksctlAgentClient) LoadBalancer(ctx context.Context, in *ReqLB, opts ...grpc.CallOption) (*ResLB, error) {
	out := new(ResLB)
	err := c.cc.Invoke(ctx, "/ksctlAgent.KsctlAgent/LoadBalancer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ksctlAgentClient) Application(ctx context.Context, in *ReqApplication, opts ...grpc.CallOption) (*ResApplication, error) {
	out := new(ResApplication)
	err := c.cc.Invoke(ctx, "/ksctlAgent.KsctlAgent/Application", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KsctlAgentServer is the server API for KsctlAgent service.
// All implementations must embed UnimplementedKsctlAgentServer
// for forward compatibility
type KsctlAgentServer interface {
	Scale(context.Context, *ReqScale) (*ResScale, error)
	LoadBalancer(context.Context, *ReqLB) (*ResLB, error)
	Application(context.Context, *ReqApplication) (*ResApplication, error)
	mustEmbedUnimplementedKsctlAgentServer()
}

// UnimplementedKsctlAgentServer must be embedded to have forward compatible implementations.
type UnimplementedKsctlAgentServer struct {
}

func (UnimplementedKsctlAgentServer) Scale(context.Context, *ReqScale) (*ResScale, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Scale not implemented")
}
func (UnimplementedKsctlAgentServer) LoadBalancer(context.Context, *ReqLB) (*ResLB, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadBalancer not implemented")
}
func (UnimplementedKsctlAgentServer) Application(context.Context, *ReqApplication) (*ResApplication, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Application not implemented")
}
func (UnimplementedKsctlAgentServer) mustEmbedUnimplementedKsctlAgentServer() {}

// UnsafeKsctlAgentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KsctlAgentServer will
// result in compilation errors.
type UnsafeKsctlAgentServer interface {
	mustEmbedUnimplementedKsctlAgentServer()
}

func RegisterKsctlAgentServer(s grpc.ServiceRegistrar, srv KsctlAgentServer) {
	s.RegisterService(&KsctlAgent_ServiceDesc, srv)
}

func _KsctlAgent_Scale_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReqScale)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KsctlAgentServer).Scale(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ksctlAgent.KsctlAgent/Scale",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KsctlAgentServer).Scale(ctx, req.(*ReqScale))
	}
	return interceptor(ctx, in, info, handler)
}

func _KsctlAgent_LoadBalancer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReqLB)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KsctlAgentServer).LoadBalancer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ksctlAgent.KsctlAgent/LoadBalancer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KsctlAgentServer).LoadBalancer(ctx, req.(*ReqLB))
	}
	return interceptor(ctx, in, info, handler)
}

func _KsctlAgent_Application_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReqApplication)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KsctlAgentServer).Application(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ksctlAgent.KsctlAgent/Application",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KsctlAgentServer).Application(ctx, req.(*ReqApplication))
	}
	return interceptor(ctx, in, info, handler)
}

// KsctlAgent_ServiceDesc is the grpc.ServiceDesc for KsctlAgent service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KsctlAgent_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ksctlAgent.KsctlAgent",
	HandlerType: (*KsctlAgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Scale",
			Handler:    _KsctlAgent_Scale_Handler,
		},
		{
			MethodName: "LoadBalancer",
			Handler:    _KsctlAgent_LoadBalancer_Handler,
		},
		{
			MethodName: "Application",
			Handler:    _KsctlAgent_Application_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "agent.proto",
}
