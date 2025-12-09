package handler

import (
	"barbertime_go/internal/service"
	"barbertime_go/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService service.SearchService
}

func NewSearchHandler(searchService service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	searchType := c.DefaultQuery("type", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var latitude, longitude *float64
	var serviceName *string
	var maxDistanceKM *float64

	if latStr := c.Query("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			latitude = &lat
		}
	}

	if lonStr := c.Query("longitude"); lonStr != "" {
		if lon, err := strconv.ParseFloat(lonStr, 64); err == nil {
			longitude = &lon
		}
	}

	if serviceNameStr := c.Query("service"); serviceNameStr != "" {
		serviceNameStr = strings.TrimSpace(serviceNameStr)
		if serviceNameStr != "" {
			serviceName = &serviceNameStr
		}
	}

	if kmStr := c.Query("km"); kmStr != "" {
		if km, err := strconv.ParseFloat(kmStr, 64); err == nil && km > 0 {
			maxDistanceKM = &km
		}
	}

	results, pagination, err := h.searchService.Search(query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	response := gin.H{
		"results":    results,
		"pagination": pagination,
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}
