# Implementation Plan: Shopping List SaaS MVP

**Branch**: `001-shopping-list-mvp` | **Date**: 2026-01-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-shopping-list-mvp/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Multi-tenant Shopping List SaaS with Go gRPC backend and iOS/Android mobile clients. Core functionality includes user authentication, group-based multi-tenancy, shopping list and item management, categories, and offline-first synchronization with conflict resolution. The system enforces tenant boundaries at the Group level, uses gRPC with protobuf contracts, and implements version-based conflict detection for reliable offline sync.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: 
- gRPC: `google.golang.org/grpc`
- Protobuf: `google.golang.org/protobuf`
- API contracts: `shopping.v1` protos with buf for linting/breaking checks
- Config: viper or envconfig
- Logging: zap (structured logging)
- Observability: OpenTelemetry (tracing/metrics)
- Auth: JWT (access + refresh tokens)
- Database: PostgreSQL
- Migrations: goose or golang-migrate
- DB layer: sqlc (recommended) or pgx + repository pattern
- Validation: protobuf validation (PGV) or server-side manual validation

**Storage**: PostgreSQL with core tables: users, groups, group_members, categories, lists, items, invites (optional), mutation_log/sync_cursor (optional). All entities include version (BIGINT) and updated_at for conflict detection.

**Testing**: 
- Domain unit tests
- gRPC integration tests (in-memory or dockerized Postgres)
- Contract tests for proto compliance

**Target Platform**: 
- Backend: Linux server (containerized with Docker)
- Clients: iOS + Android (offline-first with SQLite local storage)

**Project Type**: Mobile + API (backend API + mobile clients)

**Performance Goals**: 
- Handle 1000 concurrent users across 500 groups without degradation
- Response times: <500ms for item toggle, <1s for conflict detection, <5s for sync batches up to 100 changes
- Registration/login: <30 seconds end-to-end

**Constraints**: 
- Offline-first client capability (local queue + sync)
- Mobile network resilience: deadline enforcement, idempotency for safe retries
- Multi-tenant isolation enforced on every request
- Version-based conflict resolution (FailedPrecondition on mismatch)

**Scale/Scope**: 
- MVP: 1000 concurrent users, 500 groups
- Mobile clients: iOS 15+, Android API 24+
- gRPC over HTTP/2 transport
- No gateway/web panel in MVP (mobile-only)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Validation

✅ **SaaS Multi-Tenancy & Authorization**: 
- Tenant boundary = Group (aligned)
- Membership enforcement on every request (aligned)
- JWT auth with Bearer tokens (aligned)
- Central authorization interceptor required (aligned)

✅ **API Contracts: gRPC-First, Versioned, Idempotent**:
- gRPC with protobuf package `shopping.v1` (aligned)
- Services: AuthService, GroupService, ListService, SyncService (aligned)
- x-request-id metadata (aligned)
- idempotency-key for writes (aligned)
- Pagination with page_size/page_token/next_page_token (aligned)

✅ **Offline-First Sync & Conflict Resolution**:
- Client offline operation with local queue (aligned)
- SyncService for delta push/pull (aligned)
- Version-based conflict detection (aligned)
- FailedPrecondition on version mismatch (aligned)
- Client deadlines required (aligned)

✅ **Data Model & Migrations Discipline**:
- Core entities: User, Group, GroupMember, Category, List, Item (aligned)
- All entities have created_at, updated_at, updated_by, version (aligned)
- Database migrations mandatory (aligned)
- Item sorting: unpurchased first, priority desc, sort_order asc (aligned)

✅ **Observability, Security & Operations**:
- Structured logging with masked PII/tokens (aligned)
- Metrics: request count, error rate, latency p95 (aligned)
- Rate limits stricter on Auth RPCs (aligned)
- gRPC reflection: dev only (aligned)
- Centralized authorization interceptor (aligned)

**Status**: ✅ All constitution gates pass. Proceeding to Phase 0.

### Post-Phase 1 Design Validation

✅ **SaaS Multi-Tenancy & Authorization**: 
- Data model enforces Group as tenant boundary (all resources have group_id)
- GroupMember table tracks membership with roles and status
- Authorization checks validated in data model queries
- JWT auth flow defined in AuthService proto

✅ **API Contracts: gRPC-First, Versioned, Idempotent**:
- All services defined in protobuf: AuthService, GroupService, ListService, CategoryService, SyncService
- Package `shopping.v1` used consistently
- Pagination implemented via common.proto (PaginationRequest/Response)
- Error handling via gRPC status codes (proto definitions include error scenarios)
- Metadata requirements documented (x-request-id, idempotency-key)

✅ **Offline-First Sync & Conflict Resolution**:
- SyncService proto defines GetDelta and PushMutations
- Version field included in all entity protos (List, Item, Category, Group, GroupMember)
- ExpectedVersion parameter in update operations
- Mutation queue pattern defined in PushMutationsRequest
- Cursor-based delta sync via GetDeltaRequest/Response

✅ **Data Model & Migrations Discipline**:
- All entities include created_at, updated_at, updated_by, version
- Core entities match constitution: User, Group, GroupMember, Category, List, Item
- Item sorting rules documented (is_purchased=false first, priority desc, sort_order asc)
- Database migrations structure defined (goose/golang-migrate)

✅ **Observability, Security & Operations**:
- Structured logging approach documented (zap)
- OpenTelemetry integration planned
- Request ID propagation via x-request-id metadata
- PII masking requirements noted
- Rate limiting strategy defined (stricter on Auth)

**Status**: ✅ All constitution gates pass after Phase 1 design. Ready for implementation.

## Project Structure

### Documentation (this feature)

```text
specs/001-shopping-list-mvp/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/          # Phase 1 output (/speckit.plan command)
│   ├── auth.proto
│   ├── group.proto
│   ├── list.proto
│   ├── category.proto
│   └── sync.proto
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/          # Business logic
│   │   ├── auth/
│   │   ├── group/
│   │   ├── list/
│   │   ├── category/
│   │   └── sync/
│   ├── transport/       # gRPC handlers
│   │   ├── grpc/
│   │   │   ├── auth_handler.go
│   │   │   ├── group_handler.go
│   │   │   ├── list_handler.go
│   │   │   ├── category_handler.go
│   │   │   └── sync_handler.go
│   │   └── interceptors/
│   │       ├── request_id.go
│   │       ├── logging.go
│   │       ├── auth.go
│   │       ├── authorization.go
│   │       ├── idempotency.go
│   │       └── metrics.go
│   ├── storage/         # Repository interfaces + implementations
│   │   ├── repositories/
│   │   │   ├── user_repo.go
│   │   │   ├── group_repo.go
│   │   │   ├── list_repo.go
│   │   │   ├── item_repo.go
│   │   │   ├── category_repo.go
│   │   │   └── sync_repo.go
│   │   └── postgres/    # sqlc generated or pgx implementations
│   │       ├── queries.sql
│   │       └── models.go
│   └── config/          # Configuration management
│       └── config.go
├── pkg/                 # Shared packages
│   ├── auth/           # JWT handling
│   ├── errors/         # Error mapping
│   └── validation/     # Input validation
├── migrations/         # Database migrations (goose/golang-migrate)
│   └── *.sql
├── proto/              # Protobuf definitions
│   └── shopping/
│       └── v1/
│           ├── auth.proto
│           ├── group.proto
│           ├── list.proto
│           ├── category.proto
│           └── sync.proto
├── buf.yaml            # Buf configuration
├── buf.gen.yaml        # Code generation config
└── go.mod

tests/
├── contract/           # Contract tests for proto compliance
├── integration/        # gRPC integration tests
└── unit/               # Domain unit tests

ios/                    # iOS client (future)
└── [platform-specific structure]

android/                # Android client (future)
└── [platform-specific structure]
```

**Structure Decision**: Mobile + API structure selected. Backend API is a Go gRPC service with layered architecture (transport → domain → storage). Mobile clients (iOS/Android) will be separate projects with offline-first SQLite storage and sync engine. The API follows clean architecture principles with clear separation of concerns.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. All technical choices align with constitution principles.
