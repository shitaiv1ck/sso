package redis

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" reqiured:"true"`
	Port     string        `envconfig:"PORT" reqiured:"true"`
	Password string        `envconfig:"PASSWORD" reqiured:"true"`
	DB       int           `envconfig:"DB" reqiured:"true"`
	Timeout  time.Duration `envconfig:"TIMEOUT" reqiured:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("REDIS", &config); err != nil {
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
