package main

import (
	"context"
	"fmt"
	"net/http"
	"syscall"

	"github.com/onemoreteam/http_server_framework"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

type GreeterServer struct {
	helloworld.UnimplementedGreeterServer
}

func (gs *GreeterServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + req.Name}, nil
}

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	srv := http_server_framework.NewServer(&http.Server{
		Addr:    ":50050",
		Handler: m,
	})

	helloworld.RegisterGreeterServer(srv, &GreeterServer{})

	//srv.RegisterGatewayService(helloworld.RegisterGreeterHandler)

	fmt.Println(srv.ListenAndServe())

	http_server_framework.IgnoreSignal(syscall.SIGPIPE)

	http_server_framework.WatchSignal(syscall.SIGINT, syscall.SIGTERM).Wait()

	srv.Shutdown(context.Background())
}
