package shopping.android.grpc

import io.grpc.ManagedChannel
import io.grpc.ManagedChannelBuilder
import io.grpc.Metadata
import io.grpc.stub.MetadataUtils
import java.util.UUID
import java.util.concurrent.TimeUnit

/**
 * gRPC adapter configuration and channel management for Android.
 * Provides interceptors for authorization, request IDs, and idempotency keys.
 */
object GrpcConfig {
    // TODO: Configure from build config or environment
    var host: String = "10.0.2.2" // Android emulator localhost
    var port: Int = 50051
    var usePlaintext: Boolean = true
    var defaultTimeoutMs: Long = 10_000
}

/**
 * Creates and manages gRPC channels with proper interceptors.
 */
class GrpcChannelManager(
    private val tokenProvider: suspend () -> String?
) {
    private var channel: ManagedChannel? = null

    /**
     * Get or create the managed channel.
     */
    fun getChannel(): ManagedChannel {
        return channel ?: createChannel().also { channel = it }
    }

    /**
     * Create a new managed channel with interceptors.
     */
    private fun createChannel(): ManagedChannel {
        val builder = ManagedChannelBuilder
            .forAddress(GrpcConfig.host, GrpcConfig.port)

        if (GrpcConfig.usePlaintext) {
            builder.usePlaintext()
        }

        return builder.build()
    }

    /**
     * Shutdown the channel gracefully.
     */
    fun shutdown() {
        channel?.shutdown()?.awaitTermination(5, TimeUnit.SECONDS)
        channel = null
    }
}

/**
 * Metadata keys for gRPC headers.
 */
object GrpcMetadataKeys {
    val AUTHORIZATION: Metadata.Key<String> =
        Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER)

    val X_REQUEST_ID: Metadata.Key<String> =
        Metadata.Key.of("x-request-id", Metadata.ASCII_STRING_MARSHALLER)

    val IDEMPOTENCY_KEY: Metadata.Key<String> =
        Metadata.Key.of("idempotency-key", Metadata.ASCII_STRING_MARSHALLER)
}

/**
 * Creates metadata with standard headers.
 */
fun createMetadata(
    accessToken: String? = null,
    idempotencyKey: String? = null
): Metadata {
    val metadata = Metadata()

    // Always add request ID for observability
    metadata.put(GrpcMetadataKeys.X_REQUEST_ID, UUID.randomUUID().toString())

    // Add authorization if available
    accessToken?.let {
        metadata.put(GrpcMetadataKeys.AUTHORIZATION, "Bearer $it")
    }

    // Add idempotency key for mutations
    idempotencyKey?.let {
        metadata.put(GrpcMetadataKeys.IDEMPOTENCY_KEY, it)
    }

    return metadata
}

/**
 * Base class for gRPC adapters with common functionality.
 */
abstract class BaseGrpcAdapter(
    protected val channelManager: GrpcChannelManager,
    protected val tokenProvider: suspend () -> String?
) {
    protected val channel: ManagedChannel
        get() = channelManager.getChannel()

    /**
     * Execute a gRPC call with standard headers and error handling.
     */
    protected suspend inline fun <T> executeCall(
        idempotencyKey: String? = null,
        crossinline block: suspend (Metadata) -> T
    ): Result<T> {
        return try {
            val token = tokenProvider()
            val metadata = createMetadata(
                accessToken = token,
                idempotencyKey = idempotencyKey
            )
            Result.success(block(metadata))
        } catch (e: Exception) {
            Result.failure(mapGrpcException(e))
        }
    }

    /**
     * Map gRPC exceptions to domain errors.
     */
    protected fun mapGrpcException(e: Exception): Exception {
        // TODO: Map specific gRPC status codes to domain errors
        // io.grpc.StatusException or io.grpc.StatusRuntimeException
        return e
    }
}
