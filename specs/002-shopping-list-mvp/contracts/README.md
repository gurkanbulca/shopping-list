# Contracts: Mobile Offline MVP

**Status**: Reference  
**Feature**: Mobile Offline MVP (Android + iOS)

## Overview

This feature uses the canonical gRPC contracts defined in `api/proto/shopping/v1/` as the single source of truth for client-server communication.

## Contract Location

All protobuf definitions are located at:

```
api/proto/shopping/v1/
├── auth.proto         # Authentication service
├── group.proto        # Group management
├── list.proto         # List operations
├── item.proto         # Item operations
├── category.proto     # Category management
├── sync.proto         # Sync service (GetDelta, PushMutations)
└── common.proto       # Shared types (pagination, etc.)
```

## Key Services

### AuthService

- `Login(LoginRequest) returns (LoginResponse)` - Authenticate user
- `RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse)` - Refresh access token
- `Logout(LogoutRequest) returns (LogoutResponse)` - Invalidate session

### GroupService

- `ListGroups(ListGroupsRequest) returns (ListGroupsResponse)` - Get user's groups

### ListService

- `ListLists(ListListsRequest) returns (ListListsResponse)` - Get lists in a group

### ItemService

- `ListItems(ListItemsRequest) returns (ListItemsResponse)` - Get items in a list

### SyncService

- `GetDelta(GetDeltaRequest) returns (GetDeltaResponse)` - Pull changes since cursor
- `PushMutations(PushMutationsRequest) returns (PushMutationsResponse)` - Push local mutations

## Key Types

### GetDeltaRequest

```protobuf
message GetDeltaRequest {
  string group_id = 1;
  int64 cursor = 2;      // 0 for initial sync
  int32 page_size = 3;   // max entities per response
}
```

### GetDeltaResponse

```protobuf
message GetDeltaResponse {
  repeated DeltaEntity entities = 1;
  int64 next_cursor = 2;
  bool has_more = 3;
}
```

### PushMutationsRequest

```protobuf
message PushMutationsRequest {
  string group_id = 1;
  repeated Mutation mutations = 2;
}
```

### Mutation

```protobuf
message Mutation {
  string idempotency_key = 1;
  string entity_type = 2;      // "list", "item", "category"
  string entity_id = 3;
  string mutation_type = 4;    // "CREATE", "UPDATE", "DELETE"
  bytes entity_data = 5;       // JSON-encoded payload
  int64 expected_version = 6;  // For conflict detection
}
```

## Client Integration

### Android

Generate Kotlin stubs using `protoc` with `grpc-kotlin` plugin.

### iOS

Generate Swift stubs using `grpc-swift` plugin or integrate via KMP shared module.

## Notes

- See `api/proto/shopping/v1/` for complete protobuf definitions
- All RPCs require `authorization` header with bearer token
- Mutation RPCs require `idempotency-key` header
- All RPCs should include `x-request-id` for observability
