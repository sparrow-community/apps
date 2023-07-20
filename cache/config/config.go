package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	mconfig "github.com/sparrow-community/pkgs/config"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/logger"
)

var (
	Conf = &Config{
		Server: mconfig.Server{
			Name:    "cache",
			Address: ":",
		},
		Redis: Redis{
			Addr: "localhost:6379",
		},
	}
)

type Redis struct {
	Addr string `json:"addr"`
}

type Config struct {
	Server mconfig.Server `json:"server"`
	Redis  Redis          `json:"redis"`

	RedisClient *redis.Client `json:"-"`
}

// Init .
func (c *Config) Init() error {
	mc, err := mconfig.New(
		mconfig.WithDefaultConfig(c),
		mconfig.WithFlags(
			&cli.StringFlag{Name: "redis_addr", Usage: "redis address", EnvVars: []string{"REDIS_ADDR"}},
		),
	)
	if err != nil {
		return err
	}

	if err := mc.Scan(&c); err != nil {
		return err
	}

	logger.Infof("Read config: %+#v", c)

	c.InitRedis()

	return nil
}

func (c *Config) InitRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	c.RedisClient = rdb
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatalf("connection redis error: %s", err)
	}
}
