# Client Architecture

**Status**: Placeholder  
**Feature**: Mobile Offline MVP (Android + iOS)

## Overview

This document describes the architecture of the mobile client applications, including the adapter boundary pattern, offline-first data flow, and integration with the shared Kotlin Multiplatform (KMP) core.

## Architecture Principles

### Offline-First

- The local SQLite database (via SQLDelight) is the single source of truth for UI rendering
- All user actions first update local state, then queue mutations for sync
- UI never waits for network responses for primary rendering

### Adapter Boundary Pattern

- Platform-specific gRPC adapters implement shared interfaces defined in `shared/data`
- The shared core never imports platform-specific code directly
- All platform dependencies flow through `expect/actual` declarations in `shared/platform`

### Layer Structure

```text
┌─────────────────────────────────────────────┐
│              Platform UI                     │
│         (Compose / SwiftUI)                  │
├─────────────────────────────────────────────┤
│            Shared Domain                     │
│    (UseCases, Models, State Flows)          │
├─────────────────────────────────────────────┤
│            Shared Data                       │
│   (Repositories, Mappers, Remote Adapters)  │
├─────────────────────────────────────────────┤
│            Shared DB / Sync                  │
│  (SQLDelight, MutationQueue, SyncEngine)    │
├─────────────────────────────────────────────┤
│           Platform Adapters                  │
│    (gRPC clients, SecureStorage, etc.)      │
└─────────────────────────────────────────────┘
```

## Build & Run Instructions

### Android

```bash
# From repository root
./gradlew :androidApp:assembleDebug

# Install on connected device
./gradlew :androidApp:installDebug
```

### iOS

```bash
# Open in Xcode
open iosApp/ShoppingListApp.xcodeproj

# Build and run from Xcode or via xcodebuild
```

## Module Responsibilities

| Module | Responsibility |
|--------|----------------|
| `shared:domain` | Domain models, use cases, observable state contracts |
| `shared:data` | Repository implementations, data mappers, remote adapter interfaces |
| `shared:db` | SQLDelight schema, database access layer |
| `shared:sync` | Sync engine, mutation queue, delta applier |
| `shared:platform` | Platform abstractions (expect/actual) for UUID, time, logging, secure storage |
| `androidApp` | Android Compose UI, gRPC adapters, secure token storage |
| `iosApp` | SwiftUI views, gRPC adapters, Keychain token storage |

## Data Flow

### Read Flow (Offline-First)

```text
UI → UseCase → Repository → DbAccess → SQLite
                   ↓
            (background) → RemoteDataSource → gRPC → Server
                   ↓
            DeltaApplier → DbAccess → SQLite
                   ↓
            Flow<List<Entity>> updates UI automatically
```

### Write Flow (Optimistic)

```text
UI → UseCase → Repository
        ↓
    1. Update local DB (optimistic)
    2. Enqueue mutation
        ↓
    UI updated immediately (with pending indicator)
        ↓
    (background) SyncEngine → PushMutations → Server
        ↓
    On success: Clear pending state
    On failure: Keep in queue for retry
```

## Error Handling

| Error Type | Response |
|------------|----------|
| Network unavailable | Queue mutations, retry on connectivity |
| 401 Unauthenticated | Refresh token, retry once |
| 403 PermissionDenied | Clear group data, show error |
| 409 Conflict | Refresh latest data, reconcile, allow retry |
| Transient errors | Exponential backoff, max 3 retries |

## Testing Strategy

### Unit Tests (shared)
- Domain models and mappers
- Use case logic
- Sync engine state machine

### Integration Tests
- Repository with in-memory database
- Sync flow with mock remote data sources

### Platform Tests
- Android: Compose UI tests, Espresso
- iOS: XCTest, SwiftUI previews

## Notes

- All timestamps are UTC milliseconds
- UUIDs are generated client-side for optimistic inserts
- Mutations are deduplicated by idempotency key on server
