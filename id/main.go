package main

import (
	msgrpc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/sparrow-community/app/id/config"
	"github.com/sparrow-community/app/id/handler"
	"github.com/sparrow-community/protos/id"
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
	logger.Infof("%s %s %s %s", Version, Commit, Date, BuiltBy)

	err := config.Conf.Init()
	if err != nil {
		logger.Errorf("read config error: %v", err)
		return
	}

	l, err := net.Listen("tcp", config.Conf.Server.Address)
	if err != nil {
		logger.Fatal(err)
		return
	}

	opts := []micro.Option{
		micro.Name(config.Conf.Server.Name),
		micro.Version(Version),
	}

	grpcServer(l, opts...)
}

func grpcServer(lst net.Listener, opts ...micro.Option) {
	grpcOpts := append(opts, micro.Server(msgrpc.NewServer(
		server.Name(config.Conf.Server.Name),
		msgrpc.Listener(lst),
	)))
	srv := micro.NewService(grpcOpts...)

	if err := id.RegisterIdHandler(srv.Server(), &handler.Id{}); err != nil {
		logger.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
