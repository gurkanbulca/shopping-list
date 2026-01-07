package shopping.sync

import kotlinx.serialization.json.Json
import kotlinx.serialization.decodeFromString
import shopping.db.DbAccess
import shopping.db.Groups
import shopping.db.Lists
import shopping.db.Items
import shopping.db.Categories
import shopping.domain.model.DeltaEntity
import shopping.domain.model.EntityType
import shopping.platform.Log

/**
 * Applies delta entities from the server to the local database.
 * Ensures idempotent application by checking (entity_type, entity_id, version).
 */
class DeltaApplier(
    private val dbAccess: DbAccess,
    private val json: Json = Json { ignoreUnknownKeys = true }
) {
    companion object {
        private const val TAG = "DeltaApplier"
    }

    /**
     * Apply a batch of delta entities to the local database.
     * Skips entities with older or equal versions.
     */
    suspend fun applyDeltas(deltas: List<DeltaEntity>) {
        Log.d(TAG, "Applying ${deltas.size} delta entities")

        for (delta in deltas) {
            try {
                applyDelta(delta)
            } catch (e: Exception) {
                Log.e(TAG, "Failed to apply delta: ${delta.entityType}/${delta.entityId}", e)
                // Continue with other deltas
            }
        }
    }

    /**
     * Apply a single delta entity to the local database.
     */
    private suspend fun applyDelta(delta: DeltaEntity) {
        when (delta.entityType) {
            EntityType.GROUP -> applyGroupDelta(delta)
            EntityType.LIST -> applyListDelta(delta)
            EntityType.ITEM -> applyItemDelta(delta)
            EntityType.CATEGORY -> applyCategoryDelta(delta)
        }
    }

    private suspend fun applyGroupDelta(delta: DeltaEntity) {
        if (delta.isDeleted) {
            Log.d(TAG, "Deleting group: ${delta.entityId}")
            dbAccess.deleteGroup(delta.entityId)
            return
        }

        val existing = dbAccess.getGroup(delta.entityId)
        if (existing != null && existing.version >= delta.version) {
            Log.d(TAG, "Skipping group ${delta.entityId}: local version ${existing.version} >= delta version ${delta.version}")
            return
        }

        val groupData = delta.data.decodeToString()
        val group = json.decodeFromString<GroupDeltaPayload>(groupData)
        
        dbAccess.insertGroup(Groups(
            id = delta.entityId,
            name = group.name,
            created_at = group.createdAt,
            updated_at = group.updatedAt,
            version = delta.version
        ))
        Log.d(TAG, "Applied group delta: ${delta.entityId}")
    }

    private suspend fun applyListDelta(delta: DeltaEntity) {
        if (delta.isDeleted) {
            Log.d(TAG, "Deleting list: ${delta.entityId}")
            dbAccess.deleteList(delta.entityId)
            return
        }

        val existing = dbAccess.getList(delta.entityId)
        if (existing != null && existing.version >= delta.version) {
            Log.d(TAG, "Skipping list ${delta.entityId}: local version ${existing.version} >= delta version ${delta.version}")
            return
        }

        val listData = delta.data.decodeToString()
        val list = json.decodeFromString<ListDeltaPayload>(listData)
        
        dbAccess.insertList(Lists(
            id = delta.entityId,
            group_id = list.groupId,
            name = list.name,
            category_id = list.categoryId,
            created_at = list.createdAt,
            updated_at = list.updatedAt,
            version = delta.version
        ))
        Log.d(TAG, "Applied list delta: ${delta.entityId}")
    }

    private suspend fun applyItemDelta(delta: DeltaEntity) {
        if (delta.isDeleted) {
            Log.d(TAG, "Deleting item: ${delta.entityId}")
            dbAccess.deleteItem(delta.entityId)
            return
        }

        val existing = dbAccess.getItem(delta.entityId)
        if (existing != null && existing.version >= delta.version) {
            Log.d(TAG, "Skipping item ${delta.entityId}: local version ${existing.version} >= delta version ${delta.version}")
            return
        }

        val itemData = delta.data.decodeToString()
        val item = json.decodeFromString<ItemDeltaPayload>(itemData)
        
        dbAccess.insertItem(Items(
            id = delta.entityId,
            list_id = item.listId,
            name = item.name,
            is_purchased = if (item.isPurchased) 1L else 0L,
            priority = item.priority.toLong(),
            sort_order = item.sortOrder.toLong(),
            category_id = item.categoryId,
            created_at = item.createdAt,
            updated_at = item.updatedAt,
            version = delta.version,
            pending_mutation_id = null // Server state clears pending
        ))
        Log.d(TAG, "Applied item delta: ${delta.entityId}")
    }

    private suspend fun applyCategoryDelta(delta: DeltaEntity) {
        if (delta.isDeleted) {
            Log.d(TAG, "Deleting category: ${delta.entityId}")
            dbAccess.deleteCategory(delta.entityId)
            return
        }

        val existing = dbAccess.getCategory(delta.entityId)
        if (existing != null && existing.version >= delta.version) {
            Log.d(TAG, "Skipping category ${delta.entityId}: local version ${existing.version} >= delta version ${delta.version}")
            return
        }

        val categoryData = delta.data.decodeToString()
        val category = json.decodeFromString<CategoryDeltaPayload>(categoryData)
        
        dbAccess.insertCategory(Categories(
            id = delta.entityId,
            group_id = category.groupId,
            name = category.name,
            color = category.color,
            created_at = category.createdAt,
            updated_at = category.updatedAt,
            version = delta.version
        ))
        Log.d(TAG, "Applied category delta: ${delta.entityId}")
    }
}

// ============================================
// Delta Payload Types
// ============================================

@kotlinx.serialization.Serializable
internal data class GroupDeltaPayload(
    val name: String,
    val createdAt: Long,
    val updatedAt: Long
)

@kotlinx.serialization.Serializable
internal data class ListDeltaPayload(
    val groupId: String,
    val name: String,
    val categoryId: String? = null,
    val createdAt: Long,
    val updatedAt: Long
)

@kotlinx.serialization.Serializable
internal data class ItemDeltaPayload(
    val listId: String,
    val name: String,
    val isPurchased: Boolean,
    val priority: Int,
    val sortOrder: Int,
    val categoryId: String? = null,
    val createdAt: Long,
    val updatedAt: Long
)

@kotlinx.serialization.Serializable
internal data class CategoryDeltaPayload(
    val groupId: String,
    val name: String,
    val color: String? = null,
    val createdAt: Long,
    val updatedAt: Long
)
