package models

import (
	"time"

	"gorm.io/gorm"
)

type BarberShopStatus string

const (
	BarberShopStatusPending  BarberShopStatus = "pending"
	BarberShopStatusApproved BarberShopStatus = "approved"
	BarberShopStatusRejected BarberShopStatus = "rejected"
)

type BarberShop struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	BarberID       uint             `gorm:"not null;index" json:"barberId"`
	Barber         Barber           `gorm:"foreignKey:BarberID" json:"barber,omitempty"`
	ShopID         uint             `gorm:"not null;index" json:"shopId"`
	Shop           Shop             `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Status         BarberShopStatus `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, approved, rejected
	IsOwner        bool             `gorm:"default:false" json:"isOwner"`                     // Shop owner mı?
	ApprovedBy     *uint            `json:"approvedBy,omitempty"`                             // Onaylayan owner'ın user ID'si
	ApprovedByUser *User            `gorm:"foreignKey:ApprovedBy" json:"approvedByUser,omitempty"`
	RejectedReason string           `json:"rejectedReason,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"-"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"-"`

	// Unique constraint: bir barber aynı shop'a birden fazla kez kayıt olamaz
	// Bu GORM'de composite unique index ile sağlanacak
}
