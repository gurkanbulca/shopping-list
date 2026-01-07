# Shopping List SaaS MVP

Multi-tenant Shopping List SaaS with Go gRPC backend.

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Docker & Docker Compose (recommended)
- grpcurl (for testing)

### Option 1: Docker Compose (Recommended)

```bash
# Start PostgreSQL
docker compose up -d postgres

# Wait for database to be ready, then run migrations
cat scripts/run-migrations.sql | docker compose exec -T postgres psql -U postgres -d shopping_list_dev

# Build and run API
cd api && go run cmd/server/main.go
```

### Option 2: Local Development

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Start PostgreSQL:
   ```bash
   docker run -d --name shopping-list-db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=shopping_list_dev \
     -p 5432:5432 postgres:15-alpine
   ```

3. Run migrations:
   ```bash
   psql -h localhost -U postgres -d shopping_list_dev -f scripts/run-migrations.sql
   ```

4. Start the server:
   ```bash
   cd api
   go run cmd/server/main.go
   ```

## Testing User Stories

Run the automated test script (requires grpcurl):

```bash
# Install grpcurl if needed
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Run tests
./scripts/test-user-stories.sh
```

### Manual Testing with grpcurl

**User Story 1 - Authentication:**

```bash
# Register
grpcurl -plaintext -d '{"email":"test@example.com","password":"Test1234","name":"Test User"}' \
  localhost:50051 shopping.v1.AuthService/Register

# Login
grpcurl -plaintext -d '{"email":"test@example.com","password":"Test1234"}' \
  localhost:50051 shopping.v1.AuthService/Login

# Get Profile (replace TOKEN)
grpcurl -plaintext -H "authorization: Bearer TOKEN" \
  -d '{}' localhost:50051 shopping.v1.AuthService/GetMe
```

**User Story 2 - Groups:**

```bash
# Create Group
grpcurl -plaintext -H "authorization: Bearer TOKEN" \
  -d '{"name":"My Family","description":"Family shopping"}' \
  localhost:50051 shopping.v1.GroupService/CreateGroup

# List My Groups
grpcurl -plaintext -H "authorization: Bearer TOKEN" \
  -d '{}' localhost:50051 shopping.v1.GroupService/ListMyGroups
```

**User Story 3 - Lists & Items:**

```bash
# Create List
grpcurl -plaintext -H "authorization: Bearer TOKEN" \
  -d '{"groupId":"GROUP_ID","name":"Groceries"}' \
  localhost:50051 shopping.v1.ListService/CreateList

# Add Item
grpcurl -plaintext -H "authorization: Bearer TOKEN" \
  -d '{"listId":"LIST_ID","name":"Milk","priority":"ITEM_PRIORITY_HIGH","quantity":"2 gallons"}' \
  localhost:50051 shopping.v1.ListService/AddItem

# Toggle Purchased
grpcurl -plaintext -H "authorization: Bearer TOKEN" \
  -d '{"itemId":"ITEM_ID","isPurchased":true,"expectedVersion":1}' \
  localhost:50051 shopping.v1.ListService/TogglePurchased
```

## Project Structure

```
api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/          # Business logic
│   │   ├── auth/        # Authentication
│   │   ├── group/       # Group management
│   │   ├── list/        # Shopping lists
│   │   ├── category/    # Categories
│   │   └── sync/        # Offline sync
│   ├── transport/       # gRPC handlers
│   └── storage/         # Data access
├── pkg/                 # Shared packages
├── proto/               # Protobuf definitions
└── migrations/          # Database migrations
scripts/
├── test-user-stories.sh # Automated test script
└── run-migrations.sql   # Combined migrations
```

## API Services

| Service | Methods | Description |
|---------|---------|-------------|
| AuthService | Register, Login, RefreshToken, GetMe | User authentication |
| GroupService | CreateGroup, ListMyGroups, InviteMember, AcceptInvite, ListMembers, UpdateMemberRole | Multi-tenant groups |
| ListService | CreateList, ListLists, UpdateList, ArchiveList, AddItem, UpdateItem, DeleteItem, TogglePurchased, ReorderItems | Shopping lists & items |
| CategoryService | UpsertCategory, ListCategories, DeleteCategory | Item categorization |
| SyncService | GetDelta, PushMutations | Offline sync |

## MVP Features

- ✅ User registration & authentication (JWT)
- ✅ Multi-tenant groups with role-based access
- ✅ Shopping lists with CRUD operations
- ✅ Items with priorities, quantities, notes
- ✅ Version-based conflict detection
- ✅ Pagination for all list operations

## Development

See `specs/001-shopping-list-mvp/quickstart.md` for detailed development guide.

## License

MIT
