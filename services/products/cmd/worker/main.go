package main

import (
	"log"

	"uni-test-repo/services/products"
	"uni-test-repo/services/products/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	products.RunWorker(cfg)
}
