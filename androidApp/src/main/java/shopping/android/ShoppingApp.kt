package shopping.android

import android.app.Application
import shopping.android.auth.AndroidSecureTokenStore
import shopping.android.grpc.*
import shopping.data.di.AppGraph
import shopping.db.DriverFactory
import shopping.db.createDatabase

class ShoppingApp : Application() {
    
    override fun onCreate() {
        super.onCreate()
        // Initialize app-level dependencies
        initializeDependencies()
    }
    
    private fun initializeDependencies() {
        // Create database
        val database = createDatabase(DriverFactory(applicationContext))
        
        // Create secure token store
        val secureTokenStore = AndroidSecureTokenStore(applicationContext)
        
        // Create gRPC channel manager
        val channelManager = GrpcChannelManager(
            tokenProvider = { secureTokenStore.getAccessToken() }
        )
        
        // Create remote data sources
        val authRemoteDataSource = AuthRemoteDataSourceAndroid(channelManager)
        val groupRemoteDataSource = GroupRemoteDataSourceAndroid(
            channelManager = channelManager,
            tokenProvider = { secureTokenStore.getAccessToken() }
        )
        val listRemoteDataSource = ListRemoteDataSourceAndroid(
            channelManager = channelManager,
            tokenProvider = { secureTokenStore.getAccessToken() }
        )
        val syncRemoteDataSource = SyncRemoteDataSourceAndroid(
            channelManager = channelManager,
            tokenProvider = { secureTokenStore.getAccessToken() }
        )
        
        // Initialize AppGraph
        AppGraph.initialize(
            database = database,
            secureTokenStore = secureTokenStore,
            authRemoteDataSource = authRemoteDataSource,
            groupRemoteDataSource = groupRemoteDataSource,
            listRemoteDataSource = listRemoteDataSource,
            syncRemoteDataSource = syncRemoteDataSource
        )
    }
}
