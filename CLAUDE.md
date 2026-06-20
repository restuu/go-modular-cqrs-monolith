# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run the web server
DATABASE_DSN="postgres://..." go run ./cmd/webserver

# Test — always use -race
go test -race ./...
go test -race -cover ./...

# Single test
go test -race -run TestName ./modules/order/...

# Format + imports (run after every .go edit)
gofmt -w .
goimports -local "go-modular-cqrs-monolith" -w .

# Lint (only new-from-rev changes)
golangci-lint run --new-from-rev=origin/master ./...
```

### Required environment variables

| Variable | When required |
|---|---|
| `DATABASE_DSN` | always |
| `JWT_SECRET` | `AUTH_MODE=local` (default) |
| `AUTH_MODE` | optional; `local` (default) or `gateway` |
| `APP_PORT` | optional; defaults to `8080` |
| `INTERNAL_API_SECRET` | optional; guards `/internal/*` routes |

## Architecture

### The module seam

Every business module lives under `modules/<domain>/` and is split into exactly two parts:

- **`api/`** — the only surface other modules or `main.go` may import. Contains `Service` (embeds `Command` + `Query`) and `api/dto/` (public data-transfer types). `api/dto/` is a **leaf package** — zero `internal/` imports so both this module's internals and external consumers can freely import it.
- **`internal/`** — everything else. The Go compiler enforces this boundary: importing `modules/order/internal/...` from another module is a **compile error**, not a convention.

```
modules/<domain>/
├── api/
│   ├── <domain>_api.go   # Service type; NewService(deps); RegisterRouter(deps)
│   └── dto/              # leaf package — safe to share across modules
└── internal/
    ├── command/          # write side: Command struct + one file per operation
    ├── query/            # read side:  Query struct + one file per operation
    ├── domain/           # entity types + single repository interface
    ├── persistence/      # repository implementation
    ├── port/             # narrow outbound interfaces (consumer-defined)
    ├── adapter/          # translators for shape/vocabulary mismatch (optional)
    └── transport/        # Fiber HTTP handler + route mounting
```

### CQRS convention (no bus)

CQRS here is a **code-organisation convention**, not a runtime dispatch mechanism. There is no command bus or mediator.

- `internal/command` — `Command` struct holds write-side dependencies; each operation (`CreateOrder`, etc.) is a method in its own file.
- `internal/query` — same shape for read-side.
- `api.Service` **embeds both** structs, auto-promoting all methods with zero delegation boilerplate.
- HTTP handlers call methods directly: `h.qry.GetById(ctx, id)`.

### Inter-module communication

Modules may only call each other through the producer's `api/` surface, using a **consumer-defined narrow interface** declared in the consumer's `internal/port/` package.

**When shapes match** — the producer's method signature and return type align with what the consumer needs — the producer's `*Service` satisfies the port structurally. No adapter required.

**When shapes diverge** — the consumer needs a different method (e.g. `FindByModelNumbers([]string)`) than the producer exposes (e.g. `Search(req)`), or translation/aggregation is required — add `internal/adapter/` and split `internal/port/` into **two** interfaces:

| Interface | Role | Named |
|---|---|---|
| `port.<Capability>` | Inbound — what `Query`/`Command` depends on | capability noun, e.g. `ProductCatalog` |
| `port.<Producer>API` | Adaptee — mirrors the producer's `api.Service` surface | `<Producer>API`, e.g. `ProductAPI` |

The adapter struct (`<Consumer><Producer>Adapter`) implements `port.<Capability>` by wrapping `port.<Producer>API`. Both interfaces live in `internal/port/`; `internal/adapter/` holds only translator structs.

```go
// order/internal/port/order_product_port.go

// ProductCatalog is the inbound capability shape that Query depends on.
type ProductCatalog interface {
    GetByModelNumber(ctx context.Context, modelNumber string) (productdto.ProductDetailResponse, error)
    FindByModelNumbers(ctx context.Context, modelNumbers []string) ([]productdto.ProductListItemResponse, error)
}

// ProductAPI mirrors product/api.Service — main.go's productSvc satisfies this structurally.
type ProductAPI interface {
    GetByModelNumber(ctx context.Context, modelNumber string) (productdto.ProductDetailResponse, error)
    Search(ctx context.Context, req productdto.ProductSearchRequest) (productdto.ProductSearchResponse, error)
}
```

Rules:
- **Never import another module's `internal/`.**
- **Cross-module data references use a primitive ID (UUID/string), never a DB-level FK.**
- The producer's `api/dto` types are the cross-module data contract.
- **Both port interfaces live in `internal/port/`.** `internal/adapter/` holds only translator structs, never interface definitions.

### Dependency injection

Manual constructor injection — no wire, fx, or DI framework. `cmd/webserver/main.go` is the single composition root. Wire modules in dependency order (producer before consumer):

```go
productSvc := productapi.NewService(...)
orderSvc   := orderapi.NewService(orderapi.ServiceDeps{..., ProductService: productSvc})
```

### `platform/` boundary

`platform/` is business-agnostic infrastructure. Domain modules must **never** import from each other through `platform/` and must **never** import Fiber types into their `domain`, `command`, or `query` packages. Fiber is confined to `transport/`.

Key packages:
- `platform/config` — env-var-only config loader; `AUTH_MODE=local|gateway` swaps JWT vs upstream-gateway auth.
- `platform/httpx/router` — builds the Fiber app; wires the middleware chain (recover → health → CORS → requestID → logger → metrics → rate limiters → `/internal` → module routes).
- `platform/httpx/response` — standard `Envelope{Success, Data, Error, Meta}` response helpers.
- `platform/errcode` — typed machine-readable error codes.
- `platform/metrics` — Prometheus registry + pgx pool collector; served at `/internal/metrics`.

### Internal projections (read models)

When a query needs an internal scan target that differs from the public `api/dto/` type — e.g. a DB JOIN result or a view not exposed externally — define it in `internal/domain/` as a plain struct in its own file:

```
internal/domain/<domain>_projection.go   ← internal read-model structs only
```

Rules:
- **Pure value types only** — no methods, no business logic.
- **Zero imports from `command/`, `query/`, or `persistence/`** — the file must remain a leaf within `internal/`.
- Both `internal/query/` and `internal/persistence/` may import domain freely (domain is already the base layer), so no new import edges are introduced.
- Promote to `api/dto/` only when another module or the HTTP layer needs the type.

```go
// order_projection.go
type OrderSummaryProjection struct {
    ID        uuid.UUID
    Status    string
    CreatedAt time.Time
}
```

## Invariants to preserve

These are the non-obvious rules from the house pattern. Breaking them collapses the architecture:

1. **DTOs never enter command handlers.** Write path: `handler → command.Method(ctx, ...) → domain entity → repo.Save()`. Read path: `handler → query.Method(ctx, ...) → DTO direct from DB`.
2. **Single repository interface per domain**, in `internal/domain/`. Command (writes) and query (reads) methods live in the same interface, separated by a blank-line comment. Do not split into two interfaces or directories.
3. **Consumer owns the port shape.** The interface lives in the *consumer's* `internal/port/`, not the producer's package. One interface per consumer need — keep them narrow (1–3 methods).
4. **No cross-module DB foreign keys.** Reference another module's entity by a primitive ID only.
5. **Domain-first file naming:** `<domain>_<concept>.go` (e.g. `order_repository.go`, `order_create_command.go`). No `_service` suffix — it's redundant.
6. **Domain layer is framework-free.** `domain/`, `command/`, `query/`, `port/` must not import `github.com/gofiber/fiber`.
7. **Projections are pure structs.** `<domain>_projection.go` in `internal/domain/` must have zero methods and zero imports from other `internal/` sub-packages.
8. **Both port interfaces live in `internal/port/`.** When a cross-module adapter is needed, the capability interface (`port.<Capability>`) and the adaptee interface (`port.<Producer>API`) both belong in `internal/port/`. `internal/adapter/` holds only translator structs — never interface definitions.

## Adding a new module

1. Create `modules/<domain>/api/<domain>_api.go` with `Service`, `ServiceDeps`, `RouterDeps`, `NewService`, `RegisterRouter`.
2. Create `modules/<domain>/api/dto/<domain>_dto.go` (leaf package — no `internal/` imports).
3. Create `internal/domain/<domain>.go` (entity) and `<domain>_repository.go` (single interface with both write + read methods). If queries need internal scan targets, add `<domain>_projection.go` (pure structs, no methods).
4. Create `internal/command/<domain>_command.go` (struct + `NewCommand`) and one file per operation.
5. Create `internal/query/<domain>_query.go` (struct + `NewQuery`) and one file per operation.
6. Create `internal/persistence/<domain>_repository_impl.go`.
7. If this module calls another, create `internal/port/<domain>_<target>_port.go` with a `<Capability>` interface. If the producer's shape differs from what you need, also add a `<Producer>API` adaptee interface in the same file and an `internal/adapter/<domain>_<target>_adapter.go` translator struct.
8. If it serves HTTP, create `internal/transport/<domain>_http_handler.go` and `<domain>_http_router.go`.
9. Wire in `cmd/webserver/main.go` in dependency order; append to `modules` slice as a `router.RouteRegistrar`.

See `ARCHITECTURE.md` for the full pattern rationale and design notes.
