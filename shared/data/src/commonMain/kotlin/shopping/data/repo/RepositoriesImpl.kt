package shopping.data.repo

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import shopping.data.mapper.*
import shopping.data.remote.*
import shopping.db.DbAccess
import shopping.db.Items
import shopping.domain.model.*
import shopping.domain.repo.*
import shopping.platform.Log
import shopping.platform.currentTimeMillis
import shopping.platform.generateUuid

/**
 * Repository implementations using local DB + remote sync.
 */

// ============================================
// Auth Repository Implementation
// ============================================

class AuthRepositoryImpl(
    private val authRemoteDataSource: AuthRemoteDataSource,
    private val secureTokenStore: SecureTokenStore
) : AuthRepository {

    companion object {
        private const val TAG = "AuthRepository"
    }

    private val _authState = MutableStateFlow<AuthState>(AuthState.Loading)
    private var cachedUser: User? = null
    private var cachedTokens: AuthTokens? = null

    override suspend fun login(credentials: LoginCredentials): Result<AuthTokens> {
        Log.d(TAG, "Attempting login")
        
        val result = authRemoteDataSource.login(credentials.email, credentials.password)
        
        return result.map { dto ->
            val tokens = dto.toDomain(currentTimeMillis())
            
            // Store tokens securely
            secureTokenStore.storeTokens(tokens)
            cachedTokens = tokens
            
            // Fetch user profile
            val userResult = authRemoteDataSource.getUserProfile(tokens.accessToken)
            userResult.onSuccess { userDto ->
                cachedUser = userDto.toDomain()
                _authState.value = AuthState.Authenticated(cachedUser!!)
            }
            
            tokens
        }
    }

    override suspend fun refreshToken(): Result<AuthTokens> {
        Log.d(TAG, "Refreshing token")
        
        val currentTokens = cachedTokens ?: secureTokenStore.getTokens()
            ?: return Result.failure(DomainError.AuthError("No refresh token available"))
        
        val result = authRemoteDataSource.refreshToken(currentTokens.refreshToken)
        
        return result.map { dto ->
            val tokens = dto.toDomain(currentTimeMillis())
            secureTokenStore.storeTokens(tokens)
            cachedTokens = tokens
            tokens
        }
    }

    override suspend fun logout(): Result<Unit> {
        Log.d(TAG, "Logging out")
        
        val tokens = cachedTokens ?: secureTokenStore.getTokens()
        
        // Best effort logout on server
        tokens?.let {
            authRemoteDataSource.logout(it.accessToken)
        }
        
        // Clear local state
        secureTokenStore.clearTokens()
        cachedTokens = null
        cachedUser = null
        _authState.value = AuthState.Unauthenticated
        
        return Result.success(Unit)
    }

    override suspend fun getCurrentUser(): User? {
        if (cachedUser != null) return cachedUser
        
        // Try to restore from stored tokens
        val tokens = secureTokenStore.getTokens() ?: return null
        
        if (tokens.isExpired(currentTimeMillis())) {
            // Try to refresh
            val refreshResult = refreshToken()
            if (refreshResult.isFailure) {
                _authState.value = AuthState.Unauthenticated
                return null
            }
        }
        
        val userResult = authRemoteDataSource.getUserProfile(tokens.accessToken)
        return userResult.getOrNull()?.toDomain()?.also {
            cachedUser = it
            _authState.value = AuthState.Authenticated(it)
        }
    }

    override fun observeAuthState(): Flow<AuthState> = _authState.asStateFlow()
}

/**
 * Interface for secure token storage.
 */
interface SecureTokenStore {
    suspend fun storeTokens(tokens: AuthTokens)
    suspend fun getTokens(): AuthTokens?
    suspend fun clearTokens()
}

// ============================================
// Group Repository Implementation
// ============================================

class GroupRepositoryImpl(
    private val dbAccess: DbAccess,
    private val groupRemoteDataSource: GroupRemoteDataSource
) : GroupRepository {

    companion object {
        private const val TAG = "GroupRepository"
    }

    override fun observeGroups(): Flow<List<Group>> =
        dbAccess.observeAllGroups().map { list -> list.map { it.toDomain() } }

    override suspend fun getGroup(groupId: GroupId): Group? =
        dbAccess.getGroup(groupId.value)?.toDomain()

    override suspend fun refreshGroups(): Result<Unit> {
        Log.d(TAG, "Refreshing groups from server")
        
        val result = groupRemoteDataSource.listGroups()
        
        return result.map { groups ->
            // Update local DB
            for (groupDto in groups) {
                dbAccess.insertGroup(groupDto.toDb())
            }
        }
    }
}

// ============================================
// List Repository Implementation
// ============================================

class ListRepositoryImpl(
    private val dbAccess: DbAccess,
    private val listRemoteDataSource: ListRemoteDataSource
) : ListRepository {

    companion object {
        private const val TAG = "ListRepository"
    }

    override fun observeLists(groupId: GroupId): Flow<List<ShoppingList>> =
        dbAccess.observeListsByGroupId(groupId.value).map { list -> list.map { it.toDomain() } }

    override suspend fun getList(listId: ListId): ShoppingList? =
        dbAccess.getList(listId.value)?.toDomain()

    override suspend fun refreshLists(groupId: GroupId): Result<Unit> {
        Log.d(TAG, "Refreshing lists for group ${groupId.value}")
        
        val result = listRemoteDataSource.listLists(groupId.value)
        
        return result.map { lists ->
            for (listDto in lists) {
                dbAccess.insertList(listDto.toDb())
            }
        }
    }
}

// ============================================
// Item Repository Implementation
// ============================================

class ItemRepositoryImpl(
    private val dbAccess: DbAccess,
    private val mutationQueueStore: MutationQueueStore,
    private val json: Json = Json { ignoreUnknownKeys = true }
) : ItemRepository {

    companion object {
        private const val TAG = "ItemRepository"
    }

    override fun observeItems(listId: ListId): Flow<List<Item>> =
        dbAccess.observeItemsByListId(listId.value).map { list -> list.map { it.toDomain() } }

    override suspend fun getItem(itemId: ItemId): Item? =
        dbAccess.getItem(itemId.value)?.toDomain()

    override suspend fun addItem(listId: ListId, name: String, priority: Int): Result<Item> {
        Log.d(TAG, "Adding item to list ${listId.value}: $name")
        
        val itemId = ItemId(generateUuid())
        val now = currentTimeMillis()
        val maxSortOrder = dbAccess.getMaxSortOrder(listId.value) ?: 0L
        
        // Create mutation payload
        val payload = json.encodeToString(CreateItemPayload(
            listId = listId.value,
            name = name,
            priority = priority,
            sortOrder = (maxSortOrder + 1).toInt()
        )).encodeToByteArray()
        
        // Enqueue mutation
        val mutationId = mutationQueueStore.enqueue(
            entityType = EntityType.ITEM,
            entityId = itemId.value,
            mutationType = MutationType.CREATE,
            payload = payload
        )
        
        // Optimistically insert into local DB
        val item = Items(
            id = itemId.value,
            list_id = listId.value,
            name = name,
            is_purchased = 0L,
            priority = priority.toLong(),
            sort_order = maxSortOrder + 1,
            category_id = null,
            created_at = now,
            updated_at = now,
            version = 0L,
            pending_mutation_id = mutationId.value
        )
        
        dbAccess.insertItem(item)
        
        return Result.success(item.toDomain())
    }

    override suspend fun togglePurchased(itemId: ItemId): Result<Item> {
        Log.d(TAG, "Toggling purchased state for item ${itemId.value}")
        
        val existing = dbAccess.getItem(itemId.value)
            ?: return Result.failure(DomainError.NotFound(EntityType.ITEM, itemId.value))
        
        val newPurchased = existing.is_purchased == 0L
        val now = currentTimeMillis()
        
        // Create mutation payload
        val payload = json.encodeToString(UpdateItemPayload(
            isPurchased = newPurchased
        )).encodeToByteArray()
        
        // Enqueue mutation
        val mutationId = mutationQueueStore.enqueue(
            entityType = EntityType.ITEM,
            entityId = itemId.value,
            mutationType = MutationType.UPDATE,
            payload = payload,
            expectedVersion = existing.version
        )
        
        // Optimistically update local DB
        dbAccess.updateItemPurchased(
            id = itemId.value,
            isPurchased = newPurchased,
            updatedAt = now,
            pendingMutationId = mutationId.value
        )
        
        // Return updated item
        val updated = dbAccess.getItem(itemId.value)!!
        return Result.success(updated.toDomain())
    }

    override suspend fun reorderItems(listId: ListId, itemIds: List<ItemId>): Result<Unit> {
        Log.d(TAG, "Reordering ${itemIds.size} items in list ${listId.value}")
        
        val now = currentTimeMillis()
        val mutations = mutableListOf<MutationParams>()
        
        // Create mutations for each item's new position
        for ((index, itemId) in itemIds.withIndex()) {
            val existing = dbAccess.getItem(itemId.value) ?: continue
            
            val payload = json.encodeToString(UpdateItemPayload(
                sortOrder = index
            )).encodeToByteArray()
            
            mutations.add(MutationParams(
                entityType = EntityType.ITEM,
                entityId = itemId.value,
                mutationType = MutationType.UPDATE,
                payload = payload,
                expectedVersion = existing.version
            ))
        }
        
        // Enqueue as a batch
        val batchId = mutationQueueStore.enqueueBatch(mutations)
        
        // Optimistically update local DB
        for ((index, itemId) in itemIds.withIndex()) {
            dbAccess.updateItemSortOrder(
                id = itemId.value,
                sortOrder = index,
                updatedAt = now,
                pendingMutationId = null // Batch doesn't track individual pending IDs
            )
        }
        
        return Result.success(Unit)
    }
}

/**
 * Mutation queue store interface (from sync module).
 */
interface MutationQueueStore {
    suspend fun enqueue(
        entityType: EntityType,
        entityId: String,
        mutationType: MutationType,
        payload: ByteArray,
        expectedVersion: Long? = null,
        batchId: String? = null
    ): MutationId

    suspend fun enqueueBatch(mutations: List<MutationParams>): String
}

data class MutationParams(
    val entityType: EntityType,
    val entityId: String,
    val mutationType: MutationType,
    val payload: ByteArray,
    val expectedVersion: Long? = null
)

// ============================================
// Category Repository Implementation
// ============================================

class CategoryRepositoryImpl(
    private val dbAccess: DbAccess
) : CategoryRepository {

    override fun observeCategories(groupId: GroupId): Flow<List<Category>> =
        dbAccess.observeCategoriesByGroupId(groupId.value).map { list -> list.map { it.toDomain() } }

    override suspend fun getCategory(categoryId: CategoryId): Category? =
        dbAccess.getCategory(categoryId.value)?.toDomain()
}

// ============================================
// Payload Types for Mutations
// ============================================

@kotlinx.serialization.Serializable
internal data class CreateItemPayload(
    val listId: String,
    val name: String,
    val priority: Int,
    val sortOrder: Int
)

@kotlinx.serialization.Serializable
internal data class UpdateItemPayload(
    val isPurchased: Boolean? = null,
    val sortOrder: Int? = null,
    val name: String? = null,
    val priority: Int? = null
)
