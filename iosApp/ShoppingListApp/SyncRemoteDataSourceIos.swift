import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import GRPCProtobuf
import SwiftProtobuf

/// iOS implementation of SyncRemoteDataSource using gRPC.
/// Communicates with the SyncService for delta sync and push mutations.
@available(iOS 18.0, macOS 15.0, *)
class SyncRemoteDataSourceIos {
    
    private let clientManager: GrpcClientManager
    
    init(clientManager: GrpcClientManager) {
        self.clientManager = clientManager
    }
    
    /// Get delta changes since the given cursor.
    func getDelta(groupId: String, cursor: Int64, pageSize: Int32) async -> GrpcResult<GetDeltaResponseDto> {
        do {
            let grpcClient = try await clientManager.getClient()
            let syncClient = Shopping_V1_SyncService.Client(wrapping: grpcClient)
            
            var message = Shopping_V1_GetDeltaRequest()
            message.groupID = groupId
            message.cursor = cursor
            message.maxChanges = pageSize
            
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions()
            
            let response: Shopping_V1_GetDeltaResponse = try await syncClient.getDelta(
                request: request,
                options: options
            )
            
            let entities = response.changes.map { change -> DeltaEntityDto in
                let entityType: String
                switch change.entityType {
                case .list: entityType = "list"
                case .item: entityType = "item"
                case .category: entityType = "category"
                case .group: entityType = "group"
                case .groupMember: entityType = "group_member"
                default: entityType = "unknown"
                }
                
                let data: Data
                switch change.entityData {
                case .list(let list):
                    data = (try? list.jsonUTF8Data()) ?? Data()
                case .item(let item):
                    data = (try? item.jsonUTF8Data()) ?? Data()
                case .category(let category):
                    data = (try? category.jsonUTF8Data()) ?? Data()
                case .group(let group):
                    data = (try? group.jsonUTF8Data()) ?? Data()
                case .groupMember(let member):
                    data = (try? member.jsonUTF8Data()) ?? Data()
                case .none:
                    data = Data()
                }
                
                return DeltaEntityDto(
                    entityType: entityType,
                    entityId: change.entityID,
                    data: data,
                    version: getVersionFromChange(change),
                    isDeleted: change.operation == .delete
                )
            }
            
            return .success(GetDeltaResponseDto(
                entities: entities,
                nextCursor: response.nextCursor,
                hasMore: response.hasMore
            ))
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
    
    /// Push a batch of mutations to the server.
    func pushMutations(groupId: String, mutations: [MutationDto]) async -> GrpcResult<PushMutationsResponseDto> {
        do {
            let grpcClient = try await clientManager.getClient()
            let syncClient = Shopping_V1_SyncService.Client(wrapping: grpcClient)
            
            var message = Shopping_V1_PushMutationsRequest()
            message.groupID = groupId
            
            message.mutations = mutations.map { mutation in
                var protoMutation = Shopping_V1_Mutation()
                protoMutation.mutationID = mutation.idempotencyKey
                protoMutation.entityType = mutation.entityType
                protoMutation.entityID = mutation.entityId
                protoMutation.entityData = mutation.entityData
                
                switch mutation.mutationType {
                case "CREATE":
                    protoMutation.type = .create
                case "UPDATE":
                    protoMutation.type = .update
                case "DELETE":
                    protoMutation.type = .delete
                default:
                    protoMutation.type = .unspecified
                }
                
                if let expectedVersion = mutation.expectedVersion {
                    protoMutation.expectedVersion = expectedVersion
                }
                
                return protoMutation
            }
            
            let idempotencyKey = mutations.first?.idempotencyKey
            let request = ClientRequest(message: message)
            let options = await clientManager.callOptions(idempotencyKey: idempotencyKey)
            
            let response: Shopping_V1_PushMutationsResponse = try await syncClient.pushMutations(
                request: request,
                options: options
            )
            
            let results = response.results.map { result in
                MutationResultDto(
                    idempotencyKey: result.mutationID,
                    success: result.success,
                    errorCode: result.errorCode.isEmpty ? nil : result.errorCode,
                    errorMessage: result.errorMessage.isEmpty ? nil : result.errorMessage,
                    newVersion: result.version > 0 ? result.version : nil
                )
            }
            
            return .success(PushMutationsResponseDto(results: results))
        } catch {
            return .failure(GrpcStatusMapper.mapError(error))
        }
    }
    
    // MARK: - Helpers
    
    private func getVersionFromChange(_ change: Shopping_V1_Change) -> Int64 {
        switch change.entityData {
        case .list(let list): return list.version
        case .item(let item): return item.version
        case .category(let category): return category.version
        case .group(let group): return group.version
        case .groupMember(let member): return member.version
        case .none: return 0
        }
    }
}

// MARK: - DTOs

struct GetDeltaResponseDto: Sendable {
    let entities: [DeltaEntityDto]
    let nextCursor: Int64
    let hasMore: Bool
}

struct DeltaEntityDto: Sendable {
    let entityType: String
    let entityId: String
    let data: Data
    let version: Int64
    let isDeleted: Bool
}

struct MutationDto: Sendable {
    let idempotencyKey: String
    let entityType: String
    let entityId: String
    let mutationType: String
    let entityData: Data
    let expectedVersion: Int64?
}

struct PushMutationsResponseDto: Sendable {
    let results: [MutationResultDto]
}

struct MutationResultDto: Sendable {
    let idempotencyKey: String
    let success: Bool
    let errorCode: String?
    let errorMessage: String?
    let newVersion: Int64?
}
