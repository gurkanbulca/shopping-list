package validation

import (
	"github.com/google/uuid"
)

// ValidateGetDeltaRequest validates a get delta request
func ValidateGetDeltaRequest(groupID string, cursor int64, maxChanges int32) (uuid.UUID, error) {
	gid, err := ValidateGroupID(groupID)
	if err != nil {
		return uuid.Nil, err
	}

	if cursor < 0 {
		return uuid.Nil, &AuthValidationError{
			Field:   "cursor",
			Message: "cursor must be non-negative",
		}
	}

	if maxChanges < 0 {
		return uuid.Nil, &AuthValidationError{
			Field:   "max_changes",
			Message: "max_changes must be non-negative",
		}
	}

	return gid, nil
}

// ValidatePushMutationsRequest validates a push mutations request
func ValidatePushMutationsRequest(groupID string, mutationsCount int) (uuid.UUID, error) {
	gid, err := ValidateGroupID(groupID)
	if err != nil {
		return uuid.Nil, err
	}

	if mutationsCount == 0 {
		return uuid.Nil, &AuthValidationError{
			Field:   "mutations",
			Message: "at least one mutation is required",
		}
	}

	if mutationsCount > 100 {
		return uuid.Nil, &AuthValidationError{
			Field:   "mutations",
			Message: "maximum 100 mutations per request",
		}
	}

	return gid, nil
}

// ValidateMutationID validates a mutation ID
func ValidateMutationID(mutationID string) (uuid.UUID, error) {
	if mutationID == "" {
		return uuid.Nil, &AuthValidationError{
			Field:   "mutation_id",
			Message: "mutation ID is required",
		}
	}
	id, err := uuid.Parse(mutationID)
	if err != nil {
		return uuid.Nil, &AuthValidationError{
			Field:   "mutation_id",
			Message: "invalid mutation ID format",
		}
	}
	return id, nil
}
