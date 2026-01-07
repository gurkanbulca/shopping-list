package shopping.sync

import shopping.data.remote.SyncRemoteDataSource
import shopping.domain.model.GroupId
import shopping.domain.model.SyncState
import shopping.platform.Log
import shopping.platform.currentTimeMillis

/**
 * Performs a full initial sync for a group.
 * Pulls all data from cursor 0 until has_more = false.
 */
class FullSync(
    private val syncRemoteDataSource: SyncRemoteDataSource,
    private val syncStateStore: SyncStateStore,
    private val deltaApplier: DeltaApplier
) {
    companion object {
        private const val TAG = "FullSync"
        private const val DEFAULT_PAGE_SIZE = 100
        private const val MAX_RETRIES = 3
    }

    /**
     * Perform a full initial sync for a group.
     * Resets the cursor to 0 and pulls all data.
     * @param groupId The group to sync
     * @return Result with the final cursor on success
     */
    suspend fun execute(groupId: GroupId): Result<Long> {
        Log.i(TAG, "Starting full sync for group ${groupId.value}")

        // Reset cursor to start from beginning
        syncStateStore.updateCursor(groupId, 0L)

        var cursor = 0L
        var hasMore = true
        var retryCount = 0
        var totalEntities = 0

        while (hasMore) {
            Log.d(TAG, "Pulling delta from cursor $cursor")

            val result = syncRemoteDataSource.getDelta(
                groupId = groupId.value,
                cursor = cursor,
                pageSize = DEFAULT_PAGE_SIZE
            )

            if (result.isFailure) {
                retryCount++
                Log.w(TAG, "Pull failed, retry $retryCount/$MAX_RETRIES", result.exceptionOrNull())

                if (retryCount >= MAX_RETRIES) {
                    val error = result.exceptionOrNull()!!
                    syncStateStore.recordSyncError(groupId, error.message ?: "Unknown error")
                    return Result.failure(error)
                }
                continue
            }

            val response = result.getOrThrow()

            // Apply deltas using the mapper
            val deltas = response.entities.map { dto ->
                shopping.domain.model.DeltaEntity(
                    entityType = shopping.domain.model.EntityType.fromWireValue(dto.entityType),
                    entityId = dto.entityId,
                    data = dto.data,
                    version = dto.version,
                    isDeleted = dto.isDeleted
                )
            }
            deltaApplier.applyDeltas(deltas)
            totalEntities += deltas.size

            // Update cursor
            cursor = response.nextCursor
            syncStateStore.updateCursor(groupId, cursor)

            hasMore = response.hasMore
            retryCount = 0 // Reset retry count on success

            Log.d(TAG, "Applied ${deltas.size} deltas, hasMore=$hasMore, nextCursor=$cursor")
        }

        // Mark sync success
        syncStateStore.markSyncSuccess(groupId, cursor)
        Log.i(TAG, "Full sync completed for group ${groupId.value}: $totalEntities entities synced")

        return Result.success(cursor)
    }
}
