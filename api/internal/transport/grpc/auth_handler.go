package grpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/auth"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/interceptors"
	pkgerrors "github.com/gurkanbulca/shopping-list/api/pkg/errors"
	"github.com/gurkanbulca/shopping-list/api/pkg/validation"
	shoppingv1 "github.com/gurkanbulca/shopping-list/api/proto/shopping/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthHandler implements the AuthServiceServer interface
type AuthHandler struct {
	shoppingv1.UnimplementedAuthServiceServer
	authService *auth.Service
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(ctx context.Context, req *shoppingv1.RegisterRequest) (*shoppingv1.RegisterResponse, error) {
	// Validate request
	if err := validation.ValidateRegisterRequest(req.Email, req.Phone, req.Password, req.Name); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Prepare input
	input := auth.RegisterInput{
		Password: req.Password,
	}

	email := strings.TrimSpace(req.Email)
	if email != "" {
		input.Email = &email
	}

	phone := strings.TrimSpace(req.Phone)
	if phone != "" {
		input.Phone = &phone
	}

	name := strings.TrimSpace(req.Name)
	if name != "" {
		input.Name = &name
	}

	// Call service
	result, err := h.authService.Register(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.RegisterResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toProtoUser(result.User),
	}, nil
}

// Login handles user login
func (h *AuthHandler) Login(ctx context.Context, req *shoppingv1.LoginRequest) (*shoppingv1.LoginResponse, error) {
	// Validate request
	if err := validation.ValidateLoginRequest(req.Email, req.Phone, req.Password); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Prepare input
	input := auth.LoginInput{
		Password: req.Password,
	}

	email := strings.TrimSpace(req.Email)
	if email != "" {
		input.Email = &email
	}

	phone := strings.TrimSpace(req.Phone)
	if phone != "" {
		input.Phone = &phone
	}

	// Call service
	result, err := h.authService.Login(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         toProtoUser(result.User),
	}, nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, req *shoppingv1.RefreshTokenRequest) (*shoppingv1.RefreshTokenResponse, error) {
	// Validate request
	if err := validation.ValidateRefreshTokenRequest(req.RefreshToken); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Call service
	result, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

// GetMe handles getting current user profile
func (h *AuthHandler) GetMe(ctx context.Context, req *shoppingv1.GetMeRequest) (*shoppingv1.GetMeResponse, error) {
	// Get user ID from context (set by auth interceptor)
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Call service
	user, err := h.authService.GetMe(ctx, userID)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.GetMeResponse{
		User: toProtoUser(user),
	}, nil
}

// toProtoUser converts a domain User to a proto User
func toProtoUser(user *auth.User) *shoppingv1.User {
	protoUser := &shoppingv1.User{
		Id:        user.ID.String(),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	if user.Email != nil {
		protoUser.Email = *user.Email
	}

	if user.Phone != nil {
		protoUser.Phone = *user.Phone
	}

	if user.Name != nil {
		protoUser.Name = *user.Name
	}

	return protoUser
}
