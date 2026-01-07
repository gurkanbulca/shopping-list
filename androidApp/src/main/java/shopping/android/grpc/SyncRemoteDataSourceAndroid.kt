package shopping.android.grpc

import com.google.protobuf.ByteString
import io.grpc.stub.MetadataUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import shopping.data.remote.*
import shopping.v1.SyncServiceGrpc
import shopping.v1.Sync.*
import java.util.concurrent.TimeUnit

/**
 * Android implementation of SyncRemoteDataSource using gRPC.
 * Communicates with the SyncService for delta sync and push mutations.
 */
class SyncRemoteDataSourceAndroid(
    private val channelManager: GrpcChannelManager,
    private val tokenProvider: suspend () -> String?
) : SyncRemoteDataSource {

    private val stub: SyncServiceGrpc.SyncServiceBlockingStub
        get() = SyncServiceGrpc.newBlockingStub(channelManager.getChannel())
            .withDeadlineAfter(GrpcConfig.defaultTimeoutMs, TimeUnit.MILLISECONDS)

    override suspend fun getDelta(groupId: String, cursor: Long, pageSize: Int): Result<GetDeltaResponseDto> {
        return withContext(Dispatchers.IO) {
            try {
                val token = tokenProvider()
                val metadata = createMetadata(accessToken = token)
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)

                val request = GetDeltaRequest.newBuilder()
                    .setGroupId(groupId)
                    .setCursor(cursor)
                    .setPageSize(pageSize)
                    .build()

                val response = stubWithMetadata.getDelta(request)

                val entities = response.entitiesList.map { entity ->
                    DeltaEntityDto(
                        entityType = entity.entityType,
                        entityId = entity.entityId,
                        data = entity.data.toByteArray(),
                        version = entity.version,
                        isDeleted = entity.isDeleted
                    )
                }

                Result.success(GetDeltaResponseDto(
                    entities = entities,
                    nextCursor = response.nextCursor,
                    hasMore = response.hasMore
                ))
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }

    override suspend fun pushMutations(groupId: String, mutations: List<MutationDto>): Result<PushMutationsResponseDto> {
        return withContext(Dispatchers.IO) {
            try {
                val token = tokenProvider()
                // Use first mutation's idempotency key for the request-level idempotency
                val idempotencyKey = mutations.firstOrNull()?.idempotencyKey
                val metadata = createMetadata(accessToken = token, idempotencyKey = idempotencyKey)
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)

                val protoMutations = mutations.map { mutation ->
                    val builder = Mutation.newBuilder()
                        .setIdempotencyKey(mutation.idempotencyKey)
                        .setEntityType(mutation.entityType)
                        .setEntityId(mutation.entityId)
                        .setMutationType(mutation.mutationType)
                        .setEntityData(ByteString.copyFrom(mutation.entityData))
                    
                    mutation.expectedVersion?.let { builder.setExpectedVersion(it) }
                    
                    builder.build()
                }

                val request = PushMutationsRequest.newBuilder()
                    .setGroupId(groupId)
                    .addAllMutations(protoMutations)
                    .build()

                val response = stubWithMetadata.pushMutations(request)

                val results = response.resultsList.map { result ->
                    MutationResultDto(
                        idempotencyKey = result.idempotencyKey,
                        success = result.success,
                        errorCode = if (result.hasErrorCode()) result.errorCode else null,
                        errorMessage = if (result.hasErrorMessage()) result.errorMessage else null,
                        newVersion = if (result.hasNewVersion()) result.newVersion else null
                    )
                }

                Result.success(PushMutationsResponseDto(results = results))
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }
}
