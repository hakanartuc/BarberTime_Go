package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

func (h *AdminHandler) GetPendingShops(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	shops, total, err := h.adminService.GetPendingShops(page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"data": shops,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *AdminHandler) ApproveShop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	if err := h.adminService.ApproveShop(uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Shop approved successfully",
	})
}

func (h *AdminHandler) RejectShop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	if err := h.adminService.RejectShop(uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Shop rejected successfully",
	})
}

type CreateServiceRequest struct {
	Name  string  `json:"name" binding:"required"`
	Image *string `json:"image"`
}

func (h *AdminHandler) CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	service, err := h.adminService.CreateService(req.Name, req.Image)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, service)
}

type UpdateServiceRequest struct {
	Name  *string `json:"name"`
	Image *string `json:"image"`
}

func (h *AdminHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid service ID")
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	service, err := h.adminService.UpdateService(uint(id), req.Name, req.Image)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, service)
}

func (h *AdminHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid service ID")
		return
	}

	if err := h.adminService.DeleteService(uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Service deleted successfully")
}
