package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"
	"time"

	"gorm.io/gorm"
)

type PhoneVerificationCodeRepository interface {
	Create(code *models.PhoneVerificationCode) error
	GetByPhoneAndCode(phoneNumber, code string) (*models.PhoneVerificationCode, error)
	GetLatestByPhoneAndDevice(phoneNumber, deviceID string) (*models.PhoneVerificationCode, error)
	MarkAsUsed(id uint) error
	DeleteExpired() error
}

type phoneVerificationCodeRepository struct {
	db *gorm.DB
}

func NewPhoneVerificationCodeRepository() PhoneVerificationCodeRepository {
	return &phoneVerificationCodeRepository{
		db: database.DB,
	}
}

func (r *phoneVerificationCodeRepository) Create(code *models.PhoneVerificationCode) error {
	return r.db.Create(code).Error
}

func (r *phoneVerificationCodeRepository) GetByPhoneAndCode(phoneNumber, code string) (*models.PhoneVerificationCode, error) {
	var verificationCode models.PhoneVerificationCode
	err := r.db.Where("phone_number = ? AND code = ? AND is_used = ? AND expires_at > ?", phoneNumber, code, false, time.Now()).
		Order("created_at DESC").
		First(&verificationCode).Error
	if err != nil {
		return nil, err
	}
	return &verificationCode, nil
}

func (r *phoneVerificationCodeRepository) GetLatestByPhoneAndDevice(phoneNumber, deviceID string) (*models.PhoneVerificationCode, error) {
	var verificationCode models.PhoneVerificationCode
	err := r.db.Where("phone_number = ? AND device_id = ? AND is_used = ? AND expires_at > ?", phoneNumber, deviceID, false, time.Now()).
		Order("created_at DESC").
		First(&verificationCode).Error
	if err != nil {
		return nil, err
	}
	return &verificationCode, nil
}

func (r *phoneVerificationCodeRepository) MarkAsUsed(id uint) error {
	return r.db.Model(&models.PhoneVerificationCode{}).Where("id = ?", id).Update("is_used", true).Error
}

func (r *phoneVerificationCodeRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? OR is_used = ?", time.Now(), true).Delete(&models.PhoneVerificationCode{}).Error
}
