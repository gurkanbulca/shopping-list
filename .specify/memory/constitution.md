<!--
- Sync Impact Report
- Version change: N/A → 1.0.0
- Modified principles: NEW (filled from template placeholders)
- Added sections: Scope & MVP Boundaries; Transport & Mobile Readiness
- Removed sections: None
- Templates reviewed:
  - ✅ .specify/templates/plan-template.md (Constitution Check aligns; no edits required)
  - ✅ .specify/templates/spec-template.md (no conflicting requirements)
  - ✅ .specify/templates/tasks-template.md (task groupings compatible)
  - ⚠ No command templates found; plan template references `.specify/templates/commands/plan.md` (follow-up)
- Deferred items:
  - TODO(RATIFICATION_DATE): Original adoption date not available
-->

# Shopping List Constitution
<!-- Example: Spec Constitution, TaskFlow Constitution, etc. -->

## Core Principles

### SaaS Multi-Tenancy & Authorization
<!-- Example: I. Library-First -->
Non‑negotiable rules:
- Tenant boundary is Group. All resources are tenant‑scoped.
- Users MAY access only Groups they are members of. Enforce on every request.
- Central authorization interceptor MUST validate membership for any write via `group_id`.
- Auth via JWT access token in `authorization: Bearer ...`. Refresh handled via dedicated RPC.
Rationale: Enforces strict isolation for a multi‑tenant SaaS while keeping the model simple.
<!-- Example: Every feature starts as a standalone library; Libraries must be self-contained, independently testable, documented; Clear purpose required - no organizational-only libraries -->

### API Contracts: gRPC-First, Versioned, Idempotent
<!-- Example: II. CLI Interface -->
Non‑negotiable rules:
- Transport is gRPC (protobuf). Package: `shopping.v1`. Services: `AuthService`, `GroupService`, `ListService`, `SyncService`.
- Error handling uses canonical gRPC status codes; app details carried via `google.rpc.Status` details.
- Each request MUST include `x-request-id` (client or server generated).
- All writes MUST accept `idempotency-key` metadata to support safe mobile retries.
- List RPCs MUST implement pagination: `page_size`, `page_token`, and `next_page_token`.
Rationale: Clear, mobile‑friendly contracts with robust error semantics and safe retries.
<!-- Example: Every library exposes functionality via CLI; Text in/out protocol: stdin/args → stdout, errors → stderr; Support JSON + human-readable formats -->

### Offline-First Sync & Conflict Resolution
<!-- Example: III. Test-First (NON-NEGOTIABLE) -->
Non‑negotiable rules:
- Clients operate offline; maintain a local operation queue.
- Sync via `SyncService` supports pushing/pulling deltas.
- Conflict policy (MVP): Last‑Write‑Wins using `updated_at` or `version`.
- On version mismatch, server returns `FailedPrecondition`; client MUST refresh then retry.
- Client MUST set reasonable deadlines; retries allowed only for idempotent operations.
Rationale: Predictable conflict behavior with simple, reliable offline operation for mobile.
<!-- Example: TDD mandatory: Tests written → User approved → Tests fail → Then implement; Red-Green-Refactor cycle strictly enforced -->

### Data Model & Migrations Discipline
<!-- Example: IV. Integration Testing -->
Non‑negotiable rules:
- Core entities: `User`, `Group`, `GroupMember`, `Category`, `List`, `Item`.
- Every entity MUST include `created_at`, `updated_at`, `updated_by`, and a `version` (int64) or `etag` (string).
- Database migrations are mandatory and version‑controlled.
- Sorting semantics for Items:
  - default ordering: `is_purchased=false` first, then `priority desc`, then `sort_order asc`.
- Item state MUST support `purchased` and `unpurchased`.
Rationale: Ensures reliable sync, auditing, and deterministic UX behavior.
<!-- Example: Focus areas requiring integration tests: New library contract tests, Contract changes, Inter-service communication, Shared schemas -->

### Observability, Security & Operations
<!-- Example: V. Observability, VI. Versioning & Breaking Changes, VII. Simplicity -->
Non‑negotiable rules:
- Structured logging with masked PII/tokens; include trace/request IDs.
- Metrics: request count, error rate, latency p95 (min).
- Rate limits: stricter on Auth RPCs.
- gRPC server reflection enabled only in dev; MUST be disabled in prod.
- Centralized authorization interceptor is mandatory.
Rationale: Operational clarity and safety in production, with secure defaults.
<!-- Example: Text I/O ensures debuggability; Structured logging required; Or: MAJOR.MINOR.BUILD format; Or: Start simple, YAGNI principles -->

## Scope & MVP Boundaries
<!-- Example: Additional Constraints, Security Requirements, Performance Standards, etc. -->

In Scope (MVP):
- Auth + session management
- Group create/invite/join and roles (owner/admin/member)
- List & Item CRUD
- Item states: purchased/unpurchased
- Sorting: priority + manual order
- Offline‑first client + server sync (MVP: LWW)

Out of Scope (MVP):
- Payments/subscriptions (infrastructure readiness only)
- Barcode/market integrations
- Real‑time live‑collab (CRDT/WebSocket)
- Advanced audit UIs
<!-- Example: Technology stack requirements, compliance standards, deployment policies, etc. -->

## Transport & Mobile Readiness
<!-- Example: Development Workflow, Review Process, Quality Gates, etc. -->

- Transport: gRPC over HTTP/2 is the primary channel.
- Clients MUST set deadlines by default for all RPCs.
- Retries are allowed only for idempotent operations and MUST be bounded.
- Server reflection: enabled in development, disabled in production.
<!-- Example: Code review requirements, testing gates, deployment approval process, etc. -->

## Governance
<!-- Example: Constitution supersedes all other practices; Amendments require documentation, approval, migration plan -->

Rules:
- Spec‑first: No feature without an updated proto and acceptance criteria.
- Contract changes:
  - Update `shopping.v1` protos and generate stubs.
  - Provide migration notes, error mapping, and pagination/idempotency implications.
- Compliance gate in PRs:
  - Multi‑tenant checks on all calls that accept `group_id`.
  - Observability (logs + metrics) present for new endpoints.
  - DB migrations accompany schema changes.
  - Tests: domain unit + gRPC integration (in‑memory or dockerized Postgres).
- Constitution versioning (semver):
  - MAJOR: incompatible governance/principle removals or redefinitions.
  - MINOR: new principle/section added or materially expanded guidance.
  - PATCH: clarifications and non‑semantic refinements.
- Amendment process:
  - Propose change with rationale and impact.
  - Update this constitution and sync dependent templates.
  - Record Last Amended date; bump version per rules.
<!-- Example: All PRs/reviews must verify compliance; Complexity must be justified; Use [GUIDANCE_FILE] for runtime development guidance -->

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): Provide original adoption date | **Last Amended**: 2026-01-07
<!-- Example: Version: 2.1.1 | Ratified: 2025-06-13 | Last Amended: 2025-07-16 -->
# [PROJECT_NAME] Constitution
<!-- Example: Spec Constitution, TaskFlow Constitution, etc. -->

## Core Principles

### [PRINCIPLE_1_NAME]
<!-- Example: I. Library-First -->
[PRINCIPLE_1_DESCRIPTION]
<!-- Example: Every feature starts as a standalone library; Libraries must be self-contained, independently testable, documented; Clear purpose required - no organizational-only libraries -->

### [PRINCIPLE_2_NAME]
<!-- Example: II. CLI Interface -->
[PRINCIPLE_2_DESCRIPTION]
<!-- Example: Every library exposes functionality via CLI; Text in/out protocol: stdin/args → stdout, errors → stderr; Support JSON + human-readable formats -->

### [PRINCIPLE_3_NAME]
<!-- Example: III. Test-First (NON-NEGOTIABLE) -->
[PRINCIPLE_3_DESCRIPTION]
<!-- Example: TDD mandatory: Tests written → User approved → Tests fail → Then implement; Red-Green-Refactor cycle strictly enforced -->

### [PRINCIPLE_4_NAME]
<!-- Example: IV. Integration Testing -->
[PRINCIPLE_4_DESCRIPTION]
<!-- Example: Focus areas requiring integration tests: New library contract tests, Contract changes, Inter-service communication, Shared schemas -->

### [PRINCIPLE_5_NAME]
<!-- Example: V. Observability, VI. Versioning & Breaking Changes, VII. Simplicity -->
[PRINCIPLE_5_DESCRIPTION]
<!-- Example: Text I/O ensures debuggability; Structured logging required; Or: MAJOR.MINOR.BUILD format; Or: Start simple, YAGNI principles -->

## [SECTION_2_NAME]
<!-- Example: Additional Constraints, Security Requirements, Performance Standards, etc. -->

[SECTION_2_CONTENT]
<!-- Example: Technology stack requirements, compliance standards, deployment policies, etc. -->

## [SECTION_3_NAME]
<!-- Example: Development Workflow, Review Process, Quality Gates, etc. -->

[SECTION_3_CONTENT]
<!-- Example: Code review requirements, testing gates, deployment approval process, etc. -->

## Governance
<!-- Example: Constitution supersedes all other practices; Amendments require documentation, approval, migration plan -->

[GOVERNANCE_RULES]
<!-- Example: All PRs/reviews must verify compliance; Complexity must be justified; Use [GUIDANCE_FILE] for runtime development guidance -->

**Version**: [CONSTITUTION_VERSION] | **Ratified**: [RATIFICATION_DATE] | **Last Amended**: [LAST_AMENDED_DATE]
<!-- Example: Version: 2.1.1 | Ratified: 2025-06-13 | Last Amended: 2025-07-16 -->
