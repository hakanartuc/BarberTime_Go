package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type TokenRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	DeviceID    string `json:"deviceId" binding:"required"`
}

func (h *AuthHandler) GetToken(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	requiresVerification, response, err := h.authService.Login(req.PhoneNumber, req.DeviceID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	if requiresVerification {
		// Verification gerekiyorsa sadece requiresVerification döndür
		c.JSON(http.StatusOK, gin.H{
			"requiresVerification": true,
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

type VerifyPhoneRequest struct {
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	DeviceID    string `json:"deviceId" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	var req VerifyPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	response, err := h.authService.VerifyPhone(req.PhoneNumber, req.DeviceID, req.Code)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}
