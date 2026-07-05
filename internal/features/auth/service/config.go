package authsrvc

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	JwtTTL     time.Duration `envconfig:"JWT_TTL" required:"true"`
	SessionTTL time.Duration `envconfig:"SESSION_TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
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
