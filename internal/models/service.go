package models

import (
	"time"

	"gorm.io/gorm"
)

type Service struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null;uniqueIndex" json:"name"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ShopServices []ShopService `gorm:"foreignKey:ServiceID" json:"-"`
	Appointments []Appointment `gorm:"many2many:appointment_services;" json:"-"`
}
