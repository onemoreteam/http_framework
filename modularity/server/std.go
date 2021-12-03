package server

import (
	"net/http"

	"github.com/onemoreteam/httpframework"
	"google.golang.org/grpc"
)

var stdServeMux = http.NewServeMux()

var stdServer = httpframework.NewServer(&http.Server{
	Handler: stdServeMux,
})

func AddGrpcServerOption(opts ...grpc.ServerOption) {
	stdServer.AddGrpcServerOption(opts...)
}
func SetGrpcServerOption(opts ...grpc.ServerOption) {
	stdServer.SetGrpcServerOption(opts...)
}

func AddGatewayRequestMatcher(matchers ...httpframework.GatewayRequestMatcherFunc) {
	stdServer.AddGatewayRequestMatcher(matchers...)
}
func SetGatewayRequestMatcher(matchers ...httpframework.GatewayRequestMatcherFunc) {
	stdServer.SetGatewayRequestMatcher(matchers...)
}

func Handle(pattern string, handler http.Handler) {
	stdServeMux.Handle(pattern, handler)
}
func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	stdServeMux.HandleFunc(pattern, handler)
}

func RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	stdServer.RegisterService(desc, impl)
}
func RegisterGatewayService(f httpframework.GatewayRegisterFunc) {
	stdServer.RegisterGatewayService(f)
}
