package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
	"fmt"
	"time"
)

type BarberProfileService interface {
	RegisterAsBarber(userID uint, name string, bio string, experience int) (*models.Barber, error)
	GetBarberProfile(userID uint) (*models.Barber, error)
	UpdateBarberProfile(userID uint, name *string, bio *string, experience *int) (*models.Barber, error)
}

type barberProfileService struct {
	barberRepo     repository.BarberRepository
	userRepo       repository.UserRepository
	cacheService   CacheService
	messageService MessageService
}

func NewBarberProfileService(barberRepo repository.BarberRepository, userRepo repository.UserRepository, cacheService CacheService, messageService MessageService) BarberProfileService {
	return &barberProfileService{
		barberRepo:     barberRepo,
		userRepo:       userRepo,
		cacheService:   cacheService,
		messageService: messageService,
	}
}

func (s *barberProfileService) RegisterAsBarber(userID uint, name string, bio string, experience int) (*models.Barber, error) {
	// Kullanıcı kontrolü
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Zaten barber mı kontrol et
	existingBarber, _ := s.barberRepo.GetByUserID(userID)
	if existingBarber != nil {
		return nil, errors.New("user is already registered as barber")
	}

	barber := &models.Barber{
		UserID:     userID,
		Name:       name,
		Bio:        bio,
		Experience: experience,
		Rating: models.Rating{
			Average:     0,
			TotalReview: 0,
		},
	}

	// 1. DB'ye kaydet
	if err := s.barberRepo.Create(barber); err != nil {
		return nil, err
	}

	createdBarber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Cache'e kaydet (barberService üzerinden)
	// Not: barberService'e erişim yok, doğrudan cache kullan
	cacheKey := fmt.Sprintf("barber:%d", createdBarber.ID)
	s.cacheService.Set(cacheKey, createdBarber, 1*time.Hour)

	// 3. RabbitMQ ile event mesajı gönder
	s.messageService.PublishBarberCreated(createdBarber.ID)

	return createdBarber, nil
}

func (s *barberProfileService) GetBarberProfile(userID uint) (*models.Barber, error) {
	// Önce userID ile barber ID'sini bulmak için DB'ye bakıyoruz
	// Çünkü cache key'i barber:ID formatında
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("barber profile not found")
	}

	// Cache'den kontrol et
	cacheKey := fmt.Sprintf("barber:%d", barber.ID)
	var cachedBarber models.Barber
	if err := s.cacheService.Get(cacheKey, &cachedBarber); err == nil {
		// Avatar'ı User'dan al (cache'deki barber'da User yok)
		cachedBarber.Avatar = barber.Avatar
		return &cachedBarber, nil
	}

	// Cache'e kaydet
	s.cacheService.Set(cacheKey, barber, 1*time.Hour)

	return barber, nil
}

func (s *barberProfileService) UpdateBarberProfile(userID uint, name *string, bio *string, experience *int) (*models.Barber, error) {
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("barber profile not found")
	}

	if name != nil {
		barber.Name = *name
	}
	if bio != nil {
		barber.Bio = *bio
	}
	if experience != nil {
		barber.Experience = *experience
	}

	// 1. DB'yi güncelle
	if err := s.barberRepo.Update(barber); err != nil {
		return nil, err
	}

	updatedBarber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Cache'i güncelle
	cacheKey := fmt.Sprintf("barber:%d", updatedBarber.ID)
	s.cacheService.Set(cacheKey, updatedBarber, 1*time.Hour)

	// 3. RabbitMQ ile event mesajı gönder
	s.messageService.PublishBarberUpdated(updatedBarber.ID)

	return updatedBarber, nil
}
