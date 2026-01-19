package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
	"fmt"
	"time"
)

type ShopProfileService interface {
	CreateShop(userID uint, name string, description string, phoneNumber string, email string, website string, latitude, longitude float64, workingHours, closedDays string) (*models.Shop, error)
	GetShopProfile(shopID uint) (*models.Shop, error)
	UpdateShopProfile(shopID, userID uint, name *string, description *string, phoneNumber *string, email *string, website *string, workingHours *string, closedDays *string, latitude, longitude *float64) (*models.Shop, error)
	AddShopOwner(shopID, ownerUserID, requesterUserID uint) error
	RemoveShopOwner(shopID, ownerUserID, requesterUserID uint) error
	GetShopOwners(shopID uint) ([]models.User, error)
	CreateShopService(shopID, userID uint, serviceID uint, duration int, price float64, description *string) (*models.ShopService, error)
	UpdateShopService(shopID, userID, shopServiceID uint, duration int, price float64, description *string) (*models.ShopService, error)
	DeleteShopService(shopID, userID, shopServiceID uint) error
}

type shopProfileService struct {
	shopRepo        repository.ShopRepository
	userRepo        repository.UserRepository
	barberRepo      repository.BarberRepository
	barberShopRepo  repository.BarberShopRepository
	shopServiceRepo repository.ShopServiceRepository
	serviceRepo     repository.ServiceRepository
	cacheService    CacheService
	messageService  MessageService
}

func NewShopProfileService(
	shopRepo repository.ShopRepository,
	userRepo repository.UserRepository,
	barberRepo repository.BarberRepository,
	barberShopRepo repository.BarberShopRepository,
	shopServiceRepo repository.ShopServiceRepository,
	serviceRepo repository.ServiceRepository,
	cacheService CacheService,
	messageService MessageService,
) ShopProfileService {
	return &shopProfileService{
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		barberRepo:      barberRepo,
		barberShopRepo:  barberShopRepo,
		shopServiceRepo: shopServiceRepo,
		serviceRepo:     serviceRepo,
		cacheService:    cacheService,
		messageService:  messageService,
	}
}

func (s *shopProfileService) CreateShop(userID uint, name string, description string, phoneNumber string, email string, website string, latitude, longitude float64, workingHours, closedDays string) (*models.Shop, error) {
	// Kullanıcı kontrolü
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	shop := &models.Shop{
		Name:        name,
		Description: description,
		PhoneNumber: phoneNumber,
		Email:       email,
		Website:     website,
		Location: models.Location{
			Latitude:  latitude,
			Longitude: longitude,
		},
		WorkingHours: workingHours,
		ClosedDays:   closedDays,
		Status:       models.ShopStatusPending,
	}

	if err := s.shopRepo.Create(shop); err != nil {
		return nil, err
	}

	// User'ın barber olup olmadığını kontrol et
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("user must be registered as barber to create a shop")
	}

	// İlk owner'ı barber_shops'a ekle (isOwner = true, status = approved)
	barberShop := &models.BarberShop{
		BarberID: barber.ID,
		ShopID:   shop.ID,
		Status:   models.BarberShopStatusApproved,
		IsOwner:  true,
	}
	if err := s.barberShopRepo.Create(barberShop); err != nil {
		return nil, err
	}

	createdShop, err := s.shopRepo.GetByID(shop.ID)
	if err != nil {
		return nil, err
	}

	// 2. Cache'e kaydet
	cacheKey := fmt.Sprintf("shop:%d", createdShop.ID)
	s.cacheService.Set(cacheKey, createdShop, 1*time.Hour)

	// 3. RabbitMQ ile event mesajı gönder
	s.messageService.PublishShopCreated(createdShop.ID)

	return createdShop, nil
}

func (s *shopProfileService) GetShopProfile(shopID uint) (*models.Shop, error) {
	// Önce cache'den kontrol et
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	var cachedShop models.Shop
	if err := s.cacheService.Get(cacheKey, &cachedShop); err == nil {
		return &cachedShop, nil
	}

	// Cache'de yoksa DB'den al
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	// Owner'ları al
	owners, err := s.barberShopRepo.GetOwners(shopID)
	if err == nil {
		shop.Owners = owners
	}

	// Cache'e kaydet
	s.cacheService.Set(cacheKey, shop, 1*time.Hour)

	return shop, nil
}

func (s *shopProfileService) UpdateShopProfile(shopID, userID uint, name *string, description *string, phoneNumber *string, email *string, website *string, workingHours *string, closedDays *string, latitude, longitude *float64) (*models.Shop, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("unauthorized: only shop owners can update shop profile")
	}

	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	if name != nil {
		shop.Name = *name
	}
	if description != nil {
		shop.Description = *description
	}
	if phoneNumber != nil {
		shop.PhoneNumber = *phoneNumber
	}
	if email != nil {
		shop.Email = *email
	}
	if website != nil {
		shop.Website = *website
	}
	if workingHours != nil {
		shop.WorkingHours = *workingHours
	}
	if closedDays != nil {
		shop.ClosedDays = *closedDays
	}
	if latitude != nil && longitude != nil {
		shop.Location.Latitude = *latitude
		shop.Location.Longitude = *longitude
	}

	// 1. DB'yi güncelle
	if err := s.shopRepo.Update(shop); err != nil {
		return nil, err
	}

	updatedShop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	// 2. Cache'i güncelle
	cacheKey := fmt.Sprintf("shop:%d", updatedShop.ID)
	s.cacheService.Set(cacheKey, updatedShop, 1*time.Hour)

	// 3. RabbitMQ ile event mesajı gönder
	s.messageService.PublishShopUpdated(updatedShop.ID)

	return updatedShop, nil
}

func (s *shopProfileService) AddShopOwner(shopID, ownerUserID, requesterUserID uint) error {
	// Requester'ın owner olduğunu kontrol et
	isOwner, err := s.barberShopRepo.IsOwner(shopID, requesterUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: only shop owners can add other owners")
	}

	// Owner olacak user'ın barber olup olmadığını kontrol et
	barber, err := s.barberRepo.GetByUserID(ownerUserID)
	if err != nil {
		return errors.New("user must be registered as barber to be a shop owner")
	}

	// BarberShop kaydını kontrol et veya oluştur
	barberShop, err := s.barberShopRepo.GetByBarberAndShop(barber.ID, shopID)
	if err != nil {
		// Yeni kayıt oluştur
		barberShop = &models.BarberShop{
			BarberID: barber.ID,
			ShopID:   shopID,
			Status:   models.BarberShopStatusApproved,
			IsOwner:  true,
		}
		return s.barberShopRepo.Create(barberShop)
	}

	// Mevcut kaydı güncelle
	barberShop.IsOwner = true
	barberShop.Status = models.BarberShopStatusApproved
	return s.barberShopRepo.Update(barberShop)
}

func (s *shopProfileService) RemoveShopOwner(shopID, ownerUserID, requesterUserID uint) error {
	// Requester'ın owner olduğunu kontrol et
	isOwner, err := s.barberShopRepo.IsOwner(shopID, requesterUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: only shop owners can remove other owners")
	}

	// Owner olacak user'ın barber olup olmadığını kontrol et
	barber, err := s.barberRepo.GetByUserID(ownerUserID)
	if err != nil {
		return errors.New("user not found as barber")
	}

	// BarberShop kaydını bul ve IsOwner'ı false yap
	barberShop, err := s.barberShopRepo.GetByBarberAndShop(barber.ID, shopID)
	if err != nil {
		return errors.New("barber is not associated with this shop")
	}

	barberShop.IsOwner = false
	return s.barberShopRepo.Update(barberShop)
}

func (s *shopProfileService) GetShopOwners(shopID uint) ([]models.User, error) {
	return s.barberShopRepo.GetOwners(shopID)
}

func (s *shopProfileService) CreateShopService(shopID, userID uint, serviceID uint, duration int, price float64, description *string) (*models.ShopService, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can add services")
	}

	// Service'in var olup olmadığını kontrol et
	_, err = s.serviceRepo.GetByID(serviceID)
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

	// Cache'i temizle ve Elasticsearch'i güncelle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)
	s.messageService.PublishShopUpdated(shopID)

	return shopService, nil
}

func (s *shopProfileService) UpdateShopService(shopID, userID, shopServiceID uint, duration int, price float64, description *string) (*models.ShopService, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can update services")
	}

	shopService, err := s.shopServiceRepo.GetByID(shopServiceID)
	if err != nil {
		return nil, errors.New("shop service not found")
	}

	if shopService.ShopID != shopID {
		return nil, errors.New("shop service does not belong to this shop")
	}

	shopService.Duration = duration
	shopService.Price = price
	shopService.Description = description

	if err := s.shopServiceRepo.Update(shopService); err != nil {
		return nil, err
	}

	// Cache'i temizle ve Elasticsearch'i güncelle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)
	s.messageService.PublishShopUpdated(shopID)

	return shopService, nil
}

func (s *shopProfileService) DeleteShopService(shopID, userID, shopServiceID uint) error {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only shop owners can delete services")
	}

	shopService, err := s.shopServiceRepo.GetByID(shopServiceID)
	if err != nil {
		return errors.New("shop service not found")
	}

	if shopService.ShopID != shopID {
		return errors.New("shop service does not belong to this shop")
	}

	if err := s.shopServiceRepo.Delete(shopServiceID); err != nil {
		return err
	}

	// Cache'i temizle ve Elasticsearch'i güncelle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)
	s.messageService.PublishShopUpdated(shopID)

	return nil
}
