package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/sync"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/interceptors"
	pkgerrors "github.com/gurkanbulca/shopping-list/api/pkg/errors"
	"github.com/gurkanbulca/shopping-list/api/pkg/validation"
	shoppingv1 "github.com/gurkanbulca/shopping-list/api/proto/shopping/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SyncHandler implements the SyncServiceServer interface
type SyncHandler struct {
	shoppingv1.UnimplementedSyncServiceServer
	syncService *sync.Service
}

// NewSyncHandler creates a new SyncHandler
func NewSyncHandler(syncService *sync.Service) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
	}
}

// GetDelta handles retrieving changes since last sync
func (h *SyncHandler) GetDelta(ctx context.Context, req *shoppingv1.GetDeltaRequest) (*shoppingv1.GetDeltaResponse, error) {
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
	groupID, err := validation.ValidateGetDeltaRequest(req.GroupId, req.Cursor, req.MaxChanges)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Prepare input
	input := sync.GetDeltaInput{
		GroupID:    groupID,
		Cursor:     req.Cursor,
		MaxChanges: int(req.MaxChanges),
		UserID:     userID,
	}

	// Call service
	result, err := h.syncService.GetDelta(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	changes := make([]*shoppingv1.Change, len(result.Changes))
	for i, c := range result.Changes {
		changes[i] = toProtoChange(c)
	}

	return &shoppingv1.GetDeltaResponse{
		Changes:    changes,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}, nil
}

// PushMutations handles pushing local mutations to server
func (h *SyncHandler) PushMutations(ctx context.Context, req *shoppingv1.PushMutationsRequest) (*shoppingv1.PushMutationsResponse, error) {
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
	groupID, err := validation.ValidatePushMutationsRequest(req.GroupId, len(req.Mutations))
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Convert proto mutations to domain mutations
	mutations := make([]*sync.Mutation, len(req.Mutations))
	for i, m := range req.Mutations {
		mutationID, err := validation.ValidateMutationID(m.MutationId)
		if err != nil {
			return nil, pkgerrors.ToGRPCError(err)
		}

		mutation := &sync.Mutation{
			MutationID:      mutationID,
			Type:            protoMutationTypeToDomain(m.Type),
			EntityType:      m.EntityType,
			EntityData:      m.EntityData,
			ExpectedVersion: m.ExpectedVersion,
		}

		// Parse entity ID if provided
		if m.EntityId != "" {
			entityID, err := uuid.Parse(m.EntityId)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid entity_id format")
			}
			mutation.EntityID = &entityID
		}

		mutations[i] = mutation
	}

	// Prepare input
	input := sync.PushMutationsInput{
		GroupID:   groupID,
		Mutations: mutations,
		UserID:    userID,
	}

	// Call service
	result, err := h.syncService.PushMutations(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	results := make([]*shoppingv1.MutationResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = toProtoMutationResult(r)
	}

	return &shoppingv1.PushMutationsResponse{
		Results: results,
	}, nil
}

// Helper functions

func toProtoChange(c *sync.Change) *shoppingv1.Change {
	return &shoppingv1.Change{
		Sequence:   c.Sequence,
		EntityType: domainEntityTypeToProto(c.EntityType),
		EntityId:   c.EntityID.String(),
		Operation:  domainOperationToProto(c.Operation),
		ChangedAt:  timestamppb.New(c.ChangedAt),
	}
}

func toProtoMutationResult(r *sync.MutationResult) *shoppingv1.MutationResult {
	return &shoppingv1.MutationResult{
		MutationId:   r.MutationID.String(),
		Success:      r.Success,
		ErrorCode:    r.ErrorCode,
		ErrorMessage: r.ErrorMessage,
		EntityId:     r.EntityID.String(),
		Version:      r.Version,
	}
}

func domainEntityTypeToProto(et sync.EntityType) shoppingv1.EntityType {
	switch et {
	case sync.EntityTypeUser:
		return shoppingv1.EntityType_ENTITY_TYPE_USER
	case sync.EntityTypeGroup:
		return shoppingv1.EntityType_ENTITY_TYPE_GROUP
	case sync.EntityTypeGroupMember:
		return shoppingv1.EntityType_ENTITY_TYPE_GROUP_MEMBER
	case sync.EntityTypeList:
		return shoppingv1.EntityType_ENTITY_TYPE_LIST
	case sync.EntityTypeItem:
		return shoppingv1.EntityType_ENTITY_TYPE_ITEM
	case sync.EntityTypeCategory:
		return shoppingv1.EntityType_ENTITY_TYPE_CATEGORY
	default:
		return shoppingv1.EntityType_ENTITY_TYPE_UNSPECIFIED
	}
}

func domainOperationToProto(op sync.ChangeOperation) shoppingv1.ChangeOperation {
	switch op {
	case sync.ChangeOperationCreate:
		return shoppingv1.ChangeOperation_CHANGE_OPERATION_CREATE
	case sync.ChangeOperationUpdate:
		return shoppingv1.ChangeOperation_CHANGE_OPERATION_UPDATE
	case sync.ChangeOperationDelete:
		return shoppingv1.ChangeOperation_CHANGE_OPERATION_DELETE
	default:
		return shoppingv1.ChangeOperation_CHANGE_OPERATION_UNSPECIFIED
	}
}

func protoMutationTypeToDomain(mt shoppingv1.MutationType) sync.MutationType {
	switch mt {
	case shoppingv1.MutationType_MUTATION_TYPE_CREATE:
		return sync.MutationTypeCreate
	case shoppingv1.MutationType_MUTATION_TYPE_UPDATE:
		return sync.MutationTypeUpdate
	case shoppingv1.MutationType_MUTATION_TYPE_DELETE:
		return sync.MutationTypeDelete
	default:
		return sync.MutationType("")
	}
}
