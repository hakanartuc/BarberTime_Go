package models

import (
	"time"

	"gorm.io/gorm"
)

type DayOfWeek int

const (
	Monday DayOfWeek = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

func (d DayOfWeek) String() string {
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	if d < 0 || int(d) >= len(days) {
		return ""
	}
	return days[d]
}

type ShopWorkingHours struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"not null;index" json:"shopId"`
	DayOfWeek DayOfWeek      `gorm:"not null" json:"dayOfWeek"`
	OpenTime  string         `gorm:"type:varchar(5)" json:"openTime"`  // Format: HH:MM
	CloseTime string         `gorm:"type:varchar(5)" json:"closeTime"` // Format: HH:MM
	IsClosed  bool           `gorm:"default:false" json:"isClosed"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
