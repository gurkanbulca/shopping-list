package repositories

import (
	"context"

	"github.com/google/uuid"
)

// Group represents a group entity
type Group struct {
	ID          uuid.UUID
	Name        string
	Description *string
	OwnerID     uuid.UUID
	CreatedAt   int64
	UpdatedAt   int64
	UpdatedBy   *uuid.UUID
	Version     int64
}

// GroupMember represents a group membership
type GroupMember struct {
	ID         uuid.UUID
	GroupID    uuid.UUID
	UserID     uuid.UUID
	Role       string // OWNER, ADMIN, MEMBER
	Status     string // ACTIVE, INVITED
	InvitedBy  *uuid.UUID
	InvitedAt  *int64
	AcceptedAt *int64
	CreatedAt  int64
	UpdatedAt  int64
	Version    int64
}

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	CreateGroup(ctx context.Context, group *Group) error
	GetGroupByID(ctx context.Context, id uuid.UUID) (*Group, error)
	ListGroupsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Group, error)
	UpdateGroup(ctx context.Context, group *Group) error
}

// GroupMemberRepository defines the interface for group membership data access
type GroupMemberRepository interface {
	CreateMember(ctx context.Context, member *GroupMember) error
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*GroupMember, error)
	ListMembersByGroupID(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*GroupMember, error)
	UpdateMember(ctx context.Context, member *GroupMember) error
	DeleteMember(ctx context.Context, groupID, userID uuid.UUID) error
}
