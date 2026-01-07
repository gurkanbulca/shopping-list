#!/bin/bash
# Generate Swift protobuf and gRPC stubs from proto files
#
# Prerequisites:
# - protoc (Protocol Buffer compiler)
# - swift-protobuf plugin: brew install swift-protobuf
# - grpc-swift plugin: brew install protoc-gen-grpc-swift
#
# This project targets iOS 18+ and uses grpc-swift 2.x

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROTO_DIR="${PROJECT_ROOT}/api/proto"
OUTPUT_DIR="${PROJECT_ROOT}/iosApp/ShoppingListApp/Generated"

echo "Generating Swift protobuf stubs..."
echo "Proto source: ${PROTO_DIR}"
echo "Output: ${OUTPUT_DIR}"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Check for protoc
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc not found. Install with: brew install protobuf"
    exit 1
fi

# Check for swift-protobuf plugin
if ! command -v protoc-gen-swift &> /dev/null; then
    echo "Error: protoc-gen-swift not found. Install with: brew install swift-protobuf"
    exit 1
fi

# Find grpc-swift 2.x plugin
GRPC_SWIFT_PLUGIN=""

# Check common Homebrew paths first (with absolute paths for protoc)
BREW_PREFIX="${HOMEBREW_PREFIX:-}"
if [ -z "${BREW_PREFIX}" ]; then
    if [ -d "/home/linuxbrew/.linuxbrew" ]; then
        BREW_PREFIX="/home/linuxbrew/.linuxbrew"
    elif [ -d "/opt/homebrew" ]; then
        BREW_PREFIX="/opt/homebrew"
    elif [ -d "/usr/local" ]; then
        BREW_PREFIX="/usr/local"
    fi
fi

# Try to find the grpc-swift 2.x plugin
if [ -f "${BREW_PREFIX}/opt/protoc-gen-grpc-swift/bin/protoc-gen-grpc-swift-2" ]; then
    GRPC_SWIFT_PLUGIN="${BREW_PREFIX}/opt/protoc-gen-grpc-swift/bin/protoc-gen-grpc-swift-2"
elif [ -f "${BREW_PREFIX}/bin/protoc-gen-grpc-swift-2" ]; then
    GRPC_SWIFT_PLUGIN="${BREW_PREFIX}/bin/protoc-gen-grpc-swift-2"
elif command -v protoc-gen-grpc-swift-2 &> /dev/null; then
    GRPC_SWIFT_PLUGIN="$(which protoc-gen-grpc-swift-2)"
elif command -v protoc-gen-grpc-swift &> /dev/null; then
    GRPC_SWIFT_PLUGIN="$(which protoc-gen-grpc-swift)"
fi

if [ -z "${GRPC_SWIFT_PLUGIN}" ] || [ ! -x "${GRPC_SWIFT_PLUGIN}" ]; then
    echo "Error: protoc-gen-grpc-swift not found."
    echo ""
    echo "Install grpc-swift 2.x plugin with:"
    echo "  brew install protoc-gen-grpc-swift"
    exit 1
fi

echo "Using grpc-swift plugin: ${GRPC_SWIFT_PLUGIN}"

# Generate Swift protobuf files
# Note: google/protobuf/timestamp.proto is included with protobuf installation
protoc \
    --proto_path="${PROTO_DIR}" \
    --swift_out="${OUTPUT_DIR}" \
    --swift_opt=Visibility=Public \
    --plugin=protoc-gen-grpc-swift="${GRPC_SWIFT_PLUGIN}" \
    --grpc-swift_out="${OUTPUT_DIR}" \
    --grpc-swift_opt=Visibility=Public \
    --grpc-swift_opt=Client=true \
    --grpc-swift_opt=Server=false \
    "${PROTO_DIR}/shopping/v1/common.proto" \
    "${PROTO_DIR}/shopping/v1/auth.proto" \
    "${PROTO_DIR}/shopping/v1/group.proto" \
    "${PROTO_DIR}/shopping/v1/list.proto" \
    "${PROTO_DIR}/shopping/v1/category.proto" \
    "${PROTO_DIR}/shopping/v1/sync.proto"

echo ""
echo "Generated Swift files:"
ls -la "${OUTPUT_DIR}"

echo ""
echo "Done! Add the generated files to your Xcode project."
echo ""
echo "Next steps:"
echo "1. Open iosApp/ShoppingListApp.xcodeproj in Xcode"
echo "2. Add Swift Package dependencies (grpc-swift, swift-protobuf)"
echo "3. Add Generated/ folder to the project"
