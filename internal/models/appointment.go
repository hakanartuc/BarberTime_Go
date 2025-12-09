package models

import (
	"time"

	"gorm.io/gorm"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusApproved  AppointmentStatus = "approved"
	StatusRejected  AppointmentStatus = "rejected"
	StatusCompleted AppointmentStatus = "completed"
)

type Appointment struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	UserID    uint            `gorm:"not null;index" json:"-"`
	User      User            `gorm:"foreignKey:UserID" json:"user"`
	BarberID  uint            `gorm:"not null;index" json:"-"`
	Barber    Barber          `gorm:"foreignKey:BarberID" json:"barber"`
	Date      string          `gorm:"type:date;not null" json:"date"` // YYYY-MM-DD formatında
	Time      string          `gorm:"type:time;not null" json:"time"` // HH:mm formatında
	Status    AppointmentStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Notes     *string         `json:"notes"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"-"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`

	Services []Service `gorm:"many2many:appointment_services;" json:"services"`
}

