package http_framework

import (
	"context"

	"google.golang.org/grpc"
)

// service wide interceptors
type GrpcStreamInterceptor interface {
	InterceptStream(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error
}
type GrpcUnaryInterceptor interface {
	InterceptUnary(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error)
}

func interceptServiceStream(
	srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) (err error) {
	if inter, ok := srv.(GrpcStreamInterceptor); ok {
		return inter.InterceptStream(srv, ss, info, handler)
	}
	return handler(srv, ss)
}
func interceptServiceUnary(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	if inter, ok := info.Server.(GrpcUnaryInterceptor); ok {
		return inter.InterceptUnary(ctx, req, info, handler)
	}
	return handler(ctx, req)
}
