package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements auth.UserRepository with PostgreSQL
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{pool: db.Pool}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (id, email, phone, password_hash, name, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
		user.Version,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Check for unique constraint violations
			if pgErr.Code == "23505" { // unique_violation
				if pgErr.ConstraintName == "idx_users_email" || pgErr.ConstraintName == "users_email_key" {
					return auth.ErrDuplicateEmail
				}
				if pgErr.ConstraintName == "idx_users_phone" || pgErr.ConstraintName == "users_phone_key" {
					return auth.ErrDuplicatePhone
				}
			}
		}
		return err
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `
		SELECT id, email, phone, password_hash, name, created_at, updated_at, version
		FROM users
		WHERE id = $1
	`

	return r.scanUser(ctx, query, id)
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, phone, password_hash, name, created_at, updated_at, version
		FROM users
		WHERE email = $1
	`

	return r.scanUser(ctx, query, email)
}

// GetByPhone retrieves a user by their phone number
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*auth.User, error) {
	query := `
		SELECT id, email, phone, password_hash, name, created_at, updated_at, version
		FROM users
		WHERE phone = $1
	`

	return r.scanUser(ctx, query, phone)
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users
		SET email = $2, phone = $3, password_hash = $4, name = $5, updated_at = $6, version = version + 1
		WHERE id = $1 AND version = $7
	`

	result, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Name,
		time.Now(),
		user.Version,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "idx_users_email" || pgErr.ConstraintName == "users_email_key" {
				return auth.ErrDuplicateEmail
			}
			if pgErr.ConstraintName == "idx_users_phone" || pgErr.ConstraintName == "users_phone_key" {
				return auth.ErrDuplicatePhone
			}
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail checks if an email is already in use
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ExistsByPhone checks if a phone number is already in use
func (r *UserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, phone).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// scanUser is a helper to scan a user from a query result
func (r *UserRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*auth.User, error) {
	var user auth.User

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
