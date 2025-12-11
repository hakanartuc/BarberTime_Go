package models

import (
	"time"

	"gorm.io/gorm"
)

type PhotoType string

const (
	PhotoTypeBarber PhotoType = "barber"
	PhotoTypeShop   PhotoType = "shop"
)

type Photo struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TargetID  uint           `gorm:"not null;index" json:"targetId"` // BarberID veya ShopID
	TargetType PhotoType     `gorm:"type:varchar(20);not null;index" json:"targetType"` // "barber" veya "shop"
	ImageURL  string         `gorm:"not null" json:"imageUrl"`
	Caption   string         `json:"caption"` // Fotoğraf açıklaması (opsiyonel)
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// İlişkiler
	Likes []PhotoLike `gorm:"foreignKey:PhotoID" json:"likes,omitempty"`
}

type PhotoLike struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PhotoID   uint           `gorm:"not null;index" json:"photoId"`
	Photo     Photo          `gorm:"foreignKey:PhotoID" json:"photo,omitempty"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

