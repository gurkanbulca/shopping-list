package grpc

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/gurkanbulca/shopping-list/api/internal/domain/list"
	"github.com/gurkanbulca/shopping-list/api/internal/transport/interceptors"
	pkgerrors "github.com/gurkanbulca/shopping-list/api/pkg/errors"
	"github.com/gurkanbulca/shopping-list/api/pkg/validation"
	shoppingv1 "github.com/gurkanbulca/shopping-list/api/proto/shopping/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListHandler implements the ListServiceServer interface
type ListHandler struct {
	shoppingv1.UnimplementedListServiceServer
	listService *list.Service
}

// NewListHandler creates a new ListHandler
func NewListHandler(listService *list.Service) *ListHandler {
	return &ListHandler{
		listService: listService,
	}
}

// CreateList handles creating a new shopping list
func (h *ListHandler) CreateList(ctx context.Context, req *shoppingv1.CreateListRequest) (*shoppingv1.CreateListResponse, error) {
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
	if err := validation.ValidateCreateListRequest(req.GroupId, req.Name, req.Description); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	groupID, _ := uuid.Parse(req.GroupId)

	// Prepare input
	input := list.CreateListInput{
		GroupID:   groupID,
		Name:      strings.TrimSpace(req.Name),
		CreatedBy: userID,
	}

	description := strings.TrimSpace(req.Description)
	if description != "" {
		input.Description = &description
	}

	// Call service
	l, err := h.listService.CreateList(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.CreateListResponse{
		List: toProtoList(l),
	}, nil
}

// ListLists handles listing shopping lists in a group
func (h *ListHandler) ListLists(ctx context.Context, req *shoppingv1.ListListsRequest) (*shoppingv1.ListListsResponse, error) {
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
	result, err := h.listService.ListLists(ctx, groupID, userID, req.IncludeArchived, limit, offset)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	lists := make([]*shoppingv1.List, len(result.Lists))
	for i, l := range result.Lists {
		lists[i] = toProtoList(l)
	}

	// Build pagination response
	var nextPageToken string
	if offset+len(lists) < result.TotalCount {
		nextPageToken = encodePageToken(offset + limit)
	}

	return &shoppingv1.ListListsResponse{
		Lists: lists,
		Pagination: &shoppingv1.PaginationResponse{
			NextPageToken: nextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// UpdateList handles updating list details
func (h *ListHandler) UpdateList(ctx context.Context, req *shoppingv1.UpdateListRequest) (*shoppingv1.UpdateListResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate list ID
	listID, err := validation.ValidateListID(req.ListId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Validate name and description
	if err := validation.ValidateListName(req.Name); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}
	if err := validation.ValidateListDescription(req.Description); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Prepare input
	input := list.UpdateListInput{
		ListID:          listID,
		Name:            strings.TrimSpace(req.Name),
		ExpectedVersion: req.ExpectedVersion,
		UpdatedBy:       userID,
	}

	description := strings.TrimSpace(req.Description)
	if description != "" {
		input.Description = &description
	}

	// Call service
	l, err := h.listService.UpdateList(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.UpdateListResponse{
		List: toProtoList(l),
	}, nil
}

// ArchiveList handles archiving/unarchiving a list
func (h *ListHandler) ArchiveList(ctx context.Context, req *shoppingv1.ArchiveListRequest) (*shoppingv1.ArchiveListResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate list ID
	listID, err := validation.ValidateListID(req.ListId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Call service
	l, err := h.listService.ArchiveList(ctx, listID, req.Archive, userID)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.ArchiveListResponse{
		List: toProtoList(l),
	}, nil
}

// AddItem handles adding an item to a list
func (h *ListHandler) AddItem(ctx context.Context, req *shoppingv1.AddItemRequest) (*shoppingv1.AddItemResponse, error) {
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
	if err := validation.ValidateAddItemRequest(req.ListId, req.Name, req.Quantity, req.Notes); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	listID, _ := uuid.Parse(req.ListId)
	categoryID, _ := validation.ValidateCategoryID(req.CategoryId)

	// Prepare input
	input := list.AddItemInput{
		ListID:     listID,
		Name:       strings.TrimSpace(req.Name),
		Priority:   protoPriorityToDomain(req.Priority),
		CategoryID: categoryID,
		CreatedBy:  userID,
	}

	quantity := strings.TrimSpace(req.Quantity)
	if quantity != "" {
		input.Quantity = &quantity
	}

	notes := strings.TrimSpace(req.Notes)
	if notes != "" {
		input.Notes = &notes
	}

	if req.SortOrder > 0 {
		sortOrder := req.SortOrder
		input.SortOrder = &sortOrder
	}

	// Call service
	item, err := h.listService.AddItem(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.AddItemResponse{
		Item: toProtoItem(item),
	}, nil
}

// UpdateItem handles updating an item
func (h *ListHandler) UpdateItem(ctx context.Context, req *shoppingv1.UpdateItemRequest) (*shoppingv1.UpdateItemResponse, error) {
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
	if err := validation.ValidateUpdateItemRequest(req.ItemId, req.Name, req.Quantity, req.Notes); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	itemID, _ := uuid.Parse(req.ItemId)
	categoryID, _ := validation.ValidateCategoryID(req.CategoryId)

	// Prepare input
	input := list.UpdateItemInput{
		ItemID:          itemID,
		Name:            strings.TrimSpace(req.Name),
		Priority:        protoPriorityToDomain(req.Priority),
		CategoryID:      categoryID,
		ExpectedVersion: req.ExpectedVersion,
		UpdatedBy:       userID,
	}

	quantity := strings.TrimSpace(req.Quantity)
	if quantity != "" {
		input.Quantity = &quantity
	}

	notes := strings.TrimSpace(req.Notes)
	if notes != "" {
		input.Notes = &notes
	}

	// Call service
	item, err := h.listService.UpdateItem(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.UpdateItemResponse{
		Item: toProtoItem(item),
	}, nil
}

// DeleteItem handles deleting an item
func (h *ListHandler) DeleteItem(ctx context.Context, req *shoppingv1.DeleteItemRequest) (*shoppingv1.DeleteItemResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate item ID
	itemID, err := validation.ValidateItemID(req.ItemId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Call service
	if err := h.listService.DeleteItem(ctx, itemID, userID); err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.DeleteItemResponse{}, nil
}

// TogglePurchased handles toggling an item's purchased status
func (h *ListHandler) TogglePurchased(ctx context.Context, req *shoppingv1.TogglePurchasedRequest) (*shoppingv1.TogglePurchasedResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate item ID
	itemID, err := validation.ValidateItemID(req.ItemId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Prepare input
	input := list.TogglePurchasedInput{
		ItemID:          itemID,
		IsPurchased:     req.IsPurchased,
		ExpectedVersion: req.ExpectedVersion,
		UpdatedBy:       userID,
	}

	// Call service
	item, err := h.listService.TogglePurchased(ctx, input)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	return &shoppingv1.TogglePurchasedResponse{
		Item: toProtoItem(item),
	}, nil
}

// ReorderItems handles reordering items in a list
func (h *ListHandler) ReorderItems(ctx context.Context, req *shoppingv1.ReorderItemsRequest) (*shoppingv1.ReorderItemsResponse, error) {
	// Get user ID from context
	userIDStr, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	// Validate list ID
	listID, err := validation.ValidateListID(req.ListId)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Parse item IDs
	itemIDs := make([]uuid.UUID, len(req.ItemIds))
	for i, idStr := range req.ItemIds {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid item ID format")
		}
		itemIDs[i] = id
	}

	// Call service
	items, err := h.listService.ReorderItems(ctx, listID, itemIDs, userID)
	if err != nil {
		return nil, pkgerrors.ToGRPCError(err)
	}

	// Build response
	protoItems := make([]*shoppingv1.Item, len(items))
	for i, item := range items {
		protoItems[i] = toProtoItem(item)
	}

	return &shoppingv1.ReorderItemsResponse{
		Items: protoItems,
	}, nil
}

// Helper functions

func toProtoList(l *list.List) *shoppingv1.List {
	protoList := &shoppingv1.List{
		Id:         l.ID.String(),
		GroupId:    l.GroupID.String(),
		Name:       l.Name,
		IsArchived: l.IsArchived,
		CreatedAt:  timestamppb.New(l.CreatedAt),
		UpdatedAt:  timestamppb.New(l.UpdatedAt),
		Version:    l.Version,
	}

	if l.Description != nil {
		protoList.Description = *l.Description
	}

	return protoList
}

func toProtoItem(i *list.Item) *shoppingv1.Item {
	protoItem := &shoppingv1.Item{
		Id:          i.ID.String(),
		ListId:      i.ListID.String(),
		Name:        i.Name,
		Priority:    domainPriorityToProto(i.Priority),
		SortOrder:   i.SortOrder,
		IsPurchased: i.IsPurchased,
		CreatedAt:   timestamppb.New(i.CreatedAt),
		UpdatedAt:   timestamppb.New(i.UpdatedAt),
		Version:     i.Version,
	}

	if i.CategoryID != nil {
		protoItem.CategoryId = i.CategoryID.String()
	}

	if i.Quantity != nil {
		protoItem.Quantity = *i.Quantity
	}

	if i.Notes != nil {
		protoItem.Notes = *i.Notes
	}

	return protoItem
}

func protoPriorityToDomain(priority shoppingv1.ItemPriority) list.Priority {
	switch priority {
	case shoppingv1.ItemPriority_ITEM_PRIORITY_LOW:
		return list.PriorityLow
	case shoppingv1.ItemPriority_ITEM_PRIORITY_MEDIUM:
		return list.PriorityMedium
	case shoppingv1.ItemPriority_ITEM_PRIORITY_HIGH:
		return list.PriorityHigh
	case shoppingv1.ItemPriority_ITEM_PRIORITY_URGENT:
		return list.PriorityUrgent
	default:
		return list.PriorityMedium
	}
}

func domainPriorityToProto(priority list.Priority) shoppingv1.ItemPriority {
	switch priority {
	case list.PriorityLow:
		return shoppingv1.ItemPriority_ITEM_PRIORITY_LOW
	case list.PriorityMedium:
		return shoppingv1.ItemPriority_ITEM_PRIORITY_MEDIUM
	case list.PriorityHigh:
		return shoppingv1.ItemPriority_ITEM_PRIORITY_HIGH
	case list.PriorityUrgent:
		return shoppingv1.ItemPriority_ITEM_PRIORITY_URGENT
	default:
		return shoppingv1.ItemPriority_ITEM_PRIORITY_UNSPECIFIED
	}
}
