package main

import (
	"context"
	mcgrpc "github.com/go-micro/plugins/v4/client/grpc"
	msgrpc "github.com/go-micro/plugins/v4/server/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sparrow-community/app/identity/config"
	"github.com/sparrow-community/app/identity/handler"
	"github.com/sparrow-community/app/identity/secret"
	"github.com/sparrow-community/pkgs/listener"
	"github.com/sparrow-community/protos/cache"
	"github.com/sparrow-community/protos/id"
	"github.com/sparrow-community/protos/identity"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	go httpServer(lst.Http(), opts...)

	_ = lst.Serve()
}

func httpServer(lst net.Listener, opts ...micro.Option) {
	mux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(ResponseFilter),
	)
	grpcOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := lst.Addr().String()
	err := identity.RegisterUserServiceHandlerFromEndpoint(context.Background(), mux, endpoint, grpcOpts)
	if err != nil {
		logger.Fatalf("initial grpc http proxy error %s", err)
	}

	httpServer := mhttp.NewServer(
		server.Name(config.Conf.Server.Name),
		mhttp.Listener(lst),
	)
	if err := httpServer.Handle(httpServer.NewHandler(mux)); err != nil {
		logger.Errorf("error creating http server: %", err)
	}
	httpOpts := append(opts, micro.Server(httpServer))
	srv := micro.NewService(httpOpts...)
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

func grpcServer(lst net.Listener, opts ...micro.Option) {
	grpcOpts := append(opts,
		micro.Server(
			msgrpc.NewServer(
				server.Name(config.Conf.Server.Name),
				msgrpc.Listener(lst),
				msgrpc.Options(
					grpc.ChainUnaryInterceptor(
						selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
							//auth.AuthFromMD()
							return ctx, nil
						}), selector.MatchFunc(func(ctx context.Context, callMeta interceptors.CallMeta) bool {
							return true
						})),
					),
					grpc.ChainStreamInterceptor(
						selector.StreamServerInterceptor(auth.StreamServerInterceptor(func(ctx context.Context) (context.Context, error) {
							return ctx, nil
						}), selector.MatchFunc(func(ctx context.Context, callMeta interceptors.CallMeta) bool {
							return true
						})),
					),
				),
			),
		),
	)
	srv := micro.NewService(grpcOpts...)

	// Init auth from cache server
	config.Conf.Cache = cache.NewCacheService("cache", srv.Client())
	if err := config.Conf.InitAuth(config.Conf.Cache); err != nil {
		logger.Fatal(err)
	}

	user := &handler.UserServiceHandler{
		IDService:      id.NewIdService("id", srv.Client()),
		PasswordSecret: secret.DefaultPasswordSecret,
		DB:             config.Conf.DB,
		Validate:       config.Conf.Validate,
		N:              config.Conf.Auth.Token.N,
	}
	if err := identity.MicroRegisterUserServiceHandler(srv.Server(), user); err != nil {
		logger.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
