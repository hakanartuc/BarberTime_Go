package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PhotoLikeHandler struct {
	photoLikeService service.PhotoLikeService
}

func NewPhotoLikeHandler(photoLikeService service.PhotoLikeService) *PhotoLikeHandler {
	return &PhotoLikeHandler{
		photoLikeService: photoLikeService,
	}
}

func (h *PhotoLikeHandler) LikePhoto(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	if err := h.photoLikeService.LikePhoto(userID.(uint), uint(photoID)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"liked": true,
	})
}

func (h *PhotoLikeHandler) UnlikePhoto(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	if err := h.photoLikeService.UnlikePhoto(userID.(uint), uint(photoID)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"liked": false,
	})
}

func (h *PhotoLikeHandler) GetPhotoLikes(c *gin.Context) {
	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	likes, err := h.photoLikeService.GetPhotoLikes(uint(photoID))
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, likes)
}

func (h *PhotoLikeHandler) GetPhotoLikeCount(c *gin.Context) {
	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	count, err := h.photoLikeService.GetPhotoLikeCount(uint(photoID))
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"count": count,
	})
}

func (h *PhotoLikeHandler) IsLiked(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid photo ID")
		return
	}

	isLiked, err := h.photoLikeService.IsLiked(userID.(uint), uint(photoID))
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"isLiked": isLiked,
	})
}

