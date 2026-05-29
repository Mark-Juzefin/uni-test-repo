package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds Products service configuration.
type Config struct {
	Port      int    `env:"PORT" envDefault:"3000"`
	PgURL     string `env:"PG_URL" required:"true"`
	PgPoolMax int    `env:"PG_POOL_MAX" envDefault:"10"`
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`

	OutboxPollInterval time.Duration `env:"OUTBOX_POLL_INTERVAL" envDefault:"2s"`
	OutboxBatchSize    int           `env:"OUTBOX_BATCH_SIZE" envDefault:"100"`
}

func New() (Config, error) {
	return env.ParseAs[Config]()
}
