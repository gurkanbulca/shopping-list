package shopping.domain.usecase

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import shopping.domain.model.Group
import shopping.domain.model.GroupId
import shopping.domain.repo.GroupRepository
import shopping.domain.repo.SyncRepository

/**
 * Use case for selecting a group and managing group-scoped UI state.
 * Clears cross-group UI state when switching groups.
 */
class SelectGroupUseCase(
    private val groupRepository: GroupRepository,
    private val syncRepository: SyncRepository
) {
    private val _selectedGroupId = MutableStateFlow<GroupId?>(null)
    private val _selectedGroup = MutableStateFlow<Group?>(null)

    /**
     * The currently selected group ID.
     */
    val selectedGroupId: Flow<GroupId?> = _selectedGroupId.asStateFlow()

    /**
     * The currently selected group.
     */
    val selectedGroup: Flow<Group?> = _selectedGroup.asStateFlow()

    /**
     * Select a group and trigger initial sync.
     * Clears previous group-scoped state.
     * @param groupId The group to select
     * @return Result indicating success or failure
     */
    suspend operator fun invoke(groupId: GroupId): Result<Unit> {
        val previousGroupId = _selectedGroupId.value

        // Update selection
        _selectedGroupId.value = groupId

        // Load group details
        val group = groupRepository.getGroup(groupId)
        _selectedGroup.value = group

        // Trigger full sync for the new group if it's different
        if (previousGroupId != groupId) {
            return syncRepository.fullSync(groupId)
        }

        return Result.success(Unit)
    }

    /**
     * Get the current selected group ID synchronously.
     */
    fun getCurrentGroupId(): GroupId? = _selectedGroupId.value

    /**
     * Get the current selected group synchronously.
     */
    fun getCurrentGroup(): Group? = _selectedGroup.value

    /**
     * Clear the group selection.
     * Called on logout or when group access is revoked.
     */
    fun clearSelection() {
        _selectedGroupId.value = null
        _selectedGroup.value = null
    }

    /**
     * Check if a group is selected.
     */
    fun hasSelection(): Boolean = _selectedGroupId.value != null
}
