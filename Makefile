-include .env
export

.PHONY: up down run run-worker run-notifications run-all test seed

up:
	docker compose up -d --wait

down:
	docker compose down --remove-orphans

run:
	go run ./services/products/cmd/api

run-worker:
	go run ./services/products/cmd/worker

run-notifications:
	go run ./services/notifications/cmd

# Bring up infra, then run all three services together (logs prefixed per process).
run-all: up
	go run github.com/mattn/goreman@latest start

test:
	go test -race ./...

# Seed sample products to test list pagination. Override count with: make seed n=50
seed:
	bash ./scripts/seed.sh $(n)
