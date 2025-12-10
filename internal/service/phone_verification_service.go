package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type PhoneVerificationService interface {
	GenerateCode(phoneNumber, deviceID string) (string, error)
	VerifyCode(phoneNumber, deviceID, code string) (bool, error)
	CleanupExpiredCodes() error
}

type phoneVerificationService struct {
	codeRepo repository.PhoneVerificationCodeRepository
	userRepo repository.UserRepository
}

func NewPhoneVerificationService(codeRepo repository.PhoneVerificationCodeRepository, userRepo repository.UserRepository) PhoneVerificationService {
	return &phoneVerificationService{
		codeRepo: codeRepo,
		userRepo: userRepo,
	}
}

// GenerateCode 6 haneli random kod üretir ve DB'ye kaydeder
func (s *phoneVerificationService) GenerateCode(phoneNumber, deviceID string) (string, error) {
	// 6 haneli random kod üret (100000-999999)
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random code: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64()+100000)

	// Geçerlilik süresi: 10 dakika
	expiresAt := time.Now().Add(10 * time.Minute)

	verificationCode := &models.PhoneVerificationCode{
		PhoneNumber: phoneNumber,
		DeviceID:    deviceID,
		Code:        code,
		ExpiresAt:   expiresAt,
		IsUsed:      false,
	}

	if err := s.codeRepo.Create(verificationCode); err != nil {
		return "", fmt.Errorf("failed to save verification code: %w", err)
	}

	// TODO: SMS provider entegrasyonu eklenecek
	// Şimdilik sadece log atıyoruz
	fmt.Printf("[TODO] SMS gönderilecek: %s numarasına kod: %s\n", phoneNumber, code)

	return code, nil
}

// VerifyCode kod doğrulaması yapar ve kullanıcının IsPhoneConfirmed alanını true yapar
func (s *phoneVerificationService) VerifyCode(phoneNumber, deviceID, code string) (bool, error) {
	// Kodu kontrol et
	verificationCode, err := s.codeRepo.GetByPhoneAndCode(phoneNumber, code)
	if err != nil {
		return false, errors.New("invalid or expired verification code")
	}

	// Device ID kontrolü
	if verificationCode.DeviceID != deviceID {
		return false, errors.New("device id mismatch")
	}

	// Süresi dolmuş mu kontrol et
	if time.Now().After(verificationCode.ExpiresAt) {
		return false, errors.New("verification code expired")
	}

	// Zaten kullanılmış mı kontrol et
	if verificationCode.IsUsed {
		return false, errors.New("verification code already used")
	}

	// Kodu kullanılmış olarak işaretle
	if err := s.codeRepo.MarkAsUsed(verificationCode.ID); err != nil {
		return false, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Kullanıcının telefon numarasını doğrulanmış olarak işaretle
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return false, errors.New("user not found")
	}

	user.IsPhoneConfirmed = true
	if err := s.userRepo.Update(user); err != nil {
		return false, fmt.Errorf("failed to update user: %w", err)
	}

	return true, nil
}

// CleanupExpiredCodes süresi dolmuş kodları temizler
func (s *phoneVerificationService) CleanupExpiredCodes() error {
	return s.codeRepo.DeleteExpired()
}
