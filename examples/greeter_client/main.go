package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func SayHelloGrpc() {
	conn, err := grpc.Dial("127.0.0.1:50050", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("failed to dial: %s\n", err)
		return
	}

	cli := helloworld.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req := &helloworld.HelloRequest{Name: "Test"}

	for i := 0; i < 10; i++ {
		res, err := cli.SayHello(ctx, req)
		if err != nil {
			fmt.Printf("failed to say hello: %s\n", err)
			return
		}

		fmt.Printf("say hello reply: %s\n", res)
	}
}

func main() {
	SayHelloGrpc()
}
