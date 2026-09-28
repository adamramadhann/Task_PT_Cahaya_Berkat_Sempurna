package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type handler struct {
	svc Service
}

// Register mounts the task routes on the given router group:
// GET/POST /tasks and PUT/DELETE /tasks/:id.
func Register(rg *gin.RouterGroup, svc Service) {
	h := &handler{svc: svc}
	rg.GET("/tasks", h.list)
	rg.POST("/tasks", h.create)
	rg.PUT("/tasks/:id", h.update)
	rg.DELETE("/tasks/:id", h.remove)
}

func (h *handler) list(c *gin.Context) {
	var q ListTasksQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(&ValidationError{Fields: map[string]string{"query": err.Error()}})
		return
	}
	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *handler) create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(&ValidationError{Fields: map[string]string{"body": err.Error()}})
		return
	}
	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, TaskData{Data: resp})
}

func (h *handler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(&ValidationError{Fields: map[string]string{"body": err.Error()}})
		return
	}
	resp, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, TaskData{Data: resp})
}

func (h *handler) remove(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		_ = c.Error(&ValidationError{Fields: map[string]string{"id": "must be a positive integer"}})
		return 0, false
	}
	return id, true
}
