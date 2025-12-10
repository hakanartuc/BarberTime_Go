package models

import (
	"time"

	"gorm.io/gorm"
)

type PhoneVerificationCode struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PhoneNumber string         `gorm:"not null;index" json:"phoneNumber"`
	DeviceID    string         `gorm:"not null;index" json:"deviceId"`
	Code        string         `gorm:"not null;index" json:"code"`
	ExpiresAt   time.Time      `gorm:"not null;index" json:"expiresAt"`
	IsUsed      bool           `gorm:"default:false" json:"isUsed"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

