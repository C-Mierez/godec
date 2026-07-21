Now I have a complete picture of your codebase. Let me give you the deep analysis you asked for.

---

# Godec: Monolith-to-Microservices Analysis

## 1. Your Current Architecture (What You Have)

Your codebase is a **modular monolith** — and honestly, it's a _good_ one. Here's what I found:

### The Dependency Graph

```
cmd/api/main.go (composition root)
  ├── config/              ← env loading
  ├── logging/             ← slog setup
  ├── middleware/           ← auth, CORS, OpenAPI validation
  ├── internal/api/        ← HTTP handlers (oapi-codegen strict server)
  │   ├── server.go        ← delegates to feature handlers
  │   ├── tenant_handlers.go
  │   ├── apikey_handlers.go
  │   ├── media_handlers.go (STUB)
  │   ├── health_handlers.go
  │   └── documentation_handlers.go
  ├── internal/tenant/     ← domain model + service + Store interface
  │   └── postgres/        ← implements Store interface
  ├── internal/apikey/     ← domain model + service + Store interface + crypto
  │   └── postgres/        ← implements Store interface
  └── internal/postgres/db/ ← SQLC-generated queries, models, transformers
```

### What's Actually Good About This

| Aspect               | Status                     | Why It Matters                                                    |
| -------------------- | -------------------------- | ----------------------------------------------------------------- |
| Dependency direction | One-way, clean             | Feature packages never import SQLC types                          |
| Interface style      | Consumer-defined, implicit | `apikey.Store` is declared by `apikey`, implemented by `postgres` |
| Composition root     | Single (`cmd/api/main.go`) | All wiring in one place                                           |
| Domain isolation     | Clean                      | `tenant.Tenant` and `apikey.APIKey` are pure Go structs           |
| API contract         | OpenAPI-first with codegen | `spec.yaml` → `oapi-codegen` → `gen.go`                           |
| Database layer       | SQLC with goose migrations | Type-safe queries, clean mappers                                  |

### The Hard Coupling Points (What Would Change)

1. **Single PostgreSQL database** — both `tenants` and `api_keys` tables live in the same DB
2. **Single process** — one `go build`, one binary, one port
3. **Shared `db.Queries`** — both `apikeypg.Store` and `tenantpg.Store` receive the same `*db.Queries` instance
4. **Shared middleware chain** — auth, CORS, OpenAPI validation all live in the same Echo instance
5. **Media is a stub** — `media_handlers.go` returns a hardcoded URL

---

## 2. Microservices Explained Through Your Codebase

### The Core Idea

A **microservice** is a independently deployable unit that:

- Owns its own data
- Has its own process
- Communicates with others via network calls (HTTP/gRPC) or messages (events)
- Can be deployed, scaled, and updated independently

**Right now, your monolith is like an apartment building** — everyone shares the same plumbing (database), electrical (process), and walls (deployment). Microservices are like a neighborhood of houses — each has its own plumbing, can be renovated independently, and neighbors don't break when you remodel your kitchen.

### What Your Codebase Would Look Like as Microservices

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                              │
│                   (routes by path prefix)                       │
└──────────┬──────────────┬──────────────────┬────────────────────┘
           │              │                  │
     ┌─────▼─────┐  ┌─────▼─────┐    ┌──────▼──────┐
     │  Tenant    │  │  API Key  │    │   Media     │
     │  Service   │  │  Service  │    │   Service   │
     │            │  │           │    │             │
     │  Port:8001 │  │  Port:8002│    │  Port:8003  │
     └─────┬──────┘  └─────┬─────┘    └──────┬──────┘
           │               │                  │
     ┌─────▼──────┐  ┌─────▼─────┐    ┌──────▼──────┐
     │  Tenant DB │  │  APIKey DB│    │  Media DB   │
     │ (Postgres) │  │ (Postgres)│    │  (S3 + meta)│
     └────────────┘  └───────────┘    └─────────────┘
```

### Concrete Decomposition — Which Package Becomes Which Service

| Current Package                   | Future Service              | Database             | Why Separate                                       |
| --------------------------------- | --------------------------- | -------------------- | -------------------------------------------------- |
| `internal/tenant`                 | **Tenant Service**          | Own `tenants` table  | Different scaling, different team ownership likely |
| `internal/apikey`                 | **Auth/Identity Service**   | Own `api_keys` table | Security-critical, different blast radius          |
| `internal/api/media_handlers.go`  | **Media Service**           | S3 + metadata DB     | CPU-intensive processing, needs different infra    |
| `internal/api/health_handlers.go` | Stays in each service       | Each has own probes  | Health is per-service in microservices             |
| `middleware/auth.go`              | **API Gateway** or sidecar  | No DB                | Cross-cutting concern, not a service itself        |
| `config/config.go`                | Each service has own config | Per-service env vars | Different DB URLs, ports, secrets                  |

---

## 3. Mental Model Changes You NEED to internalize

This is the part most tutorials skip. These are the fundamental shifts in how you think about software:

### Mental Model #1: "Database per Service" (the hardest change)

**Monolith thinking**: "I have a `tenants` table and an `api_keys` table. I can JOIN them."

**Microservices thinking**: "The Tenant Service owns `tenants`. The API Key Service owns `api_keys`. They CANNOT join. If I need data from both, I call both services and compose the result."

**Why this is hard**: Your current `api_keys` table has a `tenant_id` foreign key. In a microservice world, the API Key Service doesn't even _know_ about the Tenant Service's database. It might cache a `tenant_name` locally, or it calls the Tenant Service API to resolve it.

**Your codebase today**: `apikey.Store.CreateAPIKey(ctx, tenantID, ...)` — the `tenantID` is just a UUID. That's actually _good_ — you're already treating it as an opaque identifier, not a JOIN target. This makes the split easier.

### Mental Model #2: "Network is the new function call"

**Monolith thinking**: `service.GenerateAPIKey(ctx, tenantID, name, scopes)` — this is a function call. It's fast, reliable, and if it fails, it panics.

**Microservices thinking**: `POST http://api-key-service/v1/keys` — this is an HTTP request. It can timeout. The network can partition. The service can be down. You need retries, circuit breakers, and fallbacks.

**Your codebase today**: Your services call stores via interfaces. In microservices, those interfaces become HTTP/gRPC clients. The `Store` interface stays the same shape, but the _implementation_ changes from "direct function call" to "network call."

### Mental Model #3: "Eventual Consistency is Normal"

**Monolith thinking**: "After `CreateTenant` returns, `api_keys.tenant_id` is guaranteed valid because it's the same transaction."

**Microservices thinking**: "After the Tenant Service creates a tenant, it publishes a `TenantCreated` event. The API Key Service receives this event asynchronously and updates its local cache. For a few milliseconds, the API Key Service might not know about the new tenant."

**Why this matters**: You can't have cross-service transactions. You need **sagas** (choreography or orchestration) for multi-service operations.

### Mental Model #4: "Each Service Has Its Own API Contract"

**Monolith thinking**: One `spec.yaml` defines everything. One codegen run produces all handlers.

**Microservices thinking**: Each service has its own `spec.yaml`, its own `oapi-codegen` run, its own binary. The Tenant Service's API contract is independent of the API Key Service's contract.

**Your codebase today**: You already have a clean `spec.yaml` with clear operation tags (`tenants`, `apikeys`, `media`). Splitting this into per-service specs is straightforward.

### Mental Model #5: "Deployment is Per-Service"

**Monolith thinking**: `go build -o build/api.exe ./cmd/api` — one binary, one deployment.

**Microservices thinking**: `go build -o build/tenant-service.exe ./cmd/tenant-service` AND `go build -o build/apikey-service.exe ./cmd/apikey-service` — each service has its own `cmd/` entrypoint, its own Dockerfile, its own CI/CD pipeline, its own deployment.

### Mental Model #6: "Observability is Distributed"

**Monolith thinking**: Check one log stream, one process, one database.

**Microservices thinking**: A single user request might touch 3 services. You need **distributed tracing** (OpenTelemetry), **centralized logging** (structured logs with correlation IDs), and **per-service metrics** (Prometheus).

---

## 4. The Migration Path

### The Strangler Fig Pattern (Your Best Friend)

The industry-standard approach is the **Strangler Fig Pattern** (Martin Fowler, 2004). You don't rewrite. You _grow_ new services around the old monolith until the monolith withers away.

```
Phase 1: Monolith (YOU ARE HERE)
┌──────────────────────────────┐
│         Godec Monolith       │
│  ┌─────────┐ ┌───────────┐  │
│  │ Tenants │ │ API Keys  │  │
│  └─────────┘ └───────────┘  │
│  ┌─────────────────────────┐│
│  │   Shared PostgreSQL DB   ││
│  └─────────────────────────┘│
└──────────────────────────────┘

Phase 2: Extract Media Service (lowest risk, stub anyway)
┌──────────────┐    ┌──────────────────┐
│ Media Service│    │   Godec Monolith │
│  (new)       │    │  (shrinking)     │
│  Port: 8003  │    │  Port: 8080      │
└──────┬───────┘    └────────┬─────────┘
       │                     │
  ┌────▼─────┐         ┌────▼─────────┐
  │ Media DB │         │  Shared DB   │
  └──────────┘         └──────────────┘

Phase 3: Extract API Key Service
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Media Service│    │ API Key Svc  │    │   Monolith   │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
  ┌────▼─────┐        ┌────▼─────┐        ┌────▼─────┐
  │ Media DB │        │ Key DB   │        │ Tenant DB│
  └──────────┘        └──────────┘        └──────────┘

Phase 4: Extract Tenant Service (monolith is now empty)
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Media Service│    │ API Key Svc  │    │ Tenant Svc   │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
  ┌────▼─────┐        ┌────▼─────┐        ┌────▼─────┐
  │ Media DB │        │ Key DB   │        │ Tenant DB│
  └──────────┘        └──────────┘        └──────────┘

Phase 5: Monolith decommissioned. Full microservices.
```

### Why Media First?

Your `media_handlers.go` is a **stub** — it returns a hardcoded URL with a `// TODO: Replace with actual S3 presigned URL generation` comment. This is the _perfect_ extraction candidate:

- Zero existing business logic to break
- Clear service boundary (media upload/processing)
- Different infrastructure needs (S3, video processing, CDN)
- No data dependencies on other features yet

---

## 5. Should You Stick with Monolith or Migrate Early?

### The Honest Answer: **Stay Monolith for Now, But Architect for Migration**

Here's why:

**Arguments for staying monolith (MVP phase):**

1. You have 2 real features and 1 stub. That's not enough complexity to justify microservices overhead.
2. Microservices introduce: service discovery, distributed tracing, network failure handling, data consistency challenges, deployment complexity, and operational overhead.
3. Your clean architecture (inverse interfaces, package-by-feature, single composition root) already gives you 80% of the benefit of microservices _without_ the cost.
4. You're learning Go, testing patterns, and API design. Adding distributed systems complexity now would dilute your focus.

**Arguments for early migration:**

1. If you know Media will be CPU-intensive (video transcoding), isolating it early lets you scale independently.
2. If you're building a team, microservices let people work on different services without merge conflicts.

**My recommendation**: Use the **Modular Monolith** pattern (which you already have!) and add a **"future service boundary" checklist** to each feature:

```go
// In internal/media/ — when you build this for real:
// Future service boundary: This package will become the Media Service.
// For now, it lives in the monolith but follows the same patterns.
// When extracted:
//   - This becomes cmd/media-service/main.go
//   - The Store interface becomes an HTTP client
//   - The model stays the same
//   - The spec.yaml gets split into media-specific endpoints
```

---

## 6. Practical Migration Steps (When You're Ready)

### Step 1: Define Service Boundaries (DDD Bounded Contexts)

Map your features to bounded contexts:

| Bounded Context        | Entities               | Invariants                          | Service                  |
| ---------------------- | ---------------------- | ----------------------------------- | ------------------------ |
| Identity & Access      | APIKey, Credentials    | Key must be unique, hashed secret   | API Key Service          |
| Tenant Management      | Tenant, Status         | Email unique per active tenant      | Tenant Service           |
| Media Processing       | Upload, Process, Store | File size limits, format validation | Media Service            |
| Health & Observability | HealthCheck            | Must respond within 2s              | Embedded in each service |

### Step 2: Add Event Publishing (Before Extraction)

Before you extract anything, add event publishing to your monolith. This is the "eventual consistency" bridge:

```go
// In internal/tenant/service.go — add this now:
type EventPublisher interface {
    PublishTenantCreated(ctx context.Context, tenant *Tenant) error
    PublishTenantStatusChanged(ctx context.Context, tenant *Tenant) error
}

// In cmd/api/main.go — wire it:
// - For now, the publisher can be a no-op or log to stdout
// - Later, it becomes a Kafka/NATS publisher
```

### Step 3: Extract the First Service (Media)

1. Create `cmd/media-service/main.go` — new composition root
2. Move `internal/media/` to its own module (or keep in monorepo)
3. Add a Dockerfile for `media-service`
4. Add a route in the monolith that proxies to the new service
5. Run both side by side, gradually shift traffic

### Step 4: Split the Database

For each extracted service:

1. Create a new database for the service
2. Use Change Data Capture (CDC) or dual writes to sync data during transition
3. Once stable, cut over reads to the new database
4. Remove the old tables from the shared DB

### Step 5: Add Infrastructure

| Component               | Tool Options                             | Purpose                              |
| ----------------------- | ---------------------------------------- | ------------------------------------ |
| API Gateway             | Traefik, Kong, or custom reverse proxy   | Route requests to correct service    |
| Service Discovery       | Consul, etcd, or Kubernetes DNS          | Services find each other             |
| Message Queue           | NATS, Kafka, or RabbitMQ                 | Async communication between services |
| Distributed Tracing     | OpenTelemetry + Jaeger/Tempo             | Track requests across services       |
| Centralized Logging     | Loki, ELK, or CloudWatch                 | Aggregate logs from all services     |
| Container Orchestration | Docker Compose (dev) → Kubernetes (prod) | Deploy and manage services           |

---

## 7. Key Gotchas to Watch For

### Gotcha #1: The "Shared Database" Trap

Don't let services share a database. If Service A reads Service B's tables directly, you've built a distributed monolith — all the cost of microservices with none of the benefits.

### Gotcha #2: The "Distributed Monolith" Trap

If every request requires calling all 3 services synchronously, you've just made your system slower and more fragile. Microservices should be _independently useful_.

### Gotcha #3: The "Data Consistency" Trap

You can't have ACID transactions across services. Design for eventual consistency. Use the **Saga pattern** for multi-service operations.

### Gotcha #4: The "Testing" Trap

In a monolith, `go test ./...` tests everything. In microservices, you need:

- Unit tests per service (still `go test ./...`)
- Integration tests per service (test DB, test external deps)
- Contract tests between services (verify API contracts match)
- End-to-end tests (test the full flow)

### Gotcha #5: The "Operational Complexity" Trap

One monolith = one deployment, one log stream, one health check. Three microservices = three deployments, three log streams, three health checks, plus service discovery, plus inter-service auth, plus distributed tracing.

---

## 8. Summary: Your Action Plan

| Phase       | What to Do                                                                  | When                               |
| ----------- | --------------------------------------------------------------------------- | ---------------------------------- |
| **Now**     | Keep the monolith. Build features cleanly. Add event publishing interfaces. | MVP                                |
| **Next**    | Extract Media Service (it's a stub — easiest win).                          | When media gets real               |
| **Then**    | Extract API Key Service (security isolation).                               | When you have multiple consumers   |
| **Later**   | Extract Tenant Service (the monolith is now empty).                         | When team size or scale demands it |
| **Finally** | Decommission the monolith skeleton.                                         | When all features are services     |

Your monolith's architecture is _already_ 80% of the way to microservices. The clean interfaces, package-by-feature structure, and single composition root mean you're not fighting against a tangled mess — you're building on a solid foundation. The migration will be incremental extraction, not a rewrite.

**The best microservice architecture is the one you extract from a well-structured monolith — not the one you build from scratch on day one.**

---

# The Saga Pattern — A Deep Dive Using Your Codebase

## The Problem: Why Sagas Exist

Right now, in your monolith, when you create an API key for a tenant, you could do this in a single database transaction:

```go
// In a monolith — ONE transaction, ONE database
func CreateTenantWithAPIKey(ctx context.Context, name, email string) error {
    tx, _ := db.Begin(ctx)
    defer tx.Rollback(ctx)

    tenant, _ := tx.CreateTenant(name, email)
    tx.CreateAPIKey(tenant.ID, "default-key")

    return tx.Commit(ctx)  // Either BOTH succeed or NEITHER does
}
```

**ACID guarantees**: If `CreateAPIKey` fails, `CreateTenant` is automatically rolled back. The database handles this. It's simple.

Now imagine Godec is microservices:

```
┌──────────────┐         ┌──────────────┐
│ Tenant Svc   │         │ API Key Svc  │
│ (Port 8001)  │         │ (Port 8002)  │
│              │         │              │
│ PostgreSQL A │         │ PostgreSQL B │
└──────────────┘         └──────────────┘
```

**You cannot use a single database transaction across two databases.** There's no `BEGIN` that spans two Postgres instances. So what happens if:

1. Tenant Service creates tenant ✅
2. API Key Service fails to create the key ❌

You now have an **orphaned tenant** — a tenant exists but has no API key. The data is inconsistent. There's no automatic rollback across services.

**This is the problem Sagas solve.**

---

## The Saga Pattern: Core Idea

A **Saga** is a sequence of **local transactions**, where each step has a **compensating action** (a rollback) that undoes its effect if a later step fails.

Instead of one big ACID transaction, you get a chain:

```
Step 1: Create Tenant       → Compensate: Delete Tenant
Step 2: Create API Key      → Compensate: Delete API Key
Step 3: Send Welcome Email  → Compensate: Send "Nevermind" Email
```

If Step 2 fails, you execute the compensations in **reverse order**: delete the tenant that was just created. The system ends up in a consistent state — not through database magic, but through **explicit undo logic** you write yourself.

### The Key Difference from 2PC

| Aspect        | Two-Phase Commit (2PC)                | Saga                                  |
| ------------- | ------------------------------------- | ------------------------------------- |
| Locking       | Locks resources across all services   | No cross-service locks                |
| Rollback      | Automatic (database handles it)       | Manual (you write compensations)      |
| Isolation     | Strong (ACID)                         | Weak (eventual consistency)           |
| Performance   | Degrades with participants            | Scales horizontally                   |
| Failure model | Blocking (waits for all participants) | Non-blocking (compensate and move on) |

2PC is like a referee holding all players' jerseys until everyone finishes. Sagas are like each player finishing independently, and if someone fails, the previous players undo their moves.

---

## Two Flavors: Choreography vs Orchestration

### Choreography (Event-Driven, Decentralized)

Each service publishes events and reacts to events from other services. No central coordinator.

```
┌─────────┐    OrderCreated    ┌──────────┐    CreditReserved    ┌─────────┐
│ Order   │ ─────────────────► │ Customer │ ──────────────────►  │ Order   │
│ Service │                    │ Service  │                      │ Service │
│         │                    │          │                      │         │
│ listens │ ◄── CreditFailed ──│          │                      │ updates │
│ for     │                    │ publishes│                      │ status  │
│ events  │                    │ events   │                      │         │
└─────────┘                    └──────────┘                      └─────────┘
```

**How it works in your codebase context:**

```
1. Tenant Service creates tenant
2. Tenant Service publishes event: { type: "TenantCreated", tenant_id: "abc" }
3. API Key Service receives event, creates default API key
4. API Key Service publishes event: { type: "APIKeyCreated", tenant_id: "abc" }
5. Media Service receives event, provisions storage bucket

If step 3 fails:
3b. API Key Service publishes: { type: "APIKeyCreationFailed", tenant_id: "abc" }
4b. Tenant Service receives event, deletes the tenant
```

**Pros:**

- No single point of failure
- Services are loosely coupled
- Simple to implement for 2-3 services

**Cons:**

- Hard to understand the full flow (you have to trace events across services)
- Circular dependencies can form
- Debugging is painful — "which event caused this?"
- Business logic is spread across multiple services

### Orchestration (Central Coordinator)

A dedicated **saga orchestrator** tells each service what to do and tracks the overall state.

```
                    ┌─────────────────────┐
                    │   Saga Orchestrator  │
                    │                     │
                    │  1. Create Tenant   │
                    │  2. Create API Key  │
                    │  3. Send Welcome    │
                    │                     │
                    │  Track: PENDING →   │
                    │  RUNNING → DONE     │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
     │ Tenant Svc   │  │ API Key Svc  │  │ Email Svc    │
     └──────────────┘  └──────────────┘  └──────────────┘
```

**How it works in your codebase context:**

```
Orchestrator receives: "Create Tenant with API Key"

Step 1: Call Tenant Service → POST /v1/tenants
        → Response: { id: "abc", name: "Acme" }
        → Save: tenant_id = "abc"

Step 2: Call API Key Service → POST /v1/apikey/create_key { tenant_id: "abc" }
        → Response: { id: "key1", token_id: "gdk_xxx" }

Step 3: Done! Return combined result.

If Step 2 fails:
  → Call Tenant Service → DELETE /v1/tenants/abc  (compensation)
  → Mark saga as FAILED
```

**Pros:**

- Easy to understand the full flow (it's in one place)
- Easy to debug (check the orchestrator's state)
- Easy to add new steps
- Clear separation of concerns

**Cons:**

- Single point of coordination (not failure, but coordination)
- Orchestrator can become a bottleneck
- More moving parts to deploy

### Which Should You Use?

| Factor             | Choreography        | Orchestration             |
| ------------------ | ------------------- | ------------------------- |
| Number of services | 2-3                 | 4+                        |
| Flow complexity    | Linear, simple      | Branching, conditional    |
| Debugging          | Hard (trace events) | Easy (check orchestrator) |
| Adding new steps   | Add event listener  | Add step to orchestrator  |
| Team size          | Small               | Large                     |

**For Godec**: Start with **orchestration**. You'll have 3 services (Tenant, APIKey, Media), the flow is straightforward, and debugging will be much easier as you learn distributed systems.

---

## Concrete Implementation: Saga for Godec

Let's say you extract Tenant Service and API Key Service. When a new customer signs up, you need to:

1. Create the tenant
2. Create a default API key for that tenant
3. Send a welcome email

Here's how you'd implement this as an orchestration-based saga in Go:

### The Saga Engine

```go
// internal/saga/saga.go
package saga

import (
    "context"
    "fmt"
    "log/slog"
    "time"
)

// Step represents one action in the saga with its undo logic.
type Step struct {
    Name       string
    Action     func(ctx context.Context, data *SagaData) error
    Compensate func(ctx context.Context, data *SagaData) error
}

// SagaData is the shared context passed between steps.
// Each step reads what it needs and writes what the next step needs.
type SagaData struct {
    TenantID  string
    TenantName string
    APIKeyID  string
    TokenID   string
    // ... whatever the saga needs to pass between steps
}

// Status tracks where the saga is in its lifecycle.
type Status string

const (
    StatusPending       Status = "pending"
    StatusRunning       Status = "running"
    StatusCompleted     Status = "completed"
    StatusFailed        Status = "failed"
    StatusCompensating  Status = "compensating"
)

// Saga orchestrates a sequence of steps with automatic compensation.
type Saga struct {
    Name   string
    steps  []Step
    Status Status
    // In production, you'd persist this to a database
}

func New(name string) *Saga {
    return &Saga{
        Name:   name,
        steps:  make([]Step, 0),
        Status: StatusPending,
    }
}

func (s *Saga) AddStep(step Step) *Saga {
    s.steps = append(s.steps, step)
    return s
}

// Execute runs the saga. If any step fails, it compensates
// all completed steps in reverse order.
func (s *Saga) Execute(ctx context.Context, data *SagaData) error {
    s.Status = StatusRunning
    completedSteps := make([]Step, 0, len(s.steps))

    for _, step := range s.steps {
        slog.Info("saga: executing step",
            "saga", s.Name,
            "step", step.Name,
        )

        if err := step.Action(ctx, data); err != nil {
            s.Status = StatusFailed
            slog.Error("saga: step failed",
                "saga", s.Name,
                "step", step.Name,
                "error", err,
            )

            // Compensate completed steps in REVERSE order
            s.compensate(ctx, completedSteps, data)
            return fmt.Errorf("saga %q failed at step %q: %w", s.Name, step.Name, err)
        }

        completedSteps = append(completedSteps, step)
    }

    s.Status = StatusCompleted
    return nil
}

// compensate runs compensating actions in reverse order.
func (s *Saga) compensate(ctx context.Context, completed []Step, data *SagaData) {
    s.Status = StatusCompensating
    slog.Warn("saga: compensating", "saga", s.Name, "steps_to_undo", len(completed))

    for i := len(completed) - 1; i >= 0; i-- {
        step := completed[i]
        if step.Compensate == nil {
            continue
        }

        slog.Info("saga: compensating step", "saga", s.Name, "step", step.Name)

        // Use a timeout so compensation doesn't hang forever
        compCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
        if err := step.Compensate(compCtx, data); err != nil {
            // CRITICAL: In production, log this to a dead letter queue
            // A failed compensation is a serious operational issue
            slog.Error("saga: compensation failed",
                "saga", s.Name,
                "step", step.Name,
                "error", err,
            )
        }
        cancel()
    }
}
```

### The Concrete Saga: Create Tenant with API Key

```go
// internal/saga/create_tenant_with_key.go
package saga

import (
    "context"
    "fmt"
)

// CreateTenantWithKeySaga builds the saga for creating a tenant
// with a default API key.
func CreateTenantWithKeySaga(
    tenantClient *TenantServiceClient,
    apiKeyClient *APIKeyServiceClient,
) *Saga {
    return New("create-tenant-with-key").
        AddStep(Step{
            Name: "create-tenant",
            Action: func(ctx context.Context, data *SagaData) error {
                // Call Tenant Service
                resp, err := tenantClient.CreateTenant(ctx, CreateTenantRequest{
                    Name:  data.TenantName,
                    Email: data.Email,
                })
                if err != nil {
                    return fmt.Errorf("creating tenant: %w", err)
                }
                data.TenantID = resp.ID
                return nil
            },
            Compensate: func(ctx context.Context, data *SagaData) error {
                // Undo: Delete the tenant
                return tenantClient.DeleteTenant(ctx, data.TenantID)
            },
        }).
        AddStep(Step{
            Name: "create-api-key",
            Action: func(ctx context.Context, data *SagaData) error {
                // Call API Key Service
                resp, err := apiKeyClient.CreateAPIKey(ctx, CreateAPIKeyRequest{
                    TenantID: data.TenantID,
                    Name:     "default-key",
                    Scopes:   []string{"read:domains", "write:keys"},
                })
                if err != nil {
                    return fmt.Errorf("creating API key: %w", err)
                }
                data.APIKeyID = resp.ID
                data.TokenID = resp.TokenID
                return nil
            },
            Compensate: func(ctx context.Context, data *SagaData) error {
                // Undo: Delete the API key
                return apiKeyClient.DeleteAPIKey(ctx, data.APIKeyID)
            },
        }).
        AddStep(Step{
            Name: "send-welcome-email",
            Action: func(ctx context.Context, data *SagaData) error {
                // Call Email Service (or use a message queue)
                return emailClient.SendWelcome(ctx, WelcomeEmail{
                    TenantID:  data.TenantID,
                    Email:     data.Email,
                    APIKey:    data.TokenID, // Show once!
                })
            },
            Compensate: func(ctx context.Context, data *SagaData) error {
                // Undo: Send "nevermind" email (or just log it)
                // In practice, you might not compensate this —
                // sending an extra email is less harmful than
                // leaving orphaned data
                return nil
            },
        })
}
```

### Wiring It All Together

```go
// In the API Gateway or Tenant Service handler
func HandleSignup(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req SignupRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Create the saga
    createSaga := CreateTenantWithKeySaga(tenantClient, apiKeyClient)

    // Execute it
    data := &SagaData{
        TenantName: req.Name,
        Email:      req.Email,
    }

    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    err := createSaga.Execute(ctx, data)
    if err != nil {
        // Saga failed and compensated — return error to client
        http.Error(w, "Signup failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Success! Return the combined result
    json.NewEncoder(w).Encode(SignupResponse{
        TenantID: data.TenantID,
        TokenID:  data.TokenID,
    })
}
```

---

## The Hard Parts (What Most Tutorials Skip)

### 1. Compensations Must Be Idempotent

A crash during compensation might cause the compensation to run **twice**. Your `DeleteTenant` compensation must handle being called with an already-deleted tenant:

```go
// BAD: Fails if tenant already deleted
func (c *TenantClient) DeleteTenant(ctx context.Context, id string) error {
    return c.doDelete(ctx, "/v1/tenants/"+id)
}

// GOOD: Idempotent — deleting a non-existent tenant is not an error
func (c *TenantClient) DeleteTenant(ctx context.Context, id string) error {
    err := c.doDelete(ctx, "/v1/tenants/"+id)
    if err != nil && !isNotFoundError(err) {
        return err
    }
    return nil // Already deleted? That's fine.
}
```

### 2. Lack of Isolation (The "I" in ACID)

Between Step 1 and Step 2, another request might see the tenant without an API key. This is the **dirty read** problem in sagas.

**Countermeasures:**

- **Semantic locks**: Mark the tenant as `PENDING` until the saga completes
- **Commutative updates**: Design operations so order doesn't matter
- **Pessimistic view**: Reorder steps to minimize the window
- **Read-your-owns-writes**: The client that created the tenant also reads it (so they see their own uncommitted state)

For Godec, the simplest approach is the **semantic lock**:

```go
// In the Tenant model
type Status string
const (
    StatusPending  Status = "pending"  // Saga in progress
    StatusActive   Status = "active"   // Saga completed
    StatusFailed   Status = "failed"   // Saga failed, being compensated
)
```

### 3. Reliable Event Publishing (Transactional Outbox)

When a service updates its database AND needs to publish an event, you can't do both in one transaction (different systems). The **Transactional Outbox** pattern solves this:

```go
// 1. Write to outbox table IN THE SAME TRANSACTION as your business write
func (s *TenantStore) CreateTenant(ctx context.Context, name, email string) error {
    tx, _ := s.db.Begin(ctx)
    defer tx.Rollback(ctx)

    // Business write
    tenant, _ := tx.CreateTenant(name, email)

    // Outbox write — same transaction!
    tx.InsertOutboxEvent(OutboxEvent{
        Type:    "TenantCreated",
        Payload: marshal(tenant),
    })

    return tx.Commit(ctx)  // Both are committed atomically
}

// 2. A background process reads the outbox and publishes events
func (p *OutboxPublisher) Run(ctx context.Context) {
    for {
        events := p.store.GetUnpublishedEvents(ctx)
        for _, event := range events {
            p.publisher.Publish(event.Type, event.Payload)
            p.store.MarkPublished(ctx, event.ID)
        }
        time.Sleep(1 * time.Second)
    }
}
```

### 4. What If Compensation Fails?

This is the scariest scenario. Your `DeleteTenant` compensation fails because the Tenant Service is down. Now you have an orphaned tenant AND a failed saga.

**Strategies:**

- **Retry with backoff**: Keep retrying the compensation
- **Dead letter queue**: Move failed compensations to a queue for manual intervention
- **Alerting**: This is an operational emergency — page someone
- **Idempotent retries**: Since compensations are idempotent, you can safely retry

```go
// Production-grade compensation with retry
func (s *Saga) compensateWithRetry(ctx context.Context, step Step, data *SagaData) error {
    backoff := 100 * time.Millisecond
    maxRetries := 5

    for i := 0; i < maxRetries; i++ {
        compCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
        err := step.Compensate(compCtx, data)
        cancel()

        if err == nil {
            return nil
        }

        slog.Warn("compensation failed, retrying",
            "step", step.Name,
            "attempt", i+1,
            "error", err,
        )

        select {
        case <-time.After(backoff):
            backoff *= 2 // Exponential backoff
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    // All retries exhausted — this needs manual intervention
    return fmt.Errorf("compensation for step %q failed after %d retries", step.Name, maxRetries)
}
```

---

## How This Maps to Your Codebase

When you're ready to implement sagas, here's what changes in each package:

| Current Package                  | Changes Needed                                                                                                     |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `internal/tenant`                | Add `Status: pending/active/failed` field. Add event publishing interface. Add `DELETE` endpoint for compensation. |
| `internal/apikey`                | Add `DELETE` endpoint for compensation. Make `CreateAPIKey` idempotent (check for existing key before creating).   |
| `internal/api/media_handlers.go` | When real, add `DELETE` or `REVERT` for compensation.                                                              |
| `cmd/api/main.go`                | Wire the saga orchestrator. Wire event publishers.                                                                 |
| `internal/saga/`                 | **NEW** — Saga engine + concrete saga definitions.                                                                 |

---

## Summary: The Saga Mental Model

```
Monolith:     BEGIN TRANSACTION → do A → do B → do C → COMMIT
              (if anything fails, automatic ROLLBACK)

Saga:         Step A ✅ → Step B ✅ → Step C ❌
              → Compensate B 🔙 → Compensate A 🔙
              (you write the rollback logic yourself)
```

**The 5 rules:**

1. Every action needs a compensating action
2. Compensations must be idempotent
3. Compensations run in reverse order
4. Use the Transactional Outbox pattern for reliable event publishing
5. Persist saga state so you can recover from crashes

**The tradeoff**: You gain scalability and availability (no distributed locks, no blocking), but you lose automatic rollback and strong isolation. You have to design for eventual consistency explicitly.

**Sources**: microservices.io (Chris Richardson), AWS Strangler Fig guidance, OneUptime Go saga implementation, ByteByteGo saga analysis.
