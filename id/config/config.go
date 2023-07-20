package config

import (
	mconfig "github.com/sparrow-community/pkgs/config"
	"go-micro.dev/v4/logger"
)

var (
	Conf = &Config{
		Server: mconfig.Server{
			Name:    "id",
			Address: ":",
		},
	}
)

type Config struct {
	Server mconfig.Server `json:"server"`
}

// Init .
func (c *Config) Init() error {
	mc, err := mconfig.New(
		mconfig.WithDefaultConfig(c),
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
