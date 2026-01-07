# ShoppingList iOS App

iOS client for the Shopping List app using SwiftUI and gRPC.

## Prerequisites

- Xcode 15+
- iOS 18.0+ deployment target
- macOS 15+ (for building)

## Setup

### 1. Install gRPC Swift Tools (for proto generation)

```bash
# Install swift-protobuf and grpc-swift plugins
brew install swift-protobuf protoc-gen-grpc-swift protobuf
```

### 2. Generate Swift Protobuf Stubs

```bash
cd /path/to/shopping-list
./scripts/generate-swift-protos.sh
```

This generates Swift files in `iosApp/ShoppingListApp/Generated/`:
- `shopping/v1/*.pb.swift` - Protobuf message types
- `shopping/v1/*.grpc.swift` - gRPC service clients

### 3. Add Swift Package Dependencies

Open the Xcode project and add the following Swift Package dependencies:

**File → Add Package Dependencies...**

1. **grpc-swift**: `https://github.com/grpc/grpc-swift.git` (version 2.0.0+)
2. **swift-protobuf**: `https://github.com/apple/swift-protobuf.git` (version 1.25.0+)

### 4. Add Generated Files to Project

Drag the `Generated/` folder into your Xcode project:
1. Right-click on `ShoppingListApp` folder
2. Select "Add Files to ShoppingListApp..."
3. Select the `Generated` folder
4. Ensure "Copy items if needed" is unchecked
5. Ensure "Create folder references" is selected

## Project Structure

```
iosApp/
├── ShoppingListApp/
│   ├── ShoppingListApp.swift       # App entry point
│   ├── ContentView.swift           # Main content view
│   ├── GrpcClient.swift            # gRPC channel management
│   ├── GrpcAdapters.swift          # Service container
│   ├── AuthRemoteDataSourceIos.swift
│   ├── GroupRemoteDataSourceIos.swift
│   ├── ListRemoteDataSourceIos.swift
│   ├── SyncRemoteDataSourceIos.swift
│   ├── IosSecureTokenStore.swift   # Keychain token storage
│   └── Generated/                  # Proto-generated files
│       └── shopping/v1/
│           ├── *.pb.swift
│           └── *.grpc.swift
├── ShoppingListApp.xcodeproj/
├── Package.swift                   # SPM dependencies
└── README.md
```

## Configuration

### Backend Endpoint

Configure the backend endpoint in `GrpcClient.swift`:

```swift
// For iOS Simulator
let config = GrpcClientConfig(
    host: "localhost",
    port: 50051,
    useTLS: false,
    defaultTimeoutSeconds: 10
)

// For physical device (replace with your server IP/hostname)
let config = GrpcClientConfig(
    host: "192.168.1.100",
    port: 50051,
    useTLS: true,
    defaultTimeoutSeconds: 10
)
```

### Initialize gRPC Services

In your app startup (e.g., `ShoppingListApp.swift`):

```swift
@main
struct ShoppingListApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

Use the shared `GrpcServices` instance:

```swift
let authService = GrpcServices.shared.auth
let groupService = GrpcServices.shared.groups
let listService = GrpcServices.shared.lists
let syncService = GrpcServices.shared.sync
```

## Development

### Regenerating Proto Stubs

When proto files change:

```bash
./scripts/generate-swift-protos.sh
```

### Testing gRPC Connection

```swift
// Quick connection test
Task {
    let result = await GrpcServices.shared.auth.login(
        email: "test@example.com",
        password: "password123"
    )
    switch result {
    case .success(let tokens):
        print("Login successful: \(tokens.accessToken.prefix(20))...")
    case .failure(let error):
        print("Login failed: \(error.message)")
    }
}
```

## Troubleshooting

### "Module 'GRPCCore' not found"

Ensure Swift Package dependencies are resolved:
- File → Packages → Resolve Package Versions

### Connection Refused

1. Check if the backend server is running
2. For iOS Simulator, use `localhost`
3. For physical devices, ensure:
   - Same network as the server
   - Correct IP address configured
   - Port not blocked by firewall

### TLS Certificate Errors

For development with self-signed certificates:
```swift
// Use plaintext for development
let config = GrpcClientConfig(useTLS: false, ...)
```

For production, configure proper TLS certificates.
