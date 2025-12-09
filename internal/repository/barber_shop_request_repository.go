package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type BarberShopRequestRepository interface {
	Create(request *models.BarberShopRequest) error
	GetByID(id uint) (*models.BarberShopRequest, error)
	GetByShopID(shopID uint, status *models.RequestStatus) ([]models.BarberShopRequest, error)
	GetByBarberID(barberID uint) ([]models.BarberShopRequest, error)
	GetPendingByShopID(shopID uint) ([]models.BarberShopRequest, error)
	Update(request *models.BarberShopRequest) error
	Delete(id uint) error
}

type barberShopRequestRepository struct {
	db *gorm.DB
}

func NewBarberShopRequestRepository() BarberShopRequestRepository {
	return &barberShopRequestRepository{
		db: database.DB,
	}
}

func (r *barberShopRequestRepository) Create(request *models.BarberShopRequest) error {
	return r.db.Create(request).Error
}

func (r *barberShopRequestRepository) GetByID(id uint) (*models.BarberShopRequest, error) {
	var request models.BarberShopRequest
	err := r.db.Preload("Barber").Preload("Barber.User").Preload("Shop").Preload("ApprovedByUser").
		First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *barberShopRequestRepository) GetByShopID(shopID uint, status *models.RequestStatus) ([]models.BarberShopRequest, error) {
	var requests []models.BarberShopRequest
	query := r.db.Where("shop_id = ?", shopID).Preload("Barber").Preload("Barber.User")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

func (r *barberShopRequestRepository) GetByBarberID(barberID uint) ([]models.BarberShopRequest, error) {
	var requests []models.BarberShopRequest
	err := r.db.Where("barber_id = ?", barberID).
		Preload("Shop").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *barberShopRequestRepository) GetPendingByShopID(shopID uint) ([]models.BarberShopRequest, error) {
	status := models.RequestStatusPending
	return r.GetByShopID(shopID, &status)
}

func (r *barberShopRequestRepository) Update(request *models.BarberShopRequest) error {
	return r.db.Save(request).Error
}

func (r *barberShopRequestRepository) Delete(id uint) error {
	return r.db.Delete(&models.BarberShopRequest{}, id).Error
}
