import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCProtobuf

/// Configuration for the gRPC client connection
struct GrpcClientConfig: Sendable {
    let host: String
    let port: Int
    let useTLS: Bool
    let defaultTimeoutSeconds: TimeInterval
    
    static let `default` = GrpcClientConfig(
        host: "localhost",
        port: 50051,
        useTLS: false,
        defaultTimeoutSeconds: 10
    )
}

/// Manages gRPC client and provides service clients
@available(iOS 18.0, macOS 15.0, *)
actor GrpcClientManager {
    private let config: GrpcClientConfig
    private var client: GRPCClient<HTTP2ClientTransport.Posix>?
    
    /// Token provider for authentication
    var tokenProvider: TokenProvider?
    
    init(config: GrpcClientConfig = .default) {
        self.config = config
    }
    
    /// Get or create the gRPC client
    func getClient() async throws -> GRPCClient<HTTP2ClientTransport.Posix> {
        if let existing = client {
            return existing
        }
        
        let transport = try HTTP2ClientTransport.Posix(
            target: .ipv4(host: config.host, port: config.port),
            config: .defaults(transportSecurity: config.useTLS ? .tls : .plaintext)
        )
        
        let newClient = GRPCClient(transport: transport)
        self.client = newClient
        
        // Start the client in a background task
        Task {
            try await newClient.run()
        }
        
        return newClient
    }
    
    /// Create call options with standard metadata headers
    func callOptions(idempotencyKey: String? = nil) async -> CallOptions {
        var metadata: Metadata = [:]
        
        // Add request ID for observability
        metadata["x-request-id"] = "\(UUID().uuidString)"
        
        // Add authorization if available
        if let token = await tokenProvider?.getAccessToken() {
            metadata["authorization"] = "Bearer \(token)"
        }
        
        // Add idempotency key for mutations
        if let key = idempotencyKey {
            metadata["idempotency-key"] = "\(key)"
        }
        
        return CallOptions(
            timeout: .seconds(Int64(config.defaultTimeoutSeconds))
        )
    }
    
    /// Close the client
    func close() {
        client?.beginGracefulShutdown()
        client = nil
    }
}

/// Maps gRPC errors to application errors
@available(iOS 18.0, macOS 15.0, *)
enum GrpcStatusMapper {
    static func mapError(_ error: Error) -> GrpcError {
        if let rpcError = error as? RPCError {
            switch rpcError.code {
            case .unauthenticated:
                return .unauthenticated(rpcError.message ?? "Authentication required")
            case .permissionDenied:
                return .permissionDenied(rpcError.message ?? "Permission denied")
            case .notFound:
                return .notFound(rpcError.message ?? "Resource not found")
            case .failedPrecondition:
                return .failedPrecondition(rpcError.message ?? "Precondition failed")
            case .unavailable:
                return .unavailable(rpcError.message ?? "Service unavailable")
            case .deadlineExceeded:
                return .deadlineExceeded(rpcError.message ?? "Request timed out")
            default:
                return .unknown(rpcError.message ?? "Unknown error: \(rpcError.code)")
            }
        }
        return .unknown(error.localizedDescription)
    }
    
    static func isAuthError(_ error: Error) -> Bool {
        guard let rpcError = error as? RPCError else { return false }
        return rpcError.code == .unauthenticated
    }
    
    static func isConflict(_ error: Error) -> Bool {
        guard let rpcError = error as? RPCError else { return false }
        return rpcError.code == .failedPrecondition
    }
    
    static func isRetryable(_ error: Error) -> Bool {
        guard let rpcError = error as? RPCError else { return false }
        return rpcError.code == .unavailable || rpcError.code == .deadlineExceeded || rpcError.code == .resourceExhausted
    }
}
