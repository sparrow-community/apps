package main

import (
	mcgrpc "github.com/go-micro/plugins/v4/client/grpc"
	msgrpc "github.com/go-micro/plugins/v4/server/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/sparrow-community/app/logger/config"
	"github.com/sparrow-community/app/logger/handler"
	"github.com/sparrow-community/pkgs/listener"
	"github.com/sparrow-community/protos/logger"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"net"
	"net/http"
)

var (
	Version = "Development"
	Commit  = "Development"
	Date    = "Now"
	BuiltBy = "Development"
)

func main() {
	logger.Infof("%s %s %s %s", Version, Commit, Date, BuiltBy)

	err := config.Conf.Init()
	if err != nil {
		logger.Errorf("read config error: %v", err)
		return
	}

	lst, err := listener.New()
	if err != nil {
		logger.Errorf("error creating listener: %v", err)
	}

	opts := []micro.Option{
		micro.Name(config.Conf.Server.Name),
		micro.Client(mcgrpc.NewClient()),
		micro.Version(Version),
	}

	go httpServer(lst.Http(), opts...)
	go grpcServer(lst.Grpc(), opts...)

	if err := lst.Serve(); err != nil {
		logger.Errorf("server error: %v", err)
	}
}

func httpServer(lst net.Listener, opts ...micro.Option) {
	m := http.NewServeMux()
	m.Handle("/hello", http.HandlerFunc(handler.DefaultHttpHandler.Welcome))
	m.Handle("/logger/hello", http.HandlerFunc(handler.DefaultHttpHandler.Welcome))

	httpServer := mhttp.NewServer(
		server.Name(config.Conf.Server.Name),
		mhttp.Listener(lst),
	)
	if err := httpServer.Handle(httpServer.NewHandler(m)); err != nil {
		logger.Errorf("error creating http server: %", err)
	}
	httpOpts := append(opts, micro.Server(httpServer))
	srv := micro.NewService(httpOpts...)
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

func grpcServer(lst net.Listener, opts ...micro.Option) {
	grpcOpts := append(opts, micro.Server(msgrpc.NewServer(
		server.Name(config.Conf.Server.Name),
		msgrpc.Listener(lst),
	)))
	srv := micro.NewService(grpcOpts...)

	if err := proto.RegisterLoggerHandler(srv.Server(), handler.NewLoggerService(config.Conf.Logger.Dir)); err != nil {
		logger.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
