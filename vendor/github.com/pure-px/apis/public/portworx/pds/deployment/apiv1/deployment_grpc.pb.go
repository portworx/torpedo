// Please use the following editor setup for this file:
// Tab size=2; Tabs as spaces; Clean up trailing whitepsace
//
// In vim add: au FileType proto setl sw=2 ts=2 expandtab list
//
// In vscode install vscode-proto3 extension and add this to your settings.json:
//    "[proto3]": {
//        "editor.tabSize": 2,
//        "editor.insertSpaces": true,
//        "editor.rulers": [80],
//        "editor.detectIndentation": true,
//        "files.trimTrailingWhitespace": true
//    }
//

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: public/portworx/pds/deployment/apiv1/deployment.proto

package deployment

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	DeploymentService_CreateDeployment_FullMethodName         = "/public.portworx.pds.deployment.v1.DeploymentService/CreateDeployment"
	DeploymentService_GetDeployment_FullMethodName            = "/public.portworx.pds.deployment.v1.DeploymentService/GetDeployment"
	DeploymentService_UpdateDeployment_FullMethodName         = "/public.portworx.pds.deployment.v1.DeploymentService/UpdateDeployment"
	DeploymentService_DeleteDeployment_FullMethodName         = "/public.portworx.pds.deployment.v1.DeploymentService/DeleteDeployment"
	DeploymentService_ListDeployments_FullMethodName          = "/public.portworx.pds.deployment.v1.DeploymentService/ListDeployments"
	DeploymentService_GetDeploymentCredentials_FullMethodName = "/public.portworx.pds.deployment.v1.DeploymentService/GetDeploymentCredentials"
)

// DeploymentServiceClient is the client API for DeploymentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DeploymentServiceClient interface {
	// CreateDeployment API creates the Deployment resource.
	CreateDeployment(ctx context.Context, in *CreateDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error)
	// GetDeployment API returns the Deployment resource.
	GetDeployment(ctx context.Context, in *GetDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error)
	// UpdateDeployment API updates the Deployment resource.
	UpdateDeployment(ctx context.Context, in *UpdateDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error)
	// DeleteDeployment API deletes the Deployment resource.
	DeleteDeployment(ctx context.Context, in *DeleteDeploymentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ListDeployments API lists the Deployment resources.
	ListDeployments(ctx context.Context, in *ListDeploymentsRequest, opts ...grpc.CallOption) (*ListDeploymentsResponse, error)
	// GetDeploymentCredentials API returns the Credentials to be used to access the Deployment.
	GetDeploymentCredentials(ctx context.Context, in *GetDeploymentCredentialsRequest, opts ...grpc.CallOption) (*DeploymentCredentials, error)
}

type deploymentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDeploymentServiceClient(cc grpc.ClientConnInterface) DeploymentServiceClient {
	return &deploymentServiceClient{cc}
}

func (c *deploymentServiceClient) CreateDeployment(ctx context.Context, in *CreateDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error) {
	out := new(Deployment)
	err := c.cc.Invoke(ctx, DeploymentService_CreateDeployment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentServiceClient) GetDeployment(ctx context.Context, in *GetDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error) {
	out := new(Deployment)
	err := c.cc.Invoke(ctx, DeploymentService_GetDeployment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentServiceClient) UpdateDeployment(ctx context.Context, in *UpdateDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error) {
	out := new(Deployment)
	err := c.cc.Invoke(ctx, DeploymentService_UpdateDeployment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentServiceClient) DeleteDeployment(ctx context.Context, in *DeleteDeploymentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, DeploymentService_DeleteDeployment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentServiceClient) ListDeployments(ctx context.Context, in *ListDeploymentsRequest, opts ...grpc.CallOption) (*ListDeploymentsResponse, error) {
	out := new(ListDeploymentsResponse)
	err := c.cc.Invoke(ctx, DeploymentService_ListDeployments_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentServiceClient) GetDeploymentCredentials(ctx context.Context, in *GetDeploymentCredentialsRequest, opts ...grpc.CallOption) (*DeploymentCredentials, error) {
	out := new(DeploymentCredentials)
	err := c.cc.Invoke(ctx, DeploymentService_GetDeploymentCredentials_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeploymentServiceServer is the server API for DeploymentService service.
// All implementations must embed UnimplementedDeploymentServiceServer
// for forward compatibility
type DeploymentServiceServer interface {
	// CreateDeployment API creates the Deployment resource.
	CreateDeployment(context.Context, *CreateDeploymentRequest) (*Deployment, error)
	// GetDeployment API returns the Deployment resource.
	GetDeployment(context.Context, *GetDeploymentRequest) (*Deployment, error)
	// UpdateDeployment API updates the Deployment resource.
	UpdateDeployment(context.Context, *UpdateDeploymentRequest) (*Deployment, error)
	// DeleteDeployment API deletes the Deployment resource.
	DeleteDeployment(context.Context, *DeleteDeploymentRequest) (*emptypb.Empty, error)
	// ListDeployments API lists the Deployment resources.
	ListDeployments(context.Context, *ListDeploymentsRequest) (*ListDeploymentsResponse, error)
	// GetDeploymentCredentials API returns the Credentials to be used to access the Deployment.
	GetDeploymentCredentials(context.Context, *GetDeploymentCredentialsRequest) (*DeploymentCredentials, error)
	mustEmbedUnimplementedDeploymentServiceServer()
}

// UnimplementedDeploymentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDeploymentServiceServer struct {
}

func (UnimplementedDeploymentServiceServer) CreateDeployment(context.Context, *CreateDeploymentRequest) (*Deployment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDeployment not implemented")
}
func (UnimplementedDeploymentServiceServer) GetDeployment(context.Context, *GetDeploymentRequest) (*Deployment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeployment not implemented")
}
func (UnimplementedDeploymentServiceServer) UpdateDeployment(context.Context, *UpdateDeploymentRequest) (*Deployment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDeployment not implemented")
}
func (UnimplementedDeploymentServiceServer) DeleteDeployment(context.Context, *DeleteDeploymentRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteDeployment not implemented")
}
func (UnimplementedDeploymentServiceServer) ListDeployments(context.Context, *ListDeploymentsRequest) (*ListDeploymentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDeployments not implemented")
}
func (UnimplementedDeploymentServiceServer) GetDeploymentCredentials(context.Context, *GetDeploymentCredentialsRequest) (*DeploymentCredentials, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeploymentCredentials not implemented")
}
func (UnimplementedDeploymentServiceServer) mustEmbedUnimplementedDeploymentServiceServer() {}

// UnsafeDeploymentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DeploymentServiceServer will
// result in compilation errors.
type UnsafeDeploymentServiceServer interface {
	mustEmbedUnimplementedDeploymentServiceServer()
}

func RegisterDeploymentServiceServer(s grpc.ServiceRegistrar, srv DeploymentServiceServer) {
	s.RegisterService(&DeploymentService_ServiceDesc, srv)
}

func _DeploymentService_CreateDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentServiceServer).CreateDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeploymentService_CreateDeployment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentServiceServer).CreateDeployment(ctx, req.(*CreateDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeploymentService_GetDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentServiceServer).GetDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeploymentService_GetDeployment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentServiceServer).GetDeployment(ctx, req.(*GetDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeploymentService_UpdateDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentServiceServer).UpdateDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeploymentService_UpdateDeployment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentServiceServer).UpdateDeployment(ctx, req.(*UpdateDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeploymentService_DeleteDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentServiceServer).DeleteDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeploymentService_DeleteDeployment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentServiceServer).DeleteDeployment(ctx, req.(*DeleteDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeploymentService_ListDeployments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDeploymentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentServiceServer).ListDeployments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeploymentService_ListDeployments_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentServiceServer).ListDeployments(ctx, req.(*ListDeploymentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeploymentService_GetDeploymentCredentials_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeploymentCredentialsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentServiceServer).GetDeploymentCredentials(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeploymentService_GetDeploymentCredentials_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentServiceServer).GetDeploymentCredentials(ctx, req.(*GetDeploymentCredentialsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DeploymentService_ServiceDesc is the grpc.ServiceDesc for DeploymentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DeploymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "public.portworx.pds.deployment.v1.DeploymentService",
	HandlerType: (*DeploymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateDeployment",
			Handler:    _DeploymentService_CreateDeployment_Handler,
		},
		{
			MethodName: "GetDeployment",
			Handler:    _DeploymentService_GetDeployment_Handler,
		},
		{
			MethodName: "UpdateDeployment",
			Handler:    _DeploymentService_UpdateDeployment_Handler,
		},
		{
			MethodName: "DeleteDeployment",
			Handler:    _DeploymentService_DeleteDeployment_Handler,
		},
		{
			MethodName: "ListDeployments",
			Handler:    _DeploymentService_ListDeployments_Handler,
		},
		{
			MethodName: "GetDeploymentCredentials",
			Handler:    _DeploymentService_GetDeploymentCredentials_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "public/portworx/pds/deployment/apiv1/deployment.proto",
}