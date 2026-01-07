package category

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for category data access
type Repository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	GetByGroupIDAndName(ctx context.Context, groupID uuid.UUID, name string) (*Category, error)
	ListByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*Category, int, error)
	Update(ctx context.Context, category *Category, expectedVersion int64) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MembershipChecker defines the interface for checking group membership
type MembershipChecker interface {
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
}
