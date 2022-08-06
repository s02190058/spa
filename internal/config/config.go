package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Server   `yaml:"server"`
		Static   `yaml:"static"`
		Postgres `yaml:"postgres"`
		Logger   `yaml:"logger"`
		JWT      `yaml:"jwt"`
		Hasher   `yaml:"hasher"`
	}

	Server struct {
		Port            string        `yaml:"port" env:"SRV_PORT"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SRV_SHUTDOWN_TIMEOUT"`
	}

	Static struct {
		Path  string `yaml:"path" env:"STATIC_PATH"`
		Index string `yaml:"index" env:"STATIC_INDEX"`
	}

	Postgres struct {
		Username     string        `yaml:"username" env:"PG_USERNAME"`
		Password     string        `env:"PG_PASSWORD"`
		Host         string        `yaml:"host" env:"PG_HOST"`
		Port         string        `yaml:"port" env:"PG_PORT"`
		Database     string        `yaml:"database" env:"PG_DATABASE"`
		SSLMode      string        `yaml:"sslmode" env:"PG_SSLMODE"`
		ConnAttempts int           `yaml:"conn_attempts" env:"PG_CONN_ATTEMPTS"`
		ConnTimeout  time.Duration `yaml:"conn_timeout" env:"PG_CONN_TIMEOUT"`
		MaxOpenConns int           `yaml:"max_open_conns" env:"PG_MAX_OPEN_CONNS"`
	}

	Logger struct {
		Level string `yaml:"level" env:"LOG_LEVEL"`
	}

	JWT struct {
		SigningKey string        `env:"JWT_SIGNING_KEY"`
		TokenTTL   time.Duration `yaml:"token_ttl" env:"JWT_TOKEN_TTL"`
	}

	Hasher struct {
		Cost int `env:"HASHER_COST"`
	}
)

func New(path string) (*Config, error) {
	cfg := new(Config)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
