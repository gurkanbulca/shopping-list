# Sync State Machine

**Status**: Placeholder  
**Feature**: Mobile Offline MVP (Android + iOS)

## Overview

This document describes the sync engine state machine that manages offline-first data synchronization between the mobile clients and the backend server.

## State Machine Diagram

```text
                    ┌──────────────────┐
                    │      IDLE        │
                    └────────┬─────────┘
                             │ trigger sync
                             ▼
                    ┌──────────────────┐
                    │   PUSH_PENDING   │◄───────────┐
                    └────────┬─────────┘            │
                             │ mutations sent       │ retry on
                             ▼                      │ transient error
                    ┌──────────────────┐            │
                    │   AWAIT_PUSH     │────────────┘
                    └────────┬─────────┘
                             │ push complete
                             ▼
                    ┌──────────────────┐
                    │   PULL_DELTA     │◄───────────┐
                    └────────┬─────────┘            │
                             │ has_more=true        │
                             ├──────────────────────┘
                             │ has_more=false
                             ▼
                    ┌──────────────────┐
                    │   SYNC_COMPLETE  │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │      IDLE        │
                    └──────────────────┘
```

## States

| State | Description |
|-------|-------------|
| `IDLE` | No sync in progress. Waiting for trigger (manual, timer, connectivity change) |
| `PUSH_PENDING` | Claiming queued mutations from local mutation queue |
| `AWAIT_PUSH` | Waiting for `PushMutations` RPC response |
| `PULL_DELTA` | Fetching delta changes via `GetDelta` RPC |
| `SYNC_COMPLETE` | Sync cycle finished, transitioning back to IDLE |

## Transitions

### Trigger Sync
- **From**: IDLE
- **To**: PUSH_PENDING
- **Triggers**: Manual "sync now", connectivity restored, periodic timer

### Push Complete
- **From**: AWAIT_PUSH
- **To**: PULL_DELTA
- **Condition**: All mutations processed (success or permanent failure)

### Has More Deltas
- **From**: PULL_DELTA
- **To**: PULL_DELTA (loop)
- **Condition**: `GetDeltaResponse.has_more = true`

### Sync Done
- **From**: PULL_DELTA
- **To**: SYNC_COMPLETE
- **Condition**: `GetDeltaResponse.has_more = false`

## Error Handling

### Transient Errors
- Network timeouts, temporary unavailability
- Retry with exponential backoff (max 3 attempts)
- After max retries, transition to IDLE with error state

### Conflict Errors (FailedPrecondition)
- Pull latest delta immediately
- Reconcile using "server wins" policy
- Re-attempt mutation with updated version

### Permanent Errors
- Invalid data, schema mismatch
- Mark mutation as failed, do not retry
- Continue with remaining mutations

## Implementation Details

### SyncEngine Class

The `SyncEngine` class in `shared/sync/src/commonMain/kotlin/shopping/sync/SyncEngine.kt` implements this state machine.

Key methods:
- `sync(groupId)`: Run a full sync cycle (push + pull)
- `fullSync(groupId)`: Reset cursor and sync from beginning
- `observeSyncState()`: Observe current state as a Flow

### Configuration

| Parameter | Default | Description |
|-----------|---------|-------------|
| `DEFAULT_PAGE_SIZE` | 100 | Max entities per GetDelta response |
| `MAX_PUSH_RETRIES` | 3 | Max retries for push operations |
| `MAX_PULL_RETRIES` | 3 | Max retries for pull operations |

### Error Recovery

1. **Transient errors**: Retry with exponential backoff
2. **Auth errors**: Trigger token refresh, retry once
3. **Conflict errors**: Mark mutation as failed, continue with others
4. **Permanent errors**: Log error, skip mutation

## Notes

- The sync engine is single-threaded per group to avoid race conditions
- Cursor is persisted after each successful page, enabling resume on failure
- Push operations are idempotent via mutation ID as idempotency key
