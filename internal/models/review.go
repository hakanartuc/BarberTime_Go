package models

import (
	"time"

	"gorm.io/gorm"
)

type ReviewType string

const (
	ReviewTypeShop   ReviewType = "shop"
	ReviewTypeBarber ReviewType = "barber"
)

type Review struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TargetType    string         `gorm:"type:varchar(20);not null" json:"type"` // GORM polymorphic için gerekli, JSON'da "type" olarak görünür
	TargetID      uint           `gorm:"not null;index" json:"targetId"`
	UserID        uint           `gorm:"not null;index" json:"-"`
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Rating        int            `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment       string         `json:"comment"`
	AppointmentID *uint          `json:"appointmentId,omitempty"`
	Appointment   *Appointment   `gorm:"foreignKey:AppointmentID" json:"appointment,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
