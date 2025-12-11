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
