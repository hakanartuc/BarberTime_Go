package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ShopProfileHandler struct {
	shopProfileService service.ShopProfileService
}

func NewShopProfileHandler(shopProfileService service.ShopProfileService) *ShopProfileHandler {
	return &ShopProfileHandler{
		shopProfileService: shopProfileService,
	}
}

type CreateShopRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	PhoneNumber  string  `json:"phoneNumber"`
	Email        string  `json:"email"`
	Website      string  `json:"website"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	WorkingHours string  `json:"workingHours"`
	ClosedDays   string  `json:"closedDays"`
}

func (h *ShopProfileHandler) CreateShop(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	shop, err := h.shopProfileService.CreateShop(
		userID.(uint),
		req.Name,
		req.Description,
		req.PhoneNumber,
		req.Email,
		req.Website,
		req.Latitude,
		req.Longitude,
		req.WorkingHours,
		req.ClosedDays,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, shop)
}

func (h *ShopProfileHandler) GetShopProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	shop, err := h.shopProfileService.GetShopProfile(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, shop)
}

type UpdateShopProfileRequest struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Banner       *string  `json:"banner"`
	PhoneNumber  *string  `json:"phoneNumber"`
	Email        *string  `json:"email"`
	Website      *string  `json:"website"`
	WorkingHours *string  `json:"workingHours"`
	ClosedDays   *string  `json:"closedDays"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

func (h *ShopProfileHandler) UpdateShopProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	var req UpdateShopProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	shop, err := h.shopProfileService.UpdateShopProfile(
		uint(id),
		userID.(uint),
		req.Name,
		req.Description,
		req.Banner,
		req.PhoneNumber,
		req.Email,
		req.Website,
		req.WorkingHours,
		req.ClosedDays,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, shop)
}

func (h *ShopProfileHandler) GetShopOwners(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	owners, err := h.shopProfileService.GetShopOwners(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, owners)
}

type AddShopOwnerRequest struct {
	OwnerUserID uint `json:"ownerUserId" binding:"required"`
}

func (h *ShopProfileHandler) AddShopOwner(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	var req AddShopOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.shopProfileService.AddShopOwner(uint(id), req.OwnerUserID, userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Owner added successfully",
	})
}

func (h *ShopProfileHandler) RemoveShopOwner(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	ownerID, err := strconv.ParseUint(c.Param("ownerId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid owner ID")
		return
	}

	if err := h.shopProfileService.RemoveShopOwner(uint(shopID), uint(ownerID), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Shop owner removed successfully")
}

type CreateShopServiceRequest struct {
	ServiceID   uint    `json:"serviceId" binding:"required"`
	Duration    int     `json:"duration" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Description *string `json:"description"`
}

func (h *ShopProfileHandler) CreateShopService(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	var req CreateShopServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	shopService, err := h.shopProfileService.CreateShopService(uint(shopID), userID.(uint), req.ServiceID, req.Duration, req.Price, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, shopService)
}

type UpdateShopServiceRequest struct {
	Duration    int     `json:"duration" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Description *string `json:"description"`
}

func (h *ShopProfileHandler) UpdateShopService(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	shopServiceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop service ID")
		return
	}

	var req UpdateShopServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	shopService, err := h.shopProfileService.UpdateShopService(uint(shopID), userID.(uint), uint(shopServiceID), req.Duration, req.Price, req.Description)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, shopService)
}

func (h *ShopProfileHandler) DeleteShopService(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	shopServiceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop service ID")
		return
	}

	if err := h.shopProfileService.DeleteShopService(uint(shopID), userID.(uint), uint(shopServiceID)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Shop service deleted successfully")
}

// ==================== Closed Days ====================

func (h *ShopProfileHandler) GetClosedDays(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	closedDays, err := h.shopProfileService.GetClosedDays(uint(shopID))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, closedDays)
}

type CreateClosedDayRequest struct {
	Date   string `json:"date" binding:"required"`
	Reason string `json:"reason"`
}

func (h *ShopProfileHandler) CreateClosedDay(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	var req CreateClosedDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	closedDay, err := h.shopProfileService.CreateClosedDay(uint(shopID), userID.(uint), req.Date, req.Reason)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, closedDay)
}

type UpdateClosedDayRequest struct {
	Date   *string `json:"date"`
	Reason *string `json:"reason"`
}

func (h *ShopProfileHandler) UpdateClosedDay(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	closedDayID, err := strconv.ParseUint(c.Param("closedDayId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid closed day ID")
		return
	}

	var req UpdateClosedDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	closedDay, err := h.shopProfileService.UpdateClosedDay(uint(shopID), userID.(uint), uint(closedDayID), req.Date, req.Reason)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, closedDay)
}

func (h *ShopProfileHandler) DeleteClosedDay(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	closedDayID, err := strconv.ParseUint(c.Param("closedDayId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid closed day ID")
		return
	}

	if err := h.shopProfileService.DeleteClosedDay(uint(shopID), userID.(uint), uint(closedDayID)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Closed day deleted successfully")
}

type ToggleClosedDayRequest struct {
	IsActive bool `json:"isActive"`
}

func (h *ShopProfileHandler) ToggleClosedDay(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	closedDayID, err := strconv.ParseUint(c.Param("closedDayId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid closed day ID")
		return
	}

	var req ToggleClosedDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	closedDay, err := h.shopProfileService.ToggleClosedDay(uint(shopID), userID.(uint), uint(closedDayID), req.IsActive)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, closedDay)
}

// ==================== Working Hours ====================

func (h *ShopProfileHandler) GetWorkingHours(c *gin.Context) {
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	workingHours, err := h.shopProfileService.GetWorkingHours(uint(shopID))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, workingHours)
}

type UpdateWorkingHoursRequest struct {
	OpenTime  string `json:"openTime" binding:"required"`
	CloseTime string `json:"closeTime" binding:"required"`
}

func (h *ShopProfileHandler) UpdateWorkingHours(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	var req UpdateWorkingHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	workingHours, err := h.shopProfileService.UpdateWorkingHours(uint(shopID), userID.(uint), req.OpenTime, req.CloseTime)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, workingHours)
}

// ==================== System Status ====================

type UpdateShopSystemStatusRequest struct {
	IsSystemOff bool `json:"isSystemOff"`
}

func (h *ShopProfileHandler) UpdateSystemStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	shopID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	var req UpdateShopSystemStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	shop, err := h.shopProfileService.UpdateSystemStatus(uint(shopID), userID.(uint), req.IsSystemOff)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, shop)
}
