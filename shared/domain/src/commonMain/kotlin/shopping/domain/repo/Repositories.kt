package shopping.domain.repo

import kotlinx.coroutines.flow.Flow
import shopping.domain.model.*

/**
 * Repository interfaces for the Shopping List app.
 * These define the contracts for data access, implemented in the data layer.
 */

// ============================================
// Auth Repository
// ============================================

interface AuthRepository {
    /**
     * Authenticate user with credentials.
     * @return AuthTokens on success
     */
    suspend fun login(credentials: LoginCredentials): Result<AuthTokens>

    /**
     * Refresh the access token using the refresh token.
     * @return New AuthTokens on success
     */
    suspend fun refreshToken(): Result<AuthTokens>

    /**
     * Log out the current user, invalidating tokens.
     */
    suspend fun logout(): Result<Unit>

    /**
     * Get the current authenticated user, if any.
     */
    suspend fun getCurrentUser(): User?

    /**
     * Observe authentication state changes.
     */
    fun observeAuthState(): Flow<AuthState>
}

sealed class AuthState {
    data object Loading : AuthState()
    data object Unauthenticated : AuthState()
    data class Authenticated(val user: User) : AuthState()
}

// ============================================
// Group Repository
// ============================================

interface GroupRepository {
    /**
     * Observe all groups the current user belongs to.
     * Emits from local DB, triggers remote refresh.
     */
    fun observeGroups(): Flow<List<Group>>

    /**
     * Get a specific group by ID.
     */
    suspend fun getGroup(groupId: GroupId): Group?

    /**
     * Refresh groups from remote server.
     */
    suspend fun refreshGroups(): Result<Unit>
}

// ============================================
// List Repository
// ============================================

interface ListRepository {
    /**
     * Observe all lists in a group.
     * Emits from local DB, optionally triggers remote refresh.
     */
    fun observeLists(groupId: GroupId): Flow<List<ShoppingList>>

    /**
     * Get a specific list by ID.
     */
    suspend fun getList(listId: ListId): ShoppingList?

    /**
     * Refresh lists for a group from remote server.
     */
    suspend fun refreshLists(groupId: GroupId): Result<Unit>
}

// ============================================
// Item Repository
// ============================================

interface ItemRepository {
    /**
     * Observe all items in a list, sorted by:
     * 1. Unpurchased first
     * 2. Higher priority first
     * 3. Sort order (stable)
     */
    fun observeItems(listId: ListId): Flow<List<Item>>

    /**
     * Get a specific item by ID.
     */
    suspend fun getItem(itemId: ItemId): Item?

    /**
     * Add a new item to a list.
     * Optimistically updates local DB and queues mutation for sync.
     * @return The created item with pending state
     */
    suspend fun addItem(listId: ListId, name: String, priority: Int = 0): Result<Item>

    /**
     * Toggle the purchased state of an item.
     * Optimistically updates local DB and queues mutation for sync.
     */
    suspend fun togglePurchased(itemId: ItemId): Result<Item>

    /**
     * Reorder items in a list.
     * Optimistically updates local DB and queues mutations for sync.
     * @param itemIds Ordered list of item IDs representing new order
     */
    suspend fun reorderItems(listId: ListId, itemIds: List<ItemId>): Result<Unit>
}

// ============================================
// Category Repository
// ============================================

interface CategoryRepository {
    /**
     * Observe all categories in a group.
     */
    fun observeCategories(groupId: GroupId): Flow<List<Category>>

    /**
     * Get a specific category by ID.
     */
    suspend fun getCategory(categoryId: CategoryId): Category?
}

// ============================================
// Sync Repository
// ============================================

interface SyncRepository {
    /**
     * Observe the current sync state.
     */
    fun observeSyncState(): Flow<SyncState>

    /**
     * Observe sync cursor for a group.
     */
    fun observeSyncCursor(groupId: GroupId): Flow<SyncCursor?>

    /**
     * Get pending mutation count.
     */
    fun observePendingMutationCount(): Flow<Int>

    /**
     * Trigger a sync cycle for a group.
     * Push pending mutations, then pull delta changes.
     */
    suspend fun sync(groupId: GroupId): Result<Unit>

    /**
     * Perform initial full sync for a group.
     * Pulls all data until has_more = false.
     */
    suspend fun fullSync(groupId: GroupId): Result<Unit>
}
