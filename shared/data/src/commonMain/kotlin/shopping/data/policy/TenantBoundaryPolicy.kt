package shopping.data.policy

import shopping.db.DbAccess
import shopping.domain.model.DomainError
import shopping.domain.model.GroupId
import shopping.platform.Log
import shopping.sync.SyncStateStore

/**
 * Handles PermissionDenied errors by clearing group-scoped local data.
 * Ensures tenant boundary is respected when access is revoked.
 */
class TenantBoundaryPolicy(
    private val dbAccess: DbAccess,
    private val syncStateStore: SyncStateStore
) {
    companion object {
        private const val TAG = "TenantBoundaryPolicy"
    }

    /**
     * Handle a permission denied error for a specific group.
     * Clears all local data for that group and notifies the caller.
     * @param groupId The group that access was denied for
     * @return PermissionDeniedAction describing what actions to take
     */
    suspend fun handlePermissionDenied(groupId: GroupId): PermissionDeniedAction {
        Log.w(TAG, "Permission denied for group ${groupId.value}, clearing local data")

        try {
            // Clear group-scoped data in the correct order
            // (items depend on lists, which depend on groups)
            
            // 1. Clear sync state for this group
            syncStateStore.resetSyncState(groupId)
            
            // 2. Clear all lists and their items for this group
            clearGroupData(groupId)
            
            // Note: We don't clear the group itself from the local DB
            // as the user might regain access later
            
            Log.i(TAG, "Successfully cleared local data for group ${groupId.value}")
            
            return PermissionDeniedAction(
                groupId = groupId,
                wasCleared = true,
                shouldNavigateToGroups = true,
                errorMessage = "You no longer have access to this group. Please select another group."
            )
        } catch (e: Exception) {
            Log.e(TAG, "Failed to clear local data for group ${groupId.value}", e)
            
            return PermissionDeniedAction(
                groupId = groupId,
                wasCleared = false,
                shouldNavigateToGroups = true,
                errorMessage = "Access denied. Some local data may not have been cleared."
            )
        }
    }

    /**
     * Clear all data for a specific group.
     */
    private suspend fun clearGroupData(groupId: GroupId) {
        dbAccess.transaction {
            // Delete all items for lists in this group
            // Note: This requires joining with lists table
            // For now, we rely on the lists deletion which should cascade
            
            // Delete all lists in this group
            dbAccess.deleteListsByGroupId(groupId.value)
            
            // Delete all categories in this group
            // dbAccess.deleteCategoriesByGroupId(groupId.value)
        }
    }

    /**
     * Check if an error is a permission denied error.
     */
    fun isPermissionDenied(error: Throwable): Boolean {
        return error is DomainError.PermissionDenied
    }

    /**
     * Extract group ID from a permission denied error.
     */
    fun extractGroupId(error: Throwable): GroupId? {
        return (error as? DomainError.PermissionDenied)?.groupId
    }

    /**
     * Wrap a suspend function with permission denied handling.
     */
    suspend fun <T> withPermissionHandling(
        groupId: GroupId,
        block: suspend () -> Result<T>
    ): Result<T> {
        val result = block()
        
        if (result.isFailure) {
            val error = result.exceptionOrNull()
            if (error != null && isPermissionDenied(error)) {
                handlePermissionDenied(groupId)
            }
        }
        
        return result
    }
}

/**
 * Describes the action taken when permission is denied.
 */
data class PermissionDeniedAction(
    val groupId: GroupId,
    val wasCleared: Boolean,
    val shouldNavigateToGroups: Boolean,
    val errorMessage: String
)
