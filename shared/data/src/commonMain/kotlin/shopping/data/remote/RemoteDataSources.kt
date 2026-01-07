package shopping.data.remote

import shopping.domain.model.*

/**
 * Remote data source interfaces for gRPC communication.
 * These are implemented by platform-specific adapters (Android/iOS).
 */

// ============================================
// Auth Remote Data Source
// ============================================

interface AuthRemoteDataSource {
    /**
     * Authenticate user with email and password.
     */
    suspend fun login(email: String, password: String): Result<AuthTokensDto>

    /**
     * Refresh the access token using a refresh token.
     */
    suspend fun refreshToken(refreshToken: String): Result<AuthTokensDto>

    /**
     * Logout and invalidate the current session.
     */
    suspend fun logout(accessToken: String): Result<Unit>

    /**
     * Get the current user profile.
     */
    suspend fun getUserProfile(accessToken: String): Result<UserDto>
}

// ============================================
// Group Remote Data Source
// ============================================

interface GroupRemoteDataSource {
    /**
     * List all groups the authenticated user belongs to.
     */
    suspend fun listGroups(): Result<List<GroupDto>>
}

// ============================================
// List Remote Data Source
// ============================================

interface ListRemoteDataSource {
    /**
     * List all lists in a group.
     */
    suspend fun listLists(groupId: String): Result<List<ShoppingListDto>>
}

// ============================================
// Item Remote Data Source
// ============================================

interface ItemRemoteDataSource {
    /**
     * List all items in a list.
     */
    suspend fun listItems(listId: String): Result<List<ItemDto>>
}

// ============================================
// Sync Remote Data Source
// ============================================

interface SyncRemoteDataSource {
    /**
     * Get delta changes since the given cursor.
     */
    suspend fun getDelta(groupId: String, cursor: Long, pageSize: Int): Result<GetDeltaResponseDto>

    /**
     * Push a batch of mutations to the server.
     */
    suspend fun pushMutations(groupId: String, mutations: List<MutationDto>): Result<PushMutationsResponseDto>
}

// ============================================
// DTOs (Data Transfer Objects)
// ============================================

data class AuthTokensDto(
    val accessToken: String,
    val refreshToken: String,
    val expiresIn: Long // seconds until expiration
)

data class UserDto(
    val id: String,
    val email: String,
    val displayName: String
)

data class GroupDto(
    val id: String,
    val name: String,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

data class ShoppingListDto(
    val id: String,
    val groupId: String,
    val name: String,
    val categoryId: String?,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

data class ItemDto(
    val id: String,
    val listId: String,
    val name: String,
    val isPurchased: Boolean,
    val priority: Int,
    val sortOrder: Int,
    val categoryId: String?,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

data class CategoryDto(
    val id: String,
    val groupId: String,
    val name: String,
    val color: String?,
    val createdAt: Long,
    val updatedAt: Long,
    val version: Long
)

data class MutationDto(
    val idempotencyKey: String,
    val entityType: String,
    val entityId: String,
    val mutationType: String,
    val entityData: ByteArray,
    val expectedVersion: Long?
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || this::class != other::class) return false
        other as MutationDto
        return idempotencyKey == other.idempotencyKey
    }

    override fun hashCode(): Int = idempotencyKey.hashCode()
}

data class GetDeltaResponseDto(
    val entities: List<DeltaEntityDto>,
    val nextCursor: Long,
    val hasMore: Boolean
)

data class DeltaEntityDto(
    val entityType: String,
    val entityId: String,
    val data: ByteArray,
    val version: Long,
    val isDeleted: Boolean
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || this::class != other::class) return false
        other as DeltaEntityDto
        return entityType == other.entityType && entityId == other.entityId && version == other.version
    }

    override fun hashCode(): Int {
        var result = entityType.hashCode()
        result = 31 * result + entityId.hashCode()
        result = 31 * result + version.hashCode()
        return result
    }
}

data class PushMutationsResponseDto(
    val results: List<MutationResultDto>
)

data class MutationResultDto(
    val idempotencyKey: String,
    val success: Boolean,
    val errorCode: String?,
    val errorMessage: String?,
    val newVersion: Long?
)
