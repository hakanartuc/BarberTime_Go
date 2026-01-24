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
	GetByBarberIDAndDateRange(barberID uint, startTime, endTime time.Time) ([]models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error

	// Barber appointment methods
	GetPendingByBarberID(barberID uint, page, limit int) ([]models.Appointment, int64, error)
	GetUpcomingByBarberID(barberID uint, page, limit int) ([]models.Appointment, int64, error)
	GetPastByBarberID(barberID uint, page, limit int) ([]models.Appointment, int64, error)
	GetAllByBarberID(barberID uint, startDate, endDate *time.Time, page, limit int) ([]models.Appointment, int64, error)
	GetDailyByBarberID(barberID uint, date time.Time) ([]models.Appointment, error)
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
	err := r.db.Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").First(&appointment, id).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) GetPastByUserID(userID uint, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit
	now := time.Now().UTC()

	query := r.db.Model(&models.Appointment{}).
		Where("user_id = ? AND appointment_time < ?", userID, now).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").
		Order("appointment_time DESC")

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
	now := time.Now().UTC()

	query := r.db.Model(&models.Appointment{}).
		Where("user_id = ? AND appointment_time >= ?", userID, now).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").
		Order("appointment_time ASC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

func (r *appointmentRepository) GetByBarberIDAndDateRange(barberID uint, startTime, endTime time.Time) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("barber_id = ? AND appointment_time >= ? AND appointment_time < ?", barberID, startTime.UTC(), endTime.UTC()).
		Order("appointment_time ASC").
		Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *appointmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Appointment{}, id).Error
}

// GetPendingByBarberID returns pending appointments for a barber
func (r *appointmentRepository) GetPendingByBarberID(barberID uint, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit

	query := r.db.Model(&models.Appointment{}).
		Where("barber_id = ? AND status = ?", barberID, models.StatusPending).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").
		Order("appointment_time ASC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

// GetUpcomingByBarberID returns approved upcoming appointments for a barber
func (r *appointmentRepository) GetUpcomingByBarberID(barberID uint, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit
	now := time.Now().UTC()

	query := r.db.Model(&models.Appointment{}).
		Where("barber_id = ? AND status = ? AND appointment_time >= ?", barberID, models.StatusApproved, now).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").
		Order("appointment_time ASC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

// GetPastByBarberID returns past appointments for a barber (completed or past time)
func (r *appointmentRepository) GetPastByBarberID(barberID uint, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit
	now := time.Now().UTC()

	query := r.db.Model(&models.Appointment{}).
		Where("barber_id = ? AND (status = ? OR appointment_time < ?)", barberID, models.StatusCompleted, now).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").
		Order("appointment_time DESC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

// GetAllByBarberID returns all appointments for a barber with optional date filtering
// Returns all statuses (pending, approved, rejected, completed)
func (r *appointmentRepository) GetAllByBarberID(barberID uint, startDate, endDate *time.Time, page, limit int) ([]models.Appointment, int64, error) {
	var appointments []models.Appointment
	var total int64

	offset := (page - 1) * limit

	query := r.db.Model(&models.Appointment{}).
		Where("barber_id = ?", barberID).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services")

	// Tarih filtreleri
	if startDate != nil {
		query = query.Where("appointment_time >= ?", startDate.UTC())
	}
	if endDate != nil {
		// endDate'in sonuna kadar dahil etmek için +1 gün
		endOfDay := endDate.UTC().Add(24 * time.Hour)
		query = query.Where("appointment_time < ?", endOfDay)
	}

	query = query.Order("appointment_time DESC")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&appointments).Error
	return appointments, total, err
}

// GetDailyByBarberID returns all appointments for a barber on a specific date
// Returns all statuses (pending, approved, rejected, completed)
func (r *appointmentRepository) GetDailyByBarberID(barberID uint, date time.Time) ([]models.Appointment, error) {
	var appointments []models.Appointment

	// Günün başlangıcı ve sonu
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&models.Appointment{}).
		Where("barber_id = ? AND appointment_time >= ? AND appointment_time < ?", barberID, startOfDay, endOfDay).
		Preload("User").Preload("Barber").Preload("Barber.Shops").Preload("Services").
		Order("appointment_time ASC").
		Find(&appointments).Error

	return appointments, err
}
