package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	shopService service.ShopService
}

func NewShopHandler(shopService service.ShopService) *ShopHandler {
	return &ShopHandler{
		shopService: shopService,
	}
}

func (h *ShopHandler) GetNearbyShops(c *gin.Context) {
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

	shops, pagination, err := h.shopService.GetNearbyShops(latitude, longitude, page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"shops":     shops,
		"pagination": pagination,
	})
}

func (h *ShopHandler) GetShop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	shop, err := h.shopService.GetShopByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, shop)
}

func (h *ShopHandler) GetShopBarbers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	barbers, err := h.shopService.GetShopBarbers(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, barbers)
}

func (h *ShopHandler) GetShopReviews(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reviews, pagination, avgRating, totalReviews, err := h.shopService.GetShopReviews(uint(id), page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"reviews":       reviews,
		"averageRating": avgRating,
		"totalReviews":  totalReviews,
		"pagination":    pagination,
	})
}

