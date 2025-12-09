package middleware

import (
	"barbertime_go/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Hata varsa işle
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Error: %v", err)

			// Hata tipine göre response döndür
			utils.InternalServerError(c, err.Error())
		}
	}
}
