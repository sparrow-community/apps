// Package main
package main

import (
	"context"
	msgrpc "github.com/go-micro/plugins/v4/server/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sparrow-community/pkgs/listener"
	hello "github.com/sparrow-community/protos/hello"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

type Say struct{}

func (s *Say) Hello(ctx context.Context, req *hello.Request, rsp *hello.Response) error {
	rsp.Msg = "Hello " + req.Name
	return nil
}

func (s *Say) Message(ctx context.Context, req *hello.Request, rsp *hello.Response) error {
	rsp.Msg = "Hello " + req.Name
	return nil
}

func main() {
	lst, err := listener.New(
		listener.WithAddress(":9090"),
	)
	if err != nil {
		logger.Errorf("error creating listener: %v", err)
	}

	go grpcServer(lst.Grpc())
	go httpServer(lst.Http())

	_ = lst.Serve()
}

func httpServer(lst net.Listener) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	endpoint := lst.Addr().String()

	err := hello.RegisterSayHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		logger.Error(err)
	}
	httpServer := mhttp.NewServer(
		server.Name("go.micro.srv.greeter"),
		mhttp.Listener(lst),
	)
	if err := httpServer.Handle(httpServer.NewHandler(mux)); err != nil {
		logger.Errorf("error creating http server: %", err)
	}
	var httpOpts []micro.Option
	httpOpts = append(httpOpts, micro.Server(httpServer))
	srv := micro.NewService(httpOpts...)
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

func grpcServer(lst net.Listener) {
	service := micro.NewService(
		micro.Server(msgrpc.NewServer(
			msgrpc.Listener(lst),
		)),
		micro.Name("go.micro.srv.greeter"),
	)

	// optionally setup command line usage
	service.Init()

	// Register Handlers

	_ = hello.MicroRegisterSayHandler(service.Server(), new(Say))
	// Run server
	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}
}
