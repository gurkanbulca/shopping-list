package shopping.domain.usecase

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.onStart
import shopping.domain.model.GroupId
import shopping.domain.model.ListId
import shopping.domain.model.ShoppingList
import shopping.domain.repo.ListRepository

/**
 * Use case for observing shopping lists in a group.
 * Emits from local DB and triggers remote refresh.
 */
class ObserveListsUseCase(
    private val listRepository: ListRepository
) {
    /**
     * Observe all lists in a group.
     * Automatically triggers a remote refresh on subscription.
     * @param groupId The group to observe lists for
     * @return Flow emitting the list of shopping lists
     */
    operator fun invoke(groupId: GroupId): Flow<List<ShoppingList>> {
        return listRepository.observeLists(groupId)
            .onStart {
                // Trigger background refresh when flow is collected
                refreshLists(groupId)
            }
    }

    /**
     * Manually refresh lists from the server.
     * @param groupId The group to refresh lists for
     * @return Result indicating success or failure
     */
    suspend fun refreshLists(groupId: GroupId): Result<Unit> {
        return listRepository.refreshLists(groupId)
    }

    /**
     * Get a specific list by ID.
     * @param listId The list ID to fetch
     * @return The list, or null if not found
     */
    suspend fun getList(listId: ListId): ShoppingList? {
        return listRepository.getList(listId)
    }
}
