package logger

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Level string `envconfig:"LEVEL" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("LOG", &config); err != nil {
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
