import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCProtobuf

/// iOS implementation of GroupRemoteDataSource using gRPC.
/// Communicates with the GroupService to list user's groups.
@available(iOS 18.0, macOS 15.0, *)
class GroupRemoteDataSourceIos {
    
    private let clientManager: GrpcClientManager
    
    init(clientManager: GrpcClientManager) {
        self.clientManager = clientManager
    }
    
    /// List all groups the authenticated user belongs to.
    func listGroups() async -> GrpcResult<[GroupDto]> {
        do {
            let grpcClient = try await clientManager.getClient()
            let groupClient = Shopping_V1_GroupService.Client(wrapping: grpcClient)
            
            let message = Shopping_V1_ListMyGroupsRequest()
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions()
            
            let response: Shopping_V1_ListMyGroupsResponse = try await groupClient.listMyGroups(
                request: request,
                options: options
            )
            
            let groups = response.groups.map { group in
                GroupDto(
                    id: group.id,
                    name: group.name,
                    createdAt: group.createdAt.seconds * 1000 + Int64(group.createdAt.nanos / 1_000_000),
                    updatedAt: group.updatedAt.seconds * 1000 + Int64(group.updatedAt.nanos / 1_000_000),
                    version: group.version
                )
            }
            
            return .success(groups)
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
}

// MARK: - DTOs

struct GroupDto: Sendable {
    let id: String
    let name: String
    let createdAt: Int64
    let updatedAt: Int64
    let version: Int64
}
