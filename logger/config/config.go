package config

import (
	mconfig "github.com/sparrow-community/pkgs/config"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/logger"
)

var (
	Conf = &Config{
		Server: mconfig.Server{
			Name: "logger",
		},
		Logger: Logger{
			Dir: "./logs",
		},
	}
)

type Logger struct {
	Dir string `json:"dir"`
}

type Config struct {
	Server mconfig.Server `json:"server"`
	Logger Logger         `json:"logger"`
}

// Init .
func (c *Config) Init() error {
	mc, err := mconfig.New(
		mconfig.WithDefaultConfig(c),
		mconfig.WithFlags(
			&cli.StringFlag{Name: "logger_dir", Usage: "logger directory", EnvVars: []string{"LOGGER_DIR"}},
		),
	)
	if err != nil {
		return err
	}

	if err := mc.Scan(&c); err != nil {
		return err
	}

	logger.Infof("Read config: %+#v", c)

	return nil
}
