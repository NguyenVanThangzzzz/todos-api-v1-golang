# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A Go REST API for managing todos, built with **Gin + GORM + PostgreSQL** following Clean Architecture. Module path: `github.com/thangnguyen/todo_api_v1`. Entry point is `cmd/api/main.go`.

A long-form Vietnamese walkthrough comparing each piece to its Node.js equivalent lives in [`golang.md`](golang.md) — read it for design intent before refactoring.

## Common commands

Run from the project root (`todos-api-v1-golang/`):

```bash
go mod tidy                        # install deps after touching imports
go run ./cmd/api                   # start the server
go build -o todo-api ./cmd/api     # produce a single binary
go build ./... && go vet ./...     # full compile + vet check (no test files currently)
```

There is no test suite yet — `go test ./...` is a no-op. Add `*_test.go` files next to the code they cover.

### Runtime prerequisites

The server **will not start** without these:
- PostgreSQL reachable at `DB_HOST:DB_PORT` (defaults `localhost:5432`).
- A database named `todo_api_golang` (or whatever `DB_NAME` you set) must exist — GORM does not create the database, only tables.
- Schema is handled by `db.AutoMigrate(...)` on every startup — never run a separate migration tool.
- Copy `.env.example` → `.env` and fill in real values, or set env vars directly.

### Environment variables

Config is loaded by **Viper** in `config/config.go`. It reads from a `.env` file first, then env vars override. All keys and defaults:

| Var | Default | Notes |
|---|---|---|
| `PORT` | `3636` | HTTP port |
| `APP_ENV` | `development` | |
| `DB_HOST` | `localhost` | |
| `DB_PORT` | `5432` | |
| `DB_USER` | `postgres` | |
| `DB_PASSWORD` | `123123` | override in real environments |
| `DB_NAME` | `todo_api_golang` | |
| `DB_SSLMODE` | `disable` | |
| `JWT_SECRET` | `dev-secret-change-me` | must be changed in production |
| `JWT_ACCESS_TTL_MINUTES` | `15` | access token lifetime in minutes |

PowerShell override example: `$env:DB_PASSWORD = "..."; go run ./cmd/api`

## Architecture

### Directory structure

```
cmd/
  api/
    main.go                  ← entry point, DI wiring
config/
  config.go                  ← Viper: loads .env + env vars into Config struct
internal/
  domain/                    ← entities shared across modules
  middleware/                ← Auth, Recovery, RequestLogger
  modules/
    auth/                    ← auth module (self-contained)
      auth_dto.go
      auth_handler.go
      auth_repository.go
      auth_routes.go         ← auth.RegisterRoutes(v1, h)
      auth_usecase.go
      auth_validator.go      ← local validateStruct(), not exported
    todo/                    ← todo module (self-contained)
      todo_dto.go
      todo_handler.go
      todo_repository.go
      todo_routes.go         ← todo.RegisterRoutes(v1, h, jwtManager)
      todo_usecase.go
      todo_validator.go
  router/
    router.go                ← bootstraps engine, calls each module's RegisterRoutes
pkg/
  database/                  ← postgres.New(cfg.DB)
  hash/                      ← bcrypt helpers
  jwt/                       ← jwt.NewManager(cfg.JWT)
  logger/                    ← zap wrapper
  response/                  ← uniform JSON envelope
```

### Layering

```
handler  →  usecase  →  repository (interface)  →  domain
              ↑                ↑
           dto/validator    *gorm.DB (Postgres impl)
```

Strict inward dependency rule: `domain` knows nothing else; `usecase` depends only on the repository interface and `domain`; `handler` depends on `usecase` + local dto/validator.

DI is wired manually in `cmd/api/main.go` via `New<Type>(...)` constructors (no framework). When adding a new dependency, thread it through the constructor chain there.

### Route registration pattern

Each module owns its routes. `router.go` only calls:
```go
auth.RegisterRoutes(v1, authHandler)
todo.RegisterRoutes(v1, todoHandler, jwtManager)
```
When adding a new module, create `<module>_routes.go` with a `RegisterRoutes` func and call it in `router.go`.

### Key conventions baked into the code

- **Validation is per-module** — each module has its own `<module>_validator.go` with a package-local `validateStruct()` and `formatFieldError()`. There is no shared `pkg/validator`. When adding a new validator tag, update `formatFieldError` in the relevant module.
- **`domain.Todo` / `domain.User` are both API model and DB row** — same struct carries `json:"..."` and `gorm:"..."` tags. No separate ORM entity.
- **`domain.ErrNotFound` is the contract**, not `gorm.ErrRecordNotFound`. Postgres repos translate GORM errors; handlers match on `errors.Is(err, domain.ErrNotFound)` to return 404.
- **`Update` uses `db.Save` (not `db.Updates`)** so zero values like `Completed=false` persist. Checks `RowsAffected == 0` to surface `ErrNotFound`. `Delete` uses the same pattern.
- **PATCH uses pointer fields** — `UpdateRequest` has `*string`/`*bool` so usecase can distinguish nil (not sent) from zero value (explicitly sent). Preserve this for any new partial-update DTO.
- **All HTTP responses go through `pkg/response`** — `response.OK(c, status, message, data)` / `response.Fail(c, status, errMsg)`. Do not call `c.JSON` directly from handlers.
- **Logger is `*zap.Logger`** aliased as `logger.Logger`, currently development config (colored, console). For production, switch to `zap.NewProduction()` in `pkg/logger/logger.go`.
- **Middleware order in `router.New`**: `Recovery` → `RequestLogger`. Register new global middleware there; route-group middleware goes inside each module's `RegisterRoutes`.
- **Graceful shutdown**: `cmd/api/main.go` listens for `SIGINT`/`SIGTERM` and calls `srv.Shutdown` with a 5s timeout.

### Where to put new code

| Adding... | Goes in |
|---|---|
| A new entity | `internal/domain/<name>.go` with `json:` + `gorm:` tags + `TableName()`; register in `db.AutoMigrate(...)` in `cmd/api/main.go` |
| A new module | `internal/modules/<name>/` with `<name>_handler.go`, `<name>_usecase.go`, `<name>_repository.go`, `<name>_dto.go`, `<name>_validator.go`, `<name>_routes.go`; call `RegisterRoutes` in `router.go` |
| A new endpoint in existing module | DTO in `<module>_dto.go`, method on usecase, handler method, route in `<module>_routes.go` |
| Config value | Add default in `config/config.go`, add field to the relevant `*Config` struct, add to `.env.example` |
| Shared utility usable by other projects | `pkg/<name>/` |
| Internal-only helper | Inside `internal/` |
