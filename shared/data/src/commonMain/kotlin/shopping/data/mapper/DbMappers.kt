package shopping.data.mapper

import shopping.db.Categories
import shopping.db.Groups
import shopping.db.Items
import shopping.db.Lists
import shopping.db.Mutation_queue
import shopping.db.Sync_state
import shopping.domain.model.*
import shopping.data.remote.*

/**
 * Mappers between database entities, domain models, and DTOs.
 */

// ============================================
// Groups
// ============================================

fun Groups.toDomain(): Group = Group(
    id = GroupId(id),
    name = name,
    createdAt = created_at,
    updatedAt = updated_at,
    version = version
)

fun Group.toDb(): Groups = Groups(
    id = id.value,
    name = name,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version
)

fun GroupDto.toDomain(): Group = Group(
    id = GroupId(id),
    name = name,
    createdAt = createdAt,
    updatedAt = updatedAt,
    version = version
)

fun GroupDto.toDb(): Groups = Groups(
    id = id,
    name = name,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version
)

// ============================================
// Lists
// ============================================

fun Lists.toDomain(): ShoppingList = ShoppingList(
    id = ListId(id),
    groupId = GroupId(group_id),
    name = name,
    categoryId = category_id?.let { CategoryId(it) },
    createdAt = created_at,
    updatedAt = updated_at,
    version = version
)

fun ShoppingList.toDb(): Lists = Lists(
    id = id.value,
    group_id = groupId.value,
    name = name,
    category_id = categoryId?.value,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version
)

fun ShoppingListDto.toDomain(): ShoppingList = ShoppingList(
    id = ListId(id),
    groupId = GroupId(groupId),
    name = name,
    categoryId = categoryId?.let { CategoryId(it) },
    createdAt = createdAt,
    updatedAt = updatedAt,
    version = version
)

fun ShoppingListDto.toDb(): Lists = Lists(
    id = id,
    group_id = groupId,
    name = name,
    category_id = categoryId,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version
)

// ============================================
// Items
// ============================================

fun Items.toDomain(): Item = Item(
    id = ItemId(id),
    listId = ListId(list_id),
    name = name,
    isPurchased = is_purchased != 0L,
    priority = priority.toInt(),
    sortOrder = sort_order.toInt(),
    categoryId = category_id?.let { CategoryId(it) },
    createdAt = created_at,
    updatedAt = updated_at,
    version = version,
    pendingMutationId = pending_mutation_id?.let { MutationId(it) }
)

fun Item.toDb(): Items = Items(
    id = id.value,
    list_id = listId.value,
    name = name,
    is_purchased = if (isPurchased) 1L else 0L,
    priority = priority.toLong(),
    sort_order = sortOrder.toLong(),
    category_id = categoryId?.value,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version,
    pending_mutation_id = pendingMutationId?.value
)

fun ItemDto.toDomain(): Item = Item(
    id = ItemId(id),
    listId = ListId(listId),
    name = name,
    isPurchased = isPurchased,
    priority = priority,
    sortOrder = sortOrder,
    categoryId = categoryId?.let { CategoryId(it) },
    createdAt = createdAt,
    updatedAt = updatedAt,
    version = version
)

fun ItemDto.toDb(): Items = Items(
    id = id,
    list_id = listId,
    name = name,
    is_purchased = if (isPurchased) 1L else 0L,
    priority = priority.toLong(),
    sort_order = sortOrder.toLong(),
    category_id = categoryId,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version,
    pending_mutation_id = null
)

// ============================================
// Categories
// ============================================

fun Categories.toDomain(): Category = Category(
    id = CategoryId(id),
    groupId = GroupId(group_id),
    name = name,
    color = color,
    createdAt = created_at,
    updatedAt = updated_at,
    version = version
)

fun Category.toDb(): Categories = Categories(
    id = id.value,
    group_id = groupId.value,
    name = name,
    color = color,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version
)

fun CategoryDto.toDomain(): Category = Category(
    id = CategoryId(id),
    groupId = GroupId(groupId),
    name = name,
    color = color,
    createdAt = createdAt,
    updatedAt = updatedAt,
    version = version
)

fun CategoryDto.toDb(): Categories = Categories(
    id = id,
    group_id = groupId,
    name = name,
    color = color,
    created_at = createdAt,
    updated_at = updatedAt,
    version = version
)

// ============================================
// Mutation Queue
// ============================================

fun Mutation_queue.toDomain(): QueuedMutation = QueuedMutation(
    id = MutationId(id),
    batchId = batch_id,
    entityType = EntityType.fromWireValue(entity_type),
    entityId = entity_id,
    mutationType = MutationType.valueOf(mutation_type),
    payload = payload,
    expectedVersion = expected_version,
    status = MutationStatus.valueOf(status.uppercase()),
    retryCount = retry_count.toInt(),
    createdAt = created_at,
    claimedAt = claimed_at
)

fun QueuedMutation.toDb(): Mutation_queue = Mutation_queue(
    id = id.value,
    batch_id = batchId,
    entity_type = entityType.toWireValue(),
    entity_id = entityId,
    mutation_type = mutationType.name,
    payload = payload,
    expected_version = expectedVersion,
    status = status.name.lowercase(),
    retry_count = retryCount.toLong(),
    created_at = createdAt,
    claimed_at = claimedAt
)

fun QueuedMutation.toDto(): MutationDto = MutationDto(
    idempotencyKey = id.value,
    entityType = entityType.toWireValue(),
    entityId = entityId,
    mutationType = mutationType.name,
    entityData = payload,
    expectedVersion = expectedVersion
)

// ============================================
// Sync State
// ============================================

fun Sync_state.toDomain(): SyncCursor = SyncCursor(
    groupId = GroupId(group_id),
    cursor = cursor,
    lastSyncAt = last_sync_at,
    lastError = last_error
)

fun SyncCursor.toDb(): Sync_state = Sync_state(
    group_id = groupId.value,
    cursor = cursor,
    last_sync_at = lastSyncAt,
    last_error = lastError
)

// ============================================
// Delta Entity
// ============================================

fun DeltaEntityDto.toDomain(): DeltaEntity = DeltaEntity(
    entityType = EntityType.fromWireValue(entityType),
    entityId = entityId,
    data = data,
    version = version,
    isDeleted = isDeleted
)

// ============================================
// Auth
// ============================================

fun AuthTokensDto.toDomain(currentTimeMillis: Long): AuthTokens = AuthTokens(
    accessToken = accessToken,
    refreshToken = refreshToken,
    expiresAt = currentTimeMillis + (expiresIn * 1000)
)

fun UserDto.toDomain(): User = User(
    id = UserId(id),
    email = email,
    displayName = displayName
)
