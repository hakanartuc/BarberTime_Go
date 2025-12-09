package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	PhoneNumber string         `gorm:"uniqueIndex;not null" json:"phoneNumber"`
	Avatar      *string        `json:"avatar"`
	DeviceID    string         `gorm:"index" json:"-"`
	IsAdmin     bool           `gorm:"default:false" json:"isAdmin"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
