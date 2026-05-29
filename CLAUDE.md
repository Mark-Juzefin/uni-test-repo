# CLAUDE.md

Task spec: [task.md](task.md). Remaining work + plan: [TODO.md](TODO.md).

## Architecture

Go workspace (`go.work`), isolated modules.

```
pkg/                postgres (pgxpool+Squirrel), logger (slog), migrations (goose)
services/products/  cmd, app.go (bootstrap+DI), router, config, migrations
  internal/product/ entity, interfaces, service, errors, productrepo (pg), productcontroller (http)
```

Layers: controller (HTTP, errors→status) → service (logic) → repo (interface in domain, pg impl).

## Commands

```bash
make up / down / run / test
make seed [n=50]   # sample products for pagination
```

Build from module dir: `cd services/products && go build ./...` (root isn't a workspace module).

## Gotchas
- `go 1.26.0` everywhere; never `require uni-test-repo/pkg` (dotless path — workspace resolves it).
- Postgres uses named volume `psql_data` (not bind mount — NTFS/WSL perms).
- LocalStack disabled; pin `localstack/localstack:3.8` when adding SQS.
- `price` = int64 minor units. Pagination is lenient (service clamps, no 400).
