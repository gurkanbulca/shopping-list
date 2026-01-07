package group

import "errors"

var (
	// ErrGroupNotFound is returned when a group is not found
	ErrGroupNotFound = errors.New("group not found")

	// ErrMemberNotFound is returned when a group member is not found
	ErrMemberNotFound = errors.New("member not found")

	// ErrNotGroupMember is returned when a user is not a member of the group
	ErrNotGroupMember = errors.New("user is not a member of this group")

	// ErrNotAuthorized is returned when a user doesn't have permission
	ErrNotAuthorized = errors.New("not authorized to perform this action")

	// ErrAlreadyMember is returned when inviting someone who is already a member
	ErrAlreadyMember = errors.New("user is already a member of this group")

	// ErrAlreadyInvited is returned when inviting someone who already has a pending invitation
	ErrAlreadyInvited = errors.New("user already has a pending invitation")

	// ErrNoInvitation is returned when trying to accept a non-existent invitation
	ErrNoInvitation = errors.New("no pending invitation found")

	// ErrInvalidRole is returned when an invalid role is specified
	ErrInvalidRole = errors.New("invalid role")

	// ErrCannotChangeOwner is returned when trying to change the owner's role
	ErrCannotChangeOwner = errors.New("cannot change owner's role")

	// ErrGroupMustHaveOwner is returned when trying to remove/demote the only owner
	ErrGroupMustHaveOwner = errors.New("group must have at least one owner")

	// ErrCannotInviteAsOwner is returned when trying to invite someone as owner
	ErrCannotInviteAsOwner = errors.New("cannot invite as owner; use role transfer instead")

	// ErrUserNotFound is returned when the invited user doesn't exist
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidGroupName is returned when group name is invalid
	ErrInvalidGroupName = errors.New("group name must be 1-100 characters")
)
