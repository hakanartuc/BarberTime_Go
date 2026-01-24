package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type ShopClosedDayRepository interface {
	Create(closedDay *models.ShopClosedDay) error
	GetByID(id uint) (*models.ShopClosedDay, error)
	GetByShopID(shopID uint) ([]models.ShopClosedDay, error)
	Update(closedDay *models.ShopClosedDay) error
	Delete(id uint) error
}

type shopClosedDayRepository struct {
	db *gorm.DB
}

func NewShopClosedDayRepository() ShopClosedDayRepository {
	return &shopClosedDayRepository{
		db: database.DB,
	}
}

func (r *shopClosedDayRepository) Create(closedDay *models.ShopClosedDay) error {
	return r.db.Create(closedDay).Error
}

func (r *shopClosedDayRepository) GetByID(id uint) (*models.ShopClosedDay, error) {
	var closedDay models.ShopClosedDay
	err := r.db.First(&closedDay, id).Error
	if err != nil {
		return nil, err
	}
	return &closedDay, nil
}

func (r *shopClosedDayRepository) GetByShopID(shopID uint) ([]models.ShopClosedDay, error) {
	var closedDays []models.ShopClosedDay
	err := r.db.Where("shop_id = ?", shopID).Order("date ASC").Find(&closedDays).Error
	return closedDays, err
}

func (r *shopClosedDayRepository) Update(closedDay *models.ShopClosedDay) error {
	return r.db.Save(closedDay).Error
}

func (r *shopClosedDayRepository) Delete(id uint) error {
	return r.db.Delete(&models.ShopClosedDay{}, id).Error
}
