# Sylix Engine – AI Agent Guide
So this application is a backend for database management (specifically Postgres, but in the future it should also support several databases such as MySQL and MongoDB)

with the hope that key features such as easy installation of production-grade Postgres, multi-database server replication, auto backup, PITR, etc. will be available without any configuration. 

The main concept is that the `controlplane` acts as the master backend (installed on the main server), and the `agent` acts as the backend agent (installed on the main server and child servers).


## Architecture & Layers
- `cmd/main.go` is the only binary entrypoint: open SQLite via `internal/infra/db`, run `database.AutoMigrate`, then start a gRPC server on `:8082`; always inject the DB into services via `grpc.NewServerService(db)` so validators and persistence are wired.
- The codebase follows a clean-ish layering: `internal/common` (shared models/validators), `internal/infra` (SQLite + generated protobuf stubs), and `internal/module/controlplane` (domain entities, repositories, transport adapters).
- `internal/module/controlplane/entity/server.go` defines the canonical `Server` aggregate; other layers must translate to/from this struct rather than duplicating fields.
- Repositories live under `domain/repository`; interfaces describe CRUD with context propagation, and implementations are expected to wrap a `*gorm.DB` (see `server_impl.go`).

## gRPC & Protobuf
- Public APIs are defined in `proto/frontend/server.proto`; regenerate stubs with `make compile-proto-frontend` (invokes `protoc` with `--experimental_allow_proto3_optional` and writes to `internal/infra/proto/server`).
- Generated services/types sit in `internal/infra/proto/server`; `ServerService` in `interface/grpc/server.go` embeds `pbServer.UnimplementedServerServiceServer` and must marshal/unmarshal between protobuf messages and domain entities.
- When adding new RPCs, update the proto first, rerun the Make target, then extend the service implementation—do not hand-edit the generated files.

## Validation & Error Conventions
- `internal/common/validator` wraps `go-playground/validator` and yields `[]ValidationError`; provide user-friendly messages by registering tag overrides via `RegisterTagMessage`.
- `ServerService` composes `validator.ServerValidator`, which enforces structure tags plus domain rules (password XOR SSH key). Reuse this pattern for other aggregates instead of ad-hoc checks.
- Surface validation issues to clients in gRPC errors or response payloads consistently—collect the slice returned by `ValidateStruct` + `validateBusinessRules` before touching the DB.

## Persistence Rules
- SQLite lives in `sylix.db`; schema comes from `database.AutoMigrate` using the `entity.Model` mixin (`internal/common/model/base.go`).
- Repository implementations should call `s.db.WithContext(ctx)` and rely on GORM's error returns; don't panic inside repositories—bubble errors up so the gRPC layer can translate them.
- `ServerCredential` is stored via `gorm:"embedded;embeddedPrefix:credential_"`; ensure new migrations keep this flattened naming.

## Developer Workflows
- Run the server locally with `make run` (alias for `go run ./cmd/main.go`).
- Run lightweight checks with `go test ./...` even though suites are sparse; add focused tests near any non-trivial logic (e.g., validators or repository helpers).
- Proto compilation wipes `internal/infra/proto/*` before regenerating; stage only the relevant regenerated files to avoid accidental deletions of other infra code.

## Patterns to Follow
- Prefer constructor helpers (`NewServerService`, `NewServerRepository`) over struct literals to ensure dependencies are initialized.
- Always pass `context.Context` from gRPC handlers down into repositories/services for future cancellation and logging hooks.
- Keep transport structs (protobuf) separate from domain entities; perform conversions in the gRPC layer and return protobuf message wrappers such as `pbServer.ServerResponse`.
- When extending functionality, touch the layers in order: proto ➜ interface/grpc ➜ domain/service ➜ repository to keep boundaries consistent.

## Code Conventions
- Use `CamelCase` for Go struct fields and `snake_case` for JSON tags.
- Make sure the code follows Go idioms for error handling, avoiding panics except in unrecoverable situations.
- Write clear and concise comments for exported functions and types to enhance code readability and maintainability.
- Reuse existing components and patterns rather than creating new ones for similar functionality.
- Always format code with `go fmt` before committing.


IMPORTANT: DO NOT CREATE REDUNDANT AND NON-OPTIMAL CODE.

IMPORTANT: CREATE CODE THAT IS OPTIMAL, EFFICIENT, MAINTAINABLE, AND CLEAN.