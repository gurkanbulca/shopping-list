package interceptors

import (
	"context"
	"strings"

	"github.com/gurkanbulca/shopping-list/api/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	AuthorizationKey = "authorization"
	UserIDKey        = "user_id"
)

// contextKey is a custom type to avoid context key collisions
type contextKey string

const userIDContextKey contextKey = "user_id"

// GetUserIDFromContext extracts the user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

// AuthInterceptorWithTokenManager extracts and validates JWT token from metadata
func AuthInterceptorWithTokenManager(tm *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for public endpoints
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get(AuthorizationKey)
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := extractBearerToken(authHeaders[0])
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// Validate token and extract claims
		claims, err := tm.ValidateToken(token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				return nil, status.Error(codes.Unauthenticated, "token expired")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add user ID to context
		ctx = context.WithValue(ctx, userIDContextKey, claims.UserID)

		return handler(ctx, req)
	}
}

// AuthInterceptor is a placeholder for backwards compatibility
// Use AuthInterceptorWithTokenManager instead
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for public endpoints
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get(AuthorizationKey)
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := extractBearerToken(authHeaders[0])
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		return handler(ctx, req)
	}
}

func isPublicEndpoint(method string) bool {
	publicMethods := []string{
		"/shopping.v1.AuthService/Register",
		"/shopping.v1.AuthService/Login",
		"/shopping.v1.AuthService/RefreshToken",
	}

	for _, public := range publicMethods {
		if method == public {
			return true
		}
	}
	return false
}

func extractBearerToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
