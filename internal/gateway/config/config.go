package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTP    HTTPConfig
	Infra   InfraConfig
	Clients ClientsConfig
}

type HTTPConfig struct {
	HTTPPort    string        `envconfig:"HTTP_PORT" default:"8080"`
	HTTPTimeout time.Duration `envconfig:"HTTP_TIMEOUT" default:"5s"`
}

type InfraConfig struct {
	JWTSecret string        `envconfig:"JWT_SECRET" required:"true"`
	JWTTTL    time.Duration `envconfig:"JWT_TTL" default:"24h"`
}

type ClientsConfig struct {
	UserServiceAddr    string        `envconfig:"USER_SERVICE_ADDR" required:"true"`
	UserServiceTimeout time.Duration `envconfig:"USER_SERVICE_TIMEOUT" default:"10s"`
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("GATEWAY", &cfg); err != nil {
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
