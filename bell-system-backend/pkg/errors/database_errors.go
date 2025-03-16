package errors

import (
	"errors"
	"fmt"
)

// Database error types
var (
	// ErrNotFound is returned when a requested entity is not found
	ErrNotFound = errors.New("entity not found")

	// ErrDuplicateEntry is returned when attempting to create a duplicate entry
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrInvalidData is returned when data validation fails
	ErrInvalidData = errors.New("invalid data")

	ErrForeignKeyViolation = errors.New("foreign key violation")

	// ErrTransactionFailed is returned when a transaction operation fails
	ErrTransactionFailed = errors.New("transaction failed")
)

// NotFoundError wraps ErrNotFound with context about which entity wasn't found
type NotFoundError struct {
	EntityType string
	ID         string
}

// Error implements the error interface for NotFoundError
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.EntityType, e.ID)
}

// Is allows errors.Is to match NotFoundError with ErrNotFound
func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(entityType, id string) *NotFoundError {
	return &NotFoundError{
		EntityType: entityType,
		ID:         id,
	}
}

// DatabaseError wraps database errors with additional context
type DatabaseError struct {
	Operation string
	Entity    string
	Err       error
}

// Error implements the error interface for DatabaseError
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s operation on %s: %v", e.Operation, e.Entity, e.Err)
}

// Unwrap returns the underlying error for errors.Is and errors.As
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(operation, entity string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Entity:    entity,
		Err:       err,
	}
}
