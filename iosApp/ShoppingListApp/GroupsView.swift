import SwiftUI
import shared

@available(iOS 15.0, *)
struct GroupsView: View {
    @StateObject private var viewModel = GroupsViewModel()
    @Binding var isAuthenticated: Bool
    @Binding var selectedGroupId: String?
    
    var body: some View {
        NavigationStack {
            Group {
                if viewModel.groups.isEmpty {
                    VStack {
                        ProgressView()
                        Text("Loading groups...")
                            .foregroundColor(.secondary)
                            .padding(.top)
                    }
                } else {
                    List(viewModel.groups, id: \.id.value) { group in
                        Button(action: {
                            selectedGroupId = group.id.value
                        }) {
                            HStack {
                                Text(group.name)
                                    .font(.headline)
                                Spacer()
                                Image(systemName: "chevron.right")
                                    .foregroundColor(.secondary)
                            }
                        }
                    }
                }
            }
            .navigationTitle("My Groups")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Logout") {
                        Task {
                            await viewModel.logout()
                            isAuthenticated = false
                        }
                    }
                }
            }
        }
    }
}

@available(iOS 15.0, *)
@MainActor
final class GroupsViewModel: ObservableObject {
    @Published var groups: [Group] = []
    
    private var flowObserver: FlowObserver<[Group]>?
    private lazy var observeGroupsUseCase: ObserveGroupsUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return ObserveGroupsUseCase(groupRepository: appGraph.groupRepository)
    }()
    
    private lazy var loginUseCase: LoginUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return LoginUseCase(authRepository: appGraph.authRepository)
    }()
    
    init() {
        startObserving()
    }
    
    private func startObserving() {
        let flow = observeGroupsUseCase.invoke()
        flowObserver = FlowObserver.groups(from: flow)
        
        // Observe the flow and update published property
        Task { @MainActor in
            for await groups in flowObserver!.$value.values {
                self.groups = groups
            }
        }
    }
    
    func logout() async {
        _ = try? await loginUseCase.logout()
    }
}

@available(iOS 15.0, *)
#Preview {
    GroupsView(isAuthenticated: .constant(true), selectedGroupId: .constant(nil))
}
