package handler

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	favoriteService service.FavoriteService
}

func NewFavoriteHandler(favoriteService service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteService: favoriteService,
	}
}

func (h *FavoriteHandler) GetFavorites(c *gin.Context) {
	userID, _ := c.Get("userID")
	favoriteTypeStr := c.Query("type")

	var favoriteType *models.FavoriteType
	if favoriteTypeStr != "" {
		ft := models.FavoriteType(favoriteTypeStr)
		favoriteType = &ft
	}

	favorites, err := h.favoriteService.GetFavorites(userID.(uint), favoriteType)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, favorites)
}

func (h *FavoriteHandler) GetFavoriteShops(c *gin.Context) {
	userID, _ := c.Get("userID")
	favoriteType := models.FavoriteTypeShop

	favorites, err := h.favoriteService.GetFavorites(userID.(uint), &favoriteType)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, favorites)
}

func (h *FavoriteHandler) GetFavoriteBarbers(c *gin.Context) {
	userID, _ := c.Get("userID")
	favoriteType := models.FavoriteTypeBarber

	favorites, err := h.favoriteService.GetFavorites(userID.(uint), &favoriteType)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, favorites)
}

func (h *FavoriteHandler) AddFavoriteShop(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	if err := h.favoriteService.AddFavorite(userID.(uint), models.FavoriteTypeShop, uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"isFavorite": true,
	})
}

func (h *FavoriteHandler) RemoveFavoriteShop(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid shop ID")
		return
	}

	if err := h.favoriteService.RemoveFavorite(userID.(uint), models.FavoriteTypeShop, uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Favorite shop removed successfully")
}

func (h *FavoriteHandler) AddFavoriteBarber(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid barber ID")
		return
	}

	if err := h.favoriteService.AddFavorite(userID.(uint), models.FavoriteTypeBarber, uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"isFavorite": true,
	})
}

func (h *FavoriteHandler) RemoveFavoriteBarber(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid barber ID")
		return
	}

	if err := h.favoriteService.RemoveFavorite(userID.(uint), models.FavoriteTypeBarber, uint(id)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.NoContentResponse(c, "Favorite barber removed successfully")
}

