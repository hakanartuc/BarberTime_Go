package models

import (
	"time"

	"gorm.io/gorm"
)

type FavoriteType string

const (
	FavoriteTypeShop   FavoriteType = "shop"
	FavoriteTypeBarber FavoriteType = "barber"
)

type Favorite struct {
	ID       uint         `gorm:"primaryKey" json:"id"`
	UserID   uint         `gorm:"not null;index" json:"-"`
	User     User         `gorm:"foreignKey:UserID" json:"-"`
	Type     FavoriteType `gorm:"type:varchar(20);not null" json:"type"`
	TargetID uint         `gorm:"not null;index" json:"targetId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Unique constraint: bir kullanıcı aynı hedefi sadece bir kez favorilere ekleyebilir
	// Bu GORM'de composite index ile sağlanacak
}

