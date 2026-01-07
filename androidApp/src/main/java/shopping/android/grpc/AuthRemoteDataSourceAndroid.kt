package shopping.android.grpc

import io.grpc.ManagedChannel
import io.grpc.stub.MetadataUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import shopping.data.remote.*
import shopping.v1.AuthServiceGrpc
import shopping.v1.Auth.*
import java.util.concurrent.TimeUnit

/**
 * Android implementation of AuthRemoteDataSource using gRPC.
 * Communicates with the AuthService for login, refresh, logout, and profile.
 */
class AuthRemoteDataSourceAndroid(
    private val channelManager: GrpcChannelManager
) : AuthRemoteDataSource {

    private val stub: AuthServiceGrpc.AuthServiceBlockingStub
        get() = AuthServiceGrpc.newBlockingStub(channelManager.getChannel())
            .withDeadlineAfter(GrpcConfig.defaultTimeoutMs, TimeUnit.MILLISECONDS)

    override suspend fun login(email: String, password: String): Result<AuthTokensDto> {
        return withContext(Dispatchers.IO) {
            try {
                val request = LoginRequest.newBuilder()
                    .setEmail(email)
                    .setPassword(password)
                    .build()

                val metadata = createMetadata()
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)
                
                val response = stubWithMetadata.login(request)
                
                Result.success(AuthTokensDto(
                    accessToken = response.accessToken,
                    refreshToken = response.refreshToken,
                    expiresIn = response.expiresIn
                ))
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }

    override suspend fun refreshToken(refreshToken: String): Result<AuthTokensDto> {
        return withContext(Dispatchers.IO) {
            try {
                val request = RefreshTokenRequest.newBuilder()
                    .setRefreshToken(refreshToken)
                    .build()

                val metadata = createMetadata()
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)
                
                val response = stubWithMetadata.refreshToken(request)
                
                Result.success(AuthTokensDto(
                    accessToken = response.accessToken,
                    refreshToken = response.refreshToken,
                    expiresIn = response.expiresIn
                ))
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }

    override suspend fun logout(accessToken: String): Result<Unit> {
        return withContext(Dispatchers.IO) {
            try {
                val request = LogoutRequest.newBuilder().build()

                val metadata = createMetadata(accessToken = accessToken)
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)
                
                stubWithMetadata.logout(request)
                
                Result.success(Unit)
            } catch (e: Exception) {
                // Best effort logout - don't fail if server is unreachable
                Result.success(Unit)
            }
        }
    }

    override suspend fun getUserProfile(accessToken: String): Result<UserDto> {
        return withContext(Dispatchers.IO) {
            try {
                val request = GetUserProfileRequest.newBuilder().build()

                val metadata = createMetadata(accessToken = accessToken)
                val stubWithMetadata = MetadataUtils.attachHeaders(stub, metadata)
                
                val response = stubWithMetadata.getUserProfile(request)
                
                Result.success(UserDto(
                    id = response.user.id,
                    email = response.user.email,
                    displayName = response.user.displayName
                ))
            } catch (e: Exception) {
                Result.failure(GrpcExceptionMapper.mapException(e))
            }
        }
    }
}
