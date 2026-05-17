# Godec

Godec is a Mux-inspired media processing service.

The objetive is to create a full-fledged group of microservices for handling upload, processing, orchestration, storage and streaming of both images and videos.

## Tech

- `Echo`
- `pgx`
- `godotenv` + `caarlos0's env`
- [`validator`](https://github.com/go-playground/validator)

## Environment Sync

The environment schema lives in [internal/config/config.go](internal/config/config.go).

Lefthook is used to keep `.env` and `.env.example` aligned with that source of truth through a pre-commit `envsync` task.

# Prerequisites

## Install Lefthook

```bash
go install github.com/evilmartians/lefthook@latest
lefthook install
```

### Sync env files locally

```bash
go run ./ci/envsync/cmd fix
```

### Check env files without mutating them

```bash
go run ./ci/envsync/cmd check
```

### Check env files in GitHub Actions

```bash
go run ./ci/envsync/cmd check --file .env.example
```

### Notes

- `fix` reads [internal/config/config.go](internal/config/config.go) first, then creates or appends missing keys in `.env` and `.env.example`.
- `check` only reports missing or stale keys. In GitHub Actions, pass `--file .env.example` because `.env` is intentionally local-only.

## Install SQLC

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Set up [`sqlc.yaml`](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html)

## Install [goose](https://github.com/pressly/goose)

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Set environment variables

```
# Example
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://admin:admin@localhost:5432/admin_db
GOOSE_MIGRATION_DIR=./migrations
GOOSE_TABLE=custom.goose_migrations
```
