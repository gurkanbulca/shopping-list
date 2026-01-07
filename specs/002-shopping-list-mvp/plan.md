# Implementation Plan: Mobile Offline MVP (Android + iOS)

**Branch**: `002-shopping-list-mvp` | **Date**: 2026-01-07 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/002-shopping-list-mvp/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Deliver minimal Android + iOS client apps with a shared mobile core that provides:

- Offline-first browsing and writing (local DB is the UI source of truth)
- Mutation queue for local writes (add item, toggle purchased, reorder)
- Sync loop: push queued mutations (batch) + pull delta changes (cursor)
- Conflict recovery on version mismatch: refresh latest data, reconcile (MVP: server wins), retry with bounds

Platform UI is intentionally minimal (Compose + SwiftUI) and calls into shared use-cases only.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Kotlin (KMP) 1.9+; Android Kotlin 1.9+; iOS Swift 5.9+  
**Primary Dependencies**: Kotlin Coroutines/Flow, SQLDelight, platform gRPC clients, minimal UI frameworks (Compose, SwiftUI)  
**Storage**: On-device SQLite (via SQLDelight) as the single source of truth; secure token storage per platform  
**Testing**: Shared Kotlin unit tests (kotlin.test), Android unit/UI tests, iOS unit/UI tests (XCTest)  
**Target Platform**: Android (min SDK target to be set during implementation); iOS 15+  
**Project Type**: Mobile (shared KMP core + two thin native apps)  
**Performance Goals**: UI feels instant for local actions; scroll/reorder remains smooth in typical lists (<500 items)  
**Constraints**: Offline-capable; bounded retries; all RPCs use deadlines; strict tenant boundary (Group)  
**Scale/Scope**: MVP screen set (Login, Groups, Lists, Items); core entities (Group/List/Item/Category) + sync queue/state

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ **SaaS Multi-Tenancy & Authorization**: All reads/writes are scoped to a selected `group_id`; local persistence partitions data by `group_id`; PermissionDenied triggers group-scope cleanup and user-visible error.
- ✅ **API Contracts: gRPC-First, Versioned, Idempotent**: All remote calls use `shopping.v1` gRPC services; client sets `authorization`, `x-request-id`, and `idempotency-key` (writes); retries bounded and only for safe operations.
- ✅ **Offline-First Sync & Conflict Resolution**: Local mutation queue; push+pull sync; `FailedPrecondition` triggers refresh+reconcile+retry with a defined policy.
- ✅ **Data Model & Migrations Discipline**: Local schema includes `version` and timestamps; SQLDelight migrations are version-controlled; UI sorting matches constitution.
- ✅ **Observability, Security & Operations**: Structured logs without tokens/PII; request IDs propagated; auth refresh via dedicated RPC; deadlines on all RPCs.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
api/
├── cmd/server/
├── internal/
├── proto/shopping/v1/       # Canonical protobufs (shopping.v1)
└── migrations/

shared/
├── domain/                  # Shared domain models + use-cases (KMP)
├── db/                      # SQLDelight schema + generated DB access (KMP)
├── data/                    # Repository implementations (KMP)
├── sync/                    # Sync engine + mutation queue (KMP)
└── platform/                # expect/actual helpers (time/uuid/logging/secure hooks)

androidApp/
└── (Jetpack Compose app + Android gRPC adapter + token storage)

iosApp/
└── (SwiftUI app + iOS gRPC adapter + token storage)

docs/
└── (client architecture + sync state machine + runbook)
```

**Structure Decision**: Mobile + API monorepo. The canonical API contract remains in `api/proto/shopping/v1`. The shared mobile core lives under `shared/` and is consumed by `androidApp/` and `iosApp/`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
