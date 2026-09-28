package task

import "errors"

// Domain errors. The HTTP layer maps these to status codes in one place
// (internal/middleware/error.go), so responses stay consistent.
var (
	ErrNotFound       = errors.New("task not found")
	ErrDuplicateTitle = errors.New("task title already exists")
)

// ValidationError carries per-field validation messages (mapped to 400).
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation error" }
