package models

import (
	"time"

	"gorm.io/gorm"
)

type RequestStatus string

const (
	RequestStatusPending  RequestStatus = "pending"
	RequestStatusApproved RequestStatus = "approved"
	RequestStatusRejected RequestStatus = "rejected"
)

type BarberShopRequest struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	BarberID       uint           `gorm:"not null;index" json:"barberId"`
	Barber         Barber         `gorm:"foreignKey:BarberID" json:"barber,omitempty"`
	ShopID         uint           `gorm:"not null;index" json:"shopId"`
	Shop           Shop           `gorm:"foreignKey:ShopID" json:"shop,omitempty"`
	Status         RequestStatus  `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Message        string         `json:"message"`              // Barber'ın mesajı
	ApprovedBy     *uint          `json:"approvedBy,omitempty"` // Onaylayan owner'ın user ID'si
	ApprovedByUser *User          `gorm:"foreignKey:ApprovedBy" json:"approvedByUser,omitempty"`
	RejectedReason string         `json:"rejectedReason,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
