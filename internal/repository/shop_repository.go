package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type ShopRepository interface {
	Create(shop *models.Shop) error
	GetByID(id uint) (*models.Shop, error)
	GetNearby(latitude, longitude float64, maxDistanceKM *float64, page, limit int) ([]models.Shop, int64, error)
	GetPendingShops(page, limit int) ([]models.Shop, int64, error)
	Update(shop *models.Shop) error
	Delete(id uint) error
}

type shopRepository struct {
	db *gorm.DB
}

func NewShopRepository() ShopRepository {
	return &shopRepository{
		db: database.DB,
	}
}

func (r *shopRepository) Create(shop *models.Shop) error {
	return r.db.Create(shop).Error
}

func (r *shopRepository) GetByID(id uint) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Preload("Barbers").Preload("ShopServices.Service").Preload("Reviews").First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) GetNearby(latitude, longitude float64, maxDistanceKM *float64, page, limit int) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	offset := (page - 1) * limit

	// Mesafe limiti belirle (default 50km)
	distanceLimit := 50.0
	if maxDistanceKM != nil && *maxDistanceKM > 0 {
		distanceLimit = *maxDistanceKM
	}

	// PostgreSQL'de Haversine formülü ile mesafe hesaplama
	// WHERE clause'da distance kontrolü yapıyoruz
	distanceSQL := "(6371 * acos(cos(radians(?)) * cos(radians(location_latitude)) * cos(radians(location_longitude) - radians(?)) + sin(radians(?)) * sin(radians(location_latitude))))"

	query := r.db.Model(&models.Shop{}).
		Where("status = ?", models.ShopStatusConfirmed).
		Where(distanceSQL+" <= ?", latitude, longitude, latitude, distanceLimit).
		Select("*, "+distanceSQL+" AS distance", latitude, longitude, latitude).
		Order("distance ASC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&shops).Error
	return shops, total, err
}

func (r *shopRepository) GetPendingShops(page, limit int) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	offset := (page - 1) * limit

	query := r.db.Model(&models.Shop{}).Where("status = ?", models.ShopStatusPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&shops).Error
	return shops, total, err
}

func (r *shopRepository) Update(shop *models.Shop) error {
	return r.db.Save(shop).Error
}

func (r *shopRepository) Delete(id uint) error {
	return r.db.Delete(&models.Shop{}, id).Error
}
