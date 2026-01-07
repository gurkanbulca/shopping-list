package category

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles category business logic
type Service struct {
	categoryRepo      Repository
	membershipChecker MembershipChecker
	logger            *zap.Logger
}

// NewService creates a new category service
func NewService(categoryRepo Repository, membershipChecker MembershipChecker, logger *zap.Logger) *Service {
	return &Service{
		categoryRepo:      categoryRepo,
		membershipChecker: membershipChecker,
		logger:            logger.Named("category_service"),
	}
}

// UpsertCategoryInput contains the data needed to create or update a category
type UpsertCategoryInput struct {
	GroupID         uuid.UUID
	CategoryID      *uuid.UUID // nil for create, set for update
	Name            string
	ExpectedVersion int64 // For update only
	UserID          uuid.UUID
}

// UpsertCategory creates or updates a category
func (s *Service) UpsertCategory(ctx context.Context, input UpsertCategoryInput) (*Category, error) {
	s.logger.Info("upserting category",
		zap.String("group_id", input.GroupID.String()),
		zap.Stringp("category_id", uuidToStringPtr(input.CategoryID)),
		zap.String("name", input.Name),
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

	name := strings.TrimSpace(input.Name)

	// Check for existing category with same name (case-insensitive)
	existing, err := s.categoryRepo.GetByGroupIDAndName(ctx, input.GroupID, name)
	if err != nil && err != ErrCategoryNotFound {
		s.logger.Error("failed to check for existing category", zap.Error(err))
		return nil, err
	}

	// If updating
	if input.CategoryID != nil {
		// Get the category
		cat, err := s.categoryRepo.GetByID(ctx, *input.CategoryID)
		if err != nil {
			return nil, err
		}

		// Check if this is a different category with the same name
		if existing != nil && existing.ID != *input.CategoryID {
			return nil, ErrDuplicateCategoryName
		}

		// Update the category
		cat.Update(name, input.UserID)

		if err := s.categoryRepo.Update(ctx, cat, input.ExpectedVersion); err != nil {
			s.logger.Error("failed to update category", zap.Error(err))
			return nil, err
		}

		s.logger.Info("category updated successfully",
			zap.String("category_id", cat.ID.String()),
		)

		return cat, nil
	}

	// Creating new category - check for duplicate name
	if existing != nil {
		return nil, ErrDuplicateCategoryName
	}

	// Create new category
	cat := NewCategory(input.GroupID, name, input.UserID)

	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		s.logger.Error("failed to create category", zap.Error(err))
		return nil, err
	}

	s.logger.Info("category created successfully",
		zap.String("category_id", cat.ID.String()),
	)

	return cat, nil
}

// ListCategoriesResult contains the paginated list of categories
type ListCategoriesResult struct {
	Categories []*Category
	TotalCount int
}

// ListCategories returns categories in a group
func (s *Service) ListCategories(ctx context.Context, groupID, requesterID uuid.UUID, limit, offset int) (*ListCategoriesResult, error) {
	s.logger.Info("listing categories",
		zap.String("group_id", groupID.String()),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, groupID, requesterID)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	categories, totalCount, err := s.categoryRepo.ListByGroupID(ctx, groupID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list categories", zap.Error(err))
		return nil, err
	}

	return &ListCategoriesResult{
		Categories: categories,
		TotalCount: totalCount,
	}, nil
}

// DeleteCategory removes a category from a group (hard delete - items retain reference)
func (s *Service) DeleteCategory(ctx context.Context, categoryID, deletedBy uuid.UUID) error {
	s.logger.Info("deleting category",
		zap.String("category_id", categoryID.String()),
	)

	// Get the category to verify group membership
	cat, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, cat.GroupID, deletedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return err
	}
	if !isMember {
		return ErrNotGroupMember
	}

	if err := s.categoryRepo.Delete(ctx, categoryID); err != nil {
		s.logger.Error("failed to delete category", zap.Error(err))
		return err
	}

	s.logger.Info("category deleted successfully",
		zap.String("category_id", categoryID.String()),
	)

	return nil
}

// GetCategory retrieves a category by ID
func (s *Service) GetCategory(ctx context.Context, categoryID, requesterID uuid.UUID) (*Category, error) {
	cat, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, cat.GroupID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return cat, nil
}

// Helper function
func uuidToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
