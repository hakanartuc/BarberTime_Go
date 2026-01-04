package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"
	"time"

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
	BarberID        uint   `json:"barberId" binding:"required"`
	ServiceIDs      []uint `json:"serviceIds" binding:"required"`
	AppointmentTime string `json:"appointmentTime" binding:"required"` // ISO 8601 format: 2025-12-15T14:00:00Z
	Notes           string `json:"notes"`
}

func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// ISO 8601 formatında parse et (UTC)
	appointmentTime, err := time.Parse(time.RFC3339, req.AppointmentTime)
	if err != nil {
		utils.BadRequest(c, "Invalid appointmentTime format. Use ISO 8601 format: 2025-12-15T14:00:00Z")
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
		appointmentTime,
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

	utils.NoContentResponse(c, "Appointment cancelled successfully")
}

// ============================================
// Barber Appointment Handlers
// ============================================

// GetBarberPendingAppointments returns pending appointments for the barber
func (h *AppointmentHandler) GetBarberPendingAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	appointments, pagination, err := h.appointmentService.GetBarberPendingAppointments(userID.(uint), page, limit)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"appointments": appointments,
		"pagination":   pagination,
	})
}

// GetBarberUpcomingAppointments returns approved upcoming appointments for the barber
func (h *AppointmentHandler) GetBarberUpcomingAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	appointments, pagination, err := h.appointmentService.GetBarberUpcomingAppointments(userID.(uint), page, limit)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"appointments": appointments,
		"pagination":   pagination,
	})
}

// GetBarberPastAppointments returns past appointments for the barber
func (h *AppointmentHandler) GetBarberPastAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	appointments, pagination, err := h.appointmentService.GetBarberPastAppointments(userID.(uint), page, limit)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"appointments": appointments,
		"pagination":   pagination,
	})
}

// ApproveAppointment approves a pending appointment
func (h *AppointmentHandler) ApproveAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid appointment ID")
		return
	}

	if err := h.appointmentService.ApproveAppointment(uint(id), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Appointment approved successfully",
	})
}

// RejectAppointmentRequest represents the reject appointment request body
type RejectAppointmentRequest struct {
	Reason string `json:"reason"`
}

// RejectAppointment rejects an appointment
func (h *AppointmentHandler) RejectAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid appointment ID")
		return
	}

	var req RejectAppointmentRequest
	c.ShouldBindJSON(&req) // Opsiyonel body, hata olursa da devam et

	var reason *string
	if req.Reason != "" {
		reason = &req.Reason
	}

	if err := h.appointmentService.RejectAppointment(uint(id), userID.(uint), reason); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Appointment rejected successfully",
	})
}

// CompleteAppointment marks an appointment as completed
func (h *AppointmentHandler) CompleteAppointment(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid appointment ID")
		return
	}

	if err := h.appointmentService.CompleteAppointment(uint(id), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Appointment completed successfully",
	})
}
