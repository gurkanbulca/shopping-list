package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationInterceptor validates group membership for group-scoped operations
// This is a placeholder - actual implementation will check group membership
func AuthorizationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: Extract group_id from request
		// TODO: Extract user_id from context (set by auth interceptor)
		// TODO: Check if user is member of group
		// TODO: Return PermissionDenied if not member

		// For now, pass through - authorization will be checked in handlers
		return handler(ctx, req)
	}
}

// CheckGroupMembership is a helper function to be used in handlers
func CheckGroupMembership(ctx context.Context, groupID, userID string) error {
	// TODO: Implement actual membership check
	// This will query the database to verify user is an active member of the group
	return nil
}

// RequireGroupMember is a helper that returns PermissionDenied if not member
func RequireGroupMember(ctx context.Context, groupID, userID string) error {
	if err := CheckGroupMembership(ctx, groupID, userID); err != nil {
		return status.Error(codes.PermissionDenied, "user is not a member of this group")
	}
	return nil
}
