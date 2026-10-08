package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseDSN string        `envconfig:"DATABASE_DSN" required:"true"`
	Port        string        `envconfig:"PORT" required:"true"`
	Timeout     time.Duration `envconfig:"TIMEOUT" default:"10s"`
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("USER", &cfg); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return cfg, nil
}

func MustLoad() Config {
	config, err := New()
	if err != nil {
		panic(err)
	}
	return config
}
