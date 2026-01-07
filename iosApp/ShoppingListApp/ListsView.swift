import SwiftUI
import shared

@available(iOS 15.0, *)
struct ListsView: View {
    let groupId: String
    @StateObject private var viewModel: ListsViewModel
    @Binding var selectedListId: String?
    @Environment(\.dismiss) private var dismiss
    
    init(groupId: String, selectedListId: Binding<String?>) {
        self.groupId = groupId
        self._selectedListId = selectedListId
        self._viewModel = StateObject(wrappedValue: ListsViewModel(groupId: groupId))
    }
    
    var body: some View {
        Group {
            if viewModel.lists.isEmpty {
                VStack {
                    ProgressView()
                    Text("Loading lists...")
                        .foregroundColor(.secondary)
                        .padding(.top)
                }
            } else {
                List(viewModel.lists, id: \.id.value) { list in
                    Button(action: {
                        selectedListId = list.id.value
                    }) {
                        HStack {
                            Text(list.name)
                                .font(.headline)
                            Spacer()
                            Image(systemName: "chevron.right")
                                .foregroundColor(.secondary)
                        }
                    }
                }
            }
        }
        .navigationTitle(viewModel.groupName ?? "Lists")
        .navigationBarTitleDisplayMode(.inline)
    }
}

@available(iOS 15.0, *)
@MainActor
final class ListsViewModel: ObservableObject {
    @Published var lists: [ShoppingList] = []
    @Published var groupName: String?
    
    private let groupId: GroupId
    private var flowObserver: FlowObserver<[ShoppingList]>?
    
    private lazy var observeListsUseCase: ObserveListsUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return ObserveListsUseCase(listRepository: appGraph.listRepository)
    }()
    
    private lazy var selectGroupUseCase: SelectGroupUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return SelectGroupUseCase(
            groupRepository: appGraph.groupRepository,
            syncRepository: appGraph.syncRepository
        )
    }()
    
    init(groupId: String) {
        self.groupId = GroupId(value: groupId)
        
        Task {
            // Select the group
            _ = try? await selectGroupUseCase.invoke(groupId: self.groupId)
            
            // Get group name
            if let group = try? await selectGroupUseCase.getCurrentGroup() {
                self.groupName = group?.name
            }
            
            // Start observing lists
            startObserving()
        }
    }
    
    private func startObserving() {
        let flow = observeListsUseCase.invoke(groupId: groupId)
        flowObserver = FlowObserver.lists(from: flow)
        
        Task { @MainActor in
            for await lists in flowObserver!.$value.values {
                self.lists = lists
            }
        }
    }
}

@available(iOS 15.0, *)
#Preview {
    NavigationStack {
        ListsView(groupId: "test-group", selectedListId: .constant(nil))
    }
}
