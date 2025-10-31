## Quick, practical guidance for AI coding agents

This file captures the concrete, discoverable conventions and workflows an AI should know to be productive in this repo.

### Big picture
- Project is a Go (1.21) HTTP API using Gin. Entry point: `cmd/main.go`.
- Clean-ish hex/adapters layering: `internal/api` (HTTP layer), `internal/domain` (services, business logic), `internal/infrastructure` (adapters, database, config), and `pkg` (reusable libs).
- Routes and middleware are defined in `internal/api/routes/routes.go`. Handlers live under `internal/api/handlers` and are typically constructed with services and a logger.

### How the app wires up (important places to edit)
- Config: `internal/infrastructure/config/config.go` (uses Viper). Required env vars that override config: `DATABASE_URL`, `JWT_SECRET`, `ENCRYPTION_KEY`, `PORT`.
- DB connection + migrations: `internal/infrastructure/database/*` and `migrations/`. Note: `cmd/main.go` calls `database.RunMigrations(...)` on startup.
- Router initialization: `routes.SetupRoutes(db, cfg, log)` — middleware and route groups are configured here (auth, admin, webhooks, metrics, swagger).
- Domain services are in `internal/domain/services/*` and adapters in `internal/infrastructure/adapters/*`. Many places currently expect services to be wired (there are TODOs in `routes.go`).

### Developer workflows (concrete commands)
- Run locally (reads config from `configs/config.yaml` or env): `make run` (invokes `go run ./cmd/main.go`).
- Build: `make build` or production `make build-prod`.
- Docker: `make docker-build` and `make docker-run` (uses `docker-compose.yml`).
- Migrations: `make migrate-up` / `make migrate-down` (uses `migrate` CLI and `migrations/` folder). `cmd/main.go` also runs migrations automatically on startup.
- Tests: `make test` (runs `go test ./...`). Integration/e2e tests live under `tests/integration` and `tests/e2e`.
- Lint/format: `make lint` / `make format` (golangci-lint / goimports).

### Project-specific conventions and patterns
- Handler factories: Two patterns exist:
  - Handler constructors that accept domain service instances: e.g. `handlers.NewFundingHandlers(fundingService, log)` (see `routes.go`).
  - Handler closures that directly accept `db, cfg, log` and return `gin.HandlerFunc`: e.g. `handlers.Register(db, cfg, log)`.
  Prefer the constructor pattern when adding new feature endpoints; follow existing adjacent files.
- Middleware: central middleware lives in `internal/api/middleware/middleware.go`. Authentication is applied in `routes.go` with `middleware.Authentication(cfg, log)`.
- Config-loading: `config.Load()` constructs `Database.URL` if missing and validates `JWT.Secret` and `Security.EncryptionKey` — ensure those exist in test/dev runs.
- Swagger: generated via `swag` and `make gen-swagger`; served on `/swagger/*` in non-production.

### Integration points and external dependencies
- PostgreSQL (connection via `DATABASE_URL` or the config struct), migrations in `migrations/`.
- Redis config under `internal/infrastructure/config` — used by cache adapters (look under `internal/infrastructure/cache`).
- Blockchain adapters: `internal/infrastructure/adapters/blockchain` and `internal/infrastructure/adapters/brokerage` (RPC endpoints configured in config.yaml).
- Payment/card processors: configured in `configs/config.yaml` and wired via `internal/infrastructure/adapters/*`.

### Common quick tasks for an AI to perform (contract + examples)
- Task: Add a new authenticated GET endpoint `/api/v1/foo` that returns user-specific data.
  - Inputs: DB connection (`*sql.DB`), config (`*config.Config`), logger (`*logger.Logger`).
  - Output: `gin.HandlerFunc` registered in `routes.SetupRoutes` under an appropriate group.
  - Error modes: missing service wiring — if service doesn't exist, add an interface under `internal/domain/repositories` and a simple in-memory adapter for tests under `internal/infrastructure/adapters`.
  - Example: follow pattern in `routes.go` for `wallets` or `baskets` groups.

### Files and locations to inspect first
- App entry: `cmd/main.go`
- Router & route groups: `internal/api/routes/routes.go`
- Handlers: `internal/api/handlers/*.go`
- Middleware: `internal/api/middleware/middleware.go`
- Domain services: `internal/domain/services/*`
- Adapters: `internal/infrastructure/adapters/*`
- Config: `configs/config.yaml` and `internal/infrastructure/config/config.go`
- Migrations: `migrations/`
- Makefile: top-level `Makefile` (common dev/test commands)

### Edge cases and gotchas (observed from source)
- Many service variables in `routes.SetupRoutes` are left `nil` (there are TODOs). Don't assume services are pre-wired — wire concrete implementations or add minimal stubs for tests.
- Running the app (`make run`) will attempt DB migrations. Ensure a reachable `DATABASE_URL` or run with a disposable dev DB via `docker-compose`.
- Config validation requires `JWT_SECRET` and `ENCRYPTION_KEY` — CI/dev runs must set them.

If anything in these notes is unclear or you'd like me to expand a specific section (example wiring of a service, a template handler, or tests), tell me which bit to expand and I will update this file. 
