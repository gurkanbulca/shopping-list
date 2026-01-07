package shopping.sync

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import shopping.db.DbAccess
import shopping.db.Mutation_queue
import shopping.domain.model.*
import shopping.domain.repo.MutationQueueRepository
import shopping.domain.repo.MutationInput
import shopping.data.mapper.toDomain
import shopping.data.mapper.toDb
import shopping.platform.Log
import shopping.platform.currentTimeMillis
import shopping.platform.generateUuid

/**
 * Manages the local mutation queue for offline-first writes.
 * Mutations are queued locally and synced when connectivity is available.
 */
class MutationQueueStore(
    private val dbAccess: DbAccess
) : MutationQueueRepository {
    companion object {
        private const val TAG = "MutationQueueStore"
        private const val DEFAULT_BATCH_SIZE = 50
    }

    /**
     * Observe the count of pending mutations.
     */
    fun observePendingCount(): Flow<Int> {
        // Note: This is a simplified version. In production, use a proper reactive query.
        // For now, we'll emit changes via the DB.
        return kotlinx.coroutines.flow.flow {
            while (true) {
                emit(dbAccess.countPendingMutations().toInt())
                kotlinx.coroutines.delay(1000) // Poll every second
            }
        }
    }

    /**
     * Enqueue a new mutation.
     * @return The mutation ID (used as idempotency key)
     */
    override suspend fun enqueue(
        entityType: EntityType,
        entityId: String,
        mutationType: MutationType,
        payload: ByteArray,
        expectedVersion: Long? = null,
        batchId: String? = null
    ): MutationId {
        val mutationId = MutationId(generateUuid())
        
        Log.d(TAG, "Enqueueing mutation: ${mutationType.name} ${entityType.toWireValue()}/$entityId")
        
        dbAccess.insertMutation(Mutation_queue(
            id = mutationId.value,
            batch_id = batchId,
            entity_type = entityType.toWireValue(),
            entity_id = entityId,
            mutation_type = mutationType.name,
            payload = payload,
            expected_version = expectedVersion,
            status = MutationStatus.PENDING.name.lowercase(),
            retry_count = 0,
            created_at = currentTimeMillis(),
            claimed_at = null
        ))
        
        return mutationId
    }

    /**
     * Enqueue a batch of mutations with a shared batch ID.
     * Used for atomic operations like reordering.
     */
    override suspend fun enqueueBatch(
        mutations: List<MutationInput>
    ): String {
        val batchId = generateUuid()
        Log.d(TAG, "Enqueueing batch with ${mutations.size} mutations, batchId=$batchId")
        
        for (params in mutations) {
            enqueue(
                entityType = params.entityType,
                entityId = params.entityId,
                mutationType = params.mutationType,
                payload = params.payload,
                expectedVersion = params.expectedVersion,
                batchId = batchId
            )
        }
        
        return batchId
    }

    /**
     * Claim pending mutations for sending.
     * @return List of claimed mutations, or empty if none available
     */
    suspend fun claimPending(limit: Int = DEFAULT_BATCH_SIZE): List<QueuedMutation> {
        val pending = dbAccess.selectPendingMutations(limit)
        
        if (pending.isEmpty()) {
            return emptyList()
        }
        
        val ids = pending.map { it.id }
        val claimedAt = currentTimeMillis()
        
        dbAccess.claimMutations(ids, claimedAt)
        Log.d(TAG, "Claimed ${pending.size} mutations for sending")
        
        return pending.map { it.toDomain() }
    }

    /**
     * Claim all mutations for a batch.
     * Used for atomic batch operations like reordering.
     */
    suspend fun claimBatch(batchId: String): List<QueuedMutation> {
        val mutations = dbAccess.selectMutationsByBatchId(batchId)
        
        if (mutations.isEmpty()) {
            return emptyList()
        }
        
        val ids = mutations.map { it.id }
        val claimedAt = currentTimeMillis()
        
        dbAccess.claimMutations(ids, claimedAt)
        Log.d(TAG, "Claimed batch $batchId with ${mutations.size} mutations")
        
        return mutations.map { it.toDomain() }
    }

    /**
     * Mark a mutation as successfully sent.
     * Removes it from the queue.
     */
    suspend fun markSuccess(mutationId: MutationId) {
        Log.d(TAG, "Marking mutation success: ${mutationId.value}")
        dbAccess.markMutationSuccess(mutationId.value)
        // Also clear pending state from affected entities
        dbAccess.clearItemPendingMutation(mutationId.value)
    }

    /**
     * Mark a mutation as failed.
     * Increments retry count but keeps it in queue for potential retry.
     */
    suspend fun markFailed(mutationId: MutationId) {
        Log.w(TAG, "Marking mutation failed: ${mutationId.value}")
        dbAccess.markMutationFailed(mutationId.value)
    }

    /**
     * Reset a failed mutation to pending state for retry.
     */
    suspend fun resetForRetry(mutationId: MutationId) {
        Log.d(TAG, "Resetting mutation for retry: ${mutationId.value}")
        dbAccess.resetFailedMutation(mutationId.value)
    }

    /**
     * Get a specific mutation by ID.
     */
    suspend fun getMutation(mutationId: MutationId): QueuedMutation? =
        dbAccess.getMutation(mutationId.value)?.toDomain()

    /**
     * Get the count of pending mutations.
     */
    override suspend fun getPendingCount(): Int =
        dbAccess.countPendingMutations().toInt()

    /**
     * Check if there are any pending mutations.
     */
    override suspend fun hasPending(): Boolean =
        getPendingCount() > 0

    /**
     * Clear all mutations (e.g., on logout or permission denied).
     */
    suspend fun clearAll() {
        Log.w(TAG, "Clearing all mutations from queue")
        dbAccess.deleteAllMutations()
    }
}

