-include .env
export

.PHONY: up down run test

up:
	docker compose up -d --wait

down:
	docker compose down --remove-orphans

run:
	go run ./services/products/cmd

test:
	go test -race ./...
