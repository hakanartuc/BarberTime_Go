package handler

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PhotoHandler struct {
	photoService service.PhotoService
}

func NewPhotoHandler(photoService service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
	}
}

type UploadPhotoRequest struct {
	Caption string `form:"caption"`
}

func (h *PhotoHandler) UploadPhoto(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	// Target type ve ID'yi al
	targetTypeStr := c.Param("type") // "barber" veya "shop"
	targetIDStr := c.Param("id")

	var targetType models.PhotoType
	if targetTypeStr == "barber" {
		targetType = models.PhotoTypeBarber
	} else if targetTypeStr == "shop" {
		targetType = models.PhotoTypeShop
	} else {
		utils.BadRequest(c, "Invalid target type. Must be 'barber' or 'shop'")
		return
	}

	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid target ID")
		return
	}

	// Dosyayı al
	file, err := c.FormFile("photo")
	if err != nil {
		utils.BadRequest(c, "photo file is required")
		return
	}

	// Caption'ı al
	caption := c.PostForm("caption")

	photo, err := h.photoService.UploadPhoto(targetType, uint(targetID), userID.(uint), file, caption)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, photo)
}

func (h *PhotoHandler) GetPhotos(c *gin.Context) {
	targetTypeStr := c.Param("type")
	targetIDStr := c.Param("id")

	var targetType models.PhotoType
	if targetTypeStr == "barber" {
		targetType = models.PhotoTypeBarber
	} else if targetTypeStr == "shop" {
		targetType = models.PhotoTypeShop
	} else {
		utils.BadRequest(c, "Invalid target type. Must be 'barber' or 'shop'")
		return
	}

	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid target ID")
		return
	}

	photos, err := h.photoService.GetPhotos(targetType, uint(targetID))
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, photos)
}

func (h *PhotoHandler) GetPhoto(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	photo, err := h.photoService.GetPhoto(uint(id))
	if err != nil {
		utils.NotFound(c, "Photo not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, photo)
}

func (h *PhotoHandler) DeletePhoto(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	if err := h.photoService.DeletePhoto(uint(id), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Photo deleted successfully")
}

