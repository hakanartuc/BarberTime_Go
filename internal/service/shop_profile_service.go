package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
	"fmt"
	"time"
)

type ShopProfileService interface {
	CreateShop(userID uint, name string, description string, phoneNumber string, email string, website string, latitude, longitude float64, openTime, closeTime, closedDays string) (*models.Shop, error)
	GetShopProfile(shopID uint) (*models.Shop, error)
	UpdateShopProfile(shopID, userID uint, name *string, description *string, banner *string, phoneNumber *string, email *string, website *string, openTime *string, closeTime *string, closedDays *string, latitude, longitude *float64) (*models.Shop, error)
	AddShopOwner(shopID, ownerUserID, requesterUserID uint) error
	RemoveShopOwner(shopID, ownerUserID, requesterUserID uint) error
	GetShopOwners(shopID uint) ([]models.User, error)
	CreateShopService(shopID, userID uint, serviceID uint, duration int, price float64, description *string) (*models.ShopService, error)
	UpdateShopService(shopID, userID, shopServiceID uint, duration int, price float64, description *string) (*models.ShopService, error)
	DeleteShopService(shopID, userID, shopServiceID uint) error

	// Closed Days
	GetClosedDays(shopID uint) ([]models.ShopClosedDay, error)
	CreateClosedDay(shopID, userID uint, specificDate *string, dayOfWeek *int, isAllDay bool, startTime, endTime *string, isRecurring bool, reason string) (*models.ShopClosedDay, error)
	UpdateClosedDay(shopID, userID, closedDayID uint, specificDate *string, dayOfWeek *int, isAllDay *bool, startTime, endTime *string, isRecurring *bool, reason *string) (*models.ShopClosedDay, error)
	DeleteClosedDay(shopID, userID, closedDayID uint) error
	ToggleClosedDay(shopID, userID, closedDayID uint, isActive bool) (*models.ShopClosedDay, error)

	// Working Hours
	GetWorkingHours(shopID uint) ([]models.ShopWorkingHours, error)
	UpdateWorkingHours(shopID, userID uint, openTime, closeTime string) ([]models.ShopWorkingHours, error)

	// System Status
	UpdateSystemStatus(shopID, userID uint, isSystemOff bool) (*models.Shop, error)

	// My Shops
	GetMyShops(userID uint) ([]models.Shop, error)
}

type shopProfileService struct {
	shopRepo         repository.ShopRepository
	userRepo         repository.UserRepository
	barberRepo       repository.BarberRepository
	barberShopRepo   repository.BarberShopRepository
	shopServiceRepo  repository.ShopServiceRepository
	serviceRepo      repository.ServiceRepository
	closedDayRepo    repository.ShopClosedDayRepository
	workingHoursRepo repository.ShopWorkingHoursRepository
	cacheService     CacheService
	messageService   MessageService
}

func NewShopProfileService(
	shopRepo repository.ShopRepository,
	userRepo repository.UserRepository,
	barberRepo repository.BarberRepository,
	barberShopRepo repository.BarberShopRepository,
	shopServiceRepo repository.ShopServiceRepository,
	serviceRepo repository.ServiceRepository,
	closedDayRepo repository.ShopClosedDayRepository,
	workingHoursRepo repository.ShopWorkingHoursRepository,
	cacheService CacheService,
	messageService MessageService,
) ShopProfileService {
	return &shopProfileService{
		shopRepo:         shopRepo,
		userRepo:         userRepo,
		barberRepo:       barberRepo,
		barberShopRepo:   barberShopRepo,
		shopServiceRepo:  shopServiceRepo,
		serviceRepo:      serviceRepo,
		closedDayRepo:    closedDayRepo,
		workingHoursRepo: workingHoursRepo,
		cacheService:     cacheService,
		messageService:   messageService,
	}
}

func (s *shopProfileService) CreateShop(userID uint, name string, description string, phoneNumber string, email string, website string, latitude, longitude float64, openTime, closeTime, closedDays string) (*models.Shop, error) {
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
		OpenTime:   openTime,
		CloseTime:  closeTime,
		ClosedDays: closedDays,
		Status:     models.ShopStatusPending,
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

	// Owner'ları al
	owners, err := s.barberShopRepo.GetOwners(shop.ID)
	if err == nil {
		createdShop.Owners = owners
	}

	// 2. Cache'e kaydet
	cacheKey := fmt.Sprintf("shop:%d", createdShop.ID)
	s.cacheService.Set(cacheKey, createdShop, 1*time.Hour)

	// my-shops cache'ini temizle (yeni shop eklendi)
	myShopsCacheKey := fmt.Sprintf("my-shops:%d", userID)
	s.cacheService.Delete(myShopsCacheKey)

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

func (s *shopProfileService) UpdateShopProfile(shopID, userID uint, name *string, description *string, banner *string, phoneNumber *string, email *string, website *string, openTime *string, closeTime *string, closedDays *string, latitude, longitude *float64) (*models.Shop, error) {
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
	if banner != nil {
		shop.Banner = banner
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
	if openTime != nil {
		shop.OpenTime = *openTime
	}
	if closeTime != nil {
		shop.CloseTime = *closeTime
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

	// Owner'ları al
	owners, err := s.barberShopRepo.GetOwners(shopID)
	if err == nil {
		updatedShop.Owners = owners
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
		if err := s.barberShopRepo.Create(barberShop); err != nil {
			return err
		}
		// my-shops cache'ini temizle
		myShopsCacheKey := fmt.Sprintf("my-shops:%d", ownerUserID)
		s.cacheService.Delete(myShopsCacheKey)
		return nil
	}

	// Mevcut kaydı güncelle
	barberShop.IsOwner = true
	barberShop.Status = models.BarberShopStatusApproved
	if err := s.barberShopRepo.Update(barberShop); err != nil {
		return err
	}
	// my-shops cache'ini temizle
	myShopsCacheKey := fmt.Sprintf("my-shops:%d", ownerUserID)
	s.cacheService.Delete(myShopsCacheKey)
	return nil
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
	if err := s.barberShopRepo.Update(barberShop); err != nil {
		return err
	}
	// my-shops cache'ini temizle
	myShopsCacheKey := fmt.Sprintf("my-shops:%d", ownerUserID)
	s.cacheService.Delete(myShopsCacheKey)
	return nil
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

// ==================== Closed Days ====================

func (s *shopProfileService) GetClosedDays(shopID uint) ([]models.ShopClosedDay, error) {
	// Shop var mı kontrol et
	_, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	return s.closedDayRepo.GetByShopID(shopID)
}

func (s *shopProfileService) CreateClosedDay(shopID, userID uint, specificDate *string, dayOfWeek *int, isAllDay bool, startTime, endTime *string, isRecurring bool, reason string) (*models.ShopClosedDay, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can add closed days")
	}

	// Validasyon
	if specificDate != nil && dayOfWeek != nil {
		return nil, errors.New("specificDate and dayOfWeek cannot be set together")
	}
	if specificDate == nil && dayOfWeek == nil {
		return nil, errors.New("either specificDate or dayOfWeek is required")
	}
	if dayOfWeek != nil {
		if *dayOfWeek < 0 || *dayOfWeek > 6 {
			return nil, errors.New("dayOfWeek must be between 0 (Sunday) and 6 (Saturday)")
		}
		if !isRecurring {
			return nil, errors.New("isRecurring must be true when dayOfWeek is set")
		}
	}
	if !isAllDay && (startTime == nil || endTime == nil) {
		return nil, errors.New("startTime and endTime are required when isAllDay is false")
	}

	closedDay := &models.ShopClosedDay{
		ShopID:      shopID,
		DayOfWeek:   dayOfWeek,
		IsAllDay:    isAllDay,
		StartTime:   startTime,
		EndTime:     endTime,
		IsRecurring: isRecurring,
		Reason:      reason,
		IsActive:    true,
	}

	// specificDate parse et
	if specificDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *specificDate)
		if err != nil {
			return nil, errors.New("invalid specificDate format, use YYYY-MM-DD")
		}
		closedDay.SpecificDate = &parsedDate
	}

	if err := s.closedDayRepo.Create(closedDay); err != nil {
		return nil, err
	}

	// Cache'i temizle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)

	return closedDay, nil
}

func (s *shopProfileService) UpdateClosedDay(shopID, userID, closedDayID uint, specificDate *string, dayOfWeek *int, isAllDay *bool, startTime, endTime *string, isRecurring *bool, reason *string) (*models.ShopClosedDay, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can update closed days")
	}

	closedDay, err := s.closedDayRepo.GetByID(closedDayID)
	if err != nil {
		return nil, errors.New("closed day not found")
	}

	if closedDay.ShopID != shopID {
		return nil, errors.New("closed day does not belong to this shop")
	}

	if specificDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *specificDate)
		if err != nil {
			return nil, errors.New("invalid specificDate format, use YYYY-MM-DD")
		}
		closedDay.SpecificDate = &parsedDate
		closedDay.DayOfWeek = nil // specificDate varsa dayOfWeek null olmalı
	}

	if dayOfWeek != nil {
		if *dayOfWeek < 0 || *dayOfWeek > 6 {
			return nil, errors.New("dayOfWeek must be between 0 (Sunday) and 6 (Saturday)")
		}
		closedDay.DayOfWeek = dayOfWeek
		closedDay.SpecificDate = nil // dayOfWeek varsa specificDate null olmalı
	}

	if isAllDay != nil {
		closedDay.IsAllDay = *isAllDay
	}

	if startTime != nil {
		closedDay.StartTime = startTime
	}

	if endTime != nil {
		closedDay.EndTime = endTime
	}

	if isRecurring != nil {
		closedDay.IsRecurring = *isRecurring
	}

	if reason != nil {
		closedDay.Reason = *reason
	}

	// Post-update validasyon
	if closedDay.SpecificDate != nil && closedDay.DayOfWeek != nil {
		return nil, errors.New("specificDate and dayOfWeek cannot be set together")
	}
	if closedDay.SpecificDate == nil && closedDay.DayOfWeek == nil {
		return nil, errors.New("either specificDate or dayOfWeek is required")
	}
	if closedDay.DayOfWeek != nil && !closedDay.IsRecurring {
		return nil, errors.New("isRecurring must be true when dayOfWeek is set")
	}
	if !closedDay.IsAllDay && (closedDay.StartTime == nil || closedDay.EndTime == nil) {
		return nil, errors.New("startTime and endTime are required when isAllDay is false")
	}

	if err := s.closedDayRepo.Update(closedDay); err != nil {
		return nil, err
	}

	// Cache'i temizle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)

	return closedDay, nil
}

func (s *shopProfileService) DeleteClosedDay(shopID, userID, closedDayID uint) error {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only shop owners can delete closed days")
	}

	closedDay, err := s.closedDayRepo.GetByID(closedDayID)
	if err != nil {
		return errors.New("closed day not found")
	}

	if closedDay.ShopID != shopID {
		return errors.New("closed day does not belong to this shop")
	}

	if err := s.closedDayRepo.Delete(closedDayID); err != nil {
		return err
	}

	// Cache'i temizle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)

	return nil
}

func (s *shopProfileService) ToggleClosedDay(shopID, userID, closedDayID uint, isActive bool) (*models.ShopClosedDay, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can toggle closed days")
	}

	closedDay, err := s.closedDayRepo.GetByID(closedDayID)
	if err != nil {
		return nil, errors.New("closed day not found")
	}

	if closedDay.ShopID != shopID {
		return nil, errors.New("closed day does not belong to this shop")
	}

	closedDay.IsActive = isActive

	if err := s.closedDayRepo.Update(closedDay); err != nil {
		return nil, err
	}

	// Cache'i temizle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)

	return closedDay, nil
}

// ==================== Working Hours ====================

func (s *shopProfileService) GetWorkingHours(shopID uint) ([]models.ShopWorkingHours, error) {
	// Shop var mı kontrol et
	_, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	workingHours, err := s.workingHoursRepo.GetByShopID(shopID)
	if err != nil {
		return nil, err
	}

	// Eğer kayıt yoksa, varsayılan değerler ile dön
	if len(workingHours) == 0 {
		// Boş array dön
		return []models.ShopWorkingHours{}, nil
	}

	return workingHours, nil
}

func (s *shopProfileService) UpdateWorkingHours(shopID, userID uint, openTime, closeTime string) ([]models.ShopWorkingHours, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can update working hours")
	}

	// Zaman formatını kontrol et (HH:MM)
	if !isValidTimeFormat(openTime) || !isValidTimeFormat(closeTime) {
		return nil, errors.New("invalid time format, use HH:MM")
	}

	// Mevcut kayıtları sil
	if err := s.workingHoursRepo.DeleteByShopID(shopID); err != nil {
		return nil, err
	}

	// Tüm günler için aynı saatleri kaydet
	var workingHours []models.ShopWorkingHours
	for day := models.Monday; day <= models.Sunday; day++ {
		wh := models.ShopWorkingHours{
			ShopID:    shopID,
			DayOfWeek: day,
			OpenTime:  openTime,
			CloseTime: closeTime,
			IsClosed:  false,
		}
		workingHours = append(workingHours, wh)
	}

	if err := s.workingHoursRepo.CreateBatch(workingHours); err != nil {
		return nil, err
	}

	// Cache'i temizle
	cacheKey := fmt.Sprintf("shop:%d", shopID)
	s.cacheService.Delete(cacheKey)

	return workingHours, nil
}

// isValidTimeFormat HH:MM formatını kontrol eder
func isValidTimeFormat(timeStr string) bool {
	if len(timeStr) != 5 {
		return false
	}
	if timeStr[2] != ':' {
		return false
	}

	// Saat ve dakika parse et
	_, err := time.Parse("15:04", timeStr)
	return err == nil
}

// ==================== System Status ====================

func (s *shopProfileService) UpdateSystemStatus(shopID, userID uint, isSystemOff bool) (*models.Shop, error) {
	// Owner kontrolü
	isOwner, err := s.barberShopRepo.IsOwner(shopID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only shop owners can update system status")
	}

	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	shop.IsSystemOff = isSystemOff

	// 1. DB'yi güncelle
	if err := s.shopRepo.Update(shop); err != nil {
		return nil, err
	}

	updatedShop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	// Owner'ları al
	owners, err := s.barberShopRepo.GetOwners(shopID)
	if err == nil {
		updatedShop.Owners = owners
	}

	// 2. Cache'i güncelle
	cacheKey := fmt.Sprintf("shop:%d", updatedShop.ID)
	s.cacheService.Set(cacheKey, updatedShop, 1*time.Hour)

	// 3. RabbitMQ ile event mesajı gönder
	s.messageService.PublishShopUpdated(updatedShop.ID)

	return updatedShop, nil
}

// ==================== My Shops ====================

func (s *shopProfileService) GetMyShops(userID uint) ([]models.Shop, error) {
	// Önce cache'den kontrol et
	cacheKey := fmt.Sprintf("my-shops:%d", userID)
	var cachedShops []models.Shop
	if err := s.cacheService.Get(cacheKey, &cachedShops); err == nil {
		return cachedShops, nil
	}

	// Cache'de yoksa DB'den al
	shops, err := s.barberShopRepo.GetOwnedShopsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Her shop için owner'ları al
	for i := range shops {
		owners, err := s.barberShopRepo.GetOwners(shops[i].ID)
		if err == nil {
			shops[i].Owners = owners
		}
	}

	// Cache'e kaydet (1 saat)
	s.cacheService.Set(cacheKey, shops, 1*time.Hour)

	return shops, nil
}
