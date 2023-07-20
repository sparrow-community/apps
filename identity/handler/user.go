package handler

import (
	"context"
	"encoding/base64"
	"github.com/go-playground/validator/v10"
	mdAuth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/sparrow-community/app/identity/config"
	"github.com/sparrow-community/app/identity/secret"
	"github.com/sparrow-community/pkgs/auth"
	"github.com/sparrow-community/protos/id"
	"github.com/sparrow-community/protos/identity"
	"go-micro.dev/v4/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"time"
)

type UserServiceHandler struct {
	IDService      id.IdService
	PasswordSecret secret.PasswordSecret
	Validate       *validator.Validate
	DB             *gorm.DB
	N              *auth.Authenticate
}

func (h UserServiceHandler) SignUp(ctx context.Context, in *identity.UserSignUpRequest, _ *emptypb.Empty) error {
	// validate
	h.Validate.RegisterStructValidationMapRules(getSignUpRule(in.Type), &identity.UserSignUpRequest{})
	if err := h.Validate.Struct(in); err != nil {
		return errors.BadRequest("identity:signup", "%s", err)
	}

	now := time.Now().Unix()
	uc := &UserCredential{Type: in.Type, Identifier: in.Identifier, Verified: false, VerificationAt: now, CreatedAt: now}
	if exists, err := uc.IdentifierExists(); err != nil {
		return errors.InternalServerError("identity:signup", "identifier error %s", err)
	} else if exists {
		return errors.BadRequest("identity:signup", "identifier %s already exists ", uc.Identifier)
	}

	// id
	idRsp, err := h.IDService.Generate(ctx, &id.GenerateRequest{Type: id.Types_UUID})
	if err != nil {
		return err
	}
	user := &User{ID: idRsp.Id, Nickname: "", Avatar: "", DisabledAt: 0, DeletedAt: 0, CreatedAt: now}
	uc.UserID = user.ID
	if uc.Type == identity.UserCredentialType_USERNAME {
		uc.Verified = true
	}
	// password
	slat, err := h.PasswordSecret.RandomSalt()
	if err != nil {
		return errors.InternalServerError("identity:signup", "user slat error %s", err)
	}
	sp, err := secret.DefaultPasswordSecret.Secret([]byte(in.Password), slat)
	if err != nil {
		return errors.InternalServerError("identity:signup", "user secret password error %s", err)
	}
	uc.Salt = slat
	uc.SecretData = base64.RawStdEncoding.EncodeToString(sp)

	return h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return errors.BadRequest("identity:signup", "user error %s", err)
		}
		if err := tx.Create(uc).Error; err != nil {
			return errors.BadRequest("identity:signup", "user credential error %s", err)
		}
		return nil
	})
}

func (h UserServiceHandler) SignIn(ctx context.Context, in *identity.UserSignInRequest, out *identity.UserSignInResponse) error {
	// validate
	h.Validate.RegisterStructValidationMapRules(getSignUpRule(in.Type), &identity.UserSignInRequest{})
	if err := h.Validate.Struct(in); err != nil {
		return errors.BadRequest("identity:signIn", "%s", err)
	}

	uc := &UserCredential{Identifier: in.Identifier}
	if err := uc.FindByIdentifier(); err != nil {
		return errors.BadRequest("identity:signIn", "%s", err)
	}
	secretedBytes, err := base64.RawStdEncoding.DecodeString(uc.SecretData)
	if err != nil {
		return errors.BadRequest("identity:signIn", "%s", err)
	}
	if b, err := h.PasswordSecret.Matches([]byte(in.Password), uc.Salt, secretedBytes); err != nil {
		return errors.BadRequest("identity:signIn", "%s", err)
	} else if !b {
		return errors.BadRequest("identity:signIn", "Incorrect username or password.")
	}

	duration, err := time.ParseDuration(config.Conf.Auth.Token.Duration)
	if err != nil {
		return errors.InternalServerError("identity:signIn", "generate token duration error, %s", err)
	}

	token, err := h.N.Generate(
		auth.WithSubject(uc.UserID),
		auth.WithClaims(map[string]any{}),
		auth.WithIssuer(config.Conf.Auth.Issuer),
		auth.WithExpiration(duration),
	)
	if err != nil {
		return errors.InternalServerError("identity:signIn", "generate token error, %s", err)
	}
	_ = grpc.SendHeader(ctx, metadata.Pairs("x-user-id", uc.UserID))
	*out = identity.UserSignInResponse{
		AccessToken: token.AccessToken,
		Schema:      "bearer",
		Exp:         token.Exp.UnixMilli(),
		Iat:         token.Iat.UnixMilli(),
		Iss:         token.Iss,
	}

	return nil
}

func (h UserServiceHandler) RefreshToken(ctx context.Context, in *identity.RefreshTokenRequest, out *identity.RefreshTokenResponse) error {
	accessToken, err := mdAuth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return errors.Forbidden("identity:signIn", "access token error, %s", err)
	}

	// access token is correct
	ot, err := h.N.Parse([]byte(accessToken))
	if err != nil {
		return errors.Forbidden("identity:signIn", "access token error, %s", err)
	}

	duration, err := time.ParseDuration(config.Conf.Auth.Token.Duration)
	if err != nil {
		return errors.InternalServerError("identity:signIn", "generate token duration error, %s", err)
	}
	token, err := h.N.Generate(
		auth.WithSubject(ot.Subject),
		auth.WithClaims(map[string]any{}),
		auth.WithIssuer(config.Conf.Auth.Issuer),
		auth.WithExpiration(duration),
	)
	if err != nil {
		return errors.InternalServerError("identity:signIn", "generate token error, %s", err)
	}

	_ = grpc.SendHeader(ctx, metadata.Pairs("x-user-id", ot.Subject))
	*out = identity.RefreshTokenResponse{
		AccessToken: token.AccessToken,
		Schema:      "bearer",
		Exp:         token.Exp.UnixMilli(),
		Iat:         token.Iat.UnixMilli(),
		Iss:         token.Iss,
	}

	return nil
}
