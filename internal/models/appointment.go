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
	ID              uint              `gorm:"primaryKey" json:"id"`
	UserID          uint              `gorm:"not null;index" json:"-"`
	User            User              `gorm:"foreignKey:UserID" json:"user"`
	BarberID        uint              `gorm:"not null;index" json:"-"`
	Barber          Barber            `gorm:"foreignKey:BarberID" json:"barber"`
	AppointmentTime time.Time         `gorm:"type:timestamptz;not null;index" json:"appointmentTime"` // UTC timestamp
	Status          AppointmentStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Notes           *string           `json:"notes"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"-"`
	DeletedAt       gorm.DeletedAt    `gorm:"index" json:"-"`

	Services []Service `gorm:"many2many:appointment_services;" json:"services"`
}

// Date returns the date part of AppointmentTime in YYYY-MM-DD format (UTC)
func (a *Appointment) Date() string {
	return a.AppointmentTime.UTC().Format("2006-01-02")
}

// Time returns the time part of AppointmentTime in HH:mm format (UTC)
func (a *Appointment) Time() string {
	return a.AppointmentTime.UTC().Format("15:04")
}
