package validation

import (
	"strings"

	"github.com/google/uuid"
)

// ValidateListName validates a list name
func ValidateListName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return &AuthValidationError{
			Field:   "name",
			Message: "list name is required",
		}
	}
	if len(name) > 100 {
		return &AuthValidationError{
			Field:   "name",
			Message: "list name must be 100 characters or less",
		}
	}
	return nil
}

// ValidateListDescription validates a list description
func ValidateListDescription(description string) error {
	if len(description) > 500 {
		return &AuthValidationError{
			Field:   "description",
			Message: "list description must be 500 characters or less",
		}
	}
	return nil
}

// ValidateCreateListRequest validates a create list request
func ValidateCreateListRequest(groupID, name, description string) error {
	if _, err := ValidateGroupID(groupID); err != nil {
		return err
	}
	if err := ValidateListName(name); err != nil {
		return err
	}
	if err := ValidateListDescription(description); err != nil {
		return err
	}
	return nil
}

// ValidateListID validates a list ID
func ValidateListID(listID string) (uuid.UUID, error) {
	if listID == "" {
		return uuid.Nil, &AuthValidationError{
			Field:   "list_id",
			Message: "list ID is required",
		}
	}
	id, err := uuid.Parse(listID)
	if err != nil {
		return uuid.Nil, &AuthValidationError{
			Field:   "list_id",
			Message: "invalid list ID format",
		}
	}
	return id, nil
}

// ValidateItemName validates an item name
func ValidateItemName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return &AuthValidationError{
			Field:   "name",
			Message: "item name is required",
		}
	}
	if len(name) > 200 {
		return &AuthValidationError{
			Field:   "name",
			Message: "item name must be 200 characters or less",
		}
	}
	return nil
}

// ValidateItemQuantity validates an item quantity
func ValidateItemQuantity(quantity string) error {
	if len(quantity) > 50 {
		return &AuthValidationError{
			Field:   "quantity",
			Message: "quantity must be 50 characters or less",
		}
	}
	return nil
}

// ValidateItemNotes validates item notes
func ValidateItemNotes(notes string) error {
	if len(notes) > 1000 {
		return &AuthValidationError{
			Field:   "notes",
			Message: "notes must be 1000 characters or less",
		}
	}
	return nil
}

// ValidateItemID validates an item ID
func ValidateItemID(itemID string) (uuid.UUID, error) {
	if itemID == "" {
		return uuid.Nil, &AuthValidationError{
			Field:   "item_id",
			Message: "item ID is required",
		}
	}
	id, err := uuid.Parse(itemID)
	if err != nil {
		return uuid.Nil, &AuthValidationError{
			Field:   "item_id",
			Message: "invalid item ID format",
		}
	}
	return id, nil
}

// ValidateAddItemRequest validates an add item request
func ValidateAddItemRequest(listID, name, quantity, notes string) error {
	if _, err := ValidateListID(listID); err != nil {
		return err
	}
	if err := ValidateItemName(name); err != nil {
		return err
	}
	if err := ValidateItemQuantity(quantity); err != nil {
		return err
	}
	if err := ValidateItemNotes(notes); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateItemRequest validates an update item request
func ValidateUpdateItemRequest(itemID, name, quantity, notes string) error {
	if _, err := ValidateItemID(itemID); err != nil {
		return err
	}
	if name != "" {
		if err := ValidateItemName(name); err != nil {
			return err
		}
	}
	if err := ValidateItemQuantity(quantity); err != nil {
		return err
	}
	if err := ValidateItemNotes(notes); err != nil {
		return err
	}
	return nil
}

// ValidateCategoryID validates a category ID (optional)
func ValidateCategoryID(categoryID string) (*uuid.UUID, error) {
	if categoryID == "" {
		return nil, nil // Optional field
	}
	id, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, &AuthValidationError{
			Field:   "category_id",
			Message: "invalid category ID format",
		}
	}
	return &id, nil
}
