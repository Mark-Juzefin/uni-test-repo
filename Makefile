-include .env
export

.PHONY: up down run test seed

up:
	docker compose up -d --wait

down:
	docker compose down --remove-orphans

run:
	go run ./services/products/cmd

test:
	go test -race ./...

# Seed sample products to test list pagination. Override count with: make seed n=50
seed:
	bash ./scripts/seed.sh $(n)
