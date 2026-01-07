package validation

import (
	"strings"

	"github.com/google/uuid"
)

// ValidateCategoryName validates a category name
func ValidateCategoryName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return &AuthValidationError{
			Field:   "name",
			Message: "category name is required",
		}
	}
	if len(name) > 50 {
		return &AuthValidationError{
			Field:   "name",
			Message: "category name must be 50 characters or less",
		}
	}
	return nil
}

// ValidateUpsertCategoryRequest validates an upsert category request
func ValidateUpsertCategoryRequest(groupID, categoryID, name string) error {
	if _, err := ValidateGroupID(groupID); err != nil {
		return err
	}
	// categoryID is optional for create
	if categoryID != "" {
		if _, err := ValidateCategoryIDRequired(categoryID); err != nil {
			return err
		}
	}
	if err := ValidateCategoryName(name); err != nil {
		return err
	}
	return nil
}

// ValidateCategoryIDRequired validates a required category ID
func ValidateCategoryIDRequired(categoryID string) (uuid.UUID, error) {
	if categoryID == "" {
		return uuid.Nil, &AuthValidationError{
			Field:   "category_id",
			Message: "category ID is required",
		}
	}
	id, err := uuid.Parse(categoryID)
	if err != nil {
		return uuid.Nil, &AuthValidationError{
			Field:   "category_id",
			Message: "invalid category ID format",
		}
	}
	return id, nil
}

// ValidateDeleteCategoryRequest validates a delete category request
func ValidateDeleteCategoryRequest(categoryID string) (uuid.UUID, error) {
	return ValidateCategoryIDRequired(categoryID)
}

// ValidateListCategoriesRequest validates a list categories request
func ValidateListCategoriesRequest(groupID string) (uuid.UUID, error) {
	return ValidateGroupID(groupID)
}
