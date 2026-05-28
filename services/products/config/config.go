package config

import (
	"github.com/caarlos0/env/v11"
)

// Config holds Products service configuration.
type Config struct {
	Port      int    `env:"PORT" envDefault:"3000"`
	PgURL     string `env:"PG_URL" required:"true"`
	PgPoolMax int    `env:"PG_POOL_MAX" envDefault:"10"`
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
}

func New() (Config, error) {
	return env.ParseAs[Config]()
}
