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
    └── transport/        # Fiber HTTP handler + route mounting
```

### CQRS convention (no bus)

CQRS here is a **code-organisation convention**, not a runtime dispatch mechanism. There is no command bus or mediator.

- `internal/command` — `Command` struct holds write-side dependencies; each operation (`CreateOrder`, etc.) is a method in its own file.
- `internal/query` — same shape for read-side.
- `api.Service` **embeds both** structs, auto-promoting all methods with zero delegation boilerplate.
- HTTP handlers call methods directly: `h.qry.GetById(ctx, id)`.

### Inter-module communication

Modules may only call each other through the producer's `api/` surface, using a **consumer-defined narrow interface** declared in the consumer's `internal/port/` package. The producer's `*Service` satisfies the interface structurally — no adapter needed when shapes match.

```go
// order/internal/port/order_product_port.go
type ProductService interface {
    GetByModelNumber(ctx context.Context, modelNumber string) (productdto.ProductDetailResponse, error)
}
```

Rules:
- **Never import another module's `internal/`.**
- **Cross-module data references use a primitive ID (UUID/string), never a DB-level FK.**
- The producer's `api/dto` types are the cross-module data contract.

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

## Invariants to preserve

These are the non-obvious rules from the house pattern. Breaking them collapses the architecture:

1. **DTOs never enter command handlers.** Write path: `handler → command.Method(ctx, ...) → domain entity → repo.Save()`. Read path: `handler → query.Method(ctx, ...) → DTO direct from DB`.
2. **Single repository interface per domain**, in `internal/domain/`. Command (writes) and query (reads) methods live in the same interface, separated by a blank-line comment. Do not split into two interfaces or directories.
3. **Consumer owns the port shape.** The interface lives in the *consumer's* `internal/port/`, not the producer's package. One interface per consumer need — keep them narrow (1–3 methods).
4. **No cross-module DB foreign keys.** Reference another module's entity by a primitive ID only.
5. **Domain-first file naming:** `<domain>_<concept>.go` (e.g. `order_repository.go`, `order_create_command.go`). No `_service` suffix — it's redundant.
6. **Domain layer is framework-free.** `domain/`, `command/`, `query/`, `port/` must not import `github.com/gofiber/fiber`.

## Adding a new module

1. Create `modules/<domain>/api/<domain>_api.go` with `Service`, `ServiceDeps`, `RouterDeps`, `NewService`, `RegisterRouter`.
2. Create `modules/<domain>/api/dto/<domain>_dto.go` (leaf package — no `internal/` imports).
3. Create `internal/domain/<domain>.go` (entity) and `<domain>_repository.go` (single interface with both write + read methods).
4. Create `internal/command/<domain>_command.go` (struct + `NewCommand`) and one file per operation.
5. Create `internal/query/<domain>_query.go` (struct + `NewQuery`) and one file per operation.
6. Create `internal/persistence/<domain>_repository_impl.go`.
7. If this module calls another, create `internal/port/<domain>_<target>_port.go`.
8. If it serves HTTP, create `internal/transport/<domain>_http_handler.go` and `<domain>_http_router.go`.
9. Wire in `cmd/webserver/main.go` in dependency order; append to `modules` slice as a `router.RouteRegistrar`.

See `ARCHITECTURE.md` for the full pattern rationale and design notes.
