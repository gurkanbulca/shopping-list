import SwiftUI

@available(iOS 15.0, *)
@main
struct ShoppingListApp: App {
    init() {
        // Initialize app dependencies on startup
        Task {
            await AppDependencies.shared.initialize()
        }
    }
    
    var body: some Scene {
        WindowGroup {
            NavRoot()
        }
    }
}
