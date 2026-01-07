package shopping.sync

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import shopping.data.mapper.toDomain
import shopping.data.mapper.toDto
import shopping.data.remote.SyncRemoteDataSource
import shopping.domain.model.*
import shopping.platform.Log
import shopping.platform.currentTimeMillis

/**
 * Sync engine that manages the push-pull sync cycle.
 * 
 * State Machine:
 * IDLE -> PUSH_PENDING -> AWAIT_PUSH -> PULL_DELTA -> SYNC_COMPLETE -> IDLE
 * 
 * On error: -> ERROR -> IDLE (with backoff)
 */
class SyncEngine(
    private val syncRemoteDataSource: SyncRemoteDataSource,
    private val mutationQueueStore: MutationQueueStore,
    private val syncStateStore: SyncStateStore,
    private val deltaApplier: DeltaApplier
) {
    companion object {
        private const val TAG = "SyncEngine"
        private const val DEFAULT_PAGE_SIZE = 100
        private const val MAX_PUSH_RETRIES = 3
        private const val MAX_PULL_RETRIES = 3
    }

    private val _syncState = MutableStateFlow(SyncState.IDLE)
    
    /**
     * Observe the current sync state.
     */
    fun observeSyncState(): Flow<SyncState> = _syncState.asStateFlow()

    /**
     * Get the current sync state.
     */
    fun currentState(): SyncState = _syncState.value

    /**
     * Run a full sync cycle for a group:
     * 1. Push pending mutations
     * 2. Pull delta changes until has_more = false
     */
    suspend fun sync(groupId: GroupId): Result<Unit> {
        if (_syncState.value != SyncState.IDLE) {
            Log.w(TAG, "Sync already in progress, skipping")
            return Result.success(Unit)
        }

        Log.i(TAG, "Starting sync for group ${groupId.value}")

        try {
            // Phase 1: Push pending mutations
            _syncState.value = SyncState.PUSH_PENDING
            val pushResult = pushMutations(groupId)
            if (pushResult.isFailure) {
                handleSyncError(groupId, pushResult.exceptionOrNull()!!)
                return pushResult
            }

            // Phase 2: Pull delta changes
            _syncState.value = SyncState.PULL_DELTA
            val pullResult = pullDelta(groupId)
            if (pullResult.isFailure) {
                handleSyncError(groupId, pullResult.exceptionOrNull()!!)
                return pullResult
            }

            // Success
            _syncState.value = SyncState.SYNC_COMPLETE
            syncStateStore.markSyncSuccess(groupId, pullResult.getOrThrow())
            Log.i(TAG, "Sync completed successfully for group ${groupId.value}")

            _syncState.value = SyncState.IDLE
            return Result.success(Unit)

        } catch (e: Exception) {
            handleSyncError(groupId, e)
            return Result.failure(e)
        }
    }

    /**
     * Perform initial full sync for a group.
     * Pulls all data from cursor 0 until has_more = false.
     */
    suspend fun fullSync(groupId: GroupId): Result<Unit> {
        Log.i(TAG, "Starting full sync for group ${groupId.value}")

        // Reset cursor to 0 for full sync
        syncStateStore.updateCursor(groupId, 0L)
        
        return sync(groupId)
    }

    /**
     * Push pending mutations to the server.
     */
    private suspend fun pushMutations(groupId: GroupId): Result<Unit> {
        var retryCount = 0

        while (retryCount < MAX_PUSH_RETRIES) {
            val pending = mutationQueueStore.claimPending()
            
            if (pending.isEmpty()) {
                Log.d(TAG, "No pending mutations to push")
                return Result.success(Unit)
            }

            Log.d(TAG, "Pushing ${pending.size} mutations")
            _syncState.value = SyncState.AWAIT_PUSH

            val dtos = pending.map { it.toDto() }
            val result = syncRemoteDataSource.pushMutations(groupId.value, dtos)

            if (result.isFailure) {
                retryCount++
                Log.w(TAG, "Push failed, retry $retryCount/$MAX_PUSH_RETRIES", result.exceptionOrNull())
                
                // Reset claimed mutations back to pending for retry
                for (mutation in pending) {
                    mutationQueueStore.resetForRetry(mutation.id)
                }
                
                if (retryCount >= MAX_PUSH_RETRIES) {
                    return result.map { }
                }
                continue
            }

            // Process results
            val response = result.getOrThrow()
            for (mutationResult in response.results) {
                val mutationId = MutationId(mutationResult.idempotencyKey)
                
                if (mutationResult.success) {
                    mutationQueueStore.markSuccess(mutationId)
                } else {
                    Log.w(TAG, "Mutation failed: ${mutationResult.errorCode} - ${mutationResult.errorMessage}")
                    
                    // Handle specific error codes
                    when (mutationResult.errorCode) {
                        "FAILED_PRECONDITION", "VERSION_MISMATCH" -> {
                            // Conflict - mark failed, will be handled by conflict handler
                            mutationQueueStore.markFailed(mutationId)
                        }
                        else -> {
                            // Other errors - mark failed
                            mutationQueueStore.markFailed(mutationId)
                        }
                    }
                }
            }

            // Continue if there might be more pending mutations
            if (!mutationQueueStore.hasPending()) {
                break
            }
        }

        return Result.success(Unit)
    }

    /**
     * Pull delta changes from the server.
     * @return The final cursor after pulling all changes
     */
    private suspend fun pullDelta(groupId: GroupId): Result<Long> {
        var cursor = syncStateStore.getCursor(groupId)
        var retryCount = 0
        var hasMore = true

        while (hasMore && retryCount < MAX_PULL_RETRIES) {
            Log.d(TAG, "Pulling delta from cursor $cursor")

            val result = syncRemoteDataSource.getDelta(
                groupId = groupId.value,
                cursor = cursor,
                pageSize = DEFAULT_PAGE_SIZE
            )

            if (result.isFailure) {
                retryCount++
                Log.w(TAG, "Pull failed, retry $retryCount/$MAX_PULL_RETRIES", result.exceptionOrNull())
                
                if (retryCount >= MAX_PULL_RETRIES) {
                    return Result.failure(result.exceptionOrNull()!!)
                }
                continue
            }

            val response = result.getOrThrow()
            
            // Apply deltas
            val deltas = response.entities.map { it.toDomain() }
            deltaApplier.applyDeltas(deltas)

            // Update cursor
            cursor = response.nextCursor
            syncStateStore.updateCursor(groupId, cursor)

            hasMore = response.hasMore
            retryCount = 0 // Reset retry count on success

            Log.d(TAG, "Applied ${deltas.size} deltas, hasMore=$hasMore, nextCursor=$cursor")
        }

        return Result.success(cursor)
    }

    /**
     * Handle sync errors.
     */
    private suspend fun handleSyncError(groupId: GroupId, error: Throwable) {
        Log.e(TAG, "Sync error for group ${groupId.value}", error)
        _syncState.value = SyncState.ERROR
        
        syncStateStore.recordSyncError(groupId, error.message ?: "Unknown error")
        
        // Transition back to IDLE after error handling
        _syncState.value = SyncState.IDLE
    }
}
