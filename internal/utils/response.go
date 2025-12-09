package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func ErrorResponseJSON(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	ErrorResponseJSON(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	ErrorResponseJSON(c, http.StatusUnauthorized, message)
}

func NotFound(c *gin.Context, message string) {
	ErrorResponseJSON(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	ErrorResponseJSON(c, http.StatusInternalServerError, message)
}

func Forbidden(c *gin.Context, message string) {
	ErrorResponseJSON(c, http.StatusForbidden, message)
}
