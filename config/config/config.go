package config

import (
	mconfig "github.com/sparrow-community/pkgs/config"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4/logger"
)

var (
	Conf = &Config{
		Server: mconfig.Server{
			Name: "config",
		},
		Configs: Configs{
			Path: "./conf",
		},
	}
)

type Configs struct {
	Path string `json:"path"`
}

type Config struct {
	Server  mconfig.Server `json:"server"`
	Configs Configs        `json:"configs"`
}

// Init .
func (c *Config) Init() error {
	mc, err := mconfig.New(
		mconfig.WithDefaultConfig(c),
		mconfig.WithFlags(
			&cli.StringFlag{Name: "configs_path", Usage: "config files directory", EnvVars: []string{"CONFIGS_PATH"}},
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
