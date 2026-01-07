package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	pkgauth "github.com/gurkanbulca/shopping-list/api/pkg/auth"
	"go.uber.org/zap"
)

// Service handles authentication business logic
type Service struct {
	userRepo     UserRepository
	tokenManager *pkgauth.TokenManager
	logger       *zap.Logger
}

// NewService creates a new auth service
func NewService(userRepo UserRepository, tokenManager *pkgauth.TokenManager, logger *zap.Logger) *Service {
	return &Service{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		logger:       logger.Named("auth_service"),
	}
}

// RegisterInput contains the data needed to register a new user
type RegisterInput struct {
	Email    *string
	Phone    *string
	Password string
	Name     *string
}

// AuthResult contains the result of a successful authentication
type AuthResult struct {
	User         *User
	AccessToken  string
	RefreshToken string
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	s.logger.Info("registering new user",
		zap.Stringp("email", input.Email),
		zap.Stringp("phone", input.Phone),
	)

	// Validate that at least one identifier is provided
	if (input.Email == nil || *input.Email == "") && (input.Phone == nil || *input.Phone == "") {
		return nil, ErrMissingIdentifier
	}

	// Check for duplicate email
	if input.Email != nil && *input.Email != "" {
		exists, err := s.userRepo.ExistsByEmail(ctx, *input.Email)
		if err != nil {
			s.logger.Error("failed to check email existence", zap.Error(err))
			return nil, err
		}
		if exists {
			return nil, ErrDuplicateEmail
		}
	}

	// Check for duplicate phone
	if input.Phone != nil && *input.Phone != "" {
		exists, err := s.userRepo.ExistsByPhone(ctx, *input.Phone)
		if err != nil {
			s.logger.Error("failed to check phone existence", zap.Error(err))
			return nil, err
		}
		if exists {
			return nil, ErrDuplicatePhone
		}
	}

	// Hash password
	passwordHash, err := pkgauth.HashPassword(input.Password)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user
	user := NewUser(input.Email, input.Phone, passwordHash, input.Name)

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID.String())
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	s.logger.Info("user registered successfully", zap.String("user_id", user.ID.String()))

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// LoginInput contains the data needed to login
type LoginInput struct {
	Email    *string
	Phone    *string
	Password string
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	s.logger.Info("user login attempt",
		zap.Stringp("email", input.Email),
		zap.Stringp("phone", input.Phone),
	)

	var user *User
	var err error

	// Find user by email or phone
	if input.Email != nil && *input.Email != "" {
		user, err = s.userRepo.GetByEmail(ctx, *input.Email)
	} else if input.Phone != nil && *input.Phone != "" {
		user, err = s.userRepo.GetByPhone(ctx, *input.Phone)
	} else {
		return nil, ErrMissingIdentifier
	}

	if err != nil {
		if err == ErrUserNotFound {
			s.logger.Info("login failed: user not found")
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("failed to find user", zap.Error(err))
		return nil, err
	}

	// Verify password
	if !pkgauth.CheckPassword(input.Password, user.PasswordHash) {
		s.logger.Info("login failed: invalid password", zap.String("user_id", user.ID.String()))
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID.String())
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	s.logger.Info("user logged in successfully", zap.String("user_id", user.ID.String()))

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshTokenResult contains the result of a token refresh
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

// RefreshToken generates new tokens from a valid refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResult, error) {
	s.logger.Info("refreshing token")

	// Validate refresh token
	claims, err := s.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		if err == pkgauth.ErrExpiredToken {
			s.logger.Info("refresh token expired")
			return nil, ErrTokenExpired
		}
		s.logger.Info("invalid refresh token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	// Verify user still exists
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		s.logger.Error("invalid user ID in token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == ErrUserNotFound {
			s.logger.Info("user no longer exists", zap.String("user_id", claims.UserID))
			return nil, ErrInvalidToken
		}
		s.logger.Error("failed to get user", zap.Error(err))
		return nil, err
	}

	// Generate new tokens
	newAccessToken, err := s.tokenManager.GenerateAccessToken(claims.UserID)
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	newRefreshToken, err := s.tokenManager.GenerateRefreshToken(claims.UserID)
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	s.logger.Info("token refreshed successfully", zap.String("user_id", claims.UserID))

	return &RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetMe retrieves the current user's profile
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*User, error) {
	s.logger.Info("getting user profile", zap.String("user_id", userID.String()))

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	return user, nil
}

// Helper to get current time - can be mocked for testing
var now = func() time.Time {
	return time.Now()
}
