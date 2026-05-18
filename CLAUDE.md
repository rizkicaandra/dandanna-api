# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run            # Run the application (go run ./cmd/api/main.go)
make dev            # Hot-reload dev server using air (requires: make install-tools)
make build          # Build binary to bin/dandanna-api
make test           # Run all tests with race detector
make test-coverage  # Run tests and generate coverage.html
make lint           # Run golangci-lint (requires installation)
make fmt            # Format all Go code with gofmt -s
make vet            # Run go vet
make tidy           # go mod tidy + verify
make install-tools  # Install air and golangci-lint
```

Run a single test package:
```bash
go test -v -race ./internal/api/handler/...
```

## Architecture

Clean Architecture with strict inward dependency: `api → application → domain ← infrastructure`.

```
cmd/api/main.go          # Entry point: wires all layers, starts HTTP server
internal/
  api/                   # HTTP layer — depends on application/domain
    handler/             # One handler per resource (implements router interface)
    middleware/          # RequestID, Logging, Recovery, CORS, SecurityHeaders
    router/              # Route registration; middleware chain order is significant
    response/            # response.JSON() / response.Error() write the envelope
    dto/                 # Request/response shapes; never exposed to domain
  application/
    service/             # Orchestrates domain operations; no HTTP knowledge
  domain/
    entity/              # Core business objects
    repository/          # Interfaces only — no implementations here
    service/             # Domain logic that doesn't belong to a single entity
    valueobject/         # Immutable value types
    error/               # Typed domain errors (NotFound, Conflict, Validation…)
  infrastructure/
    config/              # Env-var config; validated at startup, fails fast
    logger/              # zerolog-backed Logger interface
    postgres/            # *sqlx.DB wrapper; satisfies Pinger
    redis/               # *redis.Client wrapper; satisfies Pinger
```

## Code Comments

Every file must have comments that explain the **why**, not the what. Follow these rules:

**Section dividers in multi-step files** (e.g. `main.go`) — use numbered banners:
```go
// ── 1. Configuration ──────────────────────────────────────────────────────
```

**Package-level declarations** — one-line comment above each exported type, interface, and constructor:
```go
// ArtistServicer defines the contract for artist business operations.
type ArtistServicer interface { ... }

// NewArtistService creates a new ArtistService with the given dependencies.
func NewArtistService(...) *ArtistService { ... }
```

**Non-obvious logic** — comment the why, never the what:
```go
// Fail fast: no point serving traffic without a database connection.
// argon2.IDKey parameters: time=1, memory=64MB, threads=4 — OWASP minimum.
```

**Do not comment:**
- Self-explanatory variable names or assignments
- Standard library calls that are clear from context
- Each line of a simple CRUD operation

## Key Conventions

**Response envelope** — every endpoint must use `response.JSON()` or `response.Error()`. Never write raw JSON directly.

```go
// success
response.JSON(w, r, dto.MyData{...}, http.StatusOK)

// error
response.Error(w, r, http.StatusUnprocessableEntity,
    response.NewError("field", "ERROR_CODE", "human message"),
)
```

Error codes are always `SCREAMING_SNAKE_CASE`. `errors[].field` is omitted for non-field errors (auth, server errors). `errors[].meta` is `{}` when empty, never null.

**Domain errors** — handlers map typed domain errors to HTTP status codes:

| Domain error type     | HTTP status |
|-----------------------|-------------|
| `*NotFound`           | 404         |
| `*Conflict`           | 409         |
| `*Validation`         | 422         |
| `*Unauthorized`       | 401         |
| `*Forbidden`          | 403         |
| `*Unprocessable`      | 422         |

**Logger** — use the `logger.Logger` interface everywhere; never import zerolog directly outside `internal/infrastructure/logger`. Use typed field constructors: `logger.String`, `logger.Int`, `logger.Err`, `logger.Duration`.

**Interfaces over types** — handlers depend on interfaces (`HealthHandler`, `Pinger`), not concrete structs. Follow this pattern for all new handlers and services.

**Middleware chain** (outermost → innermost): `Recovery → SecurityHeaders → CORS → RequestID → Logging`. Request ID is available via `middleware.RequestIDFromContext(ctx)`.

**Config** — all configuration comes from environment variables via `internal/infrastructure/config`. `.env` is loaded automatically in development (ignored in production). Production rejects wildcard `CORS_ALLOWED_ORIGINS`.

## HTTP

Uses Go standard library `net/http` only — no third-party router. Route patterns use Go 1.22+ method+path syntax: `"GET /api/path"`.

Endpoints:
- `GET /api/health` — liveness (always 200 if process is running)
- `GET /api/readyz` — readiness (pings Postgres + Redis, 503 if either fails)

## Database / Cache

- PostgreSQL via `github.com/jmoiron/sqlx` + `jackc/pgx/v5` driver
- Redis via `github.com/redis/go-redis/v9`
- DB schema uses DBML in `dbml/`; migrations in `migrations/`

## OpenSpec Workflow

Feature proposals and task lists live in `openspec/changes/`. Use the `/opsx:propose`, `/opsx:apply`, and `/opsx:archive` skills to manage the change lifecycle.
