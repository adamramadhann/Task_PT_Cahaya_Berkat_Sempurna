// Package middleware holds shared Gin middleware.
package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/task"
)

// ErrorHandler is the single place where errors become HTTP responses.
// Handlers only attach errors via c.Error(err); this middleware formats them
// into one consistent envelope so every endpoint responds the same way.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		var ve *task.ValidationError
		var status int
		var code string
		switch {
		case errors.As(err, &ve):
			status, code = http.StatusBadRequest, "VALIDATION_ERROR"
		case errors.Is(err, task.ErrNotFound):
			status, code = http.StatusNotFound, "NOT_FOUND"
		case errors.Is(err, task.ErrDuplicateTitle):
			status, code = http.StatusConflict, "DUPLICATE_TITLE"
		default:
			status, code = http.StatusInternalServerError, "INTERNAL_ERROR"
			log.Printf("[error] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		}

		body := gin.H{"code": code, "message": err.Error()}
		if ve != nil {
			body["fields"] = ve.Fields
		}
		c.AbortWithStatusJSON(status, gin.H{"error": body})
	}
}
