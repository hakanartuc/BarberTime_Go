package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type BarberRepository interface {
	Create(barber *models.Barber) error
	GetByID(id uint) (*models.Barber, error)
	GetByUserID(userID uint) (*models.Barber, error)
	GetNearby(latitude, longitude float64, maxDistanceKM *float64, page, limit int) ([]models.Barber, int64, error)
	Update(barber *models.Barber) error
	Delete(id uint) error
	GetByShopID(shopID uint) ([]models.Barber, error)
}

type barberRepository struct {
	db *gorm.DB
}

func NewBarberRepository() BarberRepository {
	return &barberRepository{
		db: database.DB,
	}
}

func (r *barberRepository) Create(barber *models.Barber) error {
	return r.db.Create(barber).Error
}

func (r *barberRepository) GetByID(id uint) (*models.Barber, error) {
	var barber models.Barber
	err := r.db.Preload("Shops").Preload("Reviews").First(&barber, id).Error
	if err != nil {
		return nil, err
	}
	return &barber, nil
}

func (r *barberRepository) GetByUserID(userID uint) (*models.Barber, error) {
	var barber models.Barber
	err := r.db.Where("user_id = ?", userID).
		Preload("Shops").Preload("Reviews").
		First(&barber).Error
	if err != nil {
		return nil, err
	}
	return &barber, nil
}

func (r *barberRepository) GetNearby(latitude, longitude float64, maxDistanceKM *float64, page, limit int) ([]models.Barber, int64, error) {
	var barbers []models.Barber
	var total int64

	offset := (page - 1) * limit

	// Mesafe limiti belirle (default 50km)
	distanceLimit := 50.0
	if maxDistanceKM != nil && *maxDistanceKM > 0 {
		distanceLimit = *maxDistanceKM
	}

	// PostgreSQL'de Haversine formülü ile mesafe hesaplama
	// Subquery ile distance hesaplayıp filtreleme yapıyoruz
	distanceSQL := "(6371 * acos(cos(radians(?)) * cos(radians(shops.location_latitude)) * cos(radians(shops.location_longitude) - radians(?)) + sin(radians(?)) * sin(radians(shops.location_latitude))))"

	// Base query - distance hesaplama ile
	baseQuery := r.db.Model(&models.Barber{}).
		Joins("JOIN barber_shops ON barbers.id = barber_shops.barber_id").
		Joins("JOIN shops ON barber_shops.shop_id = shops.id").
		Where("shops.status = ? AND barber_shops.status = ?", models.ShopStatusConfirmed, models.BarberShopStatusApproved).
		Where(distanceSQL+" <= ?", latitude, longitude, latitude, distanceLimit).
		Select("DISTINCT barbers.*, "+distanceSQL+" AS distance", latitude, longitude, latitude).
		Order("distance ASC")

	// Total count
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Results with pagination
	err := baseQuery.Offset(offset).Limit(limit).Preload("Shops").Find(&barbers).Error
	return barbers, total, err
}

func (r *barberRepository) Update(barber *models.Barber) error {
	return r.db.Save(barber).Error
}

func (r *barberRepository) Delete(id uint) error {
	return r.db.Delete(&models.Barber{}, id).Error
}

func (r *barberRepository) GetByShopID(shopID uint) ([]models.Barber, error) {
	var barbers []models.Barber
	err := r.db.Joins("JOIN barber_shops ON barbers.id = barber_shops.barber_id").
		Where("barber_shops.shop_id = ? AND barber_shops.status = ?", shopID, models.BarberShopStatusApproved).
		Preload("Shops").
		Find(&barbers).Error
	return barbers, err
}
