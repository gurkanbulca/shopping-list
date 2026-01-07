package shopping.domain.model

import kotlin.jvm.JvmInline

/**
 * Core domain models for the Shopping List app.
 * These are pure Kotlin data classes, independent of any persistence or network layer.
 */

// ============================================
// Value Types
// ============================================

@JvmInline
value class UserId(val value: String)

@JvmInline
value class GroupId(val value: String)

@JvmInline
value class ListId(val value: String)

@JvmInline
value class ItemId(val value: String)

@JvmInline
value class CategoryId(val value: String)

@JvmInline
value class MutationId(val value: String)

// ============================================
// Enums
// ============================================

enum class MutationType {
    CREATE,
    UPDATE,
    DELETE
}

enum class EntityType {
    GROUP,
    LIST,
    ITEM,
    CATEGORY;

    fun toWireValue(): String = name.lowercase()

    companion object {
        fun fromWireValue(value: String): EntityType =
            entries.first { it.name.equals(value, ignoreCase = true) }
    }
}

enum class MutationStatus {
    PENDING,
    SENDING,
    FAILED
}

enum class SyncState {
    IDLE,
    PUSH_PENDING,
    AWAIT_PUSH,
    PULL_DELTA,
    SYNC_COMPLETE,
    ERROR
}

// ============================================
// Domain Entities
// ============================================

data class User(
    val id: UserId,
    val email: String,
    val displayName: String
)

data class Group(
    val id: GroupId,
    val name: String,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

data class ShoppingList(
    val id: ListId,
    val groupId: GroupId,
    val name: String,
    val categoryId: CategoryId?,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

data class Item(
    val id: ItemId,
    val listId: ListId,
    val name: String,
    val isPurchased: Boolean,
    val priority: Int,
    val sortOrder: Int,
    val categoryId: CategoryId?,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long,
    val pendingMutationId: MutationId? = null
) {
    val isPending: Boolean get() = pendingMutationId != null
}

data class Category(
    val id: CategoryId,
    val groupId: GroupId,
    val name: String,
    val color: String?,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

// ============================================
// Auth Types
// ============================================

data class AuthTokens(
    val accessToken: String,
    val refreshToken: String,
    val expiresAt: Long
) {
    fun isExpired(currentTimeMillis: Long): Boolean =
        currentTimeMillis >= expiresAt
}

data class LoginCredentials(
    val email: String,
    val password: String
)

// ============================================
// Sync Types
// ============================================

data class QueuedMutation(
    val id: MutationId,
    val batchId: String?,
    val entityType: EntityType,
    val entityId: String,
    val mutationType: MutationType,
    val payload: ByteArray,
    val expectedVersion: Long?,
    val status: MutationStatus,
    val retryCount: Int,
    val createdAt: Long,
    val claimedAt: Long?
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || this::class != other::class) return false
        other as QueuedMutation
        return id == other.id
    }

    override fun hashCode(): Int = id.hashCode()
}

data class SyncCursor(
    val groupId: GroupId,
    val cursor: Long,
    val lastSyncAt: Long?,
    val lastError: String?
)

data class DeltaEntity(
    val entityType: EntityType,
    val entityId: String,
    val data: ByteArray,
    val version: Long,
    val isDeleted: Boolean
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || this::class != other::class) return false
        other as DeltaEntity
        return entityType == other.entityType && entityId == other.entityId && version == other.version
    }

    override fun hashCode(): Int {
        var result = entityType.hashCode()
        result = 31 * result + entityId.hashCode()
        result = 31 * result + version.hashCode()
        return result
    }
}

// ============================================
// Error Types
// ============================================

sealed class DomainError : Exception() {
    data class NetworkError(override val message: String) : DomainError()
    data class AuthError(override val message: String) : DomainError()
    data class PermissionDenied(val groupId: GroupId) : DomainError()
    data class ConflictError(val entityType: EntityType, val entityId: String) : DomainError()
    data class NotFound(val entityType: EntityType, val entityId: String) : DomainError()
    data class ValidationError(override val message: String) : DomainError()
    data class UnknownError(override val message: String) : DomainError()
}
