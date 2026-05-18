# Godec

Godec is a Mux-inspired media processing service.

The objetive is to create a full-fledged group of microservices for handling upload, processing, orchestration, storage and streaming of both images and videos.

## Tech

- `Echo`
- `pgx`
- `godotenv` + `caarlos0's env`
- [`validator`](https://github.com/go-playground/validator)

## Architecture Walkthrough

See [docs/architecture-walkthrough.md](docs/architecture-walkthrough.md) for the package-by-feature layout, runtime flow, and maintenance rules used by the new architecture.

## API Code Generation

See [docs/api-codegen-walkthrough.md](docs/api-codegen-walkthrough.md) for a guide on the OpenAPI specification, oapi-codegen setup, and development flow for adding or modifying endpoints.

## Environment Sync

The environment schema lives in [internal/config/config.go](internal/config/config.go).

Lefthook is used to keep `.env` and `.env.example` aligned with that source of truth through a pre-commit `envsync` task.

## Tooling

Tool binaries are pinned in [tools/go.mod](tools/go.mod). You do not need to install `lefthook`, `sqlc`, or `goose` globally.

### Install Git hooks

```bash
make hooks-install
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

### Generate SQLC output

```bash
make sqlc
```

To verify the generated SQLC files are up to date, run:

```bash
make sqlc-check
```

### Generate the API server code

```bash
make api-codegen
```

To verify the generated API file is up to date, run:

```bash
make codegen-check
```

### Set up [`sqlc.yaml`](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html)

### Goose migrations

```bash
make goose-new NAME=add_some_column
make goose-up
make goose-down
make goose-status
```

### Set environment variables

```
# Example
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://admin:admin@localhost:5432/admin_db
GOOSE_MIGRATION_DIR=./internal/postgres/migrations
GOOSE_TABLE=custom.goose_migrations
```
