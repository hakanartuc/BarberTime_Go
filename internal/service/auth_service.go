package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"time"
)

type AuthService interface {
	Login(phoneNumber, deviceID string) (bool, *AuthResponse, error)
}

type AuthResponse struct {
	RequiresVerification bool      `json:"requiresVerification"`
	AuthData             *AuthData `json:"authData,omitempty"`
}

type AuthData struct {
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Login(phoneNumber, deviceID string) (bool, *AuthResponse, error) {
	// Kullanıcıyı telefon numarasına göre bul
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)

	// Kullanıcı yoksa yeni kullanıcı oluştur
	if err != nil {
		user = &models.User{
			Name:        "",
			PhoneNumber: phoneNumber,
			DeviceID:    deviceID,
		}
		if err := s.userRepo.Create(user); err != nil {
			return false, nil, err
		}
	} else {
		// Device ID'yi güncelle
		user.DeviceID = deviceID
		if err := s.userRepo.Update(user); err != nil {
			return false, nil, err
		}
	}

	// JWT token oluştur
	token, expiresAt, err := utils.GenerateToken(user.ID, user.PhoneNumber)
	if err != nil {
		return false, nil, err
	}

	// Şimdilik verification gerektirmiyoruz (basit implementasyon)
	// Gerçek uygulamada SMS/OTP doğrulaması yapılabilir
	requiresVerification := false

	response := &AuthResponse{
		RequiresVerification: requiresVerification,
		AuthData: &AuthData{
			AccessToken: token,
			ExpiresAt:   expiresAt,
		},
	}

	return requiresVerification, response, nil
}
