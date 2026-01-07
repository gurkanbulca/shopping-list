import Foundation

/// gRPC adapter configuration for iOS.
/// Provides channel management and metadata injection.
struct GrpcConfig {
    // TODO: Configure from environment or Info.plist
    static var host: String = "localhost"
    static var port: Int = 50051
    static var useTLS: Bool = false
    static var defaultTimeoutSeconds: TimeInterval = 10
}

/// Metadata keys for gRPC headers.
enum GrpcMetadataKeys {
    static let authorization = "authorization"
    static let xRequestId = "x-request-id"
    static let idempotencyKey = "idempotency-key"
}

/// Protocol for providing authentication tokens.
protocol TokenProvider {
    func getAccessToken() async -> String?
}

/// Creates standard metadata headers for gRPC calls.
func createMetadata(
    accessToken: String? = nil,
    idempotencyKey: String? = nil
) -> [String: String] {
    var headers: [String: String] = [:]
    
    // Always add request ID for observability
    headers[GrpcMetadataKeys.xRequestId] = UUID().uuidString
    
    // Add authorization if available
    if let token = accessToken {
        headers[GrpcMetadataKeys.authorization] = "Bearer \(token)"
    }
    
    // Add idempotency key for mutations
    if let key = idempotencyKey {
        headers[GrpcMetadataKeys.idempotencyKey] = key
    }
    
    return headers
}

/// Result type for gRPC operations.
enum GrpcResult<T> {
    case success(T)
    case failure(GrpcError)
    
    var value: T? {
        if case .success(let value) = self {
            return value
        }
        return nil
    }
    
    var error: GrpcError? {
        if case .failure(let error) = self {
            return error
        }
        return nil
    }
}

/// Error types for gRPC operations.
enum GrpcError: Error {
    case unauthenticated(String)
    case permissionDenied(String)
    case notFound(String)
    case failedPrecondition(String)
    case unavailable(String)
    case deadlineExceeded(String)
    case unknown(String)
    
    var message: String {
        switch self {
        case .unauthenticated(let msg): return msg
        case .permissionDenied(let msg): return msg
        case .notFound(let msg): return msg
        case .failedPrecondition(let msg): return msg
        case .unavailable(let msg): return msg
        case .deadlineExceeded(let msg): return msg
        case .unknown(let msg): return msg
        }
    }
}

/// Base class for gRPC adapters with common functionality.
/// Subclasses implement specific service adapters.
class BaseGrpcAdapter {
    let tokenProvider: TokenProvider
    
    init(tokenProvider: TokenProvider) {
        self.tokenProvider = tokenProvider
    }
    
    /// Execute a gRPC call with standard headers and error handling.
    func executeCall<T>(
        idempotencyKey: String? = nil,
        block: @escaping ([String: String]) async throws -> T
    ) async -> GrpcResult<T> {
        do {
            let token = await tokenProvider.getAccessToken()
            let metadata = createMetadata(
                accessToken: token,
                idempotencyKey: idempotencyKey
            )
            let result = try await block(metadata)
            return .success(result)
        } catch {
            return .failure(mapError(error))
        }
    }
    
    /// Map errors to GrpcError type.
    private func mapError(_ error: Error) -> GrpcError {
        // TODO: Map specific gRPC status codes
        // For now, return unknown error
        return .unknown(error.localizedDescription)
    }
}

// MARK: - Placeholder Adapters

/// Placeholder for Auth gRPC adapter.
/// Will be implemented with actual gRPC client.
class AuthGrpcAdapter: BaseGrpcAdapter {
    // TODO: Implement AuthService RPC methods
}

/// Placeholder for Group gRPC adapter.
class GroupGrpcAdapter: BaseGrpcAdapter {
    // TODO: Implement GroupService RPC methods
}

/// Placeholder for List gRPC adapter.
class ListGrpcAdapter: BaseGrpcAdapter {
    // TODO: Implement ListService RPC methods
}

/// Placeholder for Sync gRPC adapter.
class SyncGrpcAdapter: BaseGrpcAdapter {
    // TODO: Implement SyncService RPC methods
}
