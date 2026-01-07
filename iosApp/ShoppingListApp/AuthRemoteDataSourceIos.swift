import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCProtobuf

/// iOS implementation of AuthRemoteDataSource using gRPC.
/// Communicates with the AuthService for login, refresh, logout, and profile.
@available(iOS 18.0, macOS 15.0, *)
class AuthRemoteDataSourceIos {
    
    private let clientManager: GrpcClientManager
    
    init(clientManager: GrpcClientManager) {
        self.clientManager = clientManager
    }
    
    /// Authenticate user with email and password.
    func login(email: String, password: String) async -> GrpcResult<AuthTokensDto> {
        do {
            let grpcClient = try await clientManager.getClient()
            let authClient = Shopping_V1_AuthService.Client(wrapping: grpcClient)
            
            var message = Shopping_V1_LoginRequest()
            message.email = email
            message.password = password
            
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions()
            
            let response: Shopping_V1_LoginResponse = try await authClient.login(
                request: request,
                options: options
            )
            
            return .success(AuthTokensDto(
                accessToken: response.accessToken,
                refreshToken: response.refreshToken,
                expiresIn: 3600 // Default 1 hour
            ))
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
    
    /// Refresh the access token using a refresh token.
    func refreshToken(refreshToken: String) async -> GrpcResult<AuthTokensDto> {
        do {
            let grpcClient = try await clientManager.getClient()
            let authClient = Shopping_V1_AuthService.Client(wrapping: grpcClient)
            
            var message = Shopping_V1_RefreshTokenRequest()
            message.refreshToken = refreshToken
            
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions()
            
            let response: Shopping_V1_RefreshTokenResponse = try await authClient.refreshToken(
                request: request,
                options: options
            )
            
            return .success(AuthTokensDto(
                accessToken: response.accessToken,
                refreshToken: response.refreshToken,
                expiresIn: 3600
            ))
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
    
    /// Logout and invalidate the current session.
    func logout(accessToken: String) async -> GrpcResult<Void> {
        // Best effort - clear local tokens, server-side logout not in proto
        return .success(())
    }
    
    /// Get the current user profile.
    func getUserProfile(accessToken: String) async -> GrpcResult<UserDto> {
        do {
            let grpcClient = try await clientManager.getClient()
            let authClient = Shopping_V1_AuthService.Client(wrapping: grpcClient)
            
            let message = Shopping_V1_GetMeRequest()
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions()
            
            let response: Shopping_V1_GetMeResponse = try await authClient.getMe(
                request: request,
                options: options
            )
            
            return .success(UserDto(
                id: response.user.id,
                email: response.user.email,
                displayName: response.user.name
            ))
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
}

// MARK: - DTOs

struct AuthTokensDto: Sendable {
    let accessToken: String
    let refreshToken: String
    let expiresIn: Int64
}

struct UserDto: Sendable {
    let id: String
    let email: String
    let displayName: String
}
