# Godec — Full Codebase Walkthrough

> Last verified: 2026-07-18. Every claim below was corroborated against the actual source code.

---

## 1. The Big Picture

**Godec** is a Mux-inspired media processing service — a Go-based API server designed to eventually handle upload, processing, orchestration, storage, and streaming of images and videos. The long-term vision is a full-fledged group of microservices, but what exists today is the **foundational layer**: a multi-tenant API with API key authentication, built with clean architecture principles that make it straightforward to grow.

**Current state**: A modular monolith HTTP API server with two real features (Tenants, API Keys), a stub for media upload, health probes, and embedded OpenAPI documentation. It runs against a PostgreSQL database on Neon.

**Module path**: `github.com/c-mierez/godec`
**Go version**: 1.26.2

---

## 2. Architecture

### 2.1 Pattern: Package-by-Feature with Inverse Interfaces

The codebase follows **Clean / Hexagonal architecture** principles:

- **Feature packages** (`internal/<feature>/`) own the business language — models, service interfaces, and service implementations.
- **Infrastructure packages** (`internal/postgres/`, `internal/postgres/db/`) own persistence — they implement feature-defined interfaces implicitly.
- **A single composition root** (`cmd/api/main.go`) wires everything together.
- **Transport layer** (`internal/api/`) owns HTTP concerns — generated handlers, request/response shaping, data mapping.

The dependency direction is strictly one-way:

```
cmd/api/main.go  →  internal/api (handlers)
                 →  internal/apikey, internal/tenant (services)
                 →  internal/postgres (adapters)
                 →  internal/postgres/db (SQLC-generated queries)
```

Feature packages **never** import SQLC types or PostgreSQL-specific types. Mapping happens exclusively in `internal/postgres/mappers.go`.

### 2.2 Interface Style: Consumer-Defined, Implicitly Implemented

The codebase uses **inverse interfaces** — the consumer (service) declares what it needs, and the provider (postgres adapter) implements it implicitly without an explicit `implements` clause. Names describe the action, not the technology:

```go
// In internal/apikey/service.go — consumer defines:
type Store interface {
    CreateApiKey(ctx context.Context, tenantID uuid.UUID, name, hashedKey string, scopes []string) (*ApiKey, error)
    GetApiKeyByHashedKey(ctx context.Context, hashedKey string) (*ApiKey, error)
}

// In internal/postgres/api_key.go — provider satisfies it implicitly:
type ApiKeyStore struct { queries *db.Queries }
func (s *ApiKeyStore) CreateApiKey(...) (*apikey.ApiKey, error) { ... }
```

### 2.3 Runtime Request Flow

```
1.  main.go loads config (godotenv + caarlos0/env)
2.  main.go opens pgx connection pool
3.  main.go creates SQLC Queries object from db.New(pool)
4.  main.go wraps Queries in postgres adapters (ApiKeyStore, TenantStore)
5.  main.go wraps adapters in feature services (apikey.Service, tenant.Service)
6.  main.go creates Echo server with global middleware (CORS, logging, recovery)
7.  main.go loads embedded OpenAPI spec, creates echovalidator middleware with API key auth
8.  main.go creates api.Server (composes all feature handlers)
9.  main.go wraps Server in generated StrictHandler
10. main.go registers all routes via api.RegisterHandlers(echo, strictHandler)
11. Request arrives → Echo router → StrictHandler validates against OpenAPI spec
12. If route requires ApiKeyAuth → middleware.APIKeyAuthenticator validates key
13. StrictHandler calls Server method → delegates to feature handler
14. Feature handler calls service → service calls store interface
15. Postgres adapter satisfies store, calls SQLC queries, maps rows to domain structs
16. Response flows back: adapter → service → handler → StrictHandler → Echo → client
```

---

## 3. Package Map

### 3.1 Composition Root

| Package | Path | Purpose |
|---------|------|---------|
| `main` | `cmd/api/main.go` | The **only** place where concrete dependencies are created and wired. Loads config, opens DB, creates stores, services, handlers, registers routes, starts server. |

### 3.2 Feature Packages

| Package | Path | Models | Service Interface | Service Impl | Handler |
|---------|------|--------|-------------------|--------------|---------|
| `apikey` | `internal/apikey/` | `ApiKey` struct | `Store` interface | `Service` struct | In `internal/api/apikey_handlers.go` |
| `tenant` | `internal/tenant/` | `Tenant`, `TenantStatus` | `Store` interface | `Service` struct | In `internal/api/tenant_handlers.go` |

**apikey** is the clearest example of the intended style:
- `model.go` — pure domain struct (`ApiKey` with UUID, scopes, timestamps)
- `service.go` — `Store` interface + `Service` with `GenerateApiKey` and `ValidateAPIKey`
- `crypto.go` — key generation (32 bytes entropy, `sk_godec_` prefix, SHA-256 hashing, constant-time comparison)

**tenant** follows the same shape:
- `model.go` — `Tenant` struct with `TenantStatus` enum (`active`/`inactive`)
- `service.go` — `Store` interface with 9 methods (CRUD, list by status/email, count) + `Service` thin wrapper

### 3.3 Transport-Only Features (no domain package)

| Feature | File | Status |
|---------|------|--------|
| Health | `internal/api/health_handlers.go` | Dependency-free liveness/readiness probes. Returns static `{status: "ok", timestamp}` |
| Media | `internal/api/media_handlers.go` | **Stub** — returns hardcoded S3-like URL with `// TODO: Replace with actual S3 presigned URL generation` |
| Documentation | `internal/api/documentation_handlers.go` | Serves embedded `spec.yaml` and Scalar HTML docs |

### 3.4 Infrastructure

| Package | Path | Purpose |
|---------|------|---------|
| `postgres` | `internal/postgres/` | Concrete repository implementations: `ApiKeyStore`, `TenantStore`. Maps SQLC rows to domain structs via `mappers.go` |
| `db` | `internal/postgres/db/` | **SQLC-generated** code: `Queries` struct, row models, typed query methods. Includes hand-written `transformers.go` for UUID/timestamp conversions |
| `migrations` | `internal/postgres/migrations/` | Goose SQL migration files |

### 3.5 Support Code

| Package | Path | Purpose |
|---------|------|---------|
| `config` | `internal/config/` | Environment schema (`ServerEnv`, `DatabaseEnv`, `InternalEnv`, `Config`) loaded via `caarlos0/env` + `godotenv` |
| `middleware` | `internal/middleware/` | Global middleware (CORS, logging, recovery), API key auth (`APIKeyAuthenticator`, `APIKeyValidator`), `AuthError` type |
| `echovalidator` | `internal/middleware/echovalidator/` | Local wrapper around `openapi3filter` that plugs into Echo v5. Handles request validation, security scheme dispatch, error formatting |
| `apidoc` | `internal/apidoc/` | Embeds `scalar.html` for interactive API docs UI |
| `graceful` | `pkg/graceful/` | Graceful shutdown helper using `signal.NotifyContext` |
| `utils` | `pkg/utils/` | Single utility: `TimeNow()` returning RFC3339 UTC string |
| `envsync` | `ci/envsync/` | CLI tool that parses `config.go` AST to keep `.env` and `.env.example` in sync with the Go config schema |

---

## 4. Code Generation

### 4.1 API Code Generation (oapi-codegen)

**Source of truth**: `internal/api/spec.yaml` (OpenAPI 3.0 spec)
**Generator config**: `internal/api/codegen.yaml`
**Generated output**: `internal/api/gen.go` (DO NOT EDIT)
**Trigger**: `internal/api/generate.go` contains `//go:generate` directive

Configuration:
```yaml
package: api
output: gen.go
generate:
    echo5-server: true    # Echo v5 compatible handlers
    strict-server: true   # Type-safe request/response wrappers
    models: true          # Generate all request/response types
output-options:
    skip-prune: true
```

**Command**: `make api-codegen` or `go generate ./internal/api`
**Drift check**: `make codegen-check` (regenerates then `git diff --exit-code`)

What gets generated in `gen.go`:
- Type definitions for all request/response models
- `ServerInterface` — one method per operation
- `StrictHandler` — validates requests against spec, marshals typed responses
- `RegisterHandlers` — wires everything into Echo
- Response type wrappers (e.g., `CreateApiKey201JSONResponse`, `CreateApiKey400JSONResponse`)

### 4.2 SQL Code Generation (sqlc)

**Source of truth**: `internal/postgres/db/*.sql` (named queries)
**Schema**: `internal/postgres/db/schema.sql` (pg_dump of live DB)
**Config**: `sqlc.yaml`
**Generated output**: `internal/postgres/db/*.sql.go` + `models.go` (DO NOT EDIT)
**SQLC version**: v1.31.1
**SQL package**: `pgx/v5`

**Command**: `make sqlc` (dumps schema then generates)
**Drift check**: `make sqlc-check`

SQLC generates:
- `db.go` — `Queries` struct with `DBTX` interface
- `models.go` — row-level structs (`ApiKey`, `Tenant`) using `pgtype` types
- `api_keys.sql.go` — 14 typed query methods for api_keys table
- `tenants.sql.go` — 10 typed query methods for tenants table

The hand-written `transformers.go` provides conversion helpers between `pgtype` and domain types (`PgUUIDToUUID`, `UuidToPGUUID`, `PgTimestamptzToTime`, etc.).

### 4.3 Database Migrations (goose)

**Location**: `internal/postgres/migrations/`
**Tool**: goose v3 (pinned in `tools/go.mod`)
**Current migrations**: 3 files (tenants table, api_keys table, triggers)

Migrations create:
- `tenants` table with UUIDv7 primary key, status check constraint, unique email index (partial, active only)
- `api_keys` table with FK to tenants (CASCADE delete), hashed_key unique constraint, scopes array
- `update_updated_at_column()` trigger function applied to both tables
- `expires_at` column added to api_keys

---

## 5. Dependencies

### 5.1 Direct Dependencies (go.mod)

| Package | Purpose |
|---------|---------|
| `github.com/labstack/echo/v5` | HTTP framework (v5.1.0) |
| `github.com/jackc/pgx/v5` | PostgreSQL driver (pure Go, v5.9.2) |
| `github.com/joho/godotenv` | Loads `.env` files |
| `github.com/caarlos0/env/v11` | Struct-tag-based env parsing with validation |
| `github.com/getkin/kin-openapi` | OpenAPI 3.0 spec loading + request validation |
| `github.com/oapi-codegen/runtime` | Runtime types for oapi-codegen (UUID, Email types) |
| `github.com/google/uuid` | UUID generation and manipulation |

### 5.2 Tool Dependencies (tools/go.mod)

All pinned in a separate `tools` module — you do NOT install these globally:

| Tool | Purpose | Invocation |
|------|---------|------------|
| `github.com/sqlc-dev/sqlc/cmd/sqlc` v1.31.1 | SQL to Go codegen | `make sqlc` |
| `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` | OpenAPI to Go codegen | `make api-codegen` |
| `github.com/pressly/goose/v3/cmd/goose` | Database migrations | `make goose-up/down/new/status` |
| `github.com/evilmartians/lefthook/v2` | Git hooks manager | `make hooks-install` |

Tool invocation pattern in Makefile:
```makefile
GO_TOOL := go -C $(TOOLS_DIR) tool
SQLC := $(GO_TOOL) sqlc
GOOSE := $(GO_TOOL) goose
```

### 5.3 Indirect/Transitive Dependencies

Key indirect deps worth knowing:
- `github.com/jackc/puddle/v2` — connection pool for pgx
- `github.com/go-openapi/jsonpointer`, `swag` — OpenAPI spec parsing
- `github.com/santhosh-tekuri/jsonschema/v6` — JSON schema validation
- `golang.org/x/net`, `sync`, `text`, `time` — standard extended library

---

## 6. Database Schema

Two tables, both using UUIDv7 primary keys:

### tenants
| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, default `uuidv7()` |
| name | TEXT | NOT NULL |
| email | TEXT | NOT NULL, unique per active tenant |
| status | TEXT | NOT NULL, CHECK `active`/`inactive`, default `active` |
| created_at | TIMESTAMPTZ | NOT NULL, default `NOW()` |
| updated_at | TIMESTAMPTZ | NOT NULL, auto-updated by trigger |

Indexes:
- `idx_tenants_unique_email` — UNIQUE on `email` WHERE `status = 'active'`
- `idx_tenants_active` — on `id` WHERE `status = 'active'`

### api_keys
| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, default `uuidv7()` |
| tenant_id | UUID | FK → tenants.id ON DELETE CASCADE |
| name | TEXT | NOT NULL |
| hashed_key | TEXT | NOT NULL, UNIQUE |
| scopes | TEXT[] | NOT NULL, default `'{}'` |
| created_at | TIMESTAMPTZ | NOT NULL, default `NOW()` |
| updated_at | TIMESTAMPTZ | NOT NULL, auto-updated by trigger |
| last_used_at | TIMESTAMPTZ | NULLABLE |
| expires_at | TIMESTAMPTZ | NULLABLE |

Indexes:
- `idx_api_keys_hashed` — on `hashed_key`
- `idx_api_keys_tenant_id` — on `tenant_id`

---

## 7. API Endpoints

Defined in `internal/api/spec.yaml`, implemented via generated strict-server pattern:

| Method | Path | Operation | Auth | Description |
|--------|------|-----------|------|-------------|
| GET | `/live` | `liveness` | No | Liveness probe |
| GET | `/ready` | `readiness` | No | Readiness probe |
| GET | `/spec.yaml` | `getOpenAPISpec` | No | Returns raw OpenAPI YAML |
| GET | `/docs/api` | `getAPIDocs` | No | Scalar HTML API docs UI |
| POST | `/v1/tenants` | `createTenant` | No | Create tenant (name, email) |
| GET | `/v1/tenants` | `listTenants` | No | List tenants (paginated: limit/offset) |
| GET | `/v1/tenants/{id}` | `getTenant` | No | Get tenant by UUID |
| PATCH | `/v1/tenants/{id}/status` | `setTenantStatus` | No | Update tenant status |
| POST | `/v1/apikey/create_key` | `createApiKey` | No | Generate API key for tenant |
| POST | `/v1/media/upload-url` | `getMediaUploadURL` | **ApiKeyAuth** | Get presigned upload URL (stub) |

**Only `/v1/media/upload-url` requires authentication** via `X-API-Key` header. All other endpoints are currently unprotected.

### Auth Flow

1. Spec declares `security: - ApiKeyAuth: []` on the operation
2. `echovalidator` middleware intercepts, calls `openapi3filter.ValidateRequest`
3. `openapi3filter` calls `middleware.APIKeyAuthenticator` for `ApiKeyAuth` scheme
4. Authenticator reads `X-API-Key` header
5. Calls `middleware.APIKeyValidator` → `apikey.Service.ValidateAPIKey`
6. Service hashes key with SHA-256, looks up by hash, validates with constant-time compare
7. Checks expiration if present
8. On success, stores `*apikey.ApiKey` in request context under `ContextKeyApiKey`
9. On failure, returns `AuthError` (401 or 403) which flows through centralized error handler in `main.go`

---

## 8. Environment Configuration

Source of truth: `internal/config/config.go` struct tags.

| Variable | Default | Description |
|----------|---------|-------------|
| `ENV` | `development` | Runtime environment |
| `DATABASE_URL` | (none, required) | PostgreSQL connection string |
| `SERVER_ADDRESS` | `127.0.0.1:8080` | HTTP listen address |
| `CORS_ALLOWED_ORIGINS` | `http://127.0.0.1:8080,http://localhost:8080` | CORS origins |
| `GOOSE_DRIVER` | (none) | Goose migration driver |
| `GOOSE_DBSTRING` | (none) | Goose connection string |
| `GOOSE_MIGRATION_DIR` | `internal/postgres/migrations` | Goose migration dir |
| `GOOSE_TABLE` | (none) | Goose migrations table |

The `.env` file contains a live Neon PostgreSQL connection string. It is gitignored.

---

## 9. CI/CD

### GitHub Actions (`.github/workflows/ci.yml`)

Triggers on: push to `main`/`mvp`, all pull requests.

**Job 1: `test`**
1. Checkout + setup Go (version from go.mod)
2. Check `gofmt` formatting
3. `go test ./...`
4. `make codegen-check` (API codegen drift detection)
5. `make sqlc-check` (SQLC codegen drift detection)
6. `go mod tidy` + git diff check (ensures go.mod/go.sum are clean)

**Job 2: `envsync`**
1. Checkout + setup Go
2. `go run ./ci/envsync/cmd check --file .env.example` (validates .env.example matches config.go)

### Lefthook (pre-commit hooks)

Configured in `lefthook.yml`:

**On commit**:
1. `envsync fix` — auto-syncs `.env` and `.env.example` with `config.go`
2. `gofmt -w` — formats staged Go files
3. `go mod tidy` — tidies module files

---

## 10. Your Regular Workflow

### First Time Setup

```bash
# 1. Install git hooks (envsync + gofmt + go mod tidy on commit)
make hooks-install

# 2. Set up your .env (copy from .env.example and fill in DATABASE_URL)
cp .env.example .env
# Edit .env with your PostgreSQL connection string

# 3. Run migrations against your database
make goose-up

# 4. Dump schema (needed before sqlc generate)
make dump

# 5. Generate all code
make sqlc
make api-codegen
```

### Daily Development

```bash
# Start the server
go run ./cmd/api

# Server starts on http://127.0.0.1:8080
# API docs at http://127.0.0.1:8080/docs/api
# Raw spec at http://127.0.0.1:8080/spec.yaml
```

### Adding a New Feature

1. Create `internal/<feature>/` with `model.go` and `service.go`
2. Define `Store` interface in service.go
3. Add SQL queries to `internal/postgres/db/<feature>.sql`
4. Run `make sqlc` to generate query code
5. Implement `Store` in `internal/postgres/<feature>.go`
6. Add mappers in `internal/postgres/mappers.go`
7. Add endpoints to `internal/api/spec.yaml`
8. Run `make api-codegen` to regenerate handlers
9. Implement handler in `internal/api/<feature>_handlers.go`
10. Wire in `internal/api/server.go` and `cmd/api/main.go`
11. Run `go build ./...` and `go test ./...`

### Adding a New Database Query

1. Add query to `internal/postgres/db/*.sql` with sqlc comment annotations
2. Run `make sqlc`
3. Add/update mapping code in `internal/postgres/mappers.go`
4. Keep feature packages free of SQLC types

### Adding a New Endpoint

1. Edit `internal/api/spec.yaml` (add path, operation, schemas)
2. Run `make api-codegen`
3. Implement the generated method in the appropriate handler file
4. Wire in `server.go` if new feature
5. Run `make codegen-check` to verify drift

### After Any Schema Change

```bash
# Create migration
make goose-new NAME=description_of_change

# Edit the generated migration file in internal/postgres/migrations/

# Apply migration
make goose-up

# Dump updated schema
make dump

# Regenerate SQLC
make sqlc
```

### Verification Commands

```bash
go build ./...          # Compile check
go test ./...           # Run all tests
make codegen-check      # API codegen up to date?
make sqlc-check         # SQLC codegen up to date?
go run ./ci/envsync/cmd check   # .env files in sync?
gofmt -l .              # Any unformatted files?
```

### Key Rules to Remember

- `cmd/api/main.go` is the **only** composition root
- Feature packages must **never** import `internal/postgres/db` directly
- Feature packages must **never** import PostgreSQL or SQLC types
- Domain structs use `uuid.UUID` and `*time.Time` — not `pgtype` types
- Mapping happens **only** in `internal/postgres/mappers.go`
- `internal/api/gen.go` is **generated** — never edit manually
- `internal/postgres/db/*.sql.go` is **generated** — never edit manually
- The `tools/go.mod` pins tool versions — don't install tools globally

---

## 11. Testing

### Current Test Coverage

| File | Tests | What's Tested |
|------|-------|---------------|
| `internal/middleware/auth_test.go` | 4 tests | `APIKeyAuthenticator`: ignores other schemes, returns missing key error, returns expired key error, sets key in context |
| `pkg/graceful/graceful_test.go` | 4 tests | Graceful shutdown: execution called, shutdown called after execution, concurrent invocations, nil execution panics |
| `pkg/utils/utils_test.go` | 1 test | `TimeNow()` returns valid RFC3339 |
| `ci/envsync/internal/envsync_test.go` | 3 tests | Schema loading from AST, fix creates/appends missing keys, check reports missing/stale keys |

**No integration tests, no e2e tests, no HTTP handler tests yet.** The `openspec/config.yaml` reflects this: `integration: false`, `e2e: false`.

### Running Tests

```bash
go test ./...
go test -cover ./...    # With coverage
```

---

## 12. File Structure Reference

```
godec/
├── cmd/api/main.go                          # Composition root
├── internal/
│   ├── api/                                 # Transport layer
│   │   ├── spec.yaml                        # OpenAPI 3.0 spec (source of truth)
│   │   ├── codegen.yaml                     # oapi-codegen config
│   │   ├── generate.go                      # go:generate directive
│   │   ├── gen.go                           # GENERATED — do not edit
│   │   ├── server.go                        # Server composition (implements ServerInterface)
│   │   ├── health_handlers.go               # Liveness/readiness
│   │   ├── tenant_handlers.go               # Tenant CRUD handlers
│   │   ├── apikey_handlers.go               # API key creation handler
│   │   ├── media_handlers.go                # Media upload stub
│   │   └── documentation_handlers.go        # OpenAPI spec + Scalar UI serving
│   ├── apikey/                              # API key feature
│   │   ├── model.go                         # ApiKey domain struct
│   │   ├── service.go                       # Store interface + Service
│   │   └── crypto.go                        # Key generation + hashing
│   ├── tenant/                              # Tenant feature
│   │   ├── model.go                         # Tenant domain struct + TenantStatus enum
│   │   └── service.go                       # Store interface + Service
│   ├── postgres/                            # Infrastructure
│   │   ├── api_key.go                       # ApiKeyStore (implements apikey.Store)
│   │   ├── tenant.go                        # TenantStore (implements tenant.Store)
│   │   ├── mappers.go                       # SQLC row → domain struct mapping
│   │   ├── db/                              # SQLC-generated
│   │   │   ├── db.go                        # Queries struct, DBTX interface
│   │   │   ├── models.go                    # Row-level structs (pgtype)
│   │   │   ├── transformers.go              # Hand-written pgtype ↔ domain conversions
│   │   │   ├── schema.sql                   # pg_dump of live schema
│   │   │   ├── api_keys.sql                 # SQLC query definitions for api_keys
│   │   │   ├── api_keys.sql.go              # GENERATED
│   │   │   ├── tenants.sql                  # SQLC query definitions for tenants
│   │   │   └── tenants.sql.go              # GENERATED
│   │   └── migrations/                      # Goose SQL migrations
│   │       ├── 20260517144737_tenants.sql
│   │       ├── 20260517152622_create_api_keys.sql
│   │       └── 20260517153339_..._triggers.sql
│   ├── config/
│   │   └── config.go                        # Environment schema + Load()
│   ├── middleware/
│   │   ├── auth.go                          # APIKeyAuthenticator (OpenAPI auth func)
│   │   ├── auth_test.go                     # 4 tests
│   │   ├── validator.go                     # APIKeyValidator adapter
│   │   ├── global.go                        # CORS, logging, recovery middleware
│   │   └── echovalidator/
│   │       └── oapi_validate.go             # Local OpenAPI request validator for Echo v5
│   └── apidoc/
│       ├── embed.go                         # Embeds scalar.html
│       └── scalar.html                      # Scalar API docs UI
├── pkg/
│   ├── graceful/
│   │   ├── graceful.go                      # Graceful shutdown helper
│   │   └── graceful_test.go                 # 4 tests
│   └── utils/
│       ├── utils.go                         # TimeNow()
│       └── utils_test.go                    # 1 test
├── ci/envsync/                              # Environment file sync tool
│   ├── cmd/main.go                          # CLI entry point (fix/check)
│   └── internal/
│       ├── envsync.go                       # Core logic (AST parser, sync, check)
│       └── envsync_test.go                  # 3 tests
├── tools/
│   ├── go.mod                               # Tool version pinning module
│   └── go.sum
├── .github/workflows/ci.yml                 # CI pipeline
├── lefthook.yml                             # Git hooks config
├── Makefile                                 # Build/ship commands
├── sqlc.yaml                                # SQLC configuration
├── go.mod / go.sum                          # Application dependencies
├── .env.example                             # Template environment
├── .env                                     # Local environment (gitignored)
└── openspec/config.yaml                     # SDD context configuration
```

---

## 13. Key Gotchas and Non-Obvious Things

1. **UUIDv7**: The database uses `uuidv7()` for primary keys (time-ordered UUIDs). The Go side uses `uuid.UUID` from `github.com/google/uuid`. Conversion goes through `db.PgUUIDToUUID` / `db.UuidToPGUUID` in `transformers.go`.

2. **API key prefix**: Plain keys use the format `sk_godec_<base64>`. Only the SHA-256 hash is stored in the database. The plain key is shown **once** at creation time.

3. **`envsync` parses Go AST**: The `ci/envsync` tool literally parses `config.go` as a Go AST to extract `env` struct tags. If you add a new env var, you add the struct field with tags, then run `envsync fix` — it auto-generates the `.env` and `.env.example` entries.

4. **Centralized auth error handler**: Auth errors bypass the generated response types. They flow through `main.go`'s custom `HTTPErrorHandler` which catches `*middleware.AuthError` and formats it as JSON. This is intentional — see the comment block in `main.go` lines 48-59.

5. **Schema dump is manual**: `make dump` runs `pg_dump` against your live database. The schema dump feeds into `sqlc generate`. If you change the schema via migrations but don't run `make dump`, sqlc will generate against the old schema.

6. **Two Go modules**: The application has `go.mod` at root and `tools/go.mod` in the tools directory. The tools module exists so that `go -C tools tool` can invoke pinned tool versions without polluting the application's dependency graph.

7. **The `tests/` directory is empty**: No test fixtures or test databases are set up yet.

8. **`.env` contains live credentials**: Your `.env` file has a real Neon PostgreSQL connection string. It is gitignored, but be careful not to commit it or expose it.

9. **Media is a stub**: `GetMediaUploadURL` returns a hardcoded URL. No S3 integration exists yet. When real media orchestration is added, it should become its own `internal/media` feature package.

10. **No linter configured**: There is no `golangci-lint` or similar. The only quality gates are `gofmt` (in CI and pre-commit) and `go vet` (implicit via `go build`).
