package grpc_server_framework

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/cockroachdb/cmux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	// options from http.Server
	http.Server

	// options from grpc.Server
	grpcServerOptions []grpc.ServerOption

	httpServer http.Server

	grpcServer *grpc.Server

	grpcConn *grpc.ClientConn

	gatewayMux *runtime.ServeMux

	grpcServices    []*grpcService
	gatewayServices []*gatewayService

	ctx  context.Context
	quit context.CancelFunc
}

func (srv *Server) WithGrpcServerOptions(opts ...grpc.ServerOption) {
	srv.grpcServerOptions = append(srv.grpcServerOptions, opts...)
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(l net.Listener) error {
	m := cmux.New(l)

	var wg sync.WaitGroup
	defer wg.Wait()

	grpcL := m.Match(cmux.HTTP2HeaderField(
		"content-type", "application/grpc"))
	grpcS := grpc.NewServer(srv.grpcServerOptions...)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	for _, e := range srv.grpcServices {
		grpcS.RegisterService(e.desc, e.impl)
	}
	defer grpcS.GracefulStop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcS.Serve(grpcL)
	}()

	if len(srv.gatewayServices) > 0 {
		ipcF, err := ioutil.TempFile("/tmp/", "*.sock")
		if err != nil {
			return err
		}
		ipcL, err := net.Listen("unix", ipcF.Name())
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			grpcS.Serve(ipcL)
		}()

		ipc, err := grpc.Dial(ipcF.Name(), grpc.WithInsecure())
		if err != nil {
			return err
		}

		gatewayL := m.Match(cmux.HTTP1HeaderField(
			"content-type", "application/grpc-gateway"))
		gatewayM := runtime.NewServeMux()
		for _, e := range srv.gatewayServices {
			e.f(context.TODO(), gatewayM, ipc)
		}
		gatewayS := &http.Server{Handler: gatewayM}
		defer gatewayS.Shutdown(context.Background())

		wg.Add(1)
		go func() {
			defer wg.Done()
			gatewayS.Serve(gatewayL)
		}()
	}

	if srv.Handler != nil {
		httpL := m.Match(cmux.HTTP1Fast())
		httpS := &http.Server{Handler: srv.Handler}
		defer httpS.Shutdown(context.Background())
		wg.Add(1)
		go func() {
			defer wg.Done()
			httpS.Serve(httpL)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Serve()
	}()

	<-srv.ctx.Done()

	return nil
}

func (srv *Server) Close() (err error) {
	srv.quit()
	return
}

func (srv *Server) Shutdown(ctx context.Context) (err error) {
	srv.quit()
	return
}

// Implement grpc.ServiceRegistrar interface
func (srv *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	srv.grpcServices = append(
		srv.grpcServices, &grpcService{desc: desc, impl: impl})
}

func (srv *Server) RegisterGatewayService(f GatewayRegisterFunc) {
	srv.gatewayServices = append(srv.gatewayServices, &gatewayService{f: f})
}
