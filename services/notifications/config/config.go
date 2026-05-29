package config

import "github.com/caarlos0/env/v11"

type Config struct {
	KafkaBrokers []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:19092"`
	KafkaTopic   string   `env:"KAFKA_TOPIC" envDefault:"product-events"`
	GroupID      string   `env:"KAFKA_GROUP_ID" envDefault:"notifications"`
	LogLevel     string   `env:"LOG_LEVEL" envDefault:"info"`
}

func New() (Config, error) {
	return env.ParseAs[Config]()
}
