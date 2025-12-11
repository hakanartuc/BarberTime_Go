package handler

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

func (h *ReviewHandler) GetReviews(c *gin.Context) {
	reviewTypeStr := c.Query("type")
	targetIDStr := c.Query("targetId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var reviewType *models.ReviewType
	var targetID *uint

	if reviewTypeStr != "" {
		rt := models.ReviewType(reviewTypeStr)
		reviewType = &rt
	}

	if targetIDStr != "" {
		id, err := strconv.ParseUint(targetIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			targetID = &uid
		}
	}

	reviews, pagination, err := h.reviewService.GetReviews(reviewType, targetID, page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"reviews":    reviews,
		"pagination": pagination,
	})
}

type CreateReviewRequest struct {
	Type         string `json:"type" binding:"required"`
	TargetID     uint   `json:"targetId" binding:"required"`
	Rating       int    `json:"rating" binding:"required"`
	Comment      string `json:"comment"`
	AppointmentID *uint `json:"appointmentId"`
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	review, err := h.reviewService.CreateReview(
		userID.(uint),
		models.ReviewType(req.Type),
		req.TargetID,
		req.Rating,
		req.Comment,
		req.AppointmentID,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, review)
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid review ID")
		return
	}

	if err := h.reviewService.DeleteReview(uint(id), userID.(uint)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Review deleted successfully")
}

