package server

import (
	"net/http"

	"github.com/onemoreteam/httpframework"
	"google.golang.org/grpc"
)

var (
	stdServeMux          *http.ServeMux = nil
	stdGrpcServerOptions []grpc.ServerOption
	stdGrpcServices      []httpframework.GrpcService
)

func WithGrpcServerOption(opts ...grpc.ServerOption) {
	stdGrpcServerOptions = opts
}

func Handle(pattern string, handler http.Handler) {
	if stdServeMux == nil {
		stdServeMux = http.NewServeMux()
	}
	stdServeMux.Handle(pattern, handler)
}
func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if stdServeMux == nil {
		stdServeMux = http.NewServeMux()
	}
	stdServeMux.HandleFunc(pattern, handler)
}

func RegisterGrpcService(
	desc *grpc.ServiceDesc, impl interface{}) {
	stdGrpcServices = append(stdGrpcServices, httpframework.Service(desc, impl))
}
func RegisterGrpcServiceWithGateway(desc *grpc.ServiceDesc, impl interface{}, register httpframework.GrpcGatewayRegisterFunc, matchers ...httpframework.Matcher) {
	stdGrpcServices = append(stdGrpcServices, httpframework.ServiceWithGateway(desc, impl, register, matchers...))
}
