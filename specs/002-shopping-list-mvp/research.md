# Research: Mobile Offline MVP (Android + iOS)

**Created**: 2026-01-07  
**Branch**: `002-shopping-list-mvp`  
**Goal**: Resolve open technical decisions needed to produce the shared core + minimal native UIs described in the feature spec.

## Decisions

### Decision: gRPC contract source of truth

- **Chosen**: Use the existing canonical protobufs under `api/proto/shopping/v1` (`shopping.v1`) as the only source of truth for client stubs.
- **Rationale**: Aligns with the constitution’s “gRPC-first” rule and avoids contract drift between server and clients.
- **Alternatives considered**:
  - Copy/maintain a separate `proto/` tree for mobile: rejected due to drift risk.
  - Replace with REST/OpenAPI: rejected (conflicts with constitution; server is gRPC).

### Decision: Sync cursor type and semantics

- **Chosen**: Use `GetDeltaRequest.cursor` as an `int64` sequence cursor (0 for first sync) and persist per-group cursor locally.
- **Rationale**: Matches `shopping.v1.SyncService` (`GetDeltaRequest.cursor`, `GetDeltaResponse.next_cursor`, `has_more`) and supports safe incremental sync.
- **Alternatives considered**:
  - Timestamp-based cursor: rejected due to ordering ambiguity under clock skew.

### Decision: Mutation payload encoding for `PushMutations`

- **Chosen**: Encode `Mutation.entity_data` as JSON bytes for all mutation types (create/update), and keep `entity_type` as lower-case strings (`"list"`, `"item"`, `"category"`).
- **Rationale**: Server-side sync domain currently treats `EntityData` as JSON and validates it as JSON (`json.Unmarshal`), and validates entity types against those strings.
- **Alternatives considered**:
  - Protobuf-encoded `Any` payload: deferred (would require server+client changes and tighter schema coupling).

### Decision: Conflict handling policy

- **Chosen**: On `FailedPrecondition` (version mismatch) during a write:
  - Pull latest delta for the group
  - Reconcile local state using “server wins” for conflicted entity
  - Re-queue or re-attempt the user intent with updated version when applicable
- **Rationale**: Constitution mandates refresh+retry; “server wins” is the MVP-safe rule that prevents divergent state.
- **Alternatives considered**:
  - Field-level merge: rejected for MVP complexity.
  - “client wins”: rejected due to overwrite risk.

### Decision: Offline-first persistence & pending UX

- **Chosen**: UI renders from local DB only; all writes first update local DB and append a queue entry marked pending; confirmation clears pending.
- **Rationale**: Ensures deterministic offline behavior and avoids UI depending on network conditions.
- **Alternatives considered**:
  - Render directly from in-memory state: rejected due to restart durability requirements.

### Decision: Client retry policy

- **Chosen**:
  - Reads: bounded retries (e.g., 2–3) with backoff when safe
  - Writes: no automatic retry beyond idempotent re-send during sync loop; use idempotency key per queued mutation
  - All RPCs: default deadline (5–10s), override per call if needed
- **Rationale**: Matches constitution rules: deadlines required, retries bounded, and only safe operations retried.
- **Alternatives considered**:
  - Unbounded retries: rejected.

## Key Findings (from existing contracts)

- `SyncService.GetDelta` is group-scoped and cursor-based (`int64 cursor`, `int64 next_cursor`, `bool has_more`).
- `SyncService.PushMutations` is batch-based and returns per-mutation results (`MutationResult`).
- Lists/Items include `version` fields used for conflict detection (`expected_version` on update/toggle).
- Pagination exists for list endpoints (`PaginationRequest/Response`), and clients must be prepared to page where used.

## Open Questions (none blocking for Phase 1 design)

- iOS gRPC client library/transport choice will be finalized during implementation based on build/tooling constraints, but the adapter boundary isolates this decision.
