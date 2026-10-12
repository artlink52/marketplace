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
	var config Config
	if err := envconfig.Process("CATALOG", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	return config, nil
}

func MustLoad() Config {
	cfg, err := New()
	if err != nil {
		panic(err)
	}
	return cfg
}
