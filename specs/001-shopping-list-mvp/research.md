# Research & Technical Decisions: Shopping List SaaS MVP

**Created**: 2026-01-07  
**Feature**: Shopping List SaaS MVP

## Architecture Decisions

### Decision: Multi-tenant SaaS with Group-based Tenant Boundary

**Rationale**: Groups serve as the tenant boundary, providing natural isolation for collaborative shopping lists while allowing users to belong to multiple groups (personal, family, shared). This simplifies authorization checks and data isolation compared to user-scoped or organization-scoped models.

**Alternatives considered**:
- User-scoped: Rejected - doesn't support collaboration
- Organization-scoped: Rejected - too complex for MVP, overkill for personal use cases
- Hybrid: Rejected - adds complexity without clear benefit for MVP

### Decision: Go + gRPC Backend

**Rationale**: 
- Go provides excellent performance, concurrency, and type safety for backend services
- gRPC offers efficient binary protocol, strong typing via protobuf, and excellent mobile support
- HTTP/2 provides multiplexing and better performance than HTTP/1.1
- Protobuf-first approach ensures contract stability and code generation

**Alternatives considered**:
- REST API: Rejected - less efficient for mobile, weaker typing, more boilerplate
- GraphQL: Rejected - adds complexity, not ideal for mobile with offline sync needs
- gRPC-Web: Not needed for MVP (mobile-only), can add later if web admin needed

### Decision: PostgreSQL for Primary Storage

**Rationale**:
- ACID guarantees essential for multi-tenant data isolation
- Strong consistency required for conflict resolution
- Excellent support for transactions and complex queries
- Mature ecosystem with migration tools (goose/golang-migrate)
- Version column (BIGINT) for optimistic locking

**Alternatives considered**:
- MongoDB: Rejected - eventual consistency not suitable for conflict resolution
- SQLite: Rejected - not suitable for multi-tenant server-side (concurrency limits)
- DynamoDB: Rejected - adds vendor lock-in, overkill for MVP scale

### Decision: sqlc or pgx + Repository Pattern

**Rationale**:
- sqlc generates type-safe Go code from SQL, reducing boilerplate and errors
- pgx provides excellent PostgreSQL driver with connection pooling
- Repository pattern abstracts storage layer, enabling testing and future changes
- Type safety prevents SQL injection and runtime errors

**Alternatives considered**:
- GORM/ent: Rejected - adds ORM overhead, less control over queries
- Raw SQL with database/sql: Rejected - too much boilerplate, error-prone
- sqlc preferred for type safety, pgx as fallback if sqlc doesn't fit

### Decision: Version-based Conflict Resolution (LWW with Precondition)

**Rationale**:
- Simple and predictable for MVP
- Version column enables optimistic locking
- FailedPrecondition error clearly signals conflict to client
- Client can refresh and retry with latest version
- More deterministic than timestamp-only LWW

**Alternatives considered**:
- CRDT: Rejected - too complex for MVP, requires significant research
- Timestamp-only LWW: Rejected - clock skew issues, less reliable
- Operational Transform: Rejected - overkill for shopping lists, complex to implement

### Decision: Global Change Sequence for Sync Cursor

**Rationale**:
- More reliable than updated_at-based cursors (handles clock skew, concurrent updates)
- Enables precise delta sync with clear ordering
- Supports mutation_log table for audit trail
- Simpler conflict detection and resolution

**Alternatives considered**:
- updated_at cursor: Rejected - clock skew issues, less reliable ordering
- Per-table cursors: Rejected - complex to merge, ordering unclear
- Hybrid: Considered but global sequence simpler for MVP

### Decision: JWT Access + Refresh Tokens

**Rationale**:
- Stateless authentication reduces server-side session storage
- Refresh tokens enable secure token rotation
- Standard approach for mobile apps
- Short-lived access tokens limit exposure if compromised

**Alternatives considered**:
- Session-based auth: Rejected - requires server-side storage, less scalable
- OAuth2 with external provider: Rejected - adds complexity, not needed for MVP
- API keys: Rejected - no user context, not suitable for multi-user app

### Decision: zap for Structured Logging

**Rationale**:
- High performance (zero-allocation in hot paths)
- Structured logging enables better observability
- JSON output compatible with log aggregation systems
- Mature and widely used in Go ecosystem

**Alternatives considered**:
- logrus: Rejected - slower, less structured
- Standard log: Rejected - no structured logging support
- zerolog: Considered but zap more mature

### Decision: OpenTelemetry for Observability

**Rationale**:
- Industry standard for distributed tracing
- Vendor-agnostic (can switch backends)
- Supports both tracing and metrics
- Good integration with gRPC

**Alternatives considered**:
- Prometheus only: Rejected - no distributed tracing
- Jaeger only: Rejected - vendor lock-in
- Custom metrics: Rejected - reinventing the wheel

### Decision: buf for Protobuf Management

**Rationale**:
- Linting prevents common proto mistakes
- Breaking change detection ensures contract stability
- Code generation integration
- Better than raw protoc for large projects

**Alternatives considered**:
- protoc only: Rejected - no linting, manual breaking change detection
- prototool: Rejected - deprecated, buf is successor

### Decision: goose or golang-migrate for Database Migrations

**Rationale**:
- Version-controlled schema changes
- Rollback support
- CI/CD integration
- Both are mature and well-maintained

**Alternatives considered**:
- Manual SQL scripts: Rejected - no versioning, error-prone
- GORM AutoMigrate: Rejected - not suitable for production, loses control
- Flyway: Rejected - Java-based, Go-native tools preferred

### Decision: No Gateway/Web Panel in MVP

**Rationale**:
- Mobile-only MVP reduces scope
- gRPC-gateway can be added later if needed
- Focus on core mobile experience first
- Reduces infrastructure complexity

**Alternatives considered**:
- gRPC-gateway: Deferred - can add for admin/debug later
- Web admin panel: Deferred - not in MVP scope

### Decision: No Cache/Queue in MVP

**Rationale**:
- PostgreSQL sufficient for MVP scale (1000 concurrent users)
- Reduces infrastructure complexity
- Can add Redis later if needed for performance
- Simplifies deployment and operations

**Alternatives considered**:
- Redis cache: Deferred - not needed for MVP scale
- Message queue: Deferred - sync is request/response, no async processing needed

## Mobile Client Architecture (Future)

### Decision: Offline-First with SQLite

**Rationale**:
- SQLite provides local persistence
- Enables offline operation
- Sync engine queues mutations locally
- Optimistic UI updates before sync

### Decision: Mutation Queue for Offline Changes

**Rationale**:
- Queue local changes when offline
- Batch mutations for efficient sync
- Retry failed mutations
- Maintain operation order

## Security Considerations

### Decision: Centralized Authorization Interceptor

**Rationale**:
- Single point for membership checks
- Consistent enforcement across all RPCs
- Easier to audit and maintain
- Prevents authorization bugs

### Decision: Rate Limiting (Stricter on Auth)

**Rationale**:
- Prevents brute force attacks on auth endpoints
- Protects against abuse
- Standard security practice

### Decision: PII/Token Masking in Logs

**Rationale**:
- Prevents sensitive data leakage in logs
- Compliance with privacy requirements
- Security best practice

## Performance Considerations

### Decision: Client Deadlines Required

**Rationale**:
- Prevents hanging requests on mobile networks
- Better user experience
- Resource management

### Decision: Idempotency Keys for Writes

**Rationale**:
- Enables safe retries on mobile networks
- Prevents duplicate operations
- Essential for unreliable mobile connectivity

### Decision: Pagination for List Operations

**Rationale**:
- Prevents large payloads
- Better performance on mobile
- Reduces memory usage

## Testing Strategy

### Decision: Domain Unit Tests + gRPC Integration Tests

**Rationale**:
- Unit tests for business logic isolation
- Integration tests for contract compliance
- In-memory or dockerized Postgres for integration tests
- Ensures proto contracts are correct

**Alternatives considered**:
- E2E only: Rejected - too slow, harder to debug
- Unit only: Rejected - doesn't test contracts

## Deployment Considerations

### Decision: Docker Containerization

**Rationale**:
- Consistent environments
- Easy deployment to K8s/ECS/VM
- Reproducible builds

### Decision: Managed PostgreSQL

**Rationale**:
- Reduces operational overhead
- Built-in backups and monitoring
- Focus on application code

## Summary

All major technical decisions align with constitution principles. The architecture prioritizes:
1. Multi-tenant isolation via Groups
2. Mobile-friendly gRPC contracts
3. Offline-first sync with version conflicts
4. Observability and security
5. Simplicity for MVP scope

No unresolved clarifications remain. Ready for Phase 1 design.
