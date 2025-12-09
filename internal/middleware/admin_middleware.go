package middleware

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/utils"
	"barbertime_go/pkg/database"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			utils.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// User'ı kontrol et
		var user models.User
		err := database.DB.First(&user, userID).Error

		if err != nil || !user.IsAdmin {
			utils.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
