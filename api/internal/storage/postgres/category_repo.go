package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/category"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CategoryRepository implements category.Repository with PostgreSQL
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new PostgreSQL category repository
func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{pool: db.Pool}
}

// Create inserts a new category into the database
func (r *CategoryRepository) Create(ctx context.Context, c *category.Category) error {
	query := `
		INSERT INTO categories (id, group_id, name, created_at, updated_at, updated_by, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		c.ID,
		c.GroupID,
		c.Name,
		c.CreatedAt,
		c.UpdatedAt,
		c.UpdatedBy,
		c.Version,
	)

	if err != nil && isDuplicateKeyError(err) {
		return category.ErrDuplicateCategoryName
	}

	return err
}

// GetByID retrieves a category by its ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*category.Category, error) {
	query := `
		SELECT id, group_id, name, created_at, updated_at, updated_by, version
		FROM categories
		WHERE id = $1
	`

	var c category.Category
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.GroupID,
		&c.Name,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.UpdatedBy,
		&c.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, category.ErrCategoryNotFound
		}
		return nil, err
	}

	return &c, nil
}

// GetByGroupIDAndName retrieves a category by group ID and name (case-insensitive)
func (r *CategoryRepository) GetByGroupIDAndName(ctx context.Context, groupID uuid.UUID, name string) (*category.Category, error) {
	query := `
		SELECT id, group_id, name, created_at, updated_at, updated_by, version
		FROM categories
		WHERE group_id = $1 AND LOWER(name) = LOWER($2)
	`

	var c category.Category
	err := r.pool.QueryRow(ctx, query, groupID, name).Scan(
		&c.ID,
		&c.GroupID,
		&c.Name,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.UpdatedBy,
		&c.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, category.ErrCategoryNotFound
		}
		return nil, err
	}

	return &c, nil
}

// ListByGroupID retrieves all categories in a group with pagination
func (r *CategoryRepository) ListByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*category.Category, int, error) {
	query := `
		SELECT id, group_id, name, created_at, updated_at, updated_by, version
		FROM categories
		WHERE group_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var categories []*category.Category
	for rows.Next() {
		var c category.Category
		if err := rows.Scan(
			&c.ID,
			&c.GroupID,
			&c.Name,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.UpdatedBy,
			&c.Version,
		); err != nil {
			return nil, 0, err
		}
		categories = append(categories, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	totalCount, err := r.CountByGroupID(ctx, groupID)
	if err != nil {
		return nil, 0, err
	}

	return categories, totalCount, nil
}

// Update updates an existing category with version check
func (r *CategoryRepository) Update(ctx context.Context, c *category.Category, expectedVersion int64) error {
	query := `
		UPDATE categories
		SET name = $2, updated_at = $3, updated_by = $4, version = version + 1
		WHERE id = $1 AND version = $5
		RETURNING version
	`

	var newVersion int64
	err := r.pool.QueryRow(ctx, query,
		c.ID,
		c.Name,
		c.UpdatedAt,
		c.UpdatedBy,
		expectedVersion,
	).Scan(&newVersion)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Check if category exists
			_, existErr := r.GetByID(ctx, c.ID)
			if existErr != nil {
				return category.ErrCategoryNotFound
			}
			return category.ErrVersionMismatch
		}
		if isDuplicateKeyError(err) {
			return category.ErrDuplicateCategoryName
		}
		return err
	}

	c.Version = newVersion
	return nil
}

// Delete removes a category from the database
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return category.ErrCategoryNotFound
	}

	return nil
}

// CountByGroupID counts the total number of categories in a group
func (r *CategoryRepository) CountByGroupID(ctx context.Context, groupID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM categories WHERE group_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, groupID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// isDuplicateKeyError checks if the error is a unique constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "23505") // PostgreSQL error code for unique_violation
}
