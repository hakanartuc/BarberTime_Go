package models

import (
	"time"

	"gorm.io/gorm"
)

type ShopStatus string

const (
	ShopStatusPending   ShopStatus = "pending"
	ShopStatusConfirmed ShopStatus = "confirmed"
	ShopStatusRejected  ShopStatus = "rejected"
)

type Shop struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"not null" json:"name"`
	Description  string         `json:"description"`
	PhoneNumber  string         `json:"phoneNumber"`
	Email        string         `json:"email"`
	Website      string         `json:"website"`
	Location     Location       `gorm:"embedded;embeddedPrefix:location_" json:"location"`
	WorkingHours string         `gorm:"type:text" json:"workingHours"`                    // JSON formatında saklanacak
	ClosedDays   string         `gorm:"type:text" json:"closedDays"`                      // JSON formatında saklanacak
	Status       ShopStatus     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, confirmed, rejected
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Owners         []User              `gorm:"-" json:"owners,omitempty"` // DB'de yok, runtime'da doldurulur
	Barbers        []Barber            `gorm:"many2many:barber_shops;" json:"barbers,omitempty"`
	ShopServices   []ShopService       `gorm:"foreignKey:ShopID" json:"shopServices,omitempty"`
	Reviews        []Review            `gorm:"polymorphic:Target;polymorphicValue:shop" json:"reviews,omitempty"`
	BarberRequests []BarberShopRequest `gorm:"foreignKey:ShopID" json:"-"`
	BarberShops    []BarberShop        `gorm:"foreignKey:ShopID" json:"-"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}
