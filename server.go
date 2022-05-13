package httpframework

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ntons/log-go"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	// register compressors
	_ "github.com/ntons/grpc-compressor/lz4"
	_ "google.golang.org/grpc/encoding/gzip"
)

type GrpcGatewayRegisterFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

type grpcService struct {
	desc *grpc.ServiceDesc
	impl interface{}
}
type grpcGatewayService struct {
	register GrpcGatewayRegisterFunc
	matchers []Matcher
}

type Server struct {
	httpS *http.Server

	// options from grpc.Server
	grpcServerOptions []grpc.ServerOption

	// registered grpc services
	grpcServices []GrpcService

	// has grpc-gateway service registered
	hasGrpcGatewayService bool

	//
	mux cmux.CMux
}

// 为了更灵活的配置http.Server，允许这个对象直接从外部传进来
func New(httpS *http.Server, grpcServerOptions ...grpc.ServerOption) *Server {
	s := &Server{
		httpS: httpS,
		grpcServerOptions: append(grpcServerOptions,
			grpc.ChainUnaryInterceptor(interceptServiceUnary),
			grpc.ChainStreamInterceptor(interceptServiceStream),
		),
	}
	return s
}

// close server
func (s *Server) Close() (err error) {
	if s.mux != nil {
		s.mux.Close()
	}
	return
}

// Register grpc service
func (s *Server) RegisterGrpcService(services ...GrpcService) (err error) {
	for _, x := range services {
		if x.ServiceDesc == nil {
			return fmt.Errorf("service desc is required")
		}
		if x.ServiceImpl == nil {
			return fmt.Errorf("service impl is required")
		}
		if x.GatewayRegister != nil {
			s.hasGrpcGatewayService = true
		}
	}
	s.grpcServices = append(s.grpcServices, services...)
	return
}

func (s *Server) ListenAndServe() error {
	addr := s.httpS.Addr
	if addr == "" {
		addr = ":http"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	return s.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	s.mux = cmux.New(l)
	s.mux.HandleError(func(err error) bool {
		log.Warnf("cmux handle error: %v", err)
		return true
	})

	// serve grpc-gateway
	if s.hasGrpcGatewayService {
		// listen on unix domain socket
		const UnixSock = "/tmp/httpframework.sock"
		_l, err := net.Listen("unix", UnixSock)
		if err != nil {
			return fmt.Errorf("failed to listen on unix domain socket: %v", err)
		}
		defer _l.Close()

		_grpcS := grpc.NewServer(s.grpcServerOptions...)
		for _, x := range s.grpcServices {
			if x.GatewayRegister != nil {
				_grpcS.RegisterService(x.ServiceDesc, x.ServiceImpl)
			}
		}

		wg.Add(1)
		go func(l net.Listener) {
			defer wg.Done()
			_grpcS.Serve(l)
		}(_l)
		defer _grpcS.GracefulStop()

		grpcC, err := grpc.Dial("unix://"+UnixSock, grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer grpcC.Close()

		grpcGatewayH := runtime.NewServeMux()
		for _, x := range s.grpcServices {
			if x.GatewayRegister != nil {
				x.GatewayRegister(context.TODO(), grpcGatewayH, grpcC)
			}
		}

		if s.httpS.Handler == nil || reflect.ValueOf(s.httpS.Handler).IsNil() {
			s.httpS.Handler = grpcGatewayH
		} else {
			httpH := s.httpS.Handler

			s.httpS.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if grpcGatewayContentTypeMatcher(r) {
					grpcGatewayH.ServeHTTP(w, r)
					return
				}
				for _, x := range s.grpcServices {
					for _, x := range x.GatewayMatchers {
						if x(r) {
							grpcGatewayH.ServeHTTP(w, r)
							return
						}
					}
				}
				httpH.ServeHTTP(w, r)
			})
		}
	}

	// serve grpc
	grpcS := grpc.NewServer(s.grpcServerOptions...)
	grpc_health_v1.RegisterHealthServer(grpcS, health.NewServer())
	for _, x := range s.grpcServices {
		grpcS.RegisterService(x.ServiceDesc, x.ServiceImpl)
	}

	wg.Add(1)
	go func(l net.Listener) {
		defer wg.Done()
		grpcS.Serve(l)
	}(s.mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc")))
	defer grpcS.GracefulStop()

	// serve http
	wg.Add(1)
	go func(l net.Listener) {
		defer wg.Done()
		s.httpS.Serve(l)
	}(s.mux.Match(cmux.HTTP1Fast()))
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.httpS.Shutdown(ctx)
	}()

	s.mux.Serve()
	return nil
}
