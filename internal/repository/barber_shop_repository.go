package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type BarberShopRepository interface {
	Create(barberShop *models.BarberShop) error
	GetByID(id uint) (*models.BarberShop, error)
	GetByBarberID(barberID uint) ([]models.BarberShop, error)
	GetByShopID(shopID uint, status *models.BarberShopStatus) ([]models.BarberShop, error)
	GetPendingByShopID(shopID uint) ([]models.BarberShop, error)
	GetByBarberAndShop(barberID, shopID uint) (*models.BarberShop, error)
	IsOwner(shopID, userID uint) (bool, error)
	GetOwners(shopID uint) ([]models.User, error)
	Update(barberShop *models.BarberShop) error
	Delete(id uint) error
}

type barberShopRepository struct {
	db *gorm.DB
}

func NewBarberShopRepository() BarberShopRepository {
	return &barberShopRepository{
		db: database.DB,
	}
}

func (r *barberShopRepository) Create(barberShop *models.BarberShop) error {
	return r.db.Create(barberShop).Error
}

func (r *barberShopRepository) GetByID(id uint) (*models.BarberShop, error) {
	var barberShop models.BarberShop
	err := r.db.Preload("Barber").Preload("Barber.User").Preload("Shop").Preload("ApprovedByUser").
		First(&barberShop, id).Error
	if err != nil {
		return nil, err
	}
	return &barberShop, nil
}

func (r *barberShopRepository) GetByBarberID(barberID uint) ([]models.BarberShop, error) {
	var barberShops []models.BarberShop
	err := r.db.Where("barber_id = ?", barberID).
		Preload("Shop").
		Order("created_at DESC").
		Find(&barberShops).Error
	return barberShops, err
}

func (r *barberShopRepository) GetByShopID(shopID uint, status *models.BarberShopStatus) ([]models.BarberShop, error) {
	var barberShops []models.BarberShop
	query := r.db.Where("shop_id = ?", shopID).Preload("Barber").Preload("Barber.User")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("created_at DESC").Find(&barberShops).Error
	return barberShops, err
}

func (r *barberShopRepository) GetPendingByShopID(shopID uint) ([]models.BarberShop, error) {
	status := models.BarberShopStatusPending
	return r.GetByShopID(shopID, &status)
}

func (r *barberShopRepository) GetByBarberAndShop(barberID, shopID uint) (*models.BarberShop, error) {
	var barberShop models.BarberShop
	err := r.db.Where("barber_id = ? AND shop_id = ?", barberID, shopID).
		Preload("Barber").Preload("Shop").
		First(&barberShop).Error
	if err != nil {
		return nil, err
	}
	return &barberShop, nil
}

func (r *barberShopRepository) Update(barberShop *models.BarberShop) error {
	return r.db.Save(barberShop).Error
}

func (r *barberShopRepository) Delete(id uint) error {
	return r.db.Delete(&models.BarberShop{}, id).Error
}

func (r *barberShopRepository) IsOwner(shopID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.BarberShop{}).
		Joins("JOIN barbers ON barber_shops.barber_id = barbers.id").
		Where("barber_shops.shop_id = ? AND barbers.user_id = ? AND barber_shops.is_owner = ?", shopID, userID, true).
		Count(&count).Error
	return count > 0, err
}

func (r *barberShopRepository) GetOwners(shopID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Model(&models.User{}).
		Joins("JOIN barbers ON users.id = barbers.user_id").
		Joins("JOIN barber_shops ON barbers.id = barber_shops.barber_id").
		Where("barber_shops.shop_id = ? AND barber_shops.is_owner = ?", shopID, true).
		Find(&users).Error
	return users, err
}
