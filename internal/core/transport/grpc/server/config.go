package grpcserver

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr    string        `envconfig:"ADDR" required:"true"`
	Timeout time.Duration `envconfig:"TIMEOUT" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("GRPC", &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	return config
}
