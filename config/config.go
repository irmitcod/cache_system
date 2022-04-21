package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	Host          string `env:"HOST" env-default:"0.0.0.0"`
	Port          int    `env:"PORT" env-default:"8080"`
	Prefix        string `env:"PREFIX" env-default:"/argos/"`
	RedisAddress  string `env:"REDIS_ADDRESS" env-default:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" env-default:""`
	RedisDB       int    `env:"REDIS_DB" env-default:"0"`
	MaxWidth      int    `env:"MAX_WIDTH" env-default:"120"`
	MaxHeight     int    `env:"MAX_HEIGHT" env-default:"120"`
}

func GetConfig() *config {
	var cfg config
	cleanenv.ReadEnv(&cfg)
	return &cfg
}
