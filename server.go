package httpframework

import (
	"context"
	"net"
	"net/http"
	"sync"

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

type GatewayRegisterFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

type GatewayRequestMatcherFunc func(r *http.Request) bool

func gatewayRequestContentTypeMatcher(r *http.Request) bool {
	return r.Header.Get("Content-Type") == "application/grpc-gateway"
}

type grpcService struct {
	desc *grpc.ServiceDesc
	impl interface{}
}
type gatewayService struct {
	f GatewayRegisterFunc
}

type Server struct {
	// options from http.Server both for http and grpc-gateway
	httpS *http.Server

	// options from grpc.Server
	grpcServerOptions []grpc.ServerOption

	// match grpc gateway requests
	gatewayRequestMatchers []GatewayRequestMatcherFunc

	grpcServices []*grpcService
	gateServices []*gatewayService

	ctx  context.Context
	quit context.CancelFunc
	wg   sync.WaitGroup
}

func NewServer(httpS *http.Server) *Server {
	srv := &Server{httpS: httpS}
	srv.ctx, srv.quit = context.WithCancel(context.Background())
	return srv
}

func (srv *Server) AddGrpcServerOption(opts ...grpc.ServerOption) {
	srv.grpcServerOptions = append(srv.grpcServerOptions, opts...)
}
func (srv *Server) SetGrpcServerOption(opts ...grpc.ServerOption) {
	srv.grpcServerOptions = opts
}

func (srv *Server) AddGatewayRequestMatcher(matchers ...GatewayRequestMatcherFunc) {
	srv.gatewayRequestMatchers = append(srv.gatewayRequestMatchers, matchers...)
}
func (srv *Server) SetGatewayRequestMatcher(matchers ...GatewayRequestMatcherFunc) {
	srv.gatewayRequestMatchers = matchers
}

func (srv *Server) ListenAndServe() error {
	a := srv.httpS.Addr
	if a == "" {
		a = ":http"
	}
	l, err := net.Listen("tcp", a)
	if err != nil {
		return err
	}
	defer l.Close()
	return srv.Serve(l)
}

func (srv *Server) Serve(l net.Listener) error {
	m := cmux.New(l)
	m.HandleError(func(err error) bool {
		log.Warnf("cmux handle error: %v", err)
		return true
	})

	if len(srv.grpcServices) > 0 {
		grpcS := grpc.NewServer(
			append(srv.grpcServerOptions,
				grpc.ChainUnaryInterceptor(interceptServiceUnary),
				grpc.ChainStreamInterceptor(interceptServiceStream),
			)...)
		grpc_health_v1.RegisterHealthServer(grpcS, health.NewServer())
		for _, e := range srv.grpcServices {
			grpcS.RegisterService(e.desc, e.impl)
		}

		srv.wg.Add(1)
		go func(l net.Listener) {
			defer srv.wg.Done()
			grpcS.Serve(l)
		}(m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings(
			"content-type", "application/grpc")))

		defer grpcS.GracefulStop()

		if len(srv.gateServices) > 0 {
			var a = l.Addr().String()
			if a[0] == ':' {
				a = "localhost" + a
			}
			grpcC, err := grpc.Dial(a, grpc.WithInsecure())
			if err != nil {
				return err
			}

			gateM := runtime.NewServeMux()
			for _, e := range srv.gateServices {
				e.f(context.TODO(), gateM, grpcC)
			}
			if srv.httpS.Handler == nil {
				srv.httpS.Handler = gateM
			} else {
				httpM := srv.httpS.Handler

				gatewayRequestMatchers := append(
					[]GatewayRequestMatcherFunc{},
					srv.gatewayRequestMatchers...)

				srv.httpS.Handler = http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						if gatewayRequestContentTypeMatcher(r) {
							gateM.ServeHTTP(w, r)
							return
						}
						for _, f := range gatewayRequestMatchers {
							if f(r) {
								gateM.ServeHTTP(w, r)
								return
							}
						}
						httpM.ServeHTTP(w, r)
					})
			}
		}
	}

	srv.wg.Add(1)
	go func(l net.Listener) {
		defer srv.wg.Done()
		srv.httpS.Serve(l)
	}(m.Match(cmux.Any()))

	defer srv.httpS.Shutdown(context.Background())

	go m.Serve()

	<-srv.ctx.Done()

	return nil
}

func (srv *Server) Shutdown(ctx context.Context) (err error) {
	srv.quit()

	done := make(chan struct{}, 1)
	go func() { srv.wg.Wait(); done <- struct{}{} }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return
	}
}

// Implement grpc.ServiceRegistrar interface
func (srv *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	srv.grpcServices = append(srv.grpcServices, &grpcService{desc: desc, impl: impl})
}

// Register grpc gateway services
func (srv *Server) RegisterGatewayService(f GatewayRegisterFunc) {
	srv.gateServices = append(srv.gateServices, &gatewayService{f: f})
}
