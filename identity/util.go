package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sparrow-community/app/identity/config"
	"github.com/sparrow-community/pkgs/auth"
	"github.com/sparrow-community/protos/cache"
	"github.com/sparrow-community/protos/identity"
	"go-micro.dev/v4/errors"
	"google.golang.org/protobuf/proto"
	"net/http"
	"time"
)

const (
	CacheIdentityUserSession = "cache:identity:user:%s:session"
	SessionToken             = "session_token"
	AccessToken              = "access_token"
)

func ResponseFilter(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
	switch v := message.(type) {
	case *identity.UserSignInResponse:
		return writeSession(ctx, writer, v.AccessToken)
	case *identity.RefreshTokenResponse:
		return writeSession(ctx, writer, v.AccessToken)
	}
	delete(writer.Header(), "Grpc-Metadata-Content-Type")
	return nil
}

func writeSession(ctx context.Context, writer http.ResponseWriter, accessToken string) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}
	userIds := md.HeaderMD.Get("x-user-id")
	if len(userIds) == 0 {
		return errors.InternalServerError("identity:signIn", "generate session token error, user id not exist")
	}
	duration, err := time.ParseDuration(config.Conf.Auth.Session.Duration)
	if err != nil {
		return errors.InternalServerError("identity:signIn", "generate session token duration error, %s", err)
	}
	token, err := config.Conf.Auth.Session.Sess.N.Generate(
		auth.WithSubject(userIds[0]),
		auth.WithClaims(map[string]any{}),
		auth.WithIssuer(config.Conf.Auth.Issuer),
		auth.WithExpiration(duration),
	)
	if err != nil {
		return errors.InternalServerError("identity:signIn", "generate session token error, %s", err)
	}
	config.Conf.Auth.Session.Sess.WriteSessionCookie(ctx, writer, token.AccessToken)
	key := fmt.Sprintf(CacheIdentityUserSession, userIds[0])
	_, err = config.Conf.Cache.HSetMap(ctx, &cache.HSetMapRequest{Key: key, Value: map[string]string{
		SessionToken: token.AccessToken,
		AccessToken:  accessToken,
	}})
	if err != nil {
		return errors.InternalServerError("identity:signIn", "cache token error, %s", err)
	}
	delete(writer.Header(), "Grpc-Metadata-X-User-Id")
	return nil
}
