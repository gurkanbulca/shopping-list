package grpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/category"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/interceptors"
	pkgerrors "github.com/gurkanbulca/shopping-list/api/pkg/errors"
	"github.com/gurkanbulca/shopping-list/api/pkg/validation"
	shoppingv1 "github.com/gurkanbulca/shopping-list/api/proto/shopping/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CategoryHandler implements the CategoryServiceServer interface
type CategoryHandler struct {
	shoppingv1.UnimplementedCategoryServiceServer
	categoryService *category.Service
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryService *category.Service) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// UpsertCategory handles category creation and update
func (h *CategoryHandler) UpsertCategory(ctx context.Context, req *shoppingv1.UpsertCategoryRequest) (*shoppingv1.UpsertCategoryResponse, error) {
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
	if err := validation.ValidateUpsertCategoryRequest(req.GroupId, req.CategoryId, req.Name); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	groupID, _ := uuid.Parse(req.GroupId)

	// Prepare input
	input := category.UpsertCategoryInput{
		GroupID:         groupID,
		Name:            strings.TrimSpace(req.Name),
		ExpectedVersion: req.ExpectedVersion,
		UserID:          userID,
	}

	// Set category ID if provided (for update)
	if req.CategoryId != "" {
		categoryID, _ := uuid.Parse(req.CategoryId)
		input.CategoryID = &categoryID
	}

	// Call service
	cat, err := h.categoryService.UpsertCategory(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.UpsertCategoryResponse{
		Category: toProtoCategory(cat),
	}, nil
}

// ListCategories handles listing categories in a group
func (h *CategoryHandler) ListCategories(ctx context.Context, req *shoppingv1.ListCategoriesRequest) (*shoppingv1.ListCategoriesResponse, error) {
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
	groupID, err := validation.ValidateListCategoriesRequest(req.GroupId)
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
	result, err := h.categoryService.ListCategories(ctx, groupID, requesterID, limit, offset)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	categories := make([]*shoppingv1.Category, len(result.Categories))
	for i, c := range result.Categories {
		categories[i] = toProtoCategory(c)
	}

	// Build pagination response
	var nextPageToken string
	if offset+len(categories) < result.TotalCount {
		nextPageToken = encodePageToken(offset + limit)
	}

	return &shoppingv1.ListCategoriesResponse{
		Categories: categories,
		Pagination: &shoppingv1.PaginationResponse{
			NextPageToken: nextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// DeleteCategory handles category deletion
func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *shoppingv1.DeleteCategoryRequest) (*shoppingv1.DeleteCategoryResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	deletedBy, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate category ID
	categoryID, err := validation.ValidateDeleteCategoryRequest(req.CategoryId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Call service
	if err := h.categoryService.DeleteCategory(ctx, categoryID, deletedBy); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.DeleteCategoryResponse{}, nil
}

// Helper function to convert domain category to proto
func toProtoCategory(c *category.Category) *shoppingv1.Category {
	return &shoppingv1.Category{
		Id:        c.ID.String(),
		GroupId:   c.GroupID.String(),
		Name:      c.Name,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		Version:   c.Version,
	}
}
