package task

import "time"

// Status values allowed for a task.
const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

// Task is the database entity.
type Task struct {
	ID          int64
	Title       string
	Description *string
	Status      string
	Assignee    *string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// IsValidStatus reports whether s is one of the allowed status values.
func IsValidStatus(s string) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	}
	return false
}
