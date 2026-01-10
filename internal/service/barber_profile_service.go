package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type BarberProfileService interface {
	RegisterAsBarber(userID uint, name string, bio string, experience int) (*models.Barber, error)
	GetBarberProfile(userID uint) (*models.Barber, error)
	UpdateBarberProfile(userID uint, name *string, bio *string, experience *int, username *string) (*models.Barber, error)
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

	// Username oluştur (isim + random sayı)
	username := s.generateUsername(name)

	barber := &models.Barber{
		UserID:     userID,
		Username:   username,
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

func (s *barberProfileService) UpdateBarberProfile(userID uint, name *string, bio *string, experience *int, username *string) (*models.Barber, error) {
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
	if username != nil {
		// Username unique kontrolü (kendi ID'si hariç)
		exists, err := s.barberRepo.ExistsByUsername(*username, barber.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("username already taken")
		}
		barber.Username = *username
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

// generateUsername isimden username oluşturur (isim + random sayı)
func (s *barberProfileService) generateUsername(name string) string {
	// Türkçe karakterleri ASCII'ye çevir
	baseUsername := s.normalizeToASCII(name)

	// Küçük harfe çevir ve boşlukları alt çizgiye çevir
	baseUsername = strings.ToLower(baseUsername)
	baseUsername = strings.ReplaceAll(baseUsername, " ", "_")

	// Sadece harf, rakam ve alt çizgi bırak
	reg := regexp.MustCompile("[^a-z0-9_]")
	baseUsername = reg.ReplaceAllString(baseUsername, "")

	// Birden fazla alt çizgiyi teke indir
	reg = regexp.MustCompile("_+")
	baseUsername = reg.ReplaceAllString(baseUsername, "_")

	// Başta ve sonda alt çizgi varsa kaldır
	baseUsername = strings.Trim(baseUsername, "_")

	// Eğer boş kaldıysa default bir isim ver
	if baseUsername == "" {
		baseUsername = "barber"
	}

	// Random sayı ekle (1000-9999 arası)
	randomNum := rand.Intn(9000) + 1000
	username := fmt.Sprintf("%s%d", baseUsername, randomNum)

	// Unique olana kadar dene (yeni kayıt olduğu için excludeID = 0)
	for {
		exists, _ := s.barberRepo.ExistsByUsername(username, 0)
		if !exists {
			break
		}
		randomNum = rand.Intn(9000) + 1000
		username = fmt.Sprintf("%s%d", baseUsername, randomNum)
	}

	return username
}

// normalizeToASCII Türkçe karakterleri ASCII karşılıklarına çevirir
func (s *barberProfileService) normalizeToASCII(str string) string {
	// Türkçe karakter mapping
	replacer := strings.NewReplacer(
		"ı", "i", "İ", "I",
		"ğ", "g", "Ğ", "G",
		"ü", "u", "Ü", "U",
		"ş", "s", "Ş", "S",
		"ö", "o", "Ö", "O",
		"ç", "c", "Ç", "C",
	)
	str = replacer.Replace(str)

	// Unicode normalization ile diğer aksanlı karakterleri temizle
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, str)

	return result
}
