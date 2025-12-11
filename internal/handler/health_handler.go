package handler

import (
	"barbertime_go/internal/utils"
	"barbertime_go/pkg/database"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type HealthData struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	checks := make(map[string]string)
	overallStatus := "healthy"

	// PostgreSQL bağlantısını kontrol et
	if database.DB == nil {
		checks["database"] = "unhealthy - connection is nil"
		overallStatus = "unhealthy"
	} else {
		sqlDB, err := database.DB.DB()
		if err != nil {
			checks["database"] = "unhealthy - " + err.Error()
			overallStatus = "unhealthy"
		} else {
			// Bağlantıyı test et
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			// Ping ile bağlantıyı kontrol et
			if err := sqlDB.PingContext(ctx); err != nil {
				checks["database"] = "unhealthy - " + err.Error()
				overallStatus = "unhealthy"
			} else {
				// Bağlantı istatistiklerini al
				stats := sqlDB.Stats()
				checks["database"] = "healthy"
				checks["database_open_connections"] = fmt.Sprintf("%d", stats.OpenConnections)
				checks["database_in_use"] = fmt.Sprintf("%d", stats.InUse)
			}
		}
	}

	// HTTP status code belirle
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	healthData := HealthData{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	utils.SuccessResponse(c, statusCode, healthData)
}

