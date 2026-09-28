package task

import "time"

// CreateTaskRequest is the body of POST /api/tasks.
type CreateTaskRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Description *string `json:"description"`
	Status      string  `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Assignee    *string `json:"assignee" binding:"omitempty,max=255"`
	DueDate     *string `json:"due_date" binding:"omitempty,datetime=2006-01-02"`
}

// UpdateTaskRequest is the body of PUT /api/tasks/:id (full update).
type UpdateTaskRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Description *string `json:"description"`
	Status      string  `json:"status" binding:"required,oneof=todo in_progress done"`
	Assignee    *string `json:"assignee" binding:"omitempty,max=255"`
	DueDate     *string `json:"due_date" binding:"omitempty,datetime=2006-01-02"`
}

// ListTasksQuery is the query string of GET /api/tasks.
// Zero values are normalized by the service (defaults for page/limit/sort).
type ListTasksQuery struct {
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Assignee string `form:"assignee"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Sort     string `form:"sort"` // "field:asc|desc"
}

// TaskResponse includes every editable field, so the frontend edit modal can
// be filled straight from a list item without a separate detail endpoint.
type TaskResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	Assignee    *string   `json:"assignee"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskData struct {
	Data TaskResponse `json:"data"`
}

type ListMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type ListResponse struct {
	Data []TaskResponse `json:"data"`
	Meta ListMeta       `json:"meta"`
}

func newTaskResponse(t Task) TaskResponse {
	resp := TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Assignee:    t.Assignee,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DueDate != nil {
		d := t.DueDate.Format("2006-01-02")
		resp.DueDate = &d
	}
	return resp
}
