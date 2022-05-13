package httpframework

import "google.golang.org/grpc"

type GrpcService struct {
	// pb.Xxx_ServiceDesc
	ServiceDesc *grpc.ServiceDesc
	// Service implementation
	ServiceImpl interface{}
	// pb.RegisterXxxHandler
	GatewayRegister GrpcGatewayRegisterFunc
	// Gateway request matchers
	GatewayMatchers []Matcher
}

// create a service
func Service(
	desc *grpc.ServiceDesc, impl interface{},
) GrpcService {
	return GrpcService{
		ServiceDesc: desc,
		ServiceImpl: impl,
	}
}

// create service with gateway
func ServiceWithGateway(
	desc *grpc.ServiceDesc, impl interface{},
	register GrpcGatewayRegisterFunc, matchers ...Matcher,
) GrpcService {
	return GrpcService{
		ServiceDesc:     desc,
		ServiceImpl:     impl,
		GatewayRegister: register,
		GatewayMatchers: matchers,
	}
}
