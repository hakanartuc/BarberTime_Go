package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"fmt"
	"time"
)

type AuthService interface {
	Login(phoneNumber, deviceID string) (bool, *AuthResponse, error)
	VerifyPhone(phoneNumber, deviceID, code string) (*AuthResponse, error)
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
	userRepo             repository.UserRepository
	phoneVerificationSvc PhoneVerificationService
}

func NewAuthService(userRepo repository.UserRepository, phoneVerificationSvc PhoneVerificationService) AuthService {
	return &authService{
		userRepo:             userRepo,
		phoneVerificationSvc: phoneVerificationSvc,
	}
}

func (s *authService) Login(phoneNumber, deviceID string) (bool, *AuthResponse, error) {
	// Kullanıcıyı telefon numarasına göre bul
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)

	// Kullanıcı yoksa yeni kullanıcı oluştur
	if err != nil {
		user = &models.User{
			Name:             "",
			PhoneNumber:      phoneNumber,
			DeviceID:         deviceID,
			IsPhoneConfirmed: false,
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

	// Telefon doğrulanmış mı kontrol et
	if !user.IsPhoneConfirmed {
		// SMS kodu gönder
		_, err := s.phoneVerificationSvc.GenerateCode(phoneNumber, deviceID)
		if err != nil {
			return false, nil, fmt.Errorf("failed to generate verification code: %w", err)
		}

		// Token verme, verification gerekiyor
		return true, &AuthResponse{
			RequiresVerification: true,
			AuthData:             nil,
		}, nil
	}

	// Telefon doğrulanmışsa token ver
	token, expiresAt, err := utils.GenerateToken(user.ID, user.PhoneNumber)
	if err != nil {
		return false, nil, err
	}

	response := &AuthResponse{
		RequiresVerification: false,
		AuthData: &AuthData{
			AccessToken: token,
			ExpiresAt:   expiresAt,
		},
	}

	return false, response, nil
}

func (s *authService) VerifyPhone(phoneNumber, deviceID, code string) (*AuthResponse, error) {
	// Kodu doğrula
	verified, err := s.phoneVerificationSvc.VerifyCode(phoneNumber, deviceID, code)
	if err != nil {
		return nil, err
	}

	if !verified {
		return nil, fmt.Errorf("verification failed")
	}

	// Kullanıcıyı al
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// JWT token oluştur
	token, expiresAt, err := utils.GenerateToken(user.ID, user.PhoneNumber)
	if err != nil {
		return nil, err
	}

	response := &AuthResponse{
		RequiresVerification: false,
		AuthData: &AuthData{
			AccessToken: token,
			ExpiresAt:   expiresAt,
		},
	}

	return response, nil
}
