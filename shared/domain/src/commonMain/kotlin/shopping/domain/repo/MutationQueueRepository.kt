package shopping.domain.repo

import shopping.domain.model.*

/**
 * Repository interface for the mutation queue.
 * Abstracts the mutation queue storage for offline-first writes.
 */
interface MutationQueueRepository {
    /**
     * Enqueue a new mutation.
     * @return The mutation ID (used as idempotency key)
     */
    suspend fun enqueue(
        entityType: EntityType,
        entityId: String,
        mutationType: MutationType,
        payload: ByteArray,
        expectedVersion: Long? = null,
        batchId: String? = null
    ): MutationId

    /**
     * Enqueue a batch of mutations with a shared batch ID.
     * Used for atomic operations like reordering.
     * @return The batch ID
     */
    suspend fun enqueueBatch(mutations: List<MutationInput>): String

    /**
     * Get the count of pending mutations.
     */
    suspend fun getPendingCount(): Int

    /**
     * Check if there are any pending mutations.
     */
    suspend fun hasPending(): Boolean
}

/**
 * Input for creating a mutation.
 */
data class MutationInput(
    val entityType: EntityType,
    val entityId: String,
    val mutationType: MutationType,
    val payload: ByteArray,
    val expectedVersion: Long? = null
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || this::class != other::class) return false
        other as MutationInput
        return entityType == other.entityType && entityId == other.entityId
    }

    override fun hashCode(): Int {
        var result = entityType.hashCode()
        result = 31 * result + entityId.hashCode()
        return result
    }
}
