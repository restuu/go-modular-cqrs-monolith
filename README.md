# go-modular-cqrs-monolith

> **Suggested rename:** `go-modular-cqrs-monolith` → `go-modular-cqrs-monolith`.
> The modular-monolith boundary is the headline pattern; CQRS is a sub-convention inside it.
> `-template` is implied by the GitHub template flag.

A production-ready Go project template for the **Modular CQRS Monolith** architecture: the organisational discipline of a distributed system, deployed as a single binary, without the operational overhead of microservices.

Each business domain lives in its own **module** enforced by Go's compiler (`internal/` rule). Modules communicate only through narrow, consumer-defined interfaces. CQRS is a structural convention — no command bus or mediator — making the codebase straightforward to navigate and test.

## Architecture overview

```
modules/<domain>/
├── api/                  ← the ONLY importable surface
│   ├── <domain>_api.go   ← Service (embeds Command + Query), NewService, RegisterRouter
│   └── dto/              ← public types; safe to import from any module
└── internal/             ← compiler-enforced private boundary
    ├── command/           write side: one method per file
    ├── query/             read side:  one method per file
    ├── domain/            entity + single repository interface
    ├── persistence/       repository implementation
    ├── port/              narrow outbound interfaces (consumer-defined)
    └── transport/         Fiber handler + route mounting
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) for a full walkthrough of the pattern, enforcement mechanisms, data flow, and design notes.

## Tech stack

| Concern | Library |
|---|---|
| HTTP | [Fiber v3](https://github.com/gofiber/fiber) |
| Database | [pgx v5](https://github.com/jackc/pgx) + pgxpool |
| Auth | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) |
| Observability | [prometheus/client_golang](https://github.com/prometheus/client_golang) |
| Testing | [testify](https://github.com/stretchr/testify) + [testcontainers-go](https://github.com/testcontainers/testcontainers-go) |
| Config | env vars only (stdlib) |
| Logging | `log/slog` (stdlib) |
| DI | manual constructor injection — no framework |

## Getting started

### 1. Use this template

Clone (or use "Use this template" on GitHub), then rename the module path:

```bash
# replace the module path everywhere
find . -type f -name '*.go' | xargs sed -i '' \
  's|go-modular-cqrs-monolith|github.com/you/your-repo|g'
# update go.mod
go mod edit -module github.com/you/your-repo
go mod tidy
```

### 2. Set environment variables

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/mydb?sslmode=disable"
export JWT_SECRET="change-me"
export AUTH_MODE="local"   # or "gateway" to trust upstream headers
export APP_PORT="8080"     # optional, defaults to 8080
```

### 3. Run

```bash
go run ./cmd/webserver
```

### Built-in endpoints

| Path | Description |
|---|---|
| `GET /healthz` | Liveness + dependency probes (Postgres) |
| `GET /internal/metrics` | Prometheus metrics (guarded by `INTERNAL_API_SECRET`) |

## Project layout

```
cmd/
  webserver/main.go     single entry point; composition root (all manual DI here)
modules/
  order/                full exemplar — command + query + transport + port
  product/              query-only module; no HTTP routes of its own
platform/
  config/               env-var config loader
  httpx/                Fiber app factory, middleware, request/response helpers
  errcode/              typed machine-readable error codes
  logging/              slog context helpers
  metrics/              Prometheus registry + pgx pool collector
```

## Adding a business module

1. Create `modules/<domain>/api/<domain>_api.go` — `Service`, `NewService`, `RegisterRouter`.
2. Add `modules/<domain>/api/dto/` — public DTOs (leaf package, no `internal/` imports).
3. Build out `internal/` (domain, command, query, persistence; add `port/` and `transport/` as needed).
4. Wire the module in `cmd/webserver/main.go` in dependency order.

See [ARCHITECTURE.md](./ARCHITECTURE.md) and `CLAUDE.md` for the full checklist and invariants.

## Recommended follow-up scaffolding

These are not included in the template yet but align with the house pattern:

- `cmd/migrate/` — [goose](https://github.com/pressly/goose) migration binary
- `.env.example` — documented env var reference
- `Makefile` / `justfile` — build, lint, test targets
- `.mockery.yaml` — [mockery](https://github.com/vektra/mockery) config for interface mocks
- `.golangci.yml` — linter config (funlen 100/80, etc.)
- `.github/workflows/ci.yml` — lint + test pipeline
