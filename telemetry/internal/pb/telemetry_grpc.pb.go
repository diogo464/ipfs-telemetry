// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.1
// source: internal/pb/telemetry.proto

package pb

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
	Telemetry_GetSession_FullMethodName          = "/telemetry.Telemetry/GetSession"
	Telemetry_GetProperties_FullMethodName       = "/telemetry.Telemetry/GetProperties"
	Telemetry_GetMetrics_FullMethodName          = "/telemetry.Telemetry/GetMetrics"
	Telemetry_GetEventDescriptors_FullMethodName = "/telemetry.Telemetry/GetEventDescriptors"
	Telemetry_GetEvents_FullMethodName           = "/telemetry.Telemetry/GetEvents"
)

// TelemetryClient is the client API for Telemetry service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TelemetryClient interface {
	GetSession(ctx context.Context, in *GetSessionRequest, opts ...grpc.CallOption) (*GetSessionResponse, error)
	GetProperties(ctx context.Context, in *GetPropertiesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Property], error)
	GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamSegment], error)
	GetEventDescriptors(ctx context.Context, in *GetEventDescriptorsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[EventDescriptor], error)
	GetEvents(ctx context.Context, in *GetEventsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamSegment], error)
}

type telemetryClient struct {
	cc grpc.ClientConnInterface
}

func NewTelemetryClient(cc grpc.ClientConnInterface) TelemetryClient {
	return &telemetryClient{cc}
}

func (c *telemetryClient) GetSession(ctx context.Context, in *GetSessionRequest, opts ...grpc.CallOption) (*GetSessionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetSessionResponse)
	err := c.cc.Invoke(ctx, Telemetry_GetSession_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *telemetryClient) GetProperties(ctx context.Context, in *GetPropertiesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Property], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Telemetry_ServiceDesc.Streams[0], Telemetry_GetProperties_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[GetPropertiesRequest, Property]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetPropertiesClient = grpc.ServerStreamingClient[Property]

func (c *telemetryClient) GetMetrics(ctx context.Context, in *GetMetricsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamSegment], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Telemetry_ServiceDesc.Streams[1], Telemetry_GetMetrics_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[GetMetricsRequest, StreamSegment]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetMetricsClient = grpc.ServerStreamingClient[StreamSegment]

func (c *telemetryClient) GetEventDescriptors(ctx context.Context, in *GetEventDescriptorsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[EventDescriptor], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Telemetry_ServiceDesc.Streams[2], Telemetry_GetEventDescriptors_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[GetEventDescriptorsRequest, EventDescriptor]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetEventDescriptorsClient = grpc.ServerStreamingClient[EventDescriptor]

func (c *telemetryClient) GetEvents(ctx context.Context, in *GetEventsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamSegment], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Telemetry_ServiceDesc.Streams[3], Telemetry_GetEvents_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[GetEventsRequest, StreamSegment]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetEventsClient = grpc.ServerStreamingClient[StreamSegment]

// TelemetryServer is the server API for Telemetry service.
// All implementations must embed UnimplementedTelemetryServer
// for forward compatibility.
type TelemetryServer interface {
	GetSession(context.Context, *GetSessionRequest) (*GetSessionResponse, error)
	GetProperties(*GetPropertiesRequest, grpc.ServerStreamingServer[Property]) error
	GetMetrics(*GetMetricsRequest, grpc.ServerStreamingServer[StreamSegment]) error
	GetEventDescriptors(*GetEventDescriptorsRequest, grpc.ServerStreamingServer[EventDescriptor]) error
	GetEvents(*GetEventsRequest, grpc.ServerStreamingServer[StreamSegment]) error
	mustEmbedUnimplementedTelemetryServer()
}

// UnimplementedTelemetryServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTelemetryServer struct{}

func (UnimplementedTelemetryServer) GetSession(context.Context, *GetSessionRequest) (*GetSessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSession not implemented")
}
func (UnimplementedTelemetryServer) GetProperties(*GetPropertiesRequest, grpc.ServerStreamingServer[Property]) error {
	return status.Errorf(codes.Unimplemented, "method GetProperties not implemented")
}
func (UnimplementedTelemetryServer) GetMetrics(*GetMetricsRequest, grpc.ServerStreamingServer[StreamSegment]) error {
	return status.Errorf(codes.Unimplemented, "method GetMetrics not implemented")
}
func (UnimplementedTelemetryServer) GetEventDescriptors(*GetEventDescriptorsRequest, grpc.ServerStreamingServer[EventDescriptor]) error {
	return status.Errorf(codes.Unimplemented, "method GetEventDescriptors not implemented")
}
func (UnimplementedTelemetryServer) GetEvents(*GetEventsRequest, grpc.ServerStreamingServer[StreamSegment]) error {
	return status.Errorf(codes.Unimplemented, "method GetEvents not implemented")
}
func (UnimplementedTelemetryServer) mustEmbedUnimplementedTelemetryServer() {}
func (UnimplementedTelemetryServer) testEmbeddedByValue()                   {}

// UnsafeTelemetryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TelemetryServer will
// result in compilation errors.
type UnsafeTelemetryServer interface {
	mustEmbedUnimplementedTelemetryServer()
}

func RegisterTelemetryServer(s grpc.ServiceRegistrar, srv TelemetryServer) {
	// If the following call pancis, it indicates UnimplementedTelemetryServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Telemetry_ServiceDesc, srv)
}

func _Telemetry_GetSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TelemetryServer).GetSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Telemetry_GetSession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TelemetryServer).GetSession(ctx, req.(*GetSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Telemetry_GetProperties_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetPropertiesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TelemetryServer).GetProperties(m, &grpc.GenericServerStream[GetPropertiesRequest, Property]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetPropertiesServer = grpc.ServerStreamingServer[Property]

func _Telemetry_GetMetrics_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetMetricsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TelemetryServer).GetMetrics(m, &grpc.GenericServerStream[GetMetricsRequest, StreamSegment]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetMetricsServer = grpc.ServerStreamingServer[StreamSegment]

func _Telemetry_GetEventDescriptors_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetEventDescriptorsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TelemetryServer).GetEventDescriptors(m, &grpc.GenericServerStream[GetEventDescriptorsRequest, EventDescriptor]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetEventDescriptorsServer = grpc.ServerStreamingServer[EventDescriptor]

func _Telemetry_GetEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TelemetryServer).GetEvents(m, &grpc.GenericServerStream[GetEventsRequest, StreamSegment]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Telemetry_GetEventsServer = grpc.ServerStreamingServer[StreamSegment]

// Telemetry_ServiceDesc is the grpc.ServiceDesc for Telemetry service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Telemetry_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "telemetry.Telemetry",
	HandlerType: (*TelemetryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSession",
			Handler:    _Telemetry_GetSession_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetProperties",
			Handler:       _Telemetry_GetProperties_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetMetrics",
			Handler:       _Telemetry_GetMetrics_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetEventDescriptors",
			Handler:       _Telemetry_GetEventDescriptors_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetEvents",
			Handler:       _Telemetry_GetEvents_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "internal/pb/telemetry.proto",
}
