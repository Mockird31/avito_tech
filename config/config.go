package config

import (
	"errors"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port     int `env:"PORT,required"`
	Postgres PostgresConfig
}

type PostgresConfig struct {
	PostgresHost     string `env:"POSTGRES_HOST,required"`
	PostgresPort     string `env:"POSTGRES_PORT,required"`
	PostgresUser     string `env:"POSTGRES_USER,required"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,required"`
	PostgresDB       string `env:"POSTGRES_DB,required"`
	MaxOpenConns     int    `env:"POSTGRES_MAX_OPEN_CONNS,required"`
	MaxIdleConns     int    `env:"POSTGRES_MAX_IDLE_CONNS,required"`
	MaxLifetime      int    `env:"POSTGRES_MAX_LIFE_TIME,required"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, errors.New("parse config error: " + err.Error())
	}
	return cfg, nil
}
