# Shopping List SaaS API Documentation

**Version**: v1  
**Protocol**: gRPC over HTTP/2  
**Package**: `shopping.v1`

## Authentication

All endpoints except Auth methods require a valid JWT access token.

### Headers

| Header | Description |
|--------|-------------|
| `authorization` | Bearer token: `Bearer <access_token>` |
| `x-request-id` | Optional request ID for tracing |
| `idempotency-key` | Optional key for write operations |

### Token Flow

1. Register or Login to receive `access_token` and `refresh_token`
2. Use `access_token` in `authorization` header
3. When `access_token` expires (15 min), use `RefreshToken` with `refresh_token`
4. `refresh_token` expires after 7 days

## Services

### AuthService

Authentication and user management.

| Method | Description | Auth Required |
|--------|-------------|---------------|
| Register | Create new user account | No |
| Login | Authenticate user | No |
| RefreshToken | Get new tokens | No |
| GetMe | Get current user profile | Yes |

#### Register

```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse)

message RegisterRequest {
  string email = 1;     // Required if phone not provided
  string phone = 2;     // Required if email not provided (E.164 format)
  string password = 3;  // Required: min 8 chars, 1 letter, 1 number
  string name = 4;      // Optional: max 100 chars
}

message RegisterResponse {
  string access_token = 1;
  string refresh_token = 2;
  User user = 3;
}
```

#### Login

```protobuf
rpc Login(LoginRequest) returns (LoginResponse)

message LoginRequest {
  string email = 1;     // Required if phone not provided
  string phone = 2;     // Required if email not provided
  string password = 3;  // Required
}
```

---

### GroupService

Multi-tenant group management.

| Method | Description | Auth Required |
|--------|-------------|---------------|
| CreateGroup | Create new group | Yes |
| ListMyGroups | List groups user belongs to | Yes |
| InviteMember | Invite user to group | Yes (Admin+) |
| AcceptInvite | Accept group invitation | Yes |
| ListMembers | List group members | Yes |
| UpdateMemberRole | Change member role | Yes (Admin+) |

#### Roles

| Role | Value | Permissions |
|------|-------|-------------|
| OWNER | 1 | All permissions, can delete group |
| ADMIN | 2 | Invite members, manage roles |
| MEMBER | 3 | Read/write lists and items |

---

### ListService

Shopping list and item management.

| Method | Description | Auth Required |
|--------|-------------|---------------|
| CreateList | Create new list in group | Yes |
| ListLists | Get lists in group (paginated) | Yes |
| UpdateList | Update list details | Yes |
| ArchiveList | Archive/unarchive list | Yes |
| AddItem | Add item to list | Yes |
| UpdateItem | Update item details | Yes |
| DeleteItem | Delete item | Yes |
| TogglePurchased | Mark item purchased/unpurchased | Yes |
| ReorderItems | Reorder items in list | Yes |

#### Item Priority

| Priority | Value | Sort Order |
|----------|-------|------------|
| LOW | 1 | 4 (last) |
| MEDIUM | 2 | 3 |
| HIGH | 3 | 2 |
| URGENT | 4 | 1 (first) |

#### Item Sorting

Items are sorted by:
1. `is_purchased = false` first
2. `priority` descending (URGENT → LOW)
3. `sort_order` ascending

---

### CategoryService

Category management within groups.

| Method | Description | Auth Required |
|--------|-------------|---------------|
| UpsertCategory | Create or update category | Yes |
| ListCategories | List categories in group | Yes |
| DeleteCategory | Delete category (soft delete) | Yes |

---

### SyncService

Offline synchronization support.

| Method | Description | Auth Required |
|--------|-------------|---------------|
| GetDelta | Get changes since cursor | Yes |
| PushMutations | Push offline changes | Yes |

#### Conflict Resolution

- All entities have a `version` field (BIGINT)
- Updates must include `expected_version`
- Version mismatch returns `FailedPrecondition` (code 9)
- Client must refresh entity and retry

---

## Error Codes

| Code | Name | Description |
|------|------|-------------|
| 0 | OK | Success |
| 3 | InvalidArgument | Validation error |
| 5 | NotFound | Resource not found |
| 6 | AlreadyExists | Duplicate resource |
| 7 | PermissionDenied | Not authorized |
| 9 | FailedPrecondition | Version conflict |
| 8 | ResourceExhausted | Rate limit exceeded |
| 16 | Unauthenticated | Invalid/missing token |

---

## Pagination

List operations support cursor-based pagination:

```protobuf
message PaginationRequest {
  int32 page_size = 1;   // 1-100, default 20
  string page_token = 2; // Cursor from previous response
}

message PaginationResponse {
  string next_page_token = 1; // Empty if no more pages
}
```

---

## Examples

### Register User

```bash
grpcurl -plaintext \
  -d '{"email":"user@example.com","password":"SecurePass123","name":"John Doe"}' \
  localhost:50051 shopping.v1.AuthService/Register
```

### Create Group

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"name":"Family Shopping","description":"Weekly groceries"}' \
  localhost:50051 shopping.v1.GroupService/CreateGroup
```

### Add Item

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"listId":"uuid","name":"Milk","priority":"ITEM_PRIORITY_HIGH","quantity":"2 gallons"}' \
  localhost:50051 shopping.v1.ListService/AddItem
```

### Toggle Purchased

```bash
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"itemId":"uuid","isPurchased":true,"expectedVersion":1}' \
  localhost:50051 shopping.v1.ListService/TogglePurchased
```

### Sync Changes

```bash
# Get changes since last sync
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"groupId":"uuid","cursor":0,"maxChanges":100}' \
  localhost:50051 shopping.v1.SyncService/GetDelta

# Push offline changes
grpcurl -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"groupId":"uuid","mutations":[...]}' \
  localhost:50051 shopping.v1.SyncService/PushMutations
```

---

## Rate Limits

| Endpoint Type | Rate | Burst |
|---------------|------|-------|
| Auth (Register, Login, Refresh) | 5/sec | 10 |
| Other endpoints | 100/sec | 200 |

Exceeding rate limits returns `ResourceExhausted` (code 8).
