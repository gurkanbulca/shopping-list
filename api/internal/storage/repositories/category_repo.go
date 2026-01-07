package repositories

import (
	"context"

	"github.com/google/uuid"
)

// Category represents a category entity
type Category struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	Name      string
	CreatedAt int64
	UpdatedAt int64
	UpdatedBy *uuid.UUID
	Version   int64
}

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	GetByGroupIDAndName(ctx context.Context, groupID uuid.UUID, name string) (*Category, error)
	ListByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}
