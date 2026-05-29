package main

import (
	"log"

	"uni-test-repo/services/notifications"
	"uni-test-repo/services/notifications/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	notifications.Run(cfg)
}
