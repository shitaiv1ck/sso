package postgres

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	User     string        `envconfig:"USER" reqiured:"true"`
	Password string        `envconfig:"PASSWORD" reqiured:"true"`
	Host     string        `envconfig:"HOST" reqiured:"true"`
	Port     string        `envconfig:"PORT" reqiured:"true"`
	DB       string        `envconfig:"DB" reqiured:"true"`
	Timeout  time.Duration `envconfig:"TIMEOUT" reqiured:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("POSTGRES", &config); err != nil {
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
