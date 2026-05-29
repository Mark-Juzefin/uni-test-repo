-include .env
export

.PHONY: up down run run-worker test seed

up:
	docker compose up -d --wait

down:
	docker compose down --remove-orphans

run:
	go run ./services/products/cmd/api

run-worker:
	go run ./services/products/cmd/worker

test:
	go test -race ./...

# Seed sample products to test list pagination. Override count with: make seed n=50
seed:
	bash ./scripts/seed.sh $(n)
