import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCProtobuf

/// iOS implementation of ListRemoteDataSource using gRPC.
/// Communicates with the ListService to list shopping lists within a group.
@available(iOS 18.0, macOS 15.0, *)
class ListRemoteDataSourceIos {
    
    private let clientManager: GrpcClientManager
    
    init(clientManager: GrpcClientManager) {
        self.clientManager = clientManager
    }
    
    /// List all lists in a group.
    func listLists(groupId: String) async -> GrpcResult<[ShoppingListDto]> {
        do {
            let grpcClient = try await clientManager.getClient()
            let listClient = Shopping_V1_ListService.Client(wrapping: grpcClient)
            
            var message = Shopping_V1_ListListsRequest()
            message.groupID = groupId
            message.includeArchived = false
            
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions()
            
            let response: Shopping_V1_ListListsResponse = try await listClient.listLists(
                request: request,
                options: options
            )
            
            let lists = response.lists.map { list in
                ShoppingListDto(
                    id: list.id,
                    groupId: list.groupID,
                    name: list.name,
                    categoryId: nil,
                    createdAt: list.createdAt.seconds * 1000 + Int64(list.createdAt.nanos / 1_000_000),
                    updatedAt: list.updatedAt.seconds * 1000 + Int64(list.updatedAt.nanos / 1_000_000),
                    version: list.version
                )
            }
            
            return .success(lists)
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
}

// MARK: - DTOs

struct ShoppingListDto: Sendable {
    let id: String
    let groupId: String
    let name: String
    let categoryId: String?
    let createdAt: Int64
    let updatedAt: Int64
    let version: Int64
}
