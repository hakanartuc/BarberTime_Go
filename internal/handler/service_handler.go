package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	serviceService     service.ServiceService
	shopServiceService service.ShopServiceService
}

func NewServiceHandler(serviceService service.ServiceService, shopServiceService service.ShopServiceService) *ServiceHandler {
	return &ServiceHandler{
		serviceService:     serviceService,
		shopServiceService: shopServiceService,
	}
}

// GetAllServices Tüm servisleri getirir (admin için)
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	services, err := h.serviceService.GetAllServices()
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, services)
}

// GetShopServices Shop'un servislerini getirir
func (h *ServiceHandler) GetShopServices(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	shopServices, err := h.shopServiceService.GetShopServices(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, shopServices)
}
