package rabbitmq

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"HOST"`
	Port     string `envconfig:"PORT"`
	UserName string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
}

func ReadConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
