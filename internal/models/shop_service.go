package models

import (
	"time"

	"gorm.io/gorm"
)

type ShopService struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ShopID      uint           `gorm:"not null;index" json:"shopId"`
	ServiceID   uint           `gorm:"not null;index" json:"serviceId"`
	Service     Service        `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Duration    int            `gorm:"not null" json:"duration"` // Dakika cinsinden
	Price       float64        `gorm:"not null" json:"price"`
	Description *string        `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Unique constraint: bir shop aynı service'e birden fazla kez sahip olamaz
}
