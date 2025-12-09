package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BarberProfileHandler struct {
	barberProfileService service.BarberProfileService
}

func NewBarberProfileHandler(barberProfileService service.BarberProfileService) *BarberProfileHandler {
	return &BarberProfileHandler{
		barberProfileService: barberProfileService,
	}
}

type RegisterBarberRequest struct {
	Name       string `json:"name" binding:"required"`
	Bio        string `json:"bio"`
	Experience int    `json:"experience"`
}

func (h *BarberProfileHandler) RegisterAsBarber(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req RegisterBarberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	barber, err := h.barberProfileService.RegisterAsBarber(
		userID.(uint),
		req.Name,
		req.Bio,
		req.Experience,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, barber)
}

func (h *BarberProfileHandler) GetBarberProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	barber, err := h.barberProfileService.GetBarberProfile(userID.(uint))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, barber)
}

type UpdateBarberProfileRequest struct {
	Name       *string  `json:"name"`
	Bio        *string  `json:"bio"`
	Experience *int     `json:"experience"`
	Avatar     *string  `json:"avatar"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
}

func (h *BarberProfileHandler) UpdateBarberProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req UpdateBarberProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	barber, err := h.barberProfileService.UpdateBarberProfile(
		userID.(uint),
		req.Name,
		req.Bio,
		req.Experience,
		req.Avatar,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, barber)
}
