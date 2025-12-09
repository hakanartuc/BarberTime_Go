package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	appointmentService service.AppointmentService
}

func NewAppointmentHandler(appointmentService service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
	}
}

func (h *AppointmentHandler) GetPastAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	appointments, pagination, err := h.appointmentService.GetPastAppointments(userID.(uint), page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"appointments": appointments,
		"pagination":   pagination,
	})
}

func (h *AppointmentHandler) GetUpcomingAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	appointments, pagination, err := h.appointmentService.GetUpcomingAppointments(userID.(uint), page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"appointments": appointments,
		"pagination":   pagination,
	})
}

func (h *AppointmentHandler) GetAppointment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid appointment ID")
		return
	}

	appointment, err := h.appointmentService.GetAppointmentByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, appointment)
}

type CreateAppointmentRequest struct {
	BarberID   uint   `json:"barberId" binding:"required"`
	ServiceIDs []uint `json:"serviceIds" binding:"required"`
	Date       string `json:"date" binding:"required"`
	Time       string `json:"time" binding:"required"`
	Notes      string `json:"notes"`
}

func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	appointment, err := h.appointmentService.CreateAppointment(
		userID.(uint),
		req.BarberID,
		req.ServiceIDs,
		req.Date,
		req.Time,
		notes,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, appointment)
}

func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid appointment ID")
		return
	}

	if err := h.appointmentService.CancelAppointment(uint(id), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

