package main

import (
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/sparrow-community/app/gateway/config"
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

	httpServer := mhttp.NewServer(
		server.Name(config.Conf.Server.Name),
		mhttp.Listener(l),
	)
	if err := httpServer.Handle(httpServer.NewHandler(ReverseProxy())); err != nil {
		logger.Errorf("error creating http server: %", err)
	}
	var opts []micro.Option
	httpOpts := append(opts, micro.Server(httpServer))
	srv := micro.NewService(httpOpts...)
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
