package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
)

type ShopServiceService interface {
	GetShopServices(shopID uint) ([]models.ShopService, error)
	CreateShopService(shopID uint, serviceID uint, duration int, price float64, description *string) (*models.ShopService, error)
	UpdateShopService(shopServiceID uint, duration int, price float64, description *string) (*models.ShopService, error)
	DeleteShopService(shopServiceID uint) error
}

type shopServiceService struct {
	shopServiceRepo repository.ShopServiceRepository
	serviceRepo     repository.ServiceRepository
}

func NewShopServiceService(shopServiceRepo repository.ShopServiceRepository, serviceRepo repository.ServiceRepository) ShopServiceService {
	return &shopServiceService{
		shopServiceRepo: shopServiceRepo,
		serviceRepo:     serviceRepo,
	}
}

func (s *shopServiceService) GetShopServices(shopID uint) ([]models.ShopService, error) {
	return s.shopServiceRepo.GetByShopID(shopID)
}

func (s *shopServiceService) CreateShopService(shopID uint, serviceID uint, duration int, price float64, description *string) (*models.ShopService, error) {
	// Service'in var olup olmadığını kontrol et
	_, err := s.serviceRepo.GetByID(serviceID)
	if err != nil {
		return nil, errors.New("service not found")
	}

	// Aynı shop ve service kombinasyonunun zaten var olup olmadığını kontrol et
	existing, err := s.shopServiceRepo.GetByShopAndService(shopID, serviceID)
	if err == nil && existing != nil {
		return nil, errors.New("this service already exists for this shop")
	}

	shopService := &models.ShopService{
		ShopID:      shopID,
		ServiceID:   serviceID,
		Duration:    duration,
		Price:       price,
		Description: description,
	}

	if err := s.shopServiceRepo.Create(shopService); err != nil {
		return nil, err
	}

	return shopService, nil
}

func (s *shopServiceService) UpdateShopService(shopServiceID uint, duration int, price float64, description *string) (*models.ShopService, error) {
	shopService, err := s.shopServiceRepo.GetByID(shopServiceID)
	if err != nil {
		return nil, errors.New("shop service not found")
	}

	shopService.Duration = duration
	shopService.Price = price
	shopService.Description = description

	if err := s.shopServiceRepo.Update(shopService); err != nil {
		return nil, err
	}

	return shopService, nil
}

func (s *shopServiceService) DeleteShopService(shopServiceID uint) error {
	return s.shopServiceRepo.Delete(shopServiceID)
}
