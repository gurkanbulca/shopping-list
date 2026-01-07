package shopping.data.di

import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.serialization.json.Json
import shopping.data.mapper.*
import shopping.data.remote.*
import shopping.data.repo.*
import shopping.db.AppDatabase
import shopping.db.DbAccess
import shopping.domain.model.AuthTokens
import shopping.domain.repo.*
import shopping.platform.SecureTokenStore
import shopping.sync.DeltaApplier
import shopping.sync.MutationQueueStore
import shopping.sync.SyncEngine
import shopping.sync.SyncStateStore

/**
 * Manual dependency injection graph for the application.
 * Platform-specific implementations are injected via constructor.
 */
class AppGraph private constructor(
    // Platform-provided dependencies
    private val database: AppDatabase,
    private val secureTokenStore: SecureTokenStore,
    private val authRemoteDataSource: AuthRemoteDataSource,
    private val groupRemoteDataSource: GroupRemoteDataSource,
    private val listRemoteDataSource: ListRemoteDataSource,
    private val syncRemoteDataSource: SyncRemoteDataSource,
    private val dispatcher: CoroutineDispatcher = Dispatchers.Default
) {
    // ============================================
    // Core Dependencies
    // ============================================

    val json: Json = Json {
        ignoreUnknownKeys = true
        encodeDefaults = true
    }

    val dbAccess: DbAccess by lazy {
        DbAccess(database, dispatcher)
    }

    // ============================================
    // Sync Layer
    // ============================================

    val mutationQueueStore: MutationQueueStore by lazy {
        MutationQueueStore(dbAccess)
    }

    val syncStateStore: SyncStateStore by lazy {
        SyncStateStore(dbAccess)
    }

    val deltaApplier: DeltaApplier by lazy {
        DeltaApplier(dbAccess, json)
    }

    val syncEngine: SyncEngine by lazy {
        SyncEngine(
            syncRemoteDataSource = syncRemoteDataSource,
            mutationQueueStore = mutationQueueStore,
            syncStateStore = syncStateStore,
            deltaApplier = deltaApplier
        )
    }

    // ============================================
    // Token Store Adapter
    // ============================================

    /**
     * Adapter that bridges platform SecureTokenStore to the auth repository's SecureTokenStore interface.
     */
    private val authSecureTokenStore: shopping.data.repo.SecureTokenStore by lazy {
        object : shopping.data.repo.SecureTokenStore {
            override suspend fun storeTokens(tokens: AuthTokens) {
                secureTokenStore.storeAccessToken(tokens.accessToken)
                secureTokenStore.storeRefreshToken(tokens.refreshToken)
                secureTokenStore.storeExpiresAt(tokens.expiresAt)
            }

            override suspend fun getTokens(): AuthTokens? {
                val accessToken = secureTokenStore.getAccessToken() ?: return null
                val refreshToken = secureTokenStore.getRefreshToken() ?: return null
                val expiresAt = secureTokenStore.getExpiresAt() ?: return null
                return AuthTokens(
                    accessToken = accessToken,
                    refreshToken = refreshToken,
                    expiresAt = expiresAt
                )
            }

            override suspend fun clearTokens() {
                secureTokenStore.clearAll()
            }
        }
    }

    // ============================================
    // Repositories
    // ============================================

    val authRepository: AuthRepository by lazy {
        AuthRepositoryImpl(
            authRemoteDataSource = authRemoteDataSource,
            secureTokenStore = authSecureTokenStore
        )
    }

    val groupRepository: GroupRepository by lazy {
        GroupRepositoryImpl(
            dbAccess = dbAccess,
            groupRemoteDataSource = groupRemoteDataSource
        )
    }

    val listRepository: ListRepository by lazy {
        ListRepositoryImpl(
            dbAccess = dbAccess,
            listRemoteDataSource = listRemoteDataSource
        )
    }

    val itemRepository: ItemRepository by lazy {
        ItemRepositoryImpl(
            dbAccess = dbAccess,
            mutationQueueRepository = mutationQueueStore, // MutationQueueStore implements MutationQueueRepository
            json = json
        )
    }

    val categoryRepository: CategoryRepository by lazy {
        CategoryRepositoryImpl(dbAccess = dbAccess)
    }

    val syncRepository: SyncRepository by lazy {
        SyncRepositoryImpl(
            syncEngine = syncEngine,
            syncStateStore = syncStateStore,
            mutationQueueStore = mutationQueueStore
        )
    }

    // ============================================
    // Token Provider
    // ============================================

    /**
     * Provides access token for authenticated gRPC calls.
     */
    suspend fun getAccessToken(): String? {
        return secureTokenStore.getAccessToken()
    }

    companion object {
        @Volatile
        private var instance: AppGraph? = null

        /**
         * Initialize the AppGraph with platform-specific dependencies.
         * Must be called once at application startup.
         */
        fun initialize(
            database: AppDatabase,
            secureTokenStore: SecureTokenStore,
            authRemoteDataSource: AuthRemoteDataSource,
            groupRemoteDataSource: GroupRemoteDataSource,
            listRemoteDataSource: ListRemoteDataSource,
            syncRemoteDataSource: SyncRemoteDataSource,
            dispatcher: CoroutineDispatcher = Dispatchers.Default
        ): AppGraph {
            return instance ?: synchronized(this) {
                instance ?: AppGraph(
                    database = database,
                    secureTokenStore = secureTokenStore,
                    authRemoteDataSource = authRemoteDataSource,
                    groupRemoteDataSource = groupRemoteDataSource,
                    listRemoteDataSource = listRemoteDataSource,
                    syncRemoteDataSource = syncRemoteDataSource,
                    dispatcher = dispatcher
                ).also { instance = it }
            }
        }

        /**
         * Get the initialized AppGraph instance.
         * @throws IllegalStateException if not initialized
         */
        fun get(): AppGraph {
            return instance ?: throw IllegalStateException(
                "AppGraph not initialized. Call AppGraph.initialize() first."
            )
        }

        /**
         * Clear the instance (for testing).
         */
        fun clear() {
            instance = null
        }
    }
}

/**
 * Sync repository implementation using the sync engine.
 */
class SyncRepositoryImpl(
    private val syncEngine: SyncEngine,
    private val syncStateStore: SyncStateStore,
    private val mutationQueueStore: MutationQueueStore
) : SyncRepository {

    override fun observeSyncState() = syncEngine.observeSyncState()

    override fun observeSyncCursor(groupId: shopping.domain.model.GroupId) =
        syncStateStore.observeSyncCursor(groupId)

    override fun observePendingMutationCount() =
        mutationQueueStore.observePendingCount()

    override suspend fun sync(groupId: shopping.domain.model.GroupId) =
        syncEngine.sync(groupId)

    override suspend fun fullSync(groupId: shopping.domain.model.GroupId) =
        syncEngine.fullSync(groupId)
}
