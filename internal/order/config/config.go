package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseDSN string        `envconfig:"ORDER_DATABASE_DSN" required:"true"`
	Port        string        `envconfig:"ORDER_PORT" required:"true"`
	Timeout     time.Duration `envconfig:"ORDER_TIMEOUT" default:"10s"`

	CatalogServiceAddr    string        `envconfig:"CATALOG_SERVICE_ADDR" required:"true"`
	CatalogServiceTimeout time.Duration `envconfig:"CATALOG_SERVICE_TIMEOUT" default:"5s"`
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("process config: %w", err)
	}
	return cfg, nil
}

func MustLoad() Config {
	cfg, err := New()
	if err != nil {
		panic(err)
	}
	return cfg
}
