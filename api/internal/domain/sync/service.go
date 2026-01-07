package sync

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles sync business logic
type Service struct {
	syncRepo          Repository
	membershipChecker MembershipChecker
	logger            *zap.Logger
}

// NewService creates a new sync service
func NewService(syncRepo Repository, membershipChecker MembershipChecker, logger *zap.Logger) *Service {
	return &Service{
		syncRepo:          syncRepo,
		membershipChecker: membershipChecker,
		logger:            logger.Named("sync_service"),
	}
}

// GetDeltaInput contains the parameters for GetDelta
type GetDeltaInput struct {
	GroupID    uuid.UUID
	Cursor     int64
	MaxChanges int
	UserID     uuid.UUID
}

// GetDeltaResult contains the result of GetDelta
type GetDeltaResult struct {
	Changes    []*Change
	NextCursor int64
	HasMore    bool
}

// GetDelta retrieves changes since the given cursor for a group
func (s *Service) GetDelta(ctx context.Context, input GetDeltaInput) (*GetDeltaResult, error) {
	s.logger.Info("getting delta",
		zap.String("group_id", input.GroupID.String()),
		zap.Int64("cursor", input.Cursor),
		zap.Int("max_changes", input.MaxChanges),
	)

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, input.GroupID, input.UserID)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Validate max changes
	if input.MaxChanges <= 0 {
		input.MaxChanges = 100
	}
	if input.MaxChanges > 500 {
		input.MaxChanges = 500
	}

	// Request one extra to determine if there are more changes
	changes, err := s.syncRepo.GetDelta(ctx, input.GroupID, input.Cursor, input.MaxChanges+1)
	if err != nil {
		s.logger.Error("failed to get delta", zap.Error(err))
		return nil, err
	}

	// Determine if there are more changes
	hasMore := len(changes) > input.MaxChanges
	if hasMore {
		changes = changes[:input.MaxChanges]
	}

	// Calculate next cursor
	var nextCursor int64
	if len(changes) > 0 {
		nextCursor = changes[len(changes)-1].Sequence
	}

	s.logger.Info("delta retrieved",
		zap.Int("changes_count", len(changes)),
		zap.Int64("next_cursor", nextCursor),
		zap.Bool("has_more", hasMore),
	)

	return &GetDeltaResult{
		Changes:    changes,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// PushMutationsInput contains the parameters for PushMutations
type PushMutationsInput struct {
	GroupID   uuid.UUID
	Mutations []*Mutation
	UserID    uuid.UUID
}

// PushMutationsResult contains the results of all mutations
type PushMutationsResult struct {
	Results []*MutationResult
}

// PushMutations processes local mutations from client
func (s *Service) PushMutations(ctx context.Context, input PushMutationsInput) (*PushMutationsResult, error) {
	s.logger.Info("pushing mutations",
		zap.String("group_id", input.GroupID.String()),
		zap.Int("mutations_count", len(input.Mutations)),
	)

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, input.GroupID, input.UserID)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	results := make([]*MutationResult, len(input.Mutations))

	for i, mutation := range input.Mutations {
		result := s.processMutation(ctx, input.GroupID, mutation, input.UserID)
		results[i] = result
	}

	s.logger.Info("mutations processed",
		zap.Int("total", len(results)),
	)

	return &PushMutationsResult{Results: results}, nil
}

// processMutation processes a single mutation with idempotency check
func (s *Service) processMutation(ctx context.Context, groupID uuid.UUID, mutation *Mutation, userID uuid.UUID) *MutationResult {
	// Check if mutation was already processed (idempotency)
	processed, err := s.syncRepo.GetProcessedMutation(ctx, mutation.MutationID)
	if err == nil && processed != nil {
		// Return cached result
		s.logger.Info("mutation already processed (idempotent)",
			zap.String("mutation_id", mutation.MutationID.String()),
		)
		return &MutationResult{
			MutationID: mutation.MutationID,
			Success:    true,
			EntityID:   processed.EntityID,
			Version:    processed.Version,
		}
	}

	// Process the mutation based on type
	result := s.executeMutation(ctx, groupID, mutation, userID)

	// Record processed mutation if successful
	if result.Success {
		pm := NewProcessedMutation(mutation.MutationID, result.EntityID, result.Version)
		if err := s.syncRepo.RecordProcessedMutation(ctx, pm); err != nil {
			s.logger.Warn("failed to record processed mutation", zap.Error(err))
		}
	}

	return result
}

// executeMutation executes the actual mutation logic
func (s *Service) executeMutation(ctx context.Context, groupID uuid.UUID, mutation *Mutation, userID uuid.UUID) *MutationResult {
	result := &MutationResult{
		MutationID: mutation.MutationID,
		Success:    false,
	}

	// Validate mutation type
	switch mutation.Type {
	case MutationTypeCreate, MutationTypeUpdate, MutationTypeDelete:
		// Valid
	default:
		result.ErrorCode = "INVALID_MUTATION_TYPE"
		result.ErrorMessage = "invalid mutation type"
		return result
	}

	// Validate entity type
	switch mutation.EntityType {
	case "list", "item", "category":
		// Valid
	default:
		result.ErrorCode = "INVALID_ENTITY_TYPE"
		result.ErrorMessage = "invalid entity type: " + mutation.EntityType
		return result
	}

	// For now, return a basic success response
	// In a full implementation, this would call the respective domain services
	// to perform the actual create/update/delete operations
	entityID := uuid.New()
	if mutation.EntityID != nil {
		entityID = *mutation.EntityID
	}

	// Parse entity data if provided
	if len(mutation.EntityData) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(mutation.EntityData, &data); err != nil {
			result.ErrorCode = "INVALID_ENTITY_DATA"
			result.ErrorMessage = "failed to parse entity data: " + err.Error()
			return result
		}
	}

	// Record the change in mutation log
	operation := ChangeOperationCreate
	switch mutation.Type {
	case MutationTypeCreate:
		operation = ChangeOperationCreate
	case MutationTypeUpdate:
		operation = ChangeOperationUpdate
	case MutationTypeDelete:
		operation = ChangeOperationDelete
	}

	change := NewChange(
		EntityType(stringToEntityType(mutation.EntityType)),
		entityID,
		operation,
		&groupID,
		&userID,
	)

	if err := s.syncRepo.RecordChange(ctx, change); err != nil {
		s.logger.Error("failed to record change", zap.Error(err))
		result.ErrorCode = "INTERNAL_ERROR"
		result.ErrorMessage = "failed to record change"
		return result
	}

	result.Success = true
	result.EntityID = entityID
	result.Version = 1 // Would be returned from the actual mutation

	s.logger.Info("mutation executed",
		zap.String("mutation_id", mutation.MutationID.String()),
		zap.String("entity_type", mutation.EntityType),
		zap.String("operation", string(mutation.Type)),
	)

	return result
}

// stringToEntityType converts a string to EntityType
func stringToEntityType(s string) string {
	switch s {
	case "list":
		return string(EntityTypeList)
	case "item":
		return string(EntityTypeItem)
	case "category":
		return string(EntityTypeCategory)
	case "group":
		return string(EntityTypeGroup)
	case "group_member":
		return string(EntityTypeGroupMember)
	default:
		return s
	}
}
