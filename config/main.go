package main

import (
	mcgrpc "github.com/go-micro/plugins/v4/client/grpc"
	msgrpc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/sparrow-community/app/config/config"
	"github.com/sparrow-community/app/config/handler"
	lg "github.com/sparrow-community/plugins/v4/logger/grpc"
	"github.com/sparrow-community/protos/config"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
)

var (
	// build parameters
	version = "Development"
	commit  = "Development"
	date    = "Now"
	builtBy = "Development"
)

func main() {
	err := config.Conf.Init()
	if err != nil {
		logger.Errorf("read config error: %v", err)
		return
	}

	client.DefaultClient = mcgrpc.NewClient()
	lg.InitializeLogger(config.Conf.Server.Name)
	logger.Infof("%s %s %s %s", version, commit, date, builtBy)

	srv := micro.NewService(
		micro.Client(mcgrpc.NewClient()),
		micro.Server(msgrpc.NewServer()),
		micro.Name(config.Conf.Server.Name),
		micro.Version(version),
	)

	fs, err := handler.NewFileService(config.Conf.Configs.Path)
	if err != nil {
		logger.Fatal(err)
	}

	if err := proto.RegisterSourceHandler(srv.Server(), fs); err != nil {
		logger.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
