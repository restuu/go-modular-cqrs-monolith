# Architecture

This document explains the **Modular CQRS Monolith** pattern used in this template, the
enforcement mechanisms that make it structural (not merely conventional), and known gaps
to address before using this as a production baseline.

## The house pattern

> One deployable binary.
> Each business module = `api/` (public seam) + `internal/` (everything else).
> Modules call each other only through the `api/` surface — never into `internal/`.

This is a deliberate composition of three ideas:

| Pattern | What it contributes |
|---|---|
| **Modular Monolith** | Module-boundary enforcement; microservices-readiness without the ops tax |
| **Pragmatic CQRS** | Structural separation of read and write paths within each module |
| **Ports & Adapters** | Consumer-defined narrow interfaces; domain logic free from framework types |

## Module anatomy

```
modules/<domain>/
├── api/
│   ├── <domain>_api.go     Service type, NewService(ServiceDeps), RegisterRouter(RouterDeps)
│   └── dto/                public request/response types — the cross-module data contract
└── internal/
    ├── domain/             entity types + ONE repository interface (write + read methods)
    ├── command/            Command struct + one file per write operation
    ├── query/              Query struct + one file per read operation
    ├── persistence/        repository implementation(s)
    ├── port/               narrow outbound interfaces declared by this module
    └── transport/          Fiber HTTP handler + route mounting
```

`modules/order` is the full exemplar. `modules/product` shows a query-only module:
no `command/`, no `transport/`, not exposed over HTTP — it is consumed in-process by
the order module through a port.

## Enforcement mechanisms

### 1. The compiler's `internal/` rule

Any attempt to import `modules/order/internal/...` from outside the `modules/order`
directory tree is a **compile error** — no tooling config, no review process needed.
This is the primary seam.

### 2. `api/dto/` as a leaf package

`api/dto/` imports only stdlib and primitive types — never anything from `internal/`.
This breaks the deadlock that would otherwise exist if both external consumers and the
module's own internals needed to import the same DTO types.

```
OK:   modules/order/internal/query → modules/order/api/dto  (same module)
OK:   modules/product/api          → modules/order/api/dto  (public API of a module)
ERR:  modules/product/api          → modules/order/internal/query
```

### 3. Consumer-defined narrow ports

When module A needs to call module B, module A declares the interface it needs in its
**own** `internal/port/` package. Module B's `*Service` satisfies it structurally.

```go
// order/internal/port/order_product_port.go
// Order declares exactly the method it needs — not a copy of Product's full API.
type ProductService interface {
    GetByModelNumber(ctx context.Context, modelNumber string) (productdto.ProductDetailResponse, error)
}
```

Result: the producer (`product`) can evolve freely as long as it keeps satisfying the
method signature. Swapping an in-process call for an HTTP/gRPC adapter is a one-file
change in `main.go` — the consumer code is untouched.

## CQRS data-flow

CQRS here is a **structural convention** — there is no command bus, mediator, or event
store. Separation is achieved by directory and type, not by runtime dispatch.

### Write path

```
HTTP POST  →  transport.Handler
           →  command.Command.CreateOrder(ctx, ...)
           →  domain entity construction + validation
           →  domain.Repository.Save(ctx, entity)
```

The command method receives primitive inputs (or a dedicated input struct when ≥3
fields). It **never** receives a DTO — DTOs are read-path projections.

### Read path

```
HTTP GET   →  transport.Handler
           →  query.Query.GetById(ctx, id)
           →  domain.Repository.FindOrderById(ctx, id)
           →  DTO  ←  scanned directly from DB rows
```

The query method returns a DTO (or a slice/pagination result) directly — no entity
round-trip. The repository's read methods scan straight into DTOs.

### Single repository interface

Both write-side finders/saves and read-side projections live in the **same** interface
in `internal/domain/`, separated by a blank-line comment:

```go
type OrderRepository interface {
    // --- write side ---
    Save(ctx context.Context, o Order) error

    // --- read side ---
    FindOrderById(ctx context.Context, id uuid.UUID) (Order, error)
    FindOrderItemsByOrderId(ctx context.Context, id uuid.UUID) ([]OrderItem, error)
}
```

Two separate interfaces (or putting query reads in a different file) is over-engineering
for this pattern — splitting only pays off if you need separate storage backends.

### `Service` embeds both sides

```go
type Service struct {
    *command.Command
    *query.Query
}
```

All command and query methods are auto-promoted to `*Service` — zero delegation
boilerplate. The HTTP transport handler receives the whole service or individual structs
depending on clarity.

## Inter-module communication rules

1. **Import only `api/` and `api/dto/`.** Never import another module's `internal/`.
2. **Cross-module data references use a primitive ID (UUID/string), not a DB-level FK.**
   A DB foreign key welds two modules into one schema; extracting them into separate
   services later becomes a data-migration problem.
3. **No cross-module transactions.** Each module manages its own DB transaction scope.
   Cross-module write paths compensate on failure rather than rolling back.
4. **Shapes match → no adapter needed.** If the producer's method signature and return
   type align with what the consumer declared in its port, wire directly in `main.go`.
   Add `internal/adapter/` only for vocabulary/shape mismatch or aggregation.

## Dependency injection

`cmd/webserver/main.go` is the single composition root. No DI framework, no global
`init()`, no singletons. Modules are constructed in dependency order (producers first)
and injected by constructor:

```go
productSvc := productapi.NewService(productapi.ServiceDeps{DB: sqlDB})
orderSvc   := orderapi.NewService(orderapi.ServiceDeps{
    DB:             sqlDB,
    ProductService: productSvc,   // satisfies order's port.ProductService interface
})
```

When the composition root file becomes unwieldy, that's a signal the service may have
outgrown a single process — not a reason to add a DI container.

## `platform/` boundary

`platform/` is infrastructure with no business logic. Domain modules (`command/`,
`query/`, `domain/`, `port/`) must **never import Fiber** or any `platform/httpx` type —
those are confined to `transport/`.

```
platform/
  config/        env-var config loader; AUTH_MODE=local|gateway
  httpx/
    router/      Fiber app builder; middleware chain ordering
    middleware/  Auth (JWT / gateway), rate limiter, request ID, logger, metrics
    request/     validator
    response/    Envelope{Success, Data, Error, Meta} + helpers
  errcode/       typed machine-readable error codes
  logging/       slog context helpers + structured attributes
  metrics/       Prometheus registry + pgx pool collector
```

The `AUTH_MODE` switch in config allows running JWT locally and trusting an upstream
API-gateway's headers in production without changing module code — modules receive an
`AuthMiddleware fiber.Handler` in their `RouterDeps` and never import `modules/auth`.

---

## Design notes / known gaps

This section documents divergences from the house pattern. None are architectural
deal-breakers for a skeleton, but address them before this becomes a production
baseline.

### 1. External IDs are `int`, not UUID

**What:** `OrderRepository.FindOrderById(ctx, id int)`, the route `/:order_id`, and
domain entities use sequential integers on the external API surface.

**Why it matters:** Sequential integer IDs on public endpoints are an IDOR /
enumeration risk. The house pattern calls for a dual-ID scheme: `_id BIGSERIAL PRIMARY
KEY` for internal DB FK joins (never leaves the persistence layer) and `id UUID UNIQUE`
for all entities, DTOs, URLs, and cross-module references.

**Recommended fix:** Change `int` → `uuid.UUID` on every entity, DTO, repository
interface, and route param. The DB column stays `BIGSERIAL`; it's just never exposed.

### 2. Port name is producer-centric

**What:** `order/internal/port/order_product_port.go` declares `ProductService`.

**Why it matters:** The name reads like the producer's API rather than the consumer's
need. The convention is to name the interface from the consumer's perspective:
`ProductLookup`, `ProductByModelNumber`, or `ModelNumberResolver`.

**Recommended fix:** Rename to something like `ProductLookup` or `ModelNumberResolver`
in the port file. Zero behaviour change; the compiler validates the fix immediately.

### 3. `Command` has no dependencies; repository is query-only

**What:** `command.NewCommand()` takes no arguments and the repository interface holds
only `Find*` methods.

**Why it matters:** The template should demonstrate the full CQRS cycle — a `Save`
method on the repository used by a command handler. As-is, it illustrates only the
read path. Consumers of the template will add `Save` anyway but won't have a
reference example of the command path wiring.

**Recommended fix:** Add `Save(ctx, Order) error` to `OrderRepository`; have
`NewCommand(repo domain.OrderRepository)` accept it; implement `CreateOrder` to
construct an entity and call `repo.Save`.

### 4. `AUTH_MODE=gateway` path is not wired in `main.go`

**What:** `platform/config` loads `AUTH_MODE` and the `middleware` package has a
`gateway` middleware, but `cmd/webserver/main.go` always installs
`middleware.Auth(cfg.JWT.Secret)` regardless of mode.

**Why it matters:** The gateway branch is untested and the config option is a no-op
in the current template. A developer adopting this template and switching to `gateway`
mode will get JWT auth when they expect header-trust.

**Recommended fix:** Branch on `cfg.Auth.Mode` in `main.go` to select the appropriate
middleware before building `router.Deps`.

### 5. Middleware fields are transposed in `RegisterRouter`

**What:** In `order_api.go:46-49`:
```go
middlewares := transport.Middlewares{
    Auth:    deps.AuthLimiter,    // ← should be deps.AuthMiddleware
    Limiter: deps.AuthMiddleware, // ← should be deps.AuthLimiter
}
```
The `Auth` field receives the limiter and `Limiter` receives the middleware.

**Why it matters:** Out of scope for a structural review, but noted here so it doesn't
get copied forward. Fix before writing real routes.

### 6. Module path matches the template repo name

**What:** `go.mod` declares `module go-modular-cqrs-monolith` — the same
string as the GitHub repo name.

**Why it matters:** Every Go file uses this as the import prefix. Any user of the
template must do a global find-and-replace as their first step; the current README does
not mention this.

**Recommended fix:** Document the module-rename step prominently as step 1 of adoption
(done in README.md). Consider whether the initial module path should instead be
`github.com/your-org/your-repo` as a placeholder that's clearly meant to be replaced.

### 7. Missing template scaffolding

The following are expected by the house pattern but not yet present. Implement before
treating this as a production-ready template:

| Missing | Reason |
|---|---|
| `cmd/migrate/` + goose migrations | Every schema change needs an audit trail; ad-hoc schema creation is unacceptable in production |
| `.env.example` | Developers need a reference for required env vars |
| `Makefile` / `justfile` | Encodes build/lint/test commands so CI and local are identical |
| `.mockery.yaml` | Interface mocks are generated from `.mockery.yaml` in the house pattern |
| `.golangci.yml` | Encodes funlen limits and other project-specific lint rules |
| `.github/workflows/ci.yml` | Template repos should ship with a working CI baseline |
| Example table-driven test | Shows the `given_X_when_Y_then_Z` naming and `testify/suite` pattern |
