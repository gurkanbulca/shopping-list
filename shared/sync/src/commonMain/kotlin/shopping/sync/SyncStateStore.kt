package shopping.sync

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import shopping.db.DbAccess
import shopping.db.Sync_state
import shopping.domain.model.GroupId
import shopping.domain.model.SyncCursor
import shopping.data.mapper.toDomain
import shopping.platform.Log
import shopping.platform.currentTimeMillis

/**
 * Manages sync state/cursor persistence per group.
 * Tracks: cursor position, last successful sync time, and last error.
 */
class SyncStateStore(
    private val dbAccess: DbAccess
) {
    companion object {
        private const val TAG = "SyncStateStore"
    }

    /**
     * Observe the sync cursor for a group.
     */
    fun observeSyncCursor(groupId: GroupId): Flow<SyncCursor?> =
        dbAccess.observeSyncState(groupId.value).map { it?.toDomain() }

    /**
     * Get the current cursor for a group.
     * Returns 0 if no sync state exists (initial sync).
     */
    suspend fun getCursor(groupId: GroupId): Long {
        val state = dbAccess.getSyncState(groupId.value)
        return state?.cursor ?: 0L
    }

    /**
     * Get the full sync state for a group.
     */
    suspend fun getSyncState(groupId: GroupId): SyncCursor? =
        dbAccess.getSyncState(groupId.value)?.toDomain()

    /**
     * Update the cursor after a successful delta pull.
     */
    suspend fun updateCursor(groupId: GroupId, newCursor: Long) {
        Log.d(TAG, "Updating cursor for group ${groupId.value}: $newCursor")
        
        val existingState = dbAccess.getSyncState(groupId.value)
        if (existingState != null) {
            dbAccess.updateSyncCursor(groupId.value, newCursor, currentTimeMillis())
        } else {
            // Create new sync state
            dbAccess.insertOrUpdateSyncState(Sync_state(
                group_id = groupId.value,
                cursor = newCursor,
                last_sync_at = currentTimeMillis(),
                last_error = null
            ))
        }
    }

    /**
     * Mark the last sync as successful.
     */
    suspend fun markSyncSuccess(groupId: GroupId, cursor: Long) {
        Log.d(TAG, "Marking sync success for group ${groupId.value}")
        dbAccess.insertOrUpdateSyncState(Sync_state(
            group_id = groupId.value,
            cursor = cursor,
            last_sync_at = currentTimeMillis(),
            last_error = null
        ))
    }

    /**
     * Record a sync error.
     */
    suspend fun recordSyncError(groupId: GroupId, error: String) {
        Log.e(TAG, "Recording sync error for group ${groupId.value}: $error")
        
        val existingState = dbAccess.getSyncState(groupId.value)
        if (existingState != null) {
            dbAccess.updateSyncError(groupId.value, error)
        } else {
            // Create sync state with error
            dbAccess.insertOrUpdateSyncState(Sync_state(
                group_id = groupId.value,
                cursor = 0L,
                last_sync_at = null,
                last_error = error
            ))
        }
    }

    /**
     * Reset sync state for a group (e.g., on permission denied).
     */
    suspend fun resetSyncState(groupId: GroupId) {
        Log.w(TAG, "Resetting sync state for group ${groupId.value}")
        dbAccess.deleteSyncState(groupId.value)
    }

    /**
     * Get the last successful sync time for a group.
     */
    suspend fun getLastSyncTime(groupId: GroupId): Long? {
        return dbAccess.getSyncState(groupId.value)?.last_sync_at
    }
}
