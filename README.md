# Products & Notifications

Two Go microservices. **Products** is a REST API that writes each create/delete to its
DB and an outbox table in one transaction; an outbox worker relays events to Kafka;
**Notifications** consumes and logs them.

```
Products API ──tx──▶ Postgres (products + outbox) ──worker──▶ Kafka ──▶ Notifications (log)
```

Stack: Go 1.26 workspace · Gin · pgx + Squirrel · goose · Kafka (Redpanda).

## Run

Needs Docker, Go 1.26+, and `make`.

```bash
cp .env.example .env
make up          # Postgres, Redpanda (+Console), Prometheus
make run-all     # API + outbox worker + notifications in one terminal (Ctrl+C stops all)
```

```bash
curl -X POST localhost:3000/products -H 'Content-Type: application/json' \
  -d '{"name":"Widget","price":1999}'
curl 'localhost:3000/products?limit=10&offset=0'
curl -X DELETE localhost:3000/products/<id>
```

Notifications logs each event; browse the topic at <http://localhost:8080> (Redpanda Console).

## API

- `POST /products` — `{name, description?, price}`, `price` = int64 minor units → `201`
- `GET /products?limit&offset` — limit default 20 / max 100, clamped (never 400)
- `DELETE /products/:id` — `204`, or `404` if missing

## Other

- `make test` — unit tests (create / list / delete)
- `make seed [n=50]` — POST sample products (API must be running)
- `make run` / `run-worker` / `run-notifications` — run one service
- Config: see [.env.example](.env.example)
