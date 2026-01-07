package shopping.db

import app.cash.sqldelight.coroutines.asFlow
import app.cash.sqldelight.coroutines.mapToList
import app.cash.sqldelight.coroutines.mapToOneOrNull
import app.cash.sqldelight.db.SqlDriver
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext

/**
 * Database access layer providing typed access to SQLDelight-generated queries.
 * All operations are suspend functions or Flows for async-safe access.
 */
class DbAccess(
    private val database: AppDatabase,
    private val dispatcher: CoroutineDispatcher
) {
    // ============================================
    // Groups
    // ============================================

    fun observeAllGroups(): Flow<List<Groups>> =
        database.appDatabaseQueries.selectAllGroups()
            .asFlow()
            .mapToList(dispatcher)

    suspend fun getGroup(id: String): Groups? = withContext(dispatcher) {
        database.appDatabaseQueries.selectGroupById(id).executeAsOneOrNull()
    }

    suspend fun insertGroup(group: Groups) = withContext(dispatcher) {
        database.appDatabaseQueries.insertGroup(
            id = group.id,
            name = group.name,
            created_at = group.created_at,
            updated_at = group.updated_at,
            version = group.version
        )
    }

    suspend fun deleteGroup(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.deleteGroup(id)
    }

    suspend fun deleteAllGroups() = withContext(dispatcher) {
        database.appDatabaseQueries.deleteAllGroups()
    }

    // ============================================
    // Lists
    // ============================================

    fun observeListsByGroupId(groupId: String): Flow<List<Lists>> =
        database.appDatabaseQueries.selectListsByGroupId(groupId)
            .asFlow()
            .mapToList(dispatcher)

    suspend fun getList(id: String): Lists? = withContext(dispatcher) {
        database.appDatabaseQueries.selectListById(id).executeAsOneOrNull()
    }

    suspend fun insertList(list: Lists) = withContext(dispatcher) {
        database.appDatabaseQueries.insertList(
            id = list.id,
            group_id = list.group_id,
            name = list.name,
            category_id = list.category_id,
            created_at = list.created_at,
            updated_at = list.updated_at,
            version = list.version
        )
    }

    suspend fun deleteList(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.deleteList(id)
    }

    suspend fun deleteListsByGroupId(groupId: String) = withContext(dispatcher) {
        database.appDatabaseQueries.deleteListsByGroupId(groupId)
    }

    // ============================================
    // Items
    // ============================================

    fun observeItemsByListId(listId: String): Flow<List<Items>> =
        database.appDatabaseQueries.selectItemsByListId(listId)
            .asFlow()
            .mapToList(dispatcher)

    suspend fun getItem(id: String): Items? = withContext(dispatcher) {
        database.appDatabaseQueries.selectItemById(id).executeAsOneOrNull()
    }

    suspend fun insertItem(item: Items) = withContext(dispatcher) {
        database.appDatabaseQueries.insertItem(
            id = item.id,
            list_id = item.list_id,
            name = item.name,
            is_purchased = item.is_purchased,
            priority = item.priority,
            sort_order = item.sort_order,
            category_id = item.category_id,
            created_at = item.created_at,
            updated_at = item.updated_at,
            version = item.version,
            pending_mutation_id = item.pending_mutation_id
        )
    }

    suspend fun updateItemPurchased(
        id: String,
        isPurchased: Boolean,
        updatedAt: Long,
        pendingMutationId: String?
    ) = withContext(dispatcher) {
        database.appDatabaseQueries.updateItemPurchased(
            is_purchased = if (isPurchased) 1L else 0L,
            updated_at = updatedAt,
            pending_mutation_id = pendingMutationId,
            id = id
        )
    }

    suspend fun updateItemSortOrder(
        id: String,
        sortOrder: Int,
        updatedAt: Long,
        pendingMutationId: String?
    ) = withContext(dispatcher) {
        database.appDatabaseQueries.updateItemSortOrder(
            sort_order = sortOrder.toLong(),
            updated_at = updatedAt,
            pending_mutation_id = pendingMutationId,
            id = id
        )
    }

    suspend fun clearItemPendingMutation(mutationId: String) = withContext(dispatcher) {
        database.appDatabaseQueries.clearItemPendingMutation(mutationId)
    }

    suspend fun deleteItem(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.deleteItem(id)
    }

    suspend fun getMaxSortOrder(listId: String): Long? = withContext(dispatcher) {
        database.appDatabaseQueries.getMaxSortOrder(listId).executeAsOneOrNull()?.MAX
    }

    // ============================================
    // Categories
    // ============================================

    fun observeCategoriesByGroupId(groupId: String): Flow<List<Categories>> =
        database.appDatabaseQueries.selectCategoriesByGroupId(groupId)
            .asFlow()
            .mapToList(dispatcher)

    suspend fun getCategory(id: String): Categories? = withContext(dispatcher) {
        database.appDatabaseQueries.selectCategoryById(id).executeAsOneOrNull()
    }

    suspend fun insertCategory(category: Categories) = withContext(dispatcher) {
        database.appDatabaseQueries.insertCategory(
            id = category.id,
            group_id = category.group_id,
            name = category.name,
            color = category.color,
            created_at = category.created_at,
            updated_at = category.updated_at,
            version = category.version
        )
    }

    suspend fun deleteCategory(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.deleteCategory(id)
    }

    // ============================================
    // Mutation Queue
    // ============================================

    suspend fun selectPendingMutations(limit: Int): List<Mutation_queue> = withContext(dispatcher) {
        database.appDatabaseQueries.selectPendingMutations(limit.toLong()).executeAsList()
    }

    suspend fun selectMutationsByBatchId(batchId: String): List<Mutation_queue> = withContext(dispatcher) {
        database.appDatabaseQueries.selectMutationsByBatchId(batchId).executeAsList()
    }

    suspend fun getMutation(id: String): Mutation_queue? = withContext(dispatcher) {
        database.appDatabaseQueries.selectMutationById(id).executeAsOneOrNull()
    }

    suspend fun insertMutation(mutation: Mutation_queue) = withContext(dispatcher) {
        database.appDatabaseQueries.insertMutation(
            id = mutation.id,
            batch_id = mutation.batch_id,
            entity_type = mutation.entity_type,
            entity_id = mutation.entity_id,
            mutation_type = mutation.mutation_type,
            payload = mutation.payload,
            expected_version = mutation.expected_version,
            status = mutation.status,
            retry_count = mutation.retry_count,
            created_at = mutation.created_at,
            claimed_at = mutation.claimed_at
        )
    }

    suspend fun claimMutations(ids: List<String>, claimedAt: Long) = withContext(dispatcher) {
        database.appDatabaseQueries.claimMutations(claimedAt, ids)
    }

    suspend fun markMutationSuccess(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.markMutationSuccess(id)
    }

    suspend fun markMutationFailed(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.markMutationFailed(id)
    }

    suspend fun resetFailedMutation(id: String) = withContext(dispatcher) {
        database.appDatabaseQueries.resetFailedMutation(id)
    }

    suspend fun countPendingMutations(): Long = withContext(dispatcher) {
        database.appDatabaseQueries.countPendingMutations().executeAsOne()
    }

    // ============================================
    // Sync State
    // ============================================

    fun observeSyncState(groupId: String): Flow<Sync_state?> =
        database.appDatabaseQueries.selectSyncStateByGroupId(groupId)
            .asFlow()
            .mapToOneOrNull(dispatcher)

    suspend fun getSyncState(groupId: String): Sync_state? = withContext(dispatcher) {
        database.appDatabaseQueries.selectSyncStateByGroupId(groupId).executeAsOneOrNull()
    }

    suspend fun insertOrUpdateSyncState(syncState: Sync_state) = withContext(dispatcher) {
        database.appDatabaseQueries.insertOrUpdateSyncState(
            group_id = syncState.group_id,
            cursor = syncState.cursor,
            last_sync_at = syncState.last_sync_at,
            last_error = syncState.last_error
        )
    }

    suspend fun updateSyncCursor(groupId: String, cursor: Long, lastSyncAt: Long) = withContext(dispatcher) {
        database.appDatabaseQueries.updateSyncCursor(cursor, lastSyncAt, groupId)
    }

    suspend fun updateSyncError(groupId: String, error: String) = withContext(dispatcher) {
        database.appDatabaseQueries.updateSyncError(error, groupId)
    }

    suspend fun deleteSyncState(groupId: String) = withContext(dispatcher) {
        database.appDatabaseQueries.deleteSyncStateByGroupId(groupId)
    }

    // ============================================
    // Transactions
    // ============================================

    suspend fun <T> transaction(block: suspend () -> T): T = withContext(dispatcher) {
        database.transactionWithResult {
            // Note: SQLDelight transactions are synchronous within the transaction block
            // We use runBlocking here as a workaround, but the outer withContext handles threading
            kotlinx.coroutines.runBlocking { block() }
        }
    }
}

/**
 * Factory for creating database driver.
 */
expect class DriverFactory {
    fun createDriver(): SqlDriver
}

/**
 * Create and initialize the database.
 */
fun createDatabase(driverFactory: DriverFactory): AppDatabase {
    val driver = driverFactory.createDriver()
    return AppDatabase(driver)
}
