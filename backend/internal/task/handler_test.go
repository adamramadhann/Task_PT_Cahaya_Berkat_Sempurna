// Package task_test exercises the HTTP layer. It is an external test package
// because the error-envelope middleware imports this package's domain errors.
package task_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/middleware"
	"task-management-backend/internal/task"
)

type fakeService struct {
	listFn   func(context.Context, task.ListTasksQuery) (*task.ListResponse, error)
	createFn func(context.Context, task.CreateTaskRequest) (task.TaskResponse, error)
	updateFn func(context.Context, int64, task.UpdateTaskRequest) (task.TaskResponse, error)
	deleteFn func(context.Context, int64) error
}

func (f *fakeService) List(ctx context.Context, q task.ListTasksQuery) (*task.ListResponse, error) {
	return f.listFn(ctx, q)
}

func (f *fakeService) Create(ctx context.Context, req task.CreateTaskRequest) (task.TaskResponse, error) {
	return f.createFn(ctx, req)
}

func (f *fakeService) Update(ctx context.Context, id int64, req task.UpdateTaskRequest) (task.TaskResponse, error) {
	return f.updateFn(ctx, id, req)
}

func (f *fakeService) Delete(ctx context.Context, id int64) error {
	return f.deleteFn(ctx, id)
}

func newTestRouter(svc task.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	task.Register(r.Group("/api"), svc)
	return r
}

// Task 4: duplicate title must return HTTP 409 with the shared error envelope,
// never a 500.
func TestCreateDuplicateTitleReturns409(t *testing.T) {
	svc := &fakeService{createFn: func(context.Context, task.CreateTaskRequest) (task.TaskResponse, error) {
		return task.TaskResponse{}, task.ErrDuplicateTitle
	}}
	r := newTestRouter(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":"dup"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", w.Code, w.Body.String())
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not the error envelope: %v (%s)", err, w.Body.String())
	}
	if body.Error.Code != "DUPLICATE_TITLE" {
		t.Fatalf("error code = %q, want DUPLICATE_TITLE", body.Error.Code)
	}
}
