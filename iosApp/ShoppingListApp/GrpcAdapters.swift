import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2

// MARK: - Common Types

/// Result type for gRPC operations
typealias GrpcResult<T> = Result<T, GrpcError>

/// gRPC error types mapped from server responses
enum GrpcError: Error, Sendable {
    case unauthenticated(String)
    case permissionDenied(String)
    case notFound(String)
    case failedPrecondition(String)
    case unavailable(String)
    case deadlineExceeded(String)
    case cancelled(String)
    case unknown(String)
    
    var message: String {
        switch self {
        case .unauthenticated(let msg),
             .permissionDenied(let msg),
             .notFound(let msg),
             .failedPrecondition(let msg),
             .unavailable(let msg),
             .deadlineExceeded(let msg),
             .cancelled(let msg),
             .unknown(let msg):
            return msg
        }
    }
    
    var isRetryable: Bool {
        switch self {
        case .unavailable, .deadlineExceeded:
            return true
        default:
            return false
        }
    }
    
    var isAuthError: Bool {
        switch self {
        case .unauthenticated:
            return true
        default:
            return false
        }
    }
}

/// Protocol for providing authentication tokens
protocol TokenProvider: Sendable {
    func getAccessToken() async -> String?
}

// MARK: - Service Container

/// Container for all gRPC services. Initialize once and share across the app.
@available(iOS 18.0, macOS 15.0, *)
final class GrpcServices: Sendable {
    
    private let clientManager: GrpcClientManager
    
    /// Shared instance using default configuration
    static let shared = GrpcServices()
    
    init(config: GrpcClientConfig = .default) {
        self.clientManager = GrpcClientManager(config: config)
    }
    
    /// Set the token provider for authentication
    func setTokenProvider(_ provider: TokenProvider) async {
        await clientManager.setTokenProvider(provider)
    }
    
    /// Get the auth remote data source
    var auth: AuthRemoteDataSourceIos {
        AuthRemoteDataSourceIos(clientManager: clientManager)
    }
    
    /// Get the group remote data source
    var groups: GroupRemoteDataSourceIos {
        GroupRemoteDataSourceIos(clientManager: clientManager)
    }
    
    /// Get the list remote data source
    var lists: ListRemoteDataSourceIos {
        ListRemoteDataSourceIos(clientManager: clientManager)
    }
    
    /// Get the sync remote data source
    var sync: SyncRemoteDataSourceIos {
        SyncRemoteDataSourceIos(clientManager: clientManager)
    }
    
    /// Shutdown all gRPC connections
    func shutdown() async {
        await clientManager.close()
    }
}

// MARK: - GrpcClientManager Extension

@available(iOS 18.0, macOS 15.0, *)
extension GrpcClientManager {
    func setTokenProvider(_ provider: TokenProvider) {
        self.tokenProvider = provider
    }
}
