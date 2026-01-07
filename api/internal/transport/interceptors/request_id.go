package interceptors

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const RequestIDKey = "x-request-id"

// RequestIDInterceptor generates or extracts request ID from metadata
func RequestIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	requestID := getOrGenerateRequestID(md)
	
	// Add request ID to context for logging
	ctx = context.WithValue(ctx, RequestIDKey, requestID)
	
	// Add request ID to outgoing metadata
	ctx = metadata.AppendToOutgoingContext(ctx, RequestIDKey, requestID)

	return handler(ctx, req)
}

func getOrGenerateRequestID(md metadata.MD) string {
	values := md.Get(RequestIDKey)
	if len(values) > 0 && values[0] != "" {
		return values[0]
	}

	// Generate new request ID
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
