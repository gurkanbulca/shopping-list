import SwiftUI
import shared

@available(iOS 15.0, *)
struct ItemsView: View {
    let listId: String
    @StateObject private var viewModel: ItemsViewModel
    @Environment(\.dismiss) private var dismiss
    
    init(listId: String) {
        self.listId = listId
        self._viewModel = StateObject(wrappedValue: ItemsViewModel(listId: listId))
    }
    
    var body: some View {
        Group {
            if viewModel.items.isEmpty {
                VStack {
                    Text("No items in this list")
                        .foregroundColor(.secondary)
                }
            } else {
                List(viewModel.items, id: \.id.value) { item in
                    ItemRow(item: item)
                }
            }
        }
        .navigationTitle(viewModel.listName ?? "Items")
        .navigationBarTitleDisplayMode(.inline)
    }
}

@available(iOS 15.0, *)
struct ItemRow: View {
    let item: Item
    
    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(item.name)
                    .font(.body)
                    .strikethrough(item.isPurchased)
                    .foregroundColor(item.isPurchased ? .secondary : .primary)
                
                if item.isPending {
                    Text("Pending...")
                        .font(.caption)
                        .foregroundColor(.blue)
                }
                
                if item.priority > 0 {
                    Text("Priority: \(item.priority)")
                        .font(.caption)
                        .foregroundColor(.orange)
                }
            }
            
            Spacer()
            
            if item.isPurchased {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundColor(.green)
            }
        }
        .padding(.vertical, 4)
        .opacity(item.isPending ? 0.7 : 1.0)
    }
}

@available(iOS 15.0, *)
@MainActor
final class ItemsViewModel: ObservableObject {
    @Published var items: [Item] = []
    @Published var listName: String?
    
    private let listId: ListId
    private var flowObserver: FlowObserver<[Item]>?
    
    private lazy var observeItemsUseCase: ObserveItemsUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return ObserveItemsUseCase(itemRepository: appGraph.itemRepository)
    }()
    
    private lazy var observeListsUseCase: ObserveListsUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return ObserveListsUseCase(listRepository: appGraph.listRepository)
    }()
    
    init(listId: String) {
        self.listId = ListId(value: listId)
        
        Task {
            // Get list name
            if let list = try? await observeListsUseCase.getList(listId: self.listId) {
                self.listName = list?.name
            }
            
            // Start observing items
            startObserving()
        }
    }
    
    private func startObserving() {
        let flow = observeItemsUseCase.invoke(listId: listId)
        flowObserver = FlowObserver.items(from: flow)
        
        Task { @MainActor in
            for await items in flowObserver!.$value.values {
                self.items = items
            }
        }
    }
}

@available(iOS 15.0, *)
#Preview {
    NavigationStack {
        ItemsView(listId: "test-list")
    }
}
