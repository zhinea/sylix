---
trigger: always_on
---

# Sylix Engine – AI Agent Guide

Sylix Engine is a database management backend (Postgres focus) with a `controlplane` (master) and `agent` (node) architecture. It includes a React/Vite frontend (`ui/dashboard`).

## 🏗 Architecture & Boundaries

### Backend (Go)
- **Entry Points**:
  - `cmd/main.go`: Controlplane. Starts SQLite, runs migrations, wraps gRPC server with `grpc-web`, listens on `:8082`.
  - `cmd/agent/main.go`: Agent. Connects to controlplane or listens for commands.
- **Layering**:
  - `internal/module/{module}`: Vertical slices (controlplane, agent).
  - `entity`: Domain models (pure Go structs).
  - `repository`: Interfaces in `domain/repository`, GORM impls in `domain/repository/*_impl.go`.
  - `interface/grpc`: gRPC handlers. Maps proto messages <-> domain entities.
- **Data**: SQLite (`sylix.db`) via GORM.
- **Communication**: gRPC (Inter-service) & gRPC-Web (Frontend -> Controlplane).

### Frontend (React Router v7 + Vite)
- **Location**: `ui/dashboard`.
- **Framework**: React Router v7 (SPA mode with `clientLoader`/`clientAction`).
- **Network**: Custom `GrpcWebClient` (`lib/grpc-client.ts`) handling raw gRPC-Web frames.
- **State**: URL-driven state via React Router loaders/actions.

## 🛠 Developer Workflows

### Backend
- **Run Controlplane**: `make run` (starts on `:8082`).
- **Dev Mode**: `make dev` (uses `gowatch` for hot reload).
- **Proto Generation**: `make compile-proto` (requires `protoc`).
  - *Note*: Regenerates `internal/infra/proto`.

### Frontend
- **Dev Server**: `cd ui/dashboard && npm run dev` (starts on `:5173`).
- **Proto Generation**: `make compile-proto-frontend`.
  - *Note*: Generates TS types in `ui/dashboard/app/proto`.
  - *Critical*: Run this after any `.proto` change.

## 🧩 Key Patterns & Conventions

### Go (Backend)
- **Dependency Injection**: Manual wiring in `cmd/main.go`. Always inject DB/Repos into Services.
- **Validation**: `internal/common/validator`. Use `ValidateStruct` in gRPC handlers before business logic.
- **Error Handling**: Bubble errors from Repos. Translate to gRPC codes in `interface/grpc`.
- **Database**: Use `gorm` with `context`. Do not panic in repos.

### TypeScript (Frontend)
- **Data Fetching**: Use `clientLoader` for reads and `clientAction` for writes (React Router v7).
  - Example: `export async function clientLoader() { ... }` in routes.
- **API Calls**: Import services from `~/lib/api`.
  - Example: `import { serverService } from "~/lib/api";`
- **Forms**: Use `react-hook-form` + `zod` + `shadcn/ui` components.
- **gRPC Client**: `GrpcWebClient` manually constructs frames. *Do not use standard grpc-web libraries unless replacing this implementation.*

## 🔌 Integration Points
- **Frontend -> Backend**:
  - Frontend calls `http://localhost:8082` (proxied via `grpc-web`).
  - CORS enabled in `cmd/main.go` for `localhost:5173`.
- **Proto Files**:
  - Source of truth: `proto/`.
  - Changes require running **BOTH** `make compile-proto` and `make compile-proto-frontend`.

## ⚠️ Gotchas
- **gRPC-Web**: The backend manually wraps the gRPC server using `improbable-eng/grpc-web`. The frontend uses a custom binary frame parser/builder.
- **Migrations**: `database.AutoMigrate` runs on startup in `cmd/main.go`.


## Flow
`interface -> app -> (optional: domain/services) -> domain/repository -> entity`

All business logic is stored in the domain/services, such as installing agents, querying databases, calculating something, etc.

- interface: for connecting the “app” with grpc
- app: as a connector between the interface and services, where services are responsible for business logic.
- domain/services:  main business logic app, but not all of it is placed in services. The ideal portion is between the app and services.
- domain/repository: for database query implementation

## Context
Okay, as context, I want to create a Postgres management database, where there is a main (control plane) and node (agent) which by default already enable backup storage, PITR, WAL, branching, easy Postgres version changes, and other production-grade Postgres configurations.

IMPORTANT: There should be no redundant, inefficient, or duplicate code.

IMPORTANT: CREATE MAINTENABLE AND CLEAN CODE.