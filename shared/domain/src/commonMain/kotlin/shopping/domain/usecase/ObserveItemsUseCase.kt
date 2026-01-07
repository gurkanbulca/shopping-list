package shopping.domain.usecase

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import shopping.domain.model.Item
import shopping.domain.model.ItemId
import shopping.domain.model.ListId
import shopping.domain.repo.ItemRepository

/**
 * Use case for observing items in a list with proper sorting.
 * Items are sorted by:
 * 1. Unpurchased first
 * 2. Higher priority first (within each group)
 * 3. Sort order (stable ordering within priority)
 */
class ObserveItemsUseCase(
    private val itemRepository: ItemRepository
) {
    /**
     * Observe all items in a list with proper sorting applied.
     * @param listId The list to observe items for
     * @return Flow emitting the sorted list of items
     */
    operator fun invoke(listId: ListId): Flow<List<Item>> {
        return itemRepository.observeItems(listId)
            .map { items -> sortItems(items) }
    }

    /**
     * Get a specific item by ID.
     * @param itemId The item ID to fetch
     * @return The item, or null if not found
     */
    suspend fun getItem(itemId: ItemId): Item? {
        return itemRepository.getItem(itemId)
    }

    /**
     * Sort items according to the specification:
     * 1. Unpurchased items first
     * 2. Within each group, sort by priority (higher = more important, descending)
     * 3. Within same priority, sort by sort_order (ascending for stable ordering)
     * 4. Pending items maintain their position
     */
    private fun sortItems(items: List<Item>): List<Item> {
        return items.sortedWith(
            compareBy<Item> { it.isPurchased } // false (unpurchased) before true (purchased)
                .thenByDescending { it.priority } // Higher priority first
                .thenBy { it.sortOrder } // Lower sort_order first (stable)
        )
    }

    /**
     * Get items filtered by purchased state.
     * @param listId The list to observe items for
     * @param isPurchased Filter by purchased state
     * @return Flow emitting the filtered and sorted list of items
     */
    fun observeByPurchasedState(listId: ListId, isPurchased: Boolean): Flow<List<Item>> {
        return invoke(listId)
            .map { items -> items.filter { it.isPurchased == isPurchased } }
    }

    /**
     * Get count of pending items.
     * @param listId The list to check
     * @return Flow emitting the count of pending items
     */
    fun observePendingCount(listId: ListId): Flow<Int> {
        return invoke(listId)
            .map { items -> items.count { it.isPending } }
    }
}
