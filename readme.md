# MVP for Godec

Godec is a Mux-inspired media processing service.

The objetive is to create a full-fledged group of microservices for handling upload, processing, orchestration, storage and streaming of both images and videos.

## Tech

- `Echo`
- [`validator`](https://github.com/go-playground/validator)

## Environment Sync

The environment schema lives in [internal/config/config.go](internal/config/config.go).

Lefthook is used to keep `.env` and `.env.example` aligned with that source of truth through a pre-commit `envsync` task.

### Install Lefthook

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
