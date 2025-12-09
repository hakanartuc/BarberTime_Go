package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BarberShopRequestHandler struct {
	requestService       service.BarberShopRequestService
	barberProfileService service.BarberProfileService
}

func NewBarberShopRequestHandler(
	requestService service.BarberShopRequestService,
	barberProfileService service.BarberProfileService,
) *BarberShopRequestHandler {
	return &BarberShopRequestHandler{
		requestService:       requestService,
		barberProfileService: barberProfileService,
	}
}

type CreateBarberShopRequest struct {
	ShopID  uint   `json:"shopId" binding:"required"`
	Message string `json:"message"`
}

func (h *BarberShopRequestHandler) CreateRequest(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req CreateBarberShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// Barber profilini al
	barber, err := h.barberProfileService.GetBarberProfile(userID.(uint))
	if err != nil {
		utils.BadRequest(c, "User is not registered as barber")
		return
	}

	request, err := h.requestService.CreateRequest(barber.ID, req.ShopID, req.Message)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, request)
}

func (h *BarberShopRequestHandler) GetRequestsByShop(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	requests, err := h.requestService.GetRequestsByShop(uint(id), userID.(uint))
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, requests)
}

func (h *BarberShopRequestHandler) GetRequestsByBarber(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Barber profilini al
	barber, err := h.barberProfileService.GetBarberProfile(userID.(uint))
	if err != nil {
		utils.BadRequest(c, "User is not registered as barber")
		return
	}

	requests, err := h.requestService.GetRequestsByBarber(barber.ID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, requests)
}

type RejectRequestRequest struct {
	Reason string `json:"reason"`
}

func (h *BarberShopRequestHandler) ApproveRequest(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	requestID, err := strconv.ParseUint(c.Param("requestId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid request ID")
		return
	}

	if err := h.requestService.ApproveRequest(uint(requestID), uint(shopID), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Request approved successfully",
	})
}

func (h *BarberShopRequestHandler) RejectRequest(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	requestID, err := strconv.ParseUint(c.Param("requestId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid request ID")
		return
	}

	var req RejectRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.requestService.RejectRequest(uint(requestID), uint(shopID), userID.(uint), req.Reason); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Request rejected successfully",
	})
}
