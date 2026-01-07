import SwiftUI
import shared

/// Navigation coordinator for the app
@available(iOS 15.0, *)
struct NavRoot: View {
    @State private var isAuthenticated = false
    @State private var selectedGroupId: String?
    @State private var selectedListId: String?
    
    var body: some View {
        Group {
            if !isAuthenticated {
                LoginView(isAuthenticated: $isAuthenticated)
            } else if let groupId = selectedGroupId {
                if let listId = selectedListId {
                    NavigationStack {
                        ListsView(groupId: groupId, selectedListId: $selectedListId)
                            .sheet(item: Binding(
                                get: { selectedListId.map { ListIdWrapper(id: $0) } },
                                set: { selectedListId = $0?.id }
                            )) { wrapper in
                                NavigationStack {
                                    ItemsView(listId: wrapper.id)
                                }
                            }
                    }
                } else {
                    NavigationStack {
                        ListsView(groupId: groupId, selectedListId: $selectedListId)
                    }
                }
            } else {
                GroupsView(isAuthenticated: $isAuthenticated, selectedGroupId: $selectedGroupId)
            }
        }
        .onChange(of: isAuthenticated) { _, newValue in
            if !newValue {
                // Clear navigation state on logout
                selectedGroupId = nil
                selectedListId = nil
            }
        }
    }
}

/// Wrapper to make String identifiable for sheet presentation
struct ListIdWrapper: Identifiable {
    let id: String
}

@available(iOS 15.0, *)
#Preview {
    NavRoot()
}
