import Foundation
import Combine
import shared

/// Bridge for converting Kotlin Flow to SwiftUI ObservableObject
/// This allows Kotlin Flows to be observed in SwiftUI views
@available(iOS 15.0, *)
@MainActor
final class FlowObserver<T>: ObservableObject {
    @Published var value: T
    
    private var cancellable: Task<Void, Never>?
    
    init(flow: Kotlinx_coroutines_coreFlow, initialValue: T, transform: @escaping (Any?) -> T?) {
        self.value = initialValue
        
        self.cancellable = Task { @MainActor [weak self] in
            do {
                let collector = FlowCollector<T> { newValue in
                    Task { @MainActor in
                        self?.value = newValue
                    }
                }
                
                try await flow.collect(collector: collector, completionHandler: { error in
                    if let error = error {
                        print("Flow collection error: \(error)")
                    }
                })
            } catch {
                print("Flow error: \(error)")
            }
        }
    }
    
    deinit {
        cancellable?.cancel()
    }
}

/// Collector that bridges Kotlin Flow emissions to Swift
private class FlowCollector<T>: Kotlinx_coroutines_coreFlowCollector {
    let onValue: (T) -> Void
    
    init(onValue: @escaping (T) -> Void) {
        self.onValue = onValue
    }
    
    func emit(value: Any?) async throws {
        if let typedValue = value as? T {
            onValue(typedValue)
        }
    }
}

/// Helper to create observers for common types
@available(iOS 15.0, *)
extension FlowObserver {
    static func groups(from flow: Kotlinx_coroutines_coreFlow) -> FlowObserver<[Group]> {
        FlowObserver<[Group]>(flow: flow, initialValue: []) { value in
            value as? [Group]
        }
    }
    
    static func lists(from flow: Kotlinx_coroutines_coreFlow) -> FlowObserver<[ShoppingList]> {
        FlowObserver<[ShoppingList]>(flow: flow, initialValue: []) { value in
            value as? [ShoppingList]
        }
    }
    
    static func items(from flow: Kotlinx_coroutines_coreFlow) -> FlowObserver<[Item]> {
        FlowObserver<[Item]>(flow: flow, initialValue: []) { value in
            value as? [Item]
        }
    }
}
