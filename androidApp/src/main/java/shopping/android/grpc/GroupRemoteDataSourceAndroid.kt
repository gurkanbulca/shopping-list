package shopping.android.grpc

import io.grpc.stub.MetadataUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import shopping.data.remote.*
import shopping.v1.GroupServiceGrpc
import shopping.v1.Group.*
import java.util.concurrent.TimeUnit

/**
 * Android implementation of GroupRemoteDataSource using gRPC.
 * Communicates with the GroupService to list user's groups.
 */
class GroupRemoteDataSourceAndroid(
    private val channelManager: GrpcChannelManager,
    private val tokenProvider: suspend () -> String?
) : GroupRemoteDataSource {

    private val stub: GroupServiceGrpc.GroupServiceBlockingStub
        get() = GroupServiceGrpc.newBlockingStub(channelManager.getChannel())
            .withDeadlineAfter(GrpcConfig.defaultTimeoutMs, TimeUnit.MILLISECONDS)

    override suspend fun listGroups(): Result<List<GroupDto>> {
        return withContext(Dispatchers.IO) {
            try {
                val token = tokenProvider()
                val metadata = createMetadata(accessToken = token)
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)

                val request = ListGroupsRequest.newBuilder().build()
                val response = stubWithMetadata.listGroups(request)

                val groups = response.groupsList.map { group ->
                    GroupDto(
                        id = group.id,
                        name = group.name,
                        createdAt = group.createdAt,
                        updatedAt = group.updatedAt,
                        version = group.version
                    )
                }

                Result.success(groups)
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }
}
