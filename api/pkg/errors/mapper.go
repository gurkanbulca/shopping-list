package errors

import (
	"errors"

	"github.com/gurkanbulca/shopping-list/api/internal/domain/auth"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/group"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/list"
	"github.com/gurkanbulca/shopping-list/api/pkg/validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError converts domain errors to gRPC status errors
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// Auth domain errors
	switch {
	case errors.Is(err, auth.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, auth.ErrDuplicateEmail):
		return status.Error(codes.AlreadyExists, "email already exists")

	case errors.Is(err, auth.ErrDuplicatePhone):
		return status.Error(codes.AlreadyExists, "phone number already exists")

	case errors.Is(err, auth.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")

	case errors.Is(err, auth.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")

	case errors.Is(err, auth.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, "token expired")

	case errors.Is(err, auth.ErrMissingIdentifier):
		return status.Error(codes.InvalidArgument, "email or phone is required")

	case errors.Is(err, auth.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Group domain errors
	switch {
	case errors.Is(err, group.ErrGroupNotFound):
		return status.Error(codes.NotFound, "group not found")

	case errors.Is(err, group.ErrMemberNotFound):
		return status.Error(codes.NotFound, "member not found")

	case errors.Is(err, group.ErrNotGroupMember):
		return status.Error(codes.PermissionDenied, "user is not a member of this group")

	case errors.Is(err, group.ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "not authorized to perform this action")

	case errors.Is(err, group.ErrAlreadyMember):
		return status.Error(codes.AlreadyExists, "user is already a member of this group")

	case errors.Is(err, group.ErrAlreadyInvited):
		return status.Error(codes.AlreadyExists, "user already has a pending invitation")

	case errors.Is(err, group.ErrNoInvitation):
		return status.Error(codes.NotFound, "no pending invitation found")

	case errors.Is(err, group.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, "invalid role")

	case errors.Is(err, group.ErrCannotChangeOwner):
		return status.Error(codes.FailedPrecondition, "cannot change owner's role")

	case errors.Is(err, group.ErrGroupMustHaveOwner):
		return status.Error(codes.FailedPrecondition, "group must have at least one owner")

	case errors.Is(err, group.ErrCannotInviteAsOwner):
		return status.Error(codes.InvalidArgument, "cannot invite as owner")

	case errors.Is(err, group.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, group.ErrInvalidGroupName):
		return status.Error(codes.InvalidArgument, "group name must be 1-100 characters")
	}

	// List domain errors
	switch {
	case errors.Is(err, list.ErrListNotFound):
		return status.Error(codes.NotFound, "list not found")

	case errors.Is(err, list.ErrItemNotFound):
		return status.Error(codes.NotFound, "item not found")

	case errors.Is(err, list.ErrVersionMismatch):
		return status.Error(codes.FailedPrecondition, "version mismatch - item was modified by another user")

	case errors.Is(err, list.ErrNotGroupMember):
		return status.Error(codes.PermissionDenied, "user is not a member of this group")

	case errors.Is(err, list.ErrInvalidListName):
		return status.Error(codes.InvalidArgument, "list name must be 1-100 characters")

	case errors.Is(err, list.ErrInvalidItemName):
		return status.Error(codes.InvalidArgument, "item name must be 1-200 characters")

	case errors.Is(err, list.ErrInvalidPriority):
		return status.Error(codes.InvalidArgument, "invalid priority")

	case errors.Is(err, list.ErrListArchived):
		return status.Error(codes.FailedPrecondition, "cannot modify archived list")
	}

	// Validation errors
	var validationErr *validation.AuthValidationError
	if errors.As(err, &validationErr) {
		return status.Error(codes.InvalidArgument, validationErr.Error())
	}

	// Default to internal error
	return status.Error(codes.Internal, "internal server error")
}

// FromGRPCError extracts error information from a gRPC status error
func FromGRPCError(err error) (codes.Code, string) {
	if err == nil {
		return codes.OK, ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return codes.Unknown, err.Error()
	}

	return st.Code(), st.Message()
}
