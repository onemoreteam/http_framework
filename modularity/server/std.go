package server

import (
	"net/http"

	"github.com/onemoreteam/httpframework"
	"google.golang.org/grpc"
)

var Std = httpframework.NewServer(&http.Server{})

func RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	Std.RegisterService(desc, impl)
}

func RegisterGatewayService(f httpframework.GatewayRegisterFunc) {
	Std.RegisterGatewayService(f)
}
