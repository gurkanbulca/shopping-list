// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "ShoppingListDependencies",
    platforms: [
        .iOS(.v18),
        .macOS(.v15)
    ],
    products: [
        .library(
            name: "ShoppingListDependencies",
            targets: ["ShoppingListDependencies"]
        ),
    ],
    dependencies: [
        // gRPC Swift client 2.x (requires iOS 18+)
        .package(url: "https://github.com/grpc/grpc-swift.git", from: "2.0.0"),
        // Swift Protobuf
        .package(url: "https://github.com/apple/swift-protobuf.git", from: "1.25.0"),
    ],
    targets: [
        .target(
            name: "ShoppingListDependencies",
            dependencies: [
                .product(name: "GRPCCore", package: "grpc-swift"),
                .product(name: "GRPCProtobuf", package: "grpc-swift"),
                .product(name: "GRPCNIOTransportHTTP2", package: "grpc-swift"),
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
            ],
            path: "Sources"
        ),
    ]
)
