# Architecture Walkthrough

This repository now follows a package-by-feature structure with a single composition root and a strict split between domain code and PostgreSQL infrastructure.

## 1. The Big Picture

The application is organized around three rules:

1. `cmd/api/main.go` is the only place where concrete dependencies are created and wired together.
2. Each feature lives in its own `internal/<feature>/` package and owns its model, service, and HTTP handler.
3. All database code lives in `internal/postgres/` and `internal/postgres/db/`, with domain packages only seeing pure Go structs and consumer-defined interfaces.

The result is a codebase where the domain packages remain small, testable, and independent from database details.

## 2. Package Map

### Composition root

- [cmd/api/main.go](../cmd/api/main.go): loads config, creates the pgx pool, builds SQLC queries, instantiates postgres adapters, creates services, and registers HTTP handlers.

### Feature packages

- [internal/apikey](../internal/apikey): API key generation and validation.
- [internal/tenant](../internal/tenant): tenant business logic and tenant HTTP endpoints.
- [internal/health](../internal/health): liveness and readiness endpoints.
- [internal/media](../internal/media): media HTTP entrypoints and future media-facing orchestration.

### Infrastructure

- [internal/postgres](../internal/postgres): concrete repository implementations and data mappers.
- [internal/postgres/db](../internal/postgres/db): SQLC-generated query API, row models, SQL helpers, and schema snapshot.
- [internal/postgres/migrations](../internal/postgres/migrations): goose migrations.

### Support code

- [internal/config](../internal/config): environment schema and config loading.
- [pkg/graceful](../pkg/graceful): graceful shutdown helper.
- [pkg/utils](../pkg/utils): shared utility code.

## 3. Runtime Flow

The request lifecycle is intentionally simple:

1. `main.go` loads environment config.
2. `main.go` opens the PostgreSQL pool.
3. `main.go` creates a SQLC query object from `internal/postgres/db`.
4. `main.go` passes that query object into `internal/postgres` adapters.
5. `main.go` passes the adapters into feature services.
6. `main.go` passes services into HTTP handlers.
7. Echo routes requests to feature handlers.
8. Handlers call services.
9. Services call consumer-defined interfaces.
10. `internal/postgres` satisfies those interfaces implicitly and maps SQLC rows to domain structs.

That flow keeps dependency direction one-way: transport depends on service, service depends on interface, infrastructure implements interface.

## 4. How To Think About New Code

When adding or changing code, ask these questions in order:

1. What feature owns this behavior?
2. Is this domain logic, transport logic, or infrastructure logic?
3. Should the consumer define an interface for this dependency?
4. Does the code expose only pure domain types outside `internal/postgres`?
5. Can the change be expressed without importing SQLC types into the feature package?

If the answer to the last question is no, the boundary is in the wrong place.

### Good mental model

- Feature packages own the language of the business.
- PostgreSQL packages own persistence details.
- `main.go` owns object wiring, but not business decisions.
- Handlers translate HTTP into service calls and service results back into HTTP.
- Services orchestrate work and define what they need from the outside world.

## 5. Interface Style

The codebase uses inverse interfaces:

- The consumer declares the interface.
- The provider implements it implicitly.
- Names should describe the action, not the technology.

Examples of good names:

- `Creator`
- `Finder`
- `Store`
- `HashedKeyFinder`
- `TenantService`

Avoid names that leak implementation detail or unnecessary coupling, such as:

- `ApiKeyPostgresStore`
- `ITenantRepository`
- `ServiceImpl`

## 6. Data Mapping Rules

The domain packages should never import SQLC-generated types or PostgreSQL-specific types.

Mapping happens only inside `internal/postgres`:

- SQLC row types are converted into pure domain structs.
- Domain structs are converted into SQLC params.
- Nullable timestamps are represented as `*time.Time` in the domain.
- UUIDs remain `uuid.UUID` in the domain.

If a mapper feels repetitive, keep it in `internal/postgres/mappers.go` instead of spreading conversion code into handlers or services.

## 7. Development Flow

### For a new feature

1. Create a new `internal/<feature>/` package.
2. Add the domain model first.
3. Define the service interface the feature needs from storage or external systems.
4. Implement the service with the smallest possible orchestration logic.
5. Add a handler in the same package.
6. Add postgres adapter methods only if the feature actually needs persistence.
7. Wire everything in `cmd/api/main.go`.

### For a new database query

1. Add the query to `internal/postgres/db/*.sql`.
2. Regenerate SQLC output.
3. Add or update mapping code in `internal/postgres/mappers.go`.
4. Keep the feature package free of SQLC types.

### For a new HTTP endpoint

1. Put the handler in the owning feature package.
2. Make the handler depend on a small service interface.
3. Keep request parsing and response shaping in the handler.
4. Keep business rules in the service.

## 8. Consistency Checklist

Before merging future changes, verify:

- `cmd/api/main.go` is the only composition root.
- No feature package imports `internal/postgres/db` directly.
- No feature package imports PostgreSQL or SQLC types.
- The repository still builds with `go build ./...`.
- The repository still passes `go test ./...`.
- `Makefile`, `.env`, `.env.example`, and `sqlc.yaml` still point at the active paths.
- New migrations live under `internal/postgres/migrations`.

## 9. Feature Notes

### API keys

API keys are the clearest example of the intended style in this repository:

- `internal/apikey` owns the key generation and validation behavior.
- `internal/postgres` owns persistence.
- `internal/postgres/db` owns query generation.
- `cmd/api/main.go` wires the pieces together.

### Tenants

Tenants show the same split:

- `internal/tenant` owns tenant behavior and HTTP endpoints.
- `internal/postgres` owns tenant persistence.
- `internal/postgres/db` owns the SQLC query surface.

### Health and media

- `internal/health` stays lightweight and dependency-free.
- `internal/media` is intentionally thin until real media orchestration is added.

## 10. Working Rule Of Thumb

If you are unsure where new code belongs, default to this order:

1. Feature package.
2. Postgres adapter if it touches data.
3. SQLC query file if it needs a database query.
4. `main.go` for wiring only.

That keeps the architecture consistent even as the codebase grows.
