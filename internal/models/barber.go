package models

import (
	"time"

	"gorm.io/gorm"
)

type Barber struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"uniqueIndex;not null" json:"userId"`
	Name       string         `gorm:"not null" json:"name"`
	Avatar     *string        `gorm:"-" json:"avatar"` // DB'de yok, User'dan alınır
	Bio        string         `json:"bio"`
	Experience int            `json:"experience"` // Yıl cinsinden
	Rating     Rating         `gorm:"embedded;embeddedPrefix:rating_" json:"rating"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User         User                `gorm:"foreignKey:UserID" json:"-"`
	Shops        []Shop              `gorm:"many2many:barber_shops;" json:"shops,omitempty"`
	Appointments []Appointment       `gorm:"foreignKey:BarberID" json:"-"`
	Reviews      []Review            `gorm:"polymorphic:Target;polymorphicValue:barber" json:"reviews,omitempty"`
	ShopRequests []BarberShopRequest `gorm:"foreignKey:BarberID" json:"-"`
	BarberShops  []BarberShop        `gorm:"foreignKey:BarberID" json:"-"`
}

type Rating struct {
	Average     float64 `gorm:"default:0" json:"average"`
	TotalReview int     `gorm:"default:0" json:"totalReview"`
}
