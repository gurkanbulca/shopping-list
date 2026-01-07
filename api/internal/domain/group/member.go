package group

import (
	"time"

	"github.com/google/uuid"
)

// MemberRole represents the role of a member in a group
type MemberRole string

const (
	RoleOwner  MemberRole = "OWNER"
	RoleAdmin  MemberRole = "ADMIN"
	RoleMember MemberRole = "MEMBER"
)

// IsValid checks if the role is valid
func (r MemberRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

// CanManageMembers returns true if this role can invite/update members
func (r MemberRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

// MemberStatus represents the status of a group membership
type MemberStatus string

const (
	StatusActive  MemberStatus = "ACTIVE"
	StatusInvited MemberStatus = "INVITED"
)

// IsValid checks if the status is valid
func (s MemberStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInvited:
		return true
	default:
		return false
	}
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	ID         uuid.UUID
	GroupID    uuid.UUID
	UserID     uuid.UUID
	Role       MemberRole
	Status     MemberStatus
	InvitedBy  *uuid.UUID
	InvitedAt  *time.Time
	AcceptedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Version    int64
}

// NewGroupMember creates a new group member with INVITED status
func NewGroupMember(groupID, userID uuid.UUID, role MemberRole, invitedBy uuid.UUID) *GroupMember {
	now := time.Now()
	return &GroupMember{
		ID:        uuid.New(),
		GroupID:   groupID,
		UserID:    userID,
		Role:      role,
		Status:    StatusInvited,
		InvitedBy: &invitedBy,
		InvitedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

// NewOwnerMember creates a new owner member (automatically ACTIVE)
func NewOwnerMember(groupID, userID uuid.UUID) *GroupMember {
	now := time.Now()
	return &GroupMember{
		ID:         uuid.New(),
		GroupID:    groupID,
		UserID:     userID,
		Role:       RoleOwner,
		Status:     StatusActive,
		AcceptedAt: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

// Accept changes the status to ACTIVE
func (m *GroupMember) Accept() {
	now := time.Now()
	m.Status = StatusActive
	m.AcceptedAt = &now
	m.UpdatedAt = now
}

// UpdateRole changes the member's role
func (m *GroupMember) UpdateRole(role MemberRole) {
	m.Role = role
	m.UpdatedAt = time.Now()
}

// IsActive returns true if the member is active
func (m *GroupMember) IsActive() bool {
	return m.Status == StatusActive
}

// IsInvited returns true if the member has a pending invitation
func (m *GroupMember) IsInvited() bool {
	return m.Status == StatusInvited
}

// IsOwner returns true if the member is the owner
func (m *GroupMember) IsOwner() bool {
	return m.Role == RoleOwner
}

// CanManageMembers returns true if this member can invite/update other members
func (m *GroupMember) CanManageMembers() bool {
	return m.IsActive() && m.Role.CanManageMembers()
}
