// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: public/portworx/pds/deploymentconfigupdate/apiv1/deploymentconfigupdate.proto

/*
Package deploymentconfigupdate is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package deploymentconfigupdate

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

func request_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0(ctx context.Context, marshaler runtime.Marshaler, client DeploymentConfigUpdateServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq CreateDeploymentConfigUpdateRequest
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq.DeploymentConfigUpdate); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["deployment_config_update.config.deployment_meta.uid"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "deployment_config_update.config.deployment_meta.uid")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "deployment_config_update.config.deployment_meta.uid", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "deployment_config_update.config.deployment_meta.uid", err)
	}

	msg, err := client.CreateDeploymentConfigUpdate(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0(ctx context.Context, marshaler runtime.Marshaler, server DeploymentConfigUpdateServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq CreateDeploymentConfigUpdateRequest
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq.DeploymentConfigUpdate); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["deployment_config_update.config.deployment_meta.uid"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "deployment_config_update.config.deployment_meta.uid")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "deployment_config_update.config.deployment_meta.uid", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "deployment_config_update.config.deployment_meta.uid", err)
	}

	msg, err := server.CreateDeploymentConfigUpdate(ctx, &protoReq)
	return msg, metadata, err

}

func request_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0(ctx context.Context, marshaler runtime.Marshaler, client DeploymentConfigUpdateServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetDeploymentConfigUpdateRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	msg, err := client.GetDeploymentConfigUpdate(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0(ctx context.Context, marshaler runtime.Marshaler, server DeploymentConfigUpdateServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetDeploymentConfigUpdateRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	msg, err := server.GetDeploymentConfigUpdate(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0(ctx context.Context, marshaler runtime.Marshaler, client DeploymentConfigUpdateServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq ListDeploymentConfigUpdatesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.ListDeploymentConfigUpdates(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0(ctx context.Context, marshaler runtime.Marshaler, server DeploymentConfigUpdateServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq ListDeploymentConfigUpdatesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.ListDeploymentConfigUpdates(ctx, &protoReq)
	return msg, metadata, err

}

func request_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0(ctx context.Context, marshaler runtime.Marshaler, client DeploymentConfigUpdateServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq RetryDeploymentConfigUpdateRequest
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	msg, err := client.RetryDeploymentConfigUpdate(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0(ctx context.Context, marshaler runtime.Marshaler, server DeploymentConfigUpdateServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq RetryDeploymentConfigUpdateRequest
	var metadata runtime.ServerMetadata

	if err := marshaler.NewDecoder(req.Body).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	msg, err := server.RetryDeploymentConfigUpdate(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterDeploymentConfigUpdateServiceHandlerServer registers the http handlers for service DeploymentConfigUpdateService to "mux".
// UnaryRPC     :call DeploymentConfigUpdateServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterDeploymentConfigUpdateServiceHandlerFromEndpoint instead.
func RegisterDeploymentConfigUpdateServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server DeploymentConfigUpdateServiceServer) error {

	mux.Handle("PUT", pattern_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/CreateDeploymentConfigUpdate", runtime.WithHTTPPathPattern("/pds/v1/deployments/{deployment_config_update.config.deployment_meta.uid}/deploymentConfigUpdates"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/GetDeploymentConfigUpdate", runtime.WithHTTPPathPattern("/pds/v1/deploymentConfigUpdates/{id}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/ListDeploymentConfigUpdates", runtime.WithHTTPPathPattern("/pds/v1/deploymentConfigUpdates"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/RetryDeploymentConfigUpdate", runtime.WithHTTPPathPattern("/pds/v1/deploymentConfigUpdates/{id}:retry"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterDeploymentConfigUpdateServiceHandlerFromEndpoint is same as RegisterDeploymentConfigUpdateServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterDeploymentConfigUpdateServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterDeploymentConfigUpdateServiceHandler(ctx, mux, conn)
}

// RegisterDeploymentConfigUpdateServiceHandler registers the http handlers for service DeploymentConfigUpdateService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterDeploymentConfigUpdateServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterDeploymentConfigUpdateServiceHandlerClient(ctx, mux, NewDeploymentConfigUpdateServiceClient(conn))
}

// RegisterDeploymentConfigUpdateServiceHandlerClient registers the http handlers for service DeploymentConfigUpdateService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "DeploymentConfigUpdateServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "DeploymentConfigUpdateServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "DeploymentConfigUpdateServiceClient" to call the correct interceptors.
func RegisterDeploymentConfigUpdateServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client DeploymentConfigUpdateServiceClient) error {

	mux.Handle("PUT", pattern_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/CreateDeploymentConfigUpdate", runtime.WithHTTPPathPattern("/pds/v1/deployments/{deployment_config_update.config.deployment_meta.uid}/deploymentConfigUpdates"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/GetDeploymentConfigUpdate", runtime.WithHTTPPathPattern("/pds/v1/deploymentConfigUpdates/{id}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/ListDeploymentConfigUpdates", runtime.WithHTTPPathPattern("/pds/v1/deploymentConfigUpdates"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/public.portworx.pds.deploymentconfigupdate.v1.DeploymentConfigUpdateService/RetryDeploymentConfigUpdate", runtime.WithHTTPPathPattern("/pds/v1/deploymentConfigUpdates/{id}:retry"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3, 2, 4}, []string{"pds", "v1", "deployments", "deployment_config_update.config.deployment_meta.uid", "deploymentConfigUpdates"}, ""))

	pattern_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3}, []string{"pds", "v1", "deploymentConfigUpdates", "id"}, ""))

	pattern_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"pds", "v1", "deploymentConfigUpdates"}, ""))

	pattern_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3}, []string{"pds", "v1", "deploymentConfigUpdates", "id"}, "retry"))
)

var (
	forward_DeploymentConfigUpdateService_CreateDeploymentConfigUpdate_0 = runtime.ForwardResponseMessage

	forward_DeploymentConfigUpdateService_GetDeploymentConfigUpdate_0 = runtime.ForwardResponseMessage

	forward_DeploymentConfigUpdateService_ListDeploymentConfigUpdates_0 = runtime.ForwardResponseMessage

	forward_DeploymentConfigUpdateService_RetryDeploymentConfigUpdate_0 = runtime.ForwardResponseMessage
)