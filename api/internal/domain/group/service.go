package group

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles group business logic
type Service struct {
	groupRepo  GroupRepository
	memberRepo MemberRepository
	userLookup UserLookupRepository
	logger     *zap.Logger
}

// NewService creates a new group service
func NewService(groupRepo GroupRepository, memberRepo MemberRepository, userLookup UserLookupRepository, logger *zap.Logger) *Service {
	return &Service{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		userLookup: userLookup,
		logger:     logger.Named("group_service"),
	}
}

// CreateGroupInput contains the data needed to create a group
type CreateGroupInput struct {
	Name        string
	Description *string
	OwnerID     uuid.UUID
}

// CreateGroup creates a new group with the caller as owner
func (s *Service) CreateGroup(ctx context.Context, input CreateGroupInput) (*Group, *GroupMember, error) {
	s.logger.Info("creating new group",
		zap.String("name", input.Name),
		zap.String("owner_id", input.OwnerID.String()),
	)

	// Create the group
	group := NewGroup(input.Name, input.Description, input.OwnerID)

	if err := s.groupRepo.Create(ctx, group); err != nil {
		s.logger.Error("failed to create group", zap.Error(err))
		return nil, nil, err
	}

	// Create owner membership
	ownerMember := NewOwnerMember(group.ID, input.OwnerID)
	if err := s.memberRepo.Create(ctx, ownerMember); err != nil {
		s.logger.Error("failed to create owner membership", zap.Error(err))
		return nil, nil, err
	}

	s.logger.Info("group created successfully",
		zap.String("group_id", group.ID.String()),
	)

	return group, ownerMember, nil
}

// ListGroupsResult contains the paginated list of groups
type ListGroupsResult struct {
	Groups     []*Group
	TotalCount int
}

// ListMyGroups returns groups the user is a member of
func (s *Service) ListMyGroups(ctx context.Context, userID uuid.UUID, limit, offset int) (*ListGroupsResult, error) {
	s.logger.Info("listing user groups",
		zap.String("user_id", userID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	groups, totalCount, err := s.groupRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list groups", zap.Error(err))
		return nil, err
	}

	return &ListGroupsResult{
		Groups:     groups,
		TotalCount: totalCount,
	}, nil
}

// InviteMemberInput contains the data needed to invite a member
type InviteMemberInput struct {
	GroupID   uuid.UUID
	Email     *string
	Phone     *string
	Role      MemberRole
	InviterID uuid.UUID
}

// InviteMember invites a user to join a group
func (s *Service) InviteMember(ctx context.Context, input InviteMemberInput) (*GroupMember, error) {
	s.logger.Info("inviting member to group",
		zap.String("group_id", input.GroupID.String()),
		zap.Stringp("email", input.Email),
		zap.Stringp("phone", input.Phone),
		zap.String("role", string(input.Role)),
	)

	// Cannot invite as owner
	if input.Role == RoleOwner {
		return nil, ErrCannotInviteAsOwner
	}

	// Verify inviter is a member with permission
	inviterMember, err := s.memberRepo.GetByGroupAndUser(ctx, input.GroupID, input.InviterID)
	if err != nil {
		if err == ErrMemberNotFound {
			return nil, ErrNotGroupMember
		}
		s.logger.Error("failed to get inviter membership", zap.Error(err))
		return nil, err
	}

	if !inviterMember.CanManageMembers() {
		return nil, ErrNotAuthorized
	}

	// Find the user to invite
	var inviteeID uuid.UUID
	if input.Email != nil && *input.Email != "" {
		inviteeID, err = s.userLookup.GetUserIDByEmail(ctx, *input.Email)
	} else if input.Phone != nil && *input.Phone != "" {
		inviteeID, err = s.userLookup.GetUserIDByPhone(ctx, *input.Phone)
	} else {
		return nil, ErrUserNotFound
	}

	if err != nil {
		s.logger.Info("user not found for invitation", zap.Error(err))
		return nil, ErrUserNotFound
	}

	// Check if user is already a member or invited
	existingMember, err := s.memberRepo.GetByGroupAndUser(ctx, input.GroupID, inviteeID)
	if err == nil {
		if existingMember.IsActive() {
			return nil, ErrAlreadyMember
		}
		if existingMember.IsInvited() {
			return nil, ErrAlreadyInvited
		}
	} else if err != ErrMemberNotFound {
		s.logger.Error("failed to check existing membership", zap.Error(err))
		return nil, err
	}

	// Create the invitation
	member := NewGroupMember(input.GroupID, inviteeID, input.Role, input.InviterID)
	if err := s.memberRepo.Create(ctx, member); err != nil {
		s.logger.Error("failed to create member invitation", zap.Error(err))
		return nil, err
	}

	s.logger.Info("member invited successfully",
		zap.String("group_id", input.GroupID.String()),
		zap.String("member_id", member.ID.String()),
	)

	return member, nil
}

// AcceptInvite accepts a pending invitation
func (s *Service) AcceptInvite(ctx context.Context, groupID, userID uuid.UUID) (*GroupMember, error) {
	s.logger.Info("accepting group invitation",
		zap.String("group_id", groupID.String()),
		zap.String("user_id", userID.String()),
	)

	// Get the pending invitation
	member, err := s.memberRepo.GetByGroupAndUser(ctx, groupID, userID)
	if err != nil {
		if err == ErrMemberNotFound {
			return nil, ErrNoInvitation
		}
		s.logger.Error("failed to get invitation", zap.Error(err))
		return nil, err
	}

	if !member.IsInvited() {
		if member.IsActive() {
			return nil, ErrAlreadyMember
		}
		return nil, ErrNoInvitation
	}

	// Accept the invitation
	member.Accept()
	if err := s.memberRepo.Update(ctx, member); err != nil {
		s.logger.Error("failed to accept invitation", zap.Error(err))
		return nil, err
	}

	s.logger.Info("invitation accepted successfully",
		zap.String("group_id", groupID.String()),
		zap.String("member_id", member.ID.String()),
	)

	return member, nil
}

// ListMembersResult contains the paginated list of members
type ListMembersResult struct {
	Members    []*GroupMember
	TotalCount int
}

// ListMembers returns members of a group
func (s *Service) ListMembers(ctx context.Context, groupID, requesterID uuid.UUID, limit, offset int) (*ListMembersResult, error) {
	s.logger.Info("listing group members",
		zap.String("group_id", groupID.String()),
		zap.String("requester_id", requesterID.String()),
	)

	// Verify requester is a member
	_, err := s.memberRepo.GetByGroupAndUser(ctx, groupID, requesterID)
	if err != nil {
		if err == ErrMemberNotFound {
			return nil, ErrNotGroupMember
		}
		s.logger.Error("failed to verify requester membership", zap.Error(err))
		return nil, err
	}

	members, totalCount, err := s.memberRepo.ListByGroupID(ctx, groupID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list members", zap.Error(err))
		return nil, err
	}

	return &ListMembersResult{
		Members:    members,
		TotalCount: totalCount,
	}, nil
}

// UpdateMemberRoleInput contains the data needed to update a member's role
type UpdateMemberRoleInput struct {
	GroupID     uuid.UUID
	TargetID    uuid.UUID // User whose role is being updated
	NewRole     MemberRole
	RequesterID uuid.UUID
}

// UpdateMemberRole updates a member's role
func (s *Service) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (*GroupMember, error) {
	s.logger.Info("updating member role",
		zap.String("group_id", input.GroupID.String()),
		zap.String("target_id", input.TargetID.String()),
		zap.String("new_role", string(input.NewRole)),
	)

	// Validate role
	if !input.NewRole.IsValid() {
		return nil, ErrInvalidRole
	}

	// Cannot assign owner role via this method
	if input.NewRole == RoleOwner {
		return nil, ErrCannotChangeOwner
	}

	// Verify requester has permission
	requesterMember, err := s.memberRepo.GetByGroupAndUser(ctx, input.GroupID, input.RequesterID)
	if err != nil {
		if err == ErrMemberNotFound {
			return nil, ErrNotGroupMember
		}
		s.logger.Error("failed to get requester membership", zap.Error(err))
		return nil, err
	}

	if !requesterMember.CanManageMembers() {
		return nil, ErrNotAuthorized
	}

	// Get target member
	targetMember, err := s.memberRepo.GetByGroupAndUser(ctx, input.GroupID, input.TargetID)
	if err != nil {
		if err == ErrMemberNotFound {
			return nil, ErrMemberNotFound
		}
		s.logger.Error("failed to get target membership", zap.Error(err))
		return nil, err
	}

	// Cannot change owner's role
	if targetMember.IsOwner() {
		return nil, ErrCannotChangeOwner
	}

	// Only owner can manage admins
	if targetMember.Role == RoleAdmin && !requesterMember.IsOwner() {
		return nil, ErrNotAuthorized
	}

	// Update the role
	targetMember.UpdateRole(input.NewRole)
	if err := s.memberRepo.Update(ctx, targetMember); err != nil {
		s.logger.Error("failed to update member role", zap.Error(err))
		return nil, err
	}

	s.logger.Info("member role updated successfully",
		zap.String("member_id", targetMember.ID.String()),
		zap.String("new_role", string(input.NewRole)),
	)

	return targetMember, nil
}

// IsMember checks if a user is an active member of a group
func (s *Service) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	member, err := s.memberRepo.GetByGroupAndUser(ctx, groupID, userID)
	if err != nil {
		if err == ErrMemberNotFound {
			return false, nil
		}
		return false, err
	}
	return member.IsActive(), nil
}

// GetMember returns a user's membership in a group
func (s *Service) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*GroupMember, error) {
	return s.memberRepo.GetByGroupAndUser(ctx, groupID, userID)
}
