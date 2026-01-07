# Quickstart Guide: Shopping List SaaS MVP

**Created**: 2026-01-07  
**Feature**: Shopping List SaaS MVP

## Overview

This guide helps developers get started with the Shopping List SaaS MVP implementation. It covers setup, key concepts, and basic usage patterns.

## Prerequisites

- Go 1.22 or later
- PostgreSQL 14 or later
- Docker (for containerized development)
- buf CLI (for protobuf management)
- Git

## Project Structure

```
api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/         # Business logic
│   ├── transport/      # gRPC handlers & interceptors
│   └── storage/        # Repository interfaces & implementations
├── proto/              # Protobuf definitions
├── migrations/         # Database migrations
└── tests/              # Test suites

specs/001-shopping-list-mvp/
├── contracts/         # Proto files (reference)
├── data-model.md      # Database schema
└── plan.md            # Implementation plan
```

## Setup Steps

### 1. Database Setup

```bash
# Create PostgreSQL database
createdb shopping_list_dev

# Run migrations
cd api
goose -dir migrations postgres "postgres://user:pass@localhost/shopping_list_dev?sslmode=disable" up
```

### 2. Generate Protobuf Code

```bash
# Install buf (if not already installed)
go install github.com/bufbuild/buf/cmd/buf@latest

# Generate Go code from protos
cd api/proto
buf generate
```

### 3. Environment Configuration

Create `.env` file:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=shopping_list_dev
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Server
GRPC_PORT=50051
HTTP_PORT=8080

# Observability
OTEL_ENDPOINT=http://localhost:4317
LOG_LEVEL=info
```

### 4. Run Server

```bash
cd api/cmd/server
go run main.go
```

Server starts on `localhost:50051` (gRPC).

## Key Concepts

### Multi-Tenancy

- **Tenant Boundary**: Groups serve as tenant boundaries
- **Membership**: Users must be members of a group to access its resources
- **Authorization**: Every request validates group membership via interceptor

### Authentication Flow

1. User registers/logs in → receives `access_token` and `refresh_token`
2. Client includes `access_token` in `authorization: Bearer <token>` header
3. Access token expires → client uses `refresh_token` to get new `access_token`
4. All protected RPCs require valid access token

### Offline Sync Flow

1. **Client offline**: Changes queued locally in SQLite
2. **Client online**: 
   - Push mutations via `PushMutations` (batch)
   - Pull changes via `GetDelta` (cursor-based)
3. **Conflict resolution**: Version mismatch → `FailedPrecondition` → client refreshes and retries

### Conflict Resolution

- Every entity has a `version` (BIGINT)
- Update requests include `expected_version`
- Server checks: if current version != expected_version → return `FailedPrecondition`
- Client must refresh entity and retry with new version

## API Usage Examples

### 1. Register and Login

```go
// Register
registerReq := &shoppingv1.RegisterRequest{
    Email:    "user@example.com",
    Password: "securepassword123",
    Name:     "John Doe",
}
registerResp, err := authClient.Register(ctx, registerReq)

// Login
loginReq := &shoppingv1.LoginRequest{
    Email:    "user@example.com",
    Password: "securepassword123",
}
loginResp, err := authClient.Login(ctx, loginReq)

// Store tokens securely
accessToken := loginResp.AccessToken
refreshToken := loginResp.RefreshToken
```

### 2. Create Group and Invite Member

```go
// Create group (caller becomes owner)
createGroupReq := &shoppingv1.CreateGroupRequest{
    Name:        "Family Shopping",
    Description: "Weekly family shopping list",
}
groupResp, err := groupClient.CreateGroup(ctx, createGroupReq)
groupID := groupResp.Group.Id

// Invite member
inviteReq := &shoppingv1.InviteMemberRequest{
    GroupId: groupID,
    Email:   "member@example.com",
    Role:    shoppingv1.MemberRole_MEMBER_ROLE_MEMBER,
}
inviteResp, err := groupClient.InviteMember(ctx, inviteReq)
```

### 3. Create List and Add Items

```go
// Create list
createListReq := &shoppingv1.CreateListRequest{
    GroupId: groupID,
    Name:    "Weekly Groceries",
}
listResp, err := listClient.CreateList(ctx, createListReq)
listID := listResp.List.Id

// Add items
addItemReq := &shoppingv1.AddItemRequest{
    ListId:   listID,
    Name:     "Milk",
    Priority: shoppingv1.ItemPriority_ITEM_PRIORITY_HIGH,
    Quantity: "2L",
}
itemResp, err := listClient.AddItem(ctx, addItemReq)
```

### 4. Toggle Purchased Status

```go
// Mark as purchased
toggleReq := &shoppingv1.TogglePurchasedRequest{
    ItemId:          itemResp.Item.Id,
    IsPurchased:     true,
    ExpectedVersion: itemResp.Item.Version,
}
toggleResp, err := listClient.TogglePurchased(ctx, toggleReq)
```

### 5. Sync Offline Changes

```go
// Push local mutations
mutations := []*shoppingv1.Mutation{
    {
        MutationId:  "client-uuid-1",
        Type:        shoppingv1.MutationType_MUTATION_TYPE_CREATE,
        EntityType:   "item",
        EntityData:   itemJSONBytes,
    },
}
pushReq := &shoppingv1.PushMutationsRequest{
    GroupId:   groupID,
    Mutations: mutations,
}
pushResp, err := syncClient.PushMutations(ctx, pushReq)

// Get delta changes
deltaReq := &shoppingv1.GetDeltaRequest{
    GroupId:    groupID,
    Cursor:     0, // Last received sequence
    MaxChanges: 100,
}
deltaResp, err := syncClient.GetDelta(ctx, deltaReq)
```

## Testing

### Unit Tests

```bash
cd api
go test ./internal/domain/...
```

### Integration Tests

```bash
# Start test database
docker run -d -p 5433:5432 -e POSTGRES_PASSWORD=test postgres:14

# Run integration tests
go test ./tests/integration/... -tags=integration
```

### Contract Tests

```bash
# Validate proto contracts
cd api/proto
buf lint
buf breaking --against '.git#branch=main'
```

## Development Workflow

### 1. Update Protobuf Contracts

1. Edit proto files in `api/proto/shopping/v1/`
2. Run `buf generate` to regenerate Go code
3. Update handlers and domain logic
4. Update tests

### 2. Database Migrations

```bash
# Create new migration
goose -dir migrations create add_user_preferences sql

# Edit migration file
# Run migration
goose -dir migrations postgres "connection_string" up
```

### 3. Adding New RPC

1. Define RPC in proto file
2. Generate code: `buf generate`
3. Implement handler in `internal/transport/grpc/`
4. Implement domain logic in `internal/domain/`
5. Implement storage in `internal/storage/`
6. Add integration test
7. Update documentation

## Common Patterns

### Authorization Check

```go
// In interceptor or handler
func (h *Handler) checkMembership(ctx context.Context, groupID, userID string) error {
    member, err := h.repo.GetGroupMember(ctx, groupID, userID)
    if err != nil {
        return status.Error(codes.PermissionDenied, "not a group member")
    }
    if member.Status != shoppingv1.MemberStatus_MEMBER_STATUS_ACTIVE {
        return status.Error(codes.PermissionDenied, "membership not active")
    }
    return nil
}
```

### Version Conflict Check

```go
func (h *Handler) updateItem(ctx context.Context, req *shoppingv1.UpdateItemRequest) error {
    item, err := h.repo.GetItem(ctx, req.ItemId)
    if err != nil {
        return err
    }
    
    // Check version
    if item.Version != req.ExpectedVersion {
        return status.Error(codes.FailedPrecondition, "version mismatch")
    }
    
    // Update with incremented version
    item.Version++
    // ... apply updates
    return h.repo.UpdateItem(ctx, item)
}
```

### Pagination

```go
func (h *Handler) listItems(ctx context.Context, req *shoppingv1.ListItemsRequest) (*shoppingv1.ListItemsResponse, error) {
    pageSize := int(req.Pagination.PageSize)
    if pageSize == 0 {
        pageSize = 20 // default
    }
    if pageSize > 100 {
        pageSize = 100 // max
    }
    
    items, nextToken, err := h.repo.ListItems(ctx, req.ListId, pageSize, req.Pagination.PageToken)
    if err != nil {
        return nil, err
    }
    
    return &shoppingv1.ListItemsResponse{
        Items: items,
        Pagination: &shoppingv1.PaginationResponse{
            NextPageToken: nextToken,
        },
    }, nil
}
```

## Troubleshooting

### Common Issues

1. **"PermissionDenied" errors**: Check group membership and status
2. **"FailedPrecondition" errors**: Version mismatch - refresh entity and retry
3. **Connection errors**: Verify database connection and gRPC server is running
4. **Proto generation errors**: Ensure buf is installed and proto files are valid

### Debugging

- Enable debug logging: `LOG_LEVEL=debug`
- Check request IDs in logs for tracing
- Use gRPC reflection (dev only) for introspection
- Monitor OpenTelemetry traces

## Next Steps

1. Review [data-model.md](./data-model.md) for database schema details
2. Review [contracts/](./contracts/) for complete API reference
3. Review [plan.md](./plan.md) for implementation details
4. Start implementing handlers and domain logic

## Resources

- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Protobuf Guide](https://developers.google.com/protocol-buffers/docs/gotutorial)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [buf Documentation](https://docs.buf.build/)
