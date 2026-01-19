package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
	"fmt"
	"time"
)

type AdminService interface {
	GetPendingShops(page, limit int) ([]models.Shop, int, error)
	ApproveShop(shopID uint) error
	RejectShop(shopID uint) error
	CreateService(name string, image *string) (*models.Service, error)
	UpdateService(serviceID uint, name *string, image *string) (*models.Service, error)
	DeleteService(serviceID uint) error
}

type adminService struct {
	shopRepo       repository.ShopRepository
	serviceRepo    repository.ServiceRepository
	cacheService   CacheService
	messageService MessageService
}

func NewAdminService(shopRepo repository.ShopRepository, serviceRepo repository.ServiceRepository, cacheService CacheService, messageService MessageService) AdminService {
	return &adminService{
		shopRepo:       shopRepo,
		serviceRepo:    serviceRepo,
		cacheService:   cacheService,
		messageService: messageService,
	}
}

func (s *adminService) GetPendingShops(page, limit int) ([]models.Shop, int, error) {
	shops, total, err := s.shopRepo.GetPendingShops(page, limit)
	if err != nil {
		return nil, 0, err
	}
	return shops, int(total), nil
}

func (s *adminService) ApproveShop(shopID uint) error {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return errors.New("shop not found")
	}

	if shop.Status != models.ShopStatusPending {
		return errors.New("shop is not pending approval")
	}

	shop.Status = models.ShopStatusConfirmed

	// 1. DB'yi güncelle
	if err := s.shopRepo.Update(shop); err != nil {
		return err
	}

	updatedShop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return err
	}

	// 2. Cache'i güncelle
	cacheKey := fmt.Sprintf("shop:%d", updatedShop.ID)
	s.cacheService.Set(cacheKey, updatedShop, 1*time.Hour)

	// 3. RabbitMQ ile event mesajı gönder
	s.messageService.PublishShopUpdated(updatedShop.ID)

	return nil
}

func (s *adminService) RejectShop(shopID uint) error {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return errors.New("shop not found")
	}

	if shop.Status != models.ShopStatusPending {
		return errors.New("shop is not pending approval")
	}

	shop.Status = models.ShopStatusRejected
	if err := s.shopRepo.Update(shop); err != nil {
		return err
	}

	// Cache'i invalidate et
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)

	return nil
}

func (s *adminService) CreateService(name string, image *string) (*models.Service, error) {
	if name == "" {
		return nil, errors.New("service name is required")
	}

	service := &models.Service{
		Name:  name,
		Image: image,
	}

	if err := s.serviceRepo.Create(service); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return service, nil
}

func (s *adminService) UpdateService(serviceID uint, name *string, image *string) (*models.Service, error) {
	service, err := s.serviceRepo.GetByID(serviceID)
	if err != nil {
		return nil, errors.New("service not found")
	}

	if name != nil {
		service.Name = *name
	}
	if image != nil {
		service.Image = image
	}

	if err := s.serviceRepo.Update(service); err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	return service, nil
}

func (s *adminService) DeleteService(serviceID uint) error {
	_, err := s.serviceRepo.GetByID(serviceID)
	if err != nil {
		return errors.New("service not found")
	}

	return s.serviceRepo.Delete(serviceID)
}
