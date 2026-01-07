package shopping.domain.usecase

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.onStart
import shopping.domain.model.Group
import shopping.domain.repo.GroupRepository

/**
 * Use case for observing the user's groups.
 * Emits from local DB and triggers remote refresh.
 */
class ObserveGroupsUseCase(
    private val groupRepository: GroupRepository
) {
    /**
     * Observe all groups the current user belongs to.
     * Automatically triggers a remote refresh on subscription.
     * @return Flow emitting the list of groups
     */
    operator fun invoke(): Flow<List<Group>> {
        return groupRepository.observeGroups()
            .onStart {
                // Trigger background refresh when flow is collected
                refreshGroups()
            }
    }

    /**
     * Manually refresh groups from the server.
     * @return Result indicating success or failure
     */
    suspend fun refreshGroups(): Result<Unit> {
        return groupRepository.refreshGroups()
    }

    /**
     * Get a specific group by ID.
     * @param groupId The group ID to fetch
     * @return The group, or null if not found
     */
    suspend fun getGroup(groupId: shopping.domain.model.GroupId): Group? {
        return groupRepository.getGroup(groupId)
    }
}
