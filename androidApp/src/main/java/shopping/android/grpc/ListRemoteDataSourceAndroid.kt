package shopping.android.grpc

import io.grpc.stub.MetadataUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import shopping.data.remote.*
import shopping.v1.ListServiceGrpc
import shopping.v1.List.*
import java.util.concurrent.TimeUnit

/**
 * Android implementation of ListRemoteDataSource using gRPC.
 * Communicates with the ListService to list shopping lists within a group.
 */
class ListRemoteDataSourceAndroid(
    private val channelManager: GrpcChannelManager,
    private val tokenProvider: suspend () -> String?
) : ListRemoteDataSource {

    private val stub: ListServiceGrpc.ListServiceBlockingStub
        get() = ListServiceGrpc.newBlockingStub(channelManager.getChannel())
            .withDeadlineAfter(GrpcConfig.defaultTimeoutMs, TimeUnit.MILLISECONDS)

    override suspend fun listLists(groupId: String): Result<List<ShoppingListDto>> {
        return withContext(Dispatchers.IO) {
            try {
                val token = tokenProvider()
                val metadata = createMetadata(accessToken = token)
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)

                val request = ListListsRequest.newBuilder()
                    .setGroupId(groupId)
                    .build()
                    
                val response = stubWithMetadata.listLists(request)

                val lists = response.listsList.map { list ->
                    ShoppingListDto(
                        id = list.id,
                        groupId = list.groupId,
                        name = list.name,
                        categoryId = if (list.hasCategoryId()) list.categoryId else null,
                        createdAt = list.createdAt,
                        updatedAt = list.updatedAt,
                        version = list.version
                    )
                }

                Result.success(lists)
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }
}
