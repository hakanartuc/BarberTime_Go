package models

import (
	"time"

	"gorm.io/gorm"
)

type ShopClosedDay struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"not null;index" json:"shopId"`
	Date      time.Time      `gorm:"not null" json:"date"`
	Reason    string         `json:"reason"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
