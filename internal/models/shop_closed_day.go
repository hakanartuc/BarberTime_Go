package models

import (
	"time"

	"gorm.io/gorm"
)

type ShopClosedDay struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ShopID       uint           `gorm:"not null;index" json:"shopId"`
	SpecificDate *time.Time     `json:"specificDate"`
	DayOfWeek    *int           `json:"dayOfWeek"`
	IsAllDay     bool           `gorm:"default:true" json:"isAllDay"`
	StartTime    *string        `gorm:"type:varchar(5)" json:"startTime"`
	EndTime      *string        `gorm:"type:varchar(5)" json:"endTime"`
	IsRecurring  bool           `gorm:"default:false" json:"isRecurring"`
	Reason       string         `json:"reason"`
	IsActive     bool           `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
