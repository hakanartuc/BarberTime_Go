package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"
	"time"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	GetByID(id uint) (*models.Appointment, error)
	GetPastByUserID(userID uint, page, limit int) ([]models.Appointment, int64, error)
	GetUpcomingByUserID(userID uint, page, limit int) ([]models.Appointment, int64, error)
	GetByBarberIDAndDate(barberID uint, date string) ([]models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository() AppointmentRepository {
	return &appointmentRepository{
		db: database.DB,
	}
}

func (r *appointmentRepository) Create(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *appointmentRepository) GetByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("User").Preload("Barber").Preload("Barber.Shop").Preload("Services").First(&appointment, id).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) GetPastByUserID(userID uint, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit
	today := time.Now().Format("2006-01-02")

	query := r.db.Where("user_id = ? AND (date < ? OR (date = ? AND time < ?))", 
		userID, today, today, time.Now().Format("15:04")).
		Preload("User").Preload("Barber").Preload("Barber.Shop").Preload("Services").
		Order("date DESC, time DESC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

func (r *appointmentRepository) GetUpcomingByUserID(userID uint, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit
	today := time.Now().Format("2006-01-02")

	query := r.db.Where("user_id = ? AND (date > ? OR (date = ? AND time >= ?))", 
		userID, today, today, time.Now().Format("15:04")).
		Preload("User").Preload("Barber").Preload("Barber.Shop").Preload("Services").
		Order("date ASC, time ASC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

func (r *appointmentRepository) GetByBarberIDAndDate(barberID uint, date string) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("barber_id = ? AND date = ?", barberID, date).
		Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *appointmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Appointment{}, id).Error
}

