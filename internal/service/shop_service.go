package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"errors"
	"fmt"
	"time"
)

type ShopService interface {
	GetNearbyShops(latitude, longitude float64, page, limit int) ([]models.Shop, utils.Pagination, error)
	GetShopByID(id uint) (*models.Shop, error)
	GetShopBarbers(shopID uint) ([]models.Barber, error)
	GetShopReviews(shopID uint, page, limit int) ([]models.Review, utils.Pagination, float64, int, error)
}

type shopService struct {
	shopRepo     repository.ShopRepository
	barberRepo   repository.BarberRepository
	reviewRepo   repository.ReviewRepository
	cacheService CacheService
}

func NewShopService(
	shopRepo repository.ShopRepository,
	barberRepo repository.BarberRepository,
	reviewRepo repository.ReviewRepository,
	cacheService CacheService,
) ShopService {
	return &shopService{
		shopRepo:     shopRepo,
		barberRepo:   barberRepo,
		reviewRepo:   reviewRepo,
		cacheService: cacheService,
	}
}

func (s *shopService) GetNearbyShops(latitude, longitude float64, page, limit int) ([]models.Shop, utils.Pagination, error) {
	shops, total, err := s.shopRepo.GetNearby(latitude, longitude, nil, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return shops, pagination, nil
}

func (s *shopService) GetShopByID(id uint) (*models.Shop, error) {
	// Önce cache'den kontrol et
	cacheKey := fmt.Sprintf("shop:%d", id)
	var shop models.Shop
	if err := s.cacheService.Get(cacheKey, &shop); err == nil {
		return &shop, nil
	}

	// Cache'de yoksa DB'den al
	shopPtr, err := s.shopRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	// Cache'e kaydet (1 saat TTL)
	s.cacheService.Set(cacheKey, shopPtr, 1*time.Hour)

	return shopPtr, nil
}

func (s *shopService) SetShopCache(shopID uint, shop *models.Shop) {
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Set(cacheKey, shop, 1*time.Hour)
}

func (s *shopService) DeleteShopCache(shopID uint) {
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)
}

func (s *shopService) GetShopBarbers(shopID uint) ([]models.Barber, error) {
	barbers, err := s.barberRepo.GetByShopID(shopID)
	if err != nil {
		return nil, err
	}
	return barbers, nil
}

func (s *shopService) GetShopReviews(shopID uint, page, limit int) ([]models.Review, utils.Pagination, float64, int, error) {
	reviews, total, avgRating, err := s.reviewRepo.GetByTypeAndTargetID(models.ReviewTypeShop, shopID, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, 0, 0, err
	}

	pagination := utils.NewPagination(page, limit, total)

	// Toplam yorum sayısını al
	_, totalReviews, _ := s.reviewRepo.GetAverageRating(models.ReviewTypeShop, shopID)

	return reviews, pagination, avgRating, totalReviews, nil
}
