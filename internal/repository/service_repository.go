package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type ServiceRepository interface {
	Create(service *models.Service) error
	GetByID(id uint) (*models.Service, error)
	GetAll() ([]models.Service, error)
	Update(service *models.Service) error
	Delete(id uint) error
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository() ServiceRepository {
	return &serviceRepository{
		db: database.DB,
	}
}

func (r *serviceRepository) Create(service *models.Service) error {
	return r.db.Create(service).Error
}

func (r *serviceRepository) GetByID(id uint) (*models.Service, error) {
	var service models.Service
	err := r.db.First(&service, id).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) GetAll() ([]models.Service, error) {
	var services []models.Service
	err := r.db.Order("name ASC").Find(&services).Error
	return services, err
}

func (r *serviceRepository) Update(service *models.Service) error {
	return r.db.Save(service).Error
}

func (r *serviceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}
