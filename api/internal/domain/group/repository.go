package group

import (
	"context"

	"github.com/google/uuid"
)

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	// Create inserts a new group into the database
	Create(ctx context.Context, group *Group) error

	// GetByID retrieves a group by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Group, error)

	// ListByUserID retrieves all groups a user is a member of
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Group, int, error)

	// Update updates an existing group
	Update(ctx context.Context, group *Group) error

	// CountByUserID counts the total number of groups a user is a member of
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

// MemberRepository defines the interface for group member data access
type MemberRepository interface {
	// Create inserts a new group member
	Create(ctx context.Context, member *GroupMember) error

	// GetByGroupAndUser retrieves a member by group and user ID
	GetByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*GroupMember, error)

	// ListByGroupID retrieves all members of a group
	ListByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*GroupMember, int, error)

	// Update updates an existing group member
	Update(ctx context.Context, member *GroupMember) error

	// Delete removes a member from a group
	Delete(ctx context.Context, groupID, userID uuid.UUID) error

	// ExistsByGroupAndUser checks if a membership exists
	ExistsByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (bool, error)

	// GetByGroupAndEmail retrieves a member by group and email
	GetByGroupAndEmail(ctx context.Context, groupID uuid.UUID, email string) (*GroupMember, error)

	// GetByGroupAndPhone retrieves a member by group and phone
	GetByGroupAndPhone(ctx context.Context, groupID uuid.UUID, phone string) (*GroupMember, error)

	// CountByGroupID counts the total number of members in a group
	CountByGroupID(ctx context.Context, groupID uuid.UUID) (int, error)
}

// UserLookupRepository defines the interface for looking up users by email/phone
type UserLookupRepository interface {
	// GetUserIDByEmail retrieves a user's ID by email
	GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error)

	// GetUserIDByPhone retrieves a user's ID by phone
	GetUserIDByPhone(ctx context.Context, phone string) (uuid.UUID, error)
}
