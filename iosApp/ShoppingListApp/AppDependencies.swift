import Foundation
import shared

/// Manages application-level dependencies and initialization
@available(iOS 15.0, *)
@MainActor
final class AppDependencies: ObservableObject {
    static let shared = AppDependencies()
    
    private(set) var appGraph: AppGraph?
    private(set) var secureTokenStore: IosSecureTokenStore?
    
    private init() {}
    
    /// Initialize all app dependencies
    func initialize() async {
        // Create secure token store
        let tokenStore = IosSecureTokenStore()
        self.secureTokenStore = tokenStore
        
        // Create database
        let database = createDatabase(driverFactory: DriverFactory())
        
        // Create gRPC services
        let grpcServices = GrpcServices.shared
        
        // Set token provider for authenticated requests
        await grpcServices.setTokenProvider(tokenStore)
        
        // Create remote data sources
        let authRemoteDataSource = grpcServices.auth
        let groupRemoteDataSource = grpcServices.groups
        let listRemoteDataSource = grpcServices.lists
        let syncRemoteDataSource = grpcServices.sync
        
        // Initialize AppGraph
        self.appGraph = AppGraph.companion.initialize(
            database: database,
            secureTokenStore: tokenStore,
            authRemoteDataSource: authRemoteDataSource,
            groupRemoteDataSource: groupRemoteDataSource,
            listRemoteDataSource: listRemoteDataSource,
            syncRemoteDataSource: syncRemoteDataSource,
            dispatcher: Kotlinx_coroutines_coreDispatchers.shared.default
        )
    }
    
    /// Get the initialized AppGraph
    func getAppGraph() -> AppGraph {
        guard let appGraph = appGraph else {
            fatalError("AppGraph not initialized. Call initialize() first.")
        }
        return appGraph
    }
}

/// Token provider implementation
extension IosSecureTokenStore: TokenProvider {
    func getAccessToken() async -> String? {
        return try? await self.getAccessToken()
    }
}
