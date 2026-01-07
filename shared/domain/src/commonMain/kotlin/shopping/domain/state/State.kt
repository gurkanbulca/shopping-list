package shopping.domain.state

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.map
import shopping.domain.model.*

/**
 * Observable state conventions for UI binding.
 * All state is exposed as Flow-based outputs for reactive UI updates.
 */

// ============================================
// UI State Wrappers
// ============================================

/**
 * Generic loading state wrapper for async operations.
 */
sealed class LoadingState<out T> {
    data object Loading : LoadingState<Nothing>()
    data class Success<T>(val data: T) : LoadingState<T>()
    data class Error(val error: DomainError) : LoadingState<Nothing>()

    val isLoading: Boolean get() = this is Loading
    val isSuccess: Boolean get() = this is Success
    val isError: Boolean get() = this is Error

    fun dataOrNull(): T? = (this as? Success)?.data
    fun errorOrNull(): DomainError? = (this as? Error)?.error
}

/**
 * Extension to convert a Flow to a loading state flow.
 */
fun <T> Flow<T>.asLoadingState(): Flow<LoadingState<T>> = map { LoadingState.Success(it) }

// ============================================
// Screen States
// ============================================

/**
 * State for the Groups screen.
 */
data class GroupsScreenState(
    val groups: List<Group> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null
)

/**
 * State for the Lists screen.
 */
data class ListsScreenState(
    val selectedGroup: Group? = null,
    val lists: List<ShoppingList> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null
)

/**
 * State for the Items screen.
 */
data class ItemsScreenState(
    val selectedList: ShoppingList? = null,
    val items: List<Item> = emptyList(),
    val pendingCount: Int = 0,
    val isLoading: Boolean = false,
    val isSyncing: Boolean = false,
    val lastSyncAt: Long? = null,
    val error: String? = null
)

/**
 * State for the Login screen.
 */
data class LoginScreenState(
    val email: String = "",
    val password: String = "",
    val isLoading: Boolean = false,
    val error: String? = null
)

// ============================================
// Navigation State
// ============================================

/**
 * Global navigation state.
 */
sealed class NavigationState {
    data object Login : NavigationState()
    data object Groups : NavigationState()
    data class Lists(val groupId: GroupId) : NavigationState()
    data class Items(val listId: ListId) : NavigationState()
}

// ============================================
// App State
// ============================================

/**
 * Global app state container.
 */
data class AppState(
    val authState: AuthState = AuthState.Loading,
    val selectedGroupId: GroupId? = null,
    val syncState: SyncState = SyncState.IDLE,
    val pendingMutationCount: Int = 0
) {
    val isAuthenticated: Boolean
        get() = authState is AuthState.Authenticated

    val currentUser: User?
        get() = (authState as? AuthState.Authenticated)?.user
}

/**
 * Auth state from the repository layer.
 */
sealed class AuthState {
    data object Loading : AuthState()
    data object Unauthenticated : AuthState()
    data class Authenticated(val user: User) : AuthState()
}

// ============================================
// Sync Status
// ============================================

/**
 * Human-readable sync status for UI display.
 */
data class SyncStatus(
    val state: SyncState,
    val lastSyncAt: Long?,
    val pendingCount: Int,
    val errorMessage: String? = null
) {
    val statusText: String
        get() = when (state) {
            SyncState.IDLE -> if (pendingCount > 0) "Pending ($pendingCount)" else "Synced"
            SyncState.PUSH_PENDING -> "Preparing to sync..."
            SyncState.AWAIT_PUSH -> "Uploading changes..."
            SyncState.PULL_DELTA -> "Downloading updates..."
            SyncState.SYNC_COMPLETE -> "Sync complete"
            SyncState.ERROR -> errorMessage ?: "Sync error"
        }

    val isActive: Boolean
        get() = state != SyncState.IDLE && state != SyncState.ERROR
}
