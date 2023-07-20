package main

import (
	mcgrpc "github.com/go-micro/plugins/v4/client/grpc"
	msgrpc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/sparrow-community/app/cache/config"
	"github.com/sparrow-community/app/cache/handler"
	"github.com/sparrow-community/pkgs/listener"
	"github.com/sparrow-community/protos/cache"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"net"
)

var (
	Version = "Development"
	Commit  = "Development"
	Date    = "Now"
	BuiltBy = "Development"
)

func main() {
	err := config.Conf.Init()
	if err != nil {
		logger.Errorf("read config error: %v", err)
		return
	}

	logger.Infof("%s %s %s %s", Version, Commit, Date, BuiltBy)

	lst, err := listener.New(
		listener.WithAddress(config.Conf.Server.Address),
	)
	if err != nil {
		logger.Errorf("error creating listener: %v", err)
	}

	opts := []micro.Option{
		micro.Name(config.Conf.Server.Name),
		micro.Client(mcgrpc.NewClient()),
		micro.Version(Version),
	}

	go grpcServer(lst.Grpc(), opts...)

	_ = lst.Serve()
}

func grpcServer(lst net.Listener, opts ...micro.Option) {
	grpcOpts := append(opts, micro.Server(msgrpc.NewServer(
		server.Name(config.Conf.Server.Name),
		msgrpc.Listener(lst),
	)))
	srv := micro.NewService(grpcOpts...)

	cacheHandler := &handler.Cache{
		Client: config.Conf.RedisClient,
	}
	if err := cache.RegisterCacheHandler(srv.Server(), cacheHandler); err != nil {
		logger.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
