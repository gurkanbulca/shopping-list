package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const IdempotencyKey = "idempotency-key"

// IdempotencyInterceptor handles idempotency for write operations
// Note: Full implementation requires a cache/store (Redis recommended for production)
func IdempotencyInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only apply to write operations
		if !isWriteOperation(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		keys := md.Get(IdempotencyKey)
		if len(keys) == 0 || keys[0] == "" {
			return handler(ctx, req)
		}

		idempotencyKey := keys[0]
		
		// TODO: Check cache/store for existing response
		// If found, return cached response
		// If not found, execute handler and cache response with TTL

		// For MVP, we'll just pass through - idempotency will be handled at handler level
		ctx = context.WithValue(ctx, IdempotencyKey, idempotencyKey)

		return handler(ctx, req)
	}
}

func isWriteOperation(method string) bool {
	writeMethods := []string{
		"Create", "Update", "Delete", "Archive", "Add", "Toggle", "Reorder",
		"Invite", "Accept", "Upsert", "Push",
	}
	
	methodLower := strings.ToLower(method)
	for _, write := range writeMethods {
		if strings.Contains(methodLower, strings.ToLower(write)) {
			return true
		}
	}
	return false
}
