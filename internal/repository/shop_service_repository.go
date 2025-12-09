package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type ShopServiceRepository interface {
	Create(shopService *models.ShopService) error
	GetByID(id uint) (*models.ShopService, error)
	GetByShopID(shopID uint) ([]models.ShopService, error)
	GetByShopAndService(shopID, serviceID uint) (*models.ShopService, error)
	Update(shopService *models.ShopService) error
	Delete(id uint) error
}

type shopServiceRepository struct {
	db *gorm.DB
}

func NewShopServiceRepository() ShopServiceRepository {
	return &shopServiceRepository{
		db: database.DB,
	}
}

func (r *shopServiceRepository) Create(shopService *models.ShopService) error {
	return r.db.Create(shopService).Error
}

func (r *shopServiceRepository) GetByID(id uint) (*models.ShopService, error) {
	var shopService models.ShopService
	err := r.db.Preload("Service").Preload("Shop").First(&shopService, id).Error
	if err != nil {
		return nil, err
	}
	return &shopService, nil
}

func (r *shopServiceRepository) GetByShopID(shopID uint) ([]models.ShopService, error) {
	var shopServices []models.ShopService
	err := r.db.Where("shop_id = ?", shopID).
		Preload("Service").
		Order("created_at ASC").
		Find(&shopServices).Error
	return shopServices, err
}

func (r *shopServiceRepository) GetByShopAndService(shopID, serviceID uint) (*models.ShopService, error) {
	var shopService models.ShopService
	err := r.db.Where("shop_id = ? AND service_id = ?", shopID, serviceID).
		First(&shopService).Error
	if err != nil {
		return nil, err
	}
	return &shopService, nil
}

func (r *shopServiceRepository) Update(shopService *models.ShopService) error {
	return r.db.Save(shopService).Error
}

func (r *shopServiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.ShopService{}, id).Error
}
