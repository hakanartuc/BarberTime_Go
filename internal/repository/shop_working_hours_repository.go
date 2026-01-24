package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type ShopWorkingHoursRepository interface {
	Create(workingHours *models.ShopWorkingHours) error
	GetByID(id uint) (*models.ShopWorkingHours, error)
	GetByShopID(shopID uint) ([]models.ShopWorkingHours, error)
	GetByShopIDAndDay(shopID uint, dayOfWeek models.DayOfWeek) (*models.ShopWorkingHours, error)
	Update(workingHours *models.ShopWorkingHours) error
	Delete(id uint) error
	DeleteByShopID(shopID uint) error
	CreateBatch(workingHours []models.ShopWorkingHours) error
}

type shopWorkingHoursRepository struct {
	db *gorm.DB
}

func NewShopWorkingHoursRepository() ShopWorkingHoursRepository {
	return &shopWorkingHoursRepository{
		db: database.DB,
	}
}

func (r *shopWorkingHoursRepository) Create(workingHours *models.ShopWorkingHours) error {
	return r.db.Create(workingHours).Error
}

func (r *shopWorkingHoursRepository) GetByID(id uint) (*models.ShopWorkingHours, error) {
	var workingHours models.ShopWorkingHours
	err := r.db.First(&workingHours, id).Error
	if err != nil {
		return nil, err
	}
	return &workingHours, nil
}

func (r *shopWorkingHoursRepository) GetByShopID(shopID uint) ([]models.ShopWorkingHours, error) {
	var workingHours []models.ShopWorkingHours
	err := r.db.Where("shop_id = ?", shopID).Order("day_of_week ASC").Find(&workingHours).Error
	return workingHours, err
}

func (r *shopWorkingHoursRepository) GetByShopIDAndDay(shopID uint, dayOfWeek models.DayOfWeek) (*models.ShopWorkingHours, error) {
	var workingHours models.ShopWorkingHours
	err := r.db.Where("shop_id = ? AND day_of_week = ?", shopID, dayOfWeek).First(&workingHours).Error
	if err != nil {
		return nil, err
	}
	return &workingHours, nil
}

func (r *shopWorkingHoursRepository) Update(workingHours *models.ShopWorkingHours) error {
	return r.db.Save(workingHours).Error
}

func (r *shopWorkingHoursRepository) Delete(id uint) error {
	return r.db.Delete(&models.ShopWorkingHours{}, id).Error
}

func (r *shopWorkingHoursRepository) DeleteByShopID(shopID uint) error {
	return r.db.Where("shop_id = ?", shopID).Delete(&models.ShopWorkingHours{}).Error
}

func (r *shopWorkingHoursRepository) CreateBatch(workingHours []models.ShopWorkingHours) error {
	return r.db.Create(&workingHours).Error
}
