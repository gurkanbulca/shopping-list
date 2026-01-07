package grpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/group"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/interceptors"
	pkgerrors "github.com/gurkanbulca/shopping-list/api/pkg/errors"
	"github.com/gurkanbulca/shopping-list/api/pkg/validation"
	shoppingv1 "github.com/gurkanbulca/shopping-list/api/proto/shopping/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GroupHandler implements the GroupServiceServer interface
type GroupHandler struct {
	shoppingv1.UnimplementedGroupServiceServer
	groupService *group.Service
}

// NewGroupHandler creates a new GroupHandler
func NewGroupHandler(groupService *group.Service) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// CreateGroup handles group creation
func (h *GroupHandler) CreateGroup(ctx context.Context, req *shoppingv1.CreateGroupRequest) (*shoppingv1.CreateGroupResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate request
	if err := validation.ValidateCreateGroupRequest(req.Name, req.Description); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Prepare input
	input := group.CreateGroupInput{
		Name:    strings.TrimSpace(req.Name),
		OwnerID: userID,
	}

	description := strings.TrimSpace(req.Description)
	if description != "" {
		input.Description = &description
	}

	// Call service
	g, _, err := h.groupService.CreateGroup(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.CreateGroupResponse{
		Group: toProtoGroup(g),
	}, nil
}

// ListMyGroups handles listing user's groups
func (h *GroupHandler) ListMyGroups(ctx context.Context, req *shoppingv1.ListMyGroupsRequest) (*shoppingv1.ListMyGroupsResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Parse pagination
	limit := 20
	offset := 0

	if req.Pagination != nil {
		limit, _ = validation.ValidatePagination(req.Pagination.PageSize)
		if req.Pagination.PageToken != "" {
			offset, _ = decodePageToken(req.Pagination.PageToken)
		}
	}

	// Call service
	result, err := h.groupService.ListMyGroups(ctx, userID, limit, offset)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	groups := make([]*shoppingv1.Group, len(result.Groups))
	for i, g := range result.Groups {
		groups[i] = toProtoGroup(g)
	}

	// Build pagination response
	var nextPageToken string
	if offset+len(groups) < result.TotalCount {
		nextPageToken = encodePageToken(offset + limit)
	}

	return &shoppingv1.ListMyGroupsResponse{
		Groups: groups,
		Pagination: &shoppingv1.PaginationResponse{
			NextPageToken: nextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// InviteMember handles inviting a user to a group
func (h *GroupHandler) InviteMember(ctx context.Context, req *shoppingv1.InviteMemberRequest) (*shoppingv1.InviteMemberResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	inviterID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate request
	if err := validation.ValidateInviteMemberRequest(req.GroupId, req.Email, req.Phone, int32(req.Role)); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	groupID, _ := uuid.Parse(req.GroupId)

	// Prepare input
	input := group.InviteMemberInput{
		GroupID:   groupID,
		Role:      protoRoleToDomain(req.Role),
		InviterID: inviterID,
	}

	email := strings.TrimSpace(req.Email)
	if email != "" {
		input.Email = &email
	}

	phone := strings.TrimSpace(req.Phone)
	if phone != "" {
		input.Phone = &phone
	}

	// Call service
	member, err := h.groupService.InviteMember(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.InviteMemberResponse{
		Member: toProtoMember(member),
	}, nil
}

// AcceptInvite handles accepting a group invitation
func (h *GroupHandler) AcceptInvite(ctx context.Context, req *shoppingv1.AcceptInviteRequest) (*shoppingv1.AcceptInviteResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate group ID
	groupID, err := validation.ValidateGroupID(req.GroupId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Call service
	member, err := h.groupService.AcceptInvite(ctx, groupID, userID)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.AcceptInviteResponse{
		Member: toProtoMember(member),
	}, nil
}

// ListMembers handles listing group members
func (h *GroupHandler) ListMembers(ctx context.Context, req *shoppingv1.ListMembersRequest) (*shoppingv1.ListMembersResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	requesterID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate group ID
	groupID, err := validation.ValidateGroupID(req.GroupId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Parse pagination
	limit := 20
	offset := 0

	if req.Pagination != nil {
		limit, _ = validation.ValidatePagination(req.Pagination.PageSize)
		if req.Pagination.PageToken != "" {
			offset, _ = decodePageToken(req.Pagination.PageToken)
		}
	}

	// Call service
	result, err := h.groupService.ListMembers(ctx, groupID, requesterID, limit, offset)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	members := make([]*shoppingv1.GroupMember, len(result.Members))
	for i, m := range result.Members {
		members[i] = toProtoMember(m)
	}

	// Build pagination response
	var nextPageToken string
	if offset+len(members) < result.TotalCount {
		nextPageToken = encodePageToken(offset + limit)
	}

	return &shoppingv1.ListMembersResponse{
		Members: members,
		Pagination: &shoppingv1.PaginationResponse{
			NextPageToken: nextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// UpdateMemberRole handles updating a member's role
func (h *GroupHandler) UpdateMemberRole(ctx context.Context, req *shoppingv1.UpdateMemberRoleRequest) (*shoppingv1.UpdateMemberRoleResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	requesterID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate request
	if err := validation.ValidateUpdateMemberRoleRequest(req.GroupId, req.UserId, int32(req.Role)); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	groupID, _ := uuid.Parse(req.GroupId)
	targetID, _ := uuid.Parse(req.UserId)

	// Prepare input
	input := group.UpdateMemberRoleInput{
		GroupID:     groupID,
		TargetID:    targetID,
		NewRole:     protoRoleToDomain(req.Role),
		RequesterID: requesterID,
	}

	// Call service
	member, err := h.groupService.UpdateMemberRole(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.UpdateMemberRoleResponse{
		Member: toProtoMember(member),
	}, nil
}

// Helper functions

func toProtoGroup(g *group.Group) *shoppingv1.Group {
	protoGroup := &shoppingv1.Group{
		Id:        g.ID.String(),
		Name:      g.Name,
		OwnerId:   g.OwnerID.String(),
		CreatedAt: timestamppb.New(g.CreatedAt),
		UpdatedAt: timestamppb.New(g.UpdatedAt),
		Version:   g.Version,
	}

	if g.Description != nil {
		protoGroup.Description = *g.Description
	}

	return protoGroup
}

func toProtoMember(m *group.GroupMember) *shoppingv1.GroupMember {
	protoMember := &shoppingv1.GroupMember{
		Id:        m.ID.String(),
		GroupId:   m.GroupID.String(),
		UserId:    m.UserID.String(),
		Role:      domainRoleToProto(m.Role),
		Status:    domainStatusToProto(m.Status),
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
		Version:   m.Version,
	}

	if m.InvitedBy != nil {
		protoMember.InvitedBy = m.InvitedBy.String()
	}

	if m.InvitedAt != nil {
		protoMember.InvitedAt = timestamppb.New(*m.InvitedAt)
	}

	if m.AcceptedAt != nil {
		protoMember.AcceptedAt = timestamppb.New(*m.AcceptedAt)
	}

	return protoMember
}

func protoRoleToDomain(role shoppingv1.MemberRole) group.MemberRole {
	switch role {
	case shoppingv1.MemberRole_MEMBER_ROLE_OWNER:
		return group.RoleOwner
	case shoppingv1.MemberRole_MEMBER_ROLE_ADMIN:
		return group.RoleAdmin
	case shoppingv1.MemberRole_MEMBER_ROLE_MEMBER:
		return group.RoleMember
	default:
		return group.RoleMember
	}
}

func domainRoleToProto(role group.MemberRole) shoppingv1.MemberRole {
	switch role {
	case group.RoleOwner:
		return shoppingv1.MemberRole_MEMBER_ROLE_OWNER
	case group.RoleAdmin:
		return shoppingv1.MemberRole_MEMBER_ROLE_ADMIN
	case group.RoleMember:
		return shoppingv1.MemberRole_MEMBER_ROLE_MEMBER
	default:
		return shoppingv1.MemberRole_MEMBER_ROLE_UNSPECIFIED
	}
}

func domainStatusToProto(status group.MemberStatus) shoppingv1.MemberStatus {
	switch status {
	case group.StatusActive:
		return shoppingv1.MemberStatus_MEMBER_STATUS_ACTIVE
	case group.StatusInvited:
		return shoppingv1.MemberStatus_MEMBER_STATUS_INVITED
	default:
		return shoppingv1.MemberStatus_MEMBER_STATUS_UNSPECIFIED
	}
}

// Pagination helpers

func encodePageToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", offset)))
}

func decodePageToken(token string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(decoded))
}
