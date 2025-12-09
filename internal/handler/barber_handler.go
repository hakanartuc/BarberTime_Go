package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BarberHandler struct {
	barberService service.BarberService
}

func NewBarberHandler(barberService service.BarberService) *BarberHandler {
	return &BarberHandler{
		barberService: barberService,
	}
}

func (h *BarberHandler) GetNearbyBarbers(c *gin.Context) {
	latitude, err := strconv.ParseFloat(c.Query("latitude"), 64)
	if err != nil {
		utils.BadRequest(c, "Invalid latitude")
		return
	}

	longitude, err := strconv.ParseFloat(c.Query("longitude"), 64)
	if err != nil {
		utils.BadRequest(c, "Invalid longitude")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	barbers, pagination, err := h.barberService.GetNearbyBarbers(latitude, longitude, page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"barbers":   barbers,
		"pagination": pagination,
	})
}

func (h *BarberHandler) GetBarber(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid barber ID")
		return
	}

	barber, err := h.barberService.GetBarberByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, barber)
}

func (h *BarberHandler) GetAvailability(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid barber ID")
		return
	}

	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")

	if startDate == "" || endDate == "" {
		utils.BadRequest(c, "startDate and endDate are required")
		return
	}

	availability, err := h.barberService.GetAvailability(uint(id), startDate, endDate)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"availability": availability,
	})
}

