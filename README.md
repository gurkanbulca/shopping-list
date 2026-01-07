# Shopping List SaaS MVP

Multi-tenant Shopping List SaaS with Go gRPC backend and iOS/Android mobile clients.

## Architecture

- **Backend**: Go 1.22+ with gRPC (protobuf)
- **Database**: PostgreSQL
- **Clients**: iOS + Android (offline-first, future)

## Setup

### Prerequisites

- Go 1.22 or later
- PostgreSQL 14 or later
- buf CLI (for protobuf code generation)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and configure
3. Setup database:
   ```bash
   createdb shopping_list_dev
   ```
4. Run migrations:
   ```bash
   cd api
   goose -dir migrations postgres "postgres://user:pass@localhost/shopping_list_dev?sslmode=disable" up
   ```
5. Generate protobuf code:
   ```bash
   cd api/proto
   buf generate
   ```
6. Install dependencies:
   ```bash
   cd api
   go mod tidy
   ```
7. Run server:
   ```bash
   cd api/cmd/server
   go run main.go
   ```

## Project Structure

```
api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/         # Business logic
│   ├── transport/      # gRPC handlers & interceptors
│   ├── storage/        # Repository interfaces & implementations
│   └── config/        # Configuration management
├── pkg/                # Shared packages
├── migrations/         # Database migrations
└── proto/              # Protobuf definitions
```

## Development

See `specs/001-shopping-list-mvp/quickstart.md` for detailed development guide.

## License

[Add license information]
