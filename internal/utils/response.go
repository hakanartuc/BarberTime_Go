package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse tüm endpoint'ler için standart response formatı
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SuccessResponse başarılı response döndürür
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMessage mesaj ile başarılı response döndürür
func SuccessResponseWithMessage(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// ErrorResponseJSON hata response'u döndürür
func ErrorResponseJSON(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, StandardResponse{
		Success: false,
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

// NoContentResponse başarılı silme işlemleri için standart response döndürür
func NoContentResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Resource deleted successfully"
	}
	SuccessResponseWithMessage(c, http.StatusOK, nil, message)
}
