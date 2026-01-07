package list

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles shopping list business logic
type Service struct {
	listRepo          ListRepository
	itemRepo          ItemRepository
	membershipChecker MembershipChecker
	logger            *zap.Logger
}

// NewService creates a new list service
func NewService(listRepo ListRepository, itemRepo ItemRepository, membershipChecker MembershipChecker, logger *zap.Logger) *Service {
	return &Service{
		listRepo:          listRepo,
		itemRepo:          itemRepo,
		membershipChecker: membershipChecker,
		logger:            logger.Named("list_service"),
	}
}

// CreateListInput contains the data needed to create a list
type CreateListInput struct {
	GroupID     uuid.UUID
	Name        string
	Description *string
	CreatedBy   uuid.UUID
}

// CreateList creates a new shopping list
func (s *Service) CreateList(ctx context.Context, input CreateListInput) (*List, error) {
	s.logger.Info("creating new list",
		zap.String("group_id", input.GroupID.String()),
		zap.String("name", input.Name),
	)

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, input.GroupID, input.CreatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Create the list
	list := NewList(input.GroupID, input.Name, input.Description, input.CreatedBy)

	if err := s.listRepo.Create(ctx, list); err != nil {
		s.logger.Error("failed to create list", zap.Error(err))
		return nil, err
	}

	s.logger.Info("list created successfully",
		zap.String("list_id", list.ID.String()),
	)

	return list, nil
}

// ListListsResult contains the paginated list of lists
type ListListsResult struct {
	Lists      []*List
	TotalCount int
}

// ListLists returns lists in a group
func (s *Service) ListLists(ctx context.Context, groupID, requesterID uuid.UUID, includeArchived bool, limit, offset int) (*ListListsResult, error) {
	s.logger.Info("listing lists",
		zap.String("group_id", groupID.String()),
		zap.Bool("include_archived", includeArchived),
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

	lists, totalCount, err := s.listRepo.ListByGroupID(ctx, groupID, includeArchived, limit, offset)
	if err != nil {
		s.logger.Error("failed to list lists", zap.Error(err))
		return nil, err
	}

	return &ListListsResult{
		Lists:      lists,
		TotalCount: totalCount,
	}, nil
}

// UpdateListInput contains the data needed to update a list
type UpdateListInput struct {
	ListID          uuid.UUID
	Name            string
	Description     *string
	ExpectedVersion int64
	UpdatedBy       uuid.UUID
}

// UpdateList updates a list's details
func (s *Service) UpdateList(ctx context.Context, input UpdateListInput) (*List, error) {
	s.logger.Info("updating list",
		zap.String("list_id", input.ListID.String()),
	)

	// Get the list
	list, err := s.listRepo.GetByID(ctx, input.ListID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, input.UpdatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Check if archived
	if list.IsArchived {
		return nil, ErrListArchived
	}

	// Update the list
	list.Update(input.Name, input.Description, input.UpdatedBy)

	if err := s.listRepo.Update(ctx, list, input.ExpectedVersion); err != nil {
		s.logger.Error("failed to update list", zap.Error(err))
		return nil, err
	}

	s.logger.Info("list updated successfully",
		zap.String("list_id", list.ID.String()),
	)

	return list, nil
}

// ArchiveList archives or unarchives a list
func (s *Service) ArchiveList(ctx context.Context, listID uuid.UUID, archive bool, updatedBy uuid.UUID) (*List, error) {
	s.logger.Info("archiving list",
		zap.String("list_id", listID.String()),
		zap.Bool("archive", archive),
	)

	// Get the list to check membership
	list, err := s.listRepo.GetByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, updatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	updatedList, err := s.listRepo.Archive(ctx, listID, archive, updatedBy)
	if err != nil {
		s.logger.Error("failed to archive list", zap.Error(err))
		return nil, err
	}

	s.logger.Info("list archived successfully",
		zap.String("list_id", listID.String()),
		zap.Bool("archived", archive),
	)

	return updatedList, nil
}

// AddItemInput contains the data needed to add an item
type AddItemInput struct {
	ListID     uuid.UUID
	Name       string
	Priority   Priority
	CategoryID *uuid.UUID
	Quantity   *string
	Notes      *string
	SortOrder  *int32
	CreatedBy  uuid.UUID
}

// AddItem adds an item to a list
func (s *Service) AddItem(ctx context.Context, input AddItemInput) (*Item, error) {
	s.logger.Info("adding item",
		zap.String("list_id", input.ListID.String()),
		zap.String("name", input.Name),
	)

	// Get the list to verify membership and check archive status
	list, err := s.listRepo.GetByID(ctx, input.ListID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, input.CreatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Check if archived
	if list.IsArchived {
		return nil, ErrListArchived
	}

	// Get sort order
	sortOrder := int32(0)
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	} else {
		// Append to end
		maxOrder, err := s.itemRepo.GetMaxSortOrder(ctx, input.ListID)
		if err != nil {
			s.logger.Error("failed to get max sort order", zap.Error(err))
			return nil, err
		}
		sortOrder = maxOrder + 1
	}

	// Create the item
	item := NewItem(input.ListID, input.Name, input.Priority, sortOrder, input.CreatedBy)
	item.CategoryID = input.CategoryID
	item.Quantity = input.Quantity
	item.Notes = input.Notes

	if err := s.itemRepo.Create(ctx, item); err != nil {
		s.logger.Error("failed to create item", zap.Error(err))
		return nil, err
	}

	s.logger.Info("item added successfully",
		zap.String("item_id", item.ID.String()),
	)

	return item, nil
}

// UpdateItemInput contains the data needed to update an item
type UpdateItemInput struct {
	ItemID          uuid.UUID
	Name            string
	Priority        Priority
	CategoryID      *uuid.UUID
	Quantity        *string
	Notes           *string
	ExpectedVersion int64
	UpdatedBy       uuid.UUID
}

// UpdateItem updates an item's details
func (s *Service) UpdateItem(ctx context.Context, input UpdateItemInput) (*Item, error) {
	s.logger.Info("updating item",
		zap.String("item_id", input.ItemID.String()),
	)

	// Get the item
	item, err := s.itemRepo.GetByID(ctx, input.ItemID)
	if err != nil {
		return nil, err
	}

	// Get the list to verify membership
	list, err := s.listRepo.GetByID(ctx, item.ListID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, input.UpdatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Check if archived
	if list.IsArchived {
		return nil, ErrListArchived
	}

	// Update the item
	item.Update(input.Name, input.Priority, input.CategoryID, input.Quantity, input.Notes, input.UpdatedBy)

	if err := s.itemRepo.Update(ctx, item, input.ExpectedVersion); err != nil {
		s.logger.Error("failed to update item", zap.Error(err))
		return nil, err
	}

	s.logger.Info("item updated successfully",
		zap.String("item_id", item.ID.String()),
	)

	return item, nil
}

// DeleteItem removes an item from a list
func (s *Service) DeleteItem(ctx context.Context, itemID, deletedBy uuid.UUID) error {
	s.logger.Info("deleting item",
		zap.String("item_id", itemID.String()),
	)

	// Get the item
	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return err
	}

	// Get the list to verify membership
	list, err := s.listRepo.GetByID(ctx, item.ListID)
	if err != nil {
		return err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, deletedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return err
	}
	if !isMember {
		return ErrNotGroupMember
	}

	// Check if archived
	if list.IsArchived {
		return ErrListArchived
	}

	if err := s.itemRepo.Delete(ctx, itemID); err != nil {
		s.logger.Error("failed to delete item", zap.Error(err))
		return err
	}

	s.logger.Info("item deleted successfully",
		zap.String("item_id", itemID.String()),
	)

	return nil
}

// TogglePurchasedInput contains the data needed to toggle purchased status
type TogglePurchasedInput struct {
	ItemID          uuid.UUID
	IsPurchased     bool
	ExpectedVersion int64
	UpdatedBy       uuid.UUID
}

// TogglePurchased toggles an item's purchased status
func (s *Service) TogglePurchased(ctx context.Context, input TogglePurchasedInput) (*Item, error) {
	s.logger.Info("toggling purchased status",
		zap.String("item_id", input.ItemID.String()),
		zap.Bool("is_purchased", input.IsPurchased),
	)

	// Get the item
	item, err := s.itemRepo.GetByID(ctx, input.ItemID)
	if err != nil {
		return nil, err
	}

	// Get the list to verify membership
	list, err := s.listRepo.GetByID(ctx, item.ListID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, input.UpdatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Update purchased status
	item.SetPurchased(input.IsPurchased, input.UpdatedBy)

	if err := s.itemRepo.Update(ctx, item, input.ExpectedVersion); err != nil {
		s.logger.Error("failed to toggle purchased", zap.Error(err))
		return nil, err
	}

	s.logger.Info("purchased status toggled successfully",
		zap.String("item_id", item.ID.String()),
	)

	return item, nil
}

// ReorderItems reorders items in a list
func (s *Service) ReorderItems(ctx context.Context, listID uuid.UUID, itemIDs []uuid.UUID, updatedBy uuid.UUID) ([]*Item, error) {
	s.logger.Info("reordering items",
		zap.String("list_id", listID.String()),
		zap.Int("item_count", len(itemIDs)),
	)

	// Get the list to verify membership
	list, err := s.listRepo.GetByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, updatedBy)
	if err != nil {
		s.logger.Error("failed to check membership", zap.Error(err))
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	// Check if archived
	if list.IsArchived {
		return nil, ErrListArchived
	}

	if err := s.itemRepo.ReorderItems(ctx, listID, itemIDs); err != nil {
		s.logger.Error("failed to reorder items", zap.Error(err))
		return nil, err
	}

	// Get updated items
	items, err := s.itemRepo.ListByListID(ctx, listID)
	if err != nil {
		s.logger.Error("failed to get items after reorder", zap.Error(err))
		return nil, err
	}

	s.logger.Info("items reordered successfully",
		zap.String("list_id", listID.String()),
	)

	return items, nil
}

// GetList retrieves a list by ID
func (s *Service) GetList(ctx context.Context, listID, requesterID uuid.UUID) (*List, error) {
	list, err := s.listRepo.GetByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return list, nil
}

// GetItems retrieves all items in a list
func (s *Service) GetItems(ctx context.Context, listID, requesterID uuid.UUID) ([]*Item, error) {
	// Get the list to verify membership
	list, err := s.listRepo.GetByID(ctx, listID)
	if err != nil {
		return nil, err
	}

	// Verify membership
	isMember, err := s.membershipChecker.IsMember(ctx, list.GroupID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return s.itemRepo.ListByListID(ctx, listID)
}
