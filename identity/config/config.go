package config

import (
	"context"
	"embed"
	"github.com/go-playground/validator/v10"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/pressly/goose/v3"
	"github.com/sparrow-community/pkgs/auth"
	"github.com/sparrow-community/pkgs/auth/session"
	mconfig "github.com/sparrow-community/pkgs/config"
	"github.com/sparrow-community/protos/cache"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"regexp"
	"time"
)

var (
	Conf = &Config{
		Server: mconfig.Server{
			Name:    "identity",
			Address: ":9090",
		},
		Database: Database{
			DSN: "postgres://postgres:postgres@localhost:5432/postgres",
		},
		Auth: Auth{
			Issuer:  "sparrow-community",
			Session: Session{Duration: "168h0m0s"},
			Token:   Token{Duration: "24h0m0s"},
		},
	}
)

type Database struct {
	DSN string `json:"dsn"`
}

type Token struct {
	Duration string             `json:"duration"`
	N        *auth.Authenticate `json:"-"`
}

type Session struct {
	Duration string           `json:"duration"`
	Sess     *session.Session `json:"-"`
}

type Auth struct {
	Issuer  string  `json:"issuer"`
	Session Session `json:"session"`
	Token   Token   `json:"token"`
}

type Config struct {
	Server   mconfig.Server `json:"server"`
	Database Database       `json:"database"`
	Auth     Auth           `json:"auth"`

	Validate *validator.Validate `json:"-"`
	DB       *gorm.DB            `json:"-"`
	Cache    cache.CacheService  `json:"-"`
}

const (
	CacheIdentityRSASessionKey = "cache:identity:rsa:session"
	CacheIdentityRSATokenKey   = "cache:identity:rsa:token"
	CacheIdentityRSAPrivate    = "private_key"
	CacheIdentityRSAPublic     = "public_key"
	CacheIdentityRSAVersion    = "version"
)

// Init .
func (c *Config) Init() error {
	mc, err := mconfig.New(
		mconfig.WithDefaultConfig(c),
		mconfig.WithFlags(
			&cli.StringFlag{Name: "database_dsn", Usage: "database dns", EnvVars: []string{"DATABASE_DSN"}},
			&cli.StringFlag{Name: "auth_issuer", Usage: "authentication issuer", EnvVars: []string{"AUTH_ISSUER"}},
			&cli.StringFlag{Name: "auth_session_duration", Usage: "authentication session duration issuer", EnvVars: []string{"AUTH_SESSION_DURATION"}},
			&cli.StringFlag{Name: "auth_token_duration", Usage: "authentication token duration issuer", EnvVars: []string{"AUTH_TOKEN_DURATION"}},
		),
	)
	if err != nil {
		return err
	}

	if err := mc.Scan(&c); err != nil {
		return err
	}

	logger.Infof("Read config: %+#v", c)

	if err := c.InitDatabase(); err != nil {
		return err
	}

	if err := c.InitValidator(); err != nil {
		return err
	}

	return nil
}

//go:embed migrations/**/*.sql
var migrationsFs embed.FS

// databaseMigrate .
func (c *Config) databaseMigrate() error {
	goose.SetBaseFS(migrationsFs)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	db, err := Conf.DB.DB()
	if err != nil {
		return err
	}
	dbDir := "migrations/postgres"
	if err := goose.Up(db, dbDir); err != nil {
		_ = goose.Down(db, dbDir)
		logger.Fatalf("DB migrations error %s", err)
		return err
	}
	return nil
}

// InitDatabase .
func (c *Config) InitDatabase() error {
	db, err := gorm.Open(postgres.Open(c.Database.DSN), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{},
		Logger:         gl.Default.LogMode(gl.Info),
	})
	if err != nil {
		return err
	}
	c.DB = db
	return c.databaseMigrate()
}

// InitValidator .
func (c *Config) InitValidator() error {
	c.Validate = validator.New()
	err := c.Validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		hasUpper := regexp.MustCompile(`[A-Z]`)
		hasLower := regexp.MustCompile(`[a-z]`)
		hasDigit := regexp.MustCompile(`[0-9]`)
		hasSpecial := regexp.MustCompile(`[^A-Za-z0-9]`)

		return hasUpper.MatchString(password) &&
			hasLower.MatchString(password) &&
			hasDigit.MatchString(password) &&
			hasSpecial.MatchString(password)
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) InitAuth(client cache.CacheService) error {
	ctx := context.Background()

	getAuth := func(cacheKey string) (*auth.Authenticate, error) {
		ret, err := client.HGetAll(ctx, &cache.HGetAllRequest{Key: cacheKey})
		if err != nil {
			return nil, err
		}
		if len(ret.Value) <= 0 {
			authn, err := auth.New(auth.WithDefaultRsaKeyPair(true))
			if err != nil {
				return nil, err
			}
			_, err = client.HSetMap(ctx, &cache.HSetMapRequest{Key: cacheKey, Value: map[string]string{
				CacheIdentityRSAPrivate: string(authn.PrivateKeyBytes()),
				CacheIdentityRSAPublic:  string(authn.PublicKeyBytes()),
				CacheIdentityRSAVersion: time.Now().Format(time.RFC3339),
			}})
			if err != nil {
				return nil, err
			}
			return authn, nil
		}
		rsa := ret.Value
		authn, err := auth.New(
			auth.WithRsaPrivateKeyBytes([]byte(rsa[CacheIdentityRSAPrivate])),
			auth.WithRsaPublicKeyBytes([]byte(rsa[CacheIdentityRSAPublic])),
		)
		if err != nil {
			return nil, err
		}
		return authn, nil
	}

	if a, err := getAuth(CacheIdentityRSATokenKey); err != nil {
		return err
	} else {
		c.Auth.Token.N = a
	}
	if a, err := getAuth(CacheIdentityRSASessionKey); err != nil {
		return err
	} else {
		c.Auth.Session.Sess = session.New(
			session.WithCookieName("session_token"),
			session.WithAuth(a),
		)
	}

	return nil
}
