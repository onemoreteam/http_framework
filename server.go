package http_server_framework

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GatewayRegisterFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

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
	GrpcServerOptions []grpc.ServerOption

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
		return true
	})

	if len(srv.grpcServices) > 0 {
		grpcS := grpc.NewServer(srv.GrpcServerOptions...)
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
				srv.httpS.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Content-Type") == "application/grpc-gateway" {
						gateM.ServeHTTP(w, r)
					} else {
						httpM.ServeHTTP(w, r)
					}
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
