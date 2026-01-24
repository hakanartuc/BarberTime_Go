package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"errors"
	"fmt"
	"time"
)

type AppointmentService interface {
	// Customer appointment methods
	GetPastAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetUpcomingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetAppointmentByID(id uint) (*models.Appointment, error)
	CreateAppointment(userID uint, barberID uint, serviceIDs []uint, appointmentTime time.Time, notes *string) (*models.Appointment, error)
	CancelAppointment(id, userID uint) error

	// Barber appointment methods
	GetBarberPendingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetBarberUpcomingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetBarberPastAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetBarberAllAppointments(userID uint, startDate, endDate *time.Time, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetBarberDailyAppointments(userID uint, date time.Time) ([]models.Appointment, error)
	ApproveAppointment(appointmentID, userID uint) error
	RejectAppointment(appointmentID, userID uint, reason *string) error
	CompleteAppointment(appointmentID, userID uint) error
}

type appointmentService struct {
	appointmentRepo repository.AppointmentRepository
	barberRepo      repository.BarberRepository
	barberShopRepo  repository.BarberShopRepository
	shopRepo        repository.ShopRepository
	serviceRepo     repository.ServiceRepository
	messageService  MessageService
}

func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	barberRepo repository.BarberRepository,
	barberShopRepo repository.BarberShopRepository,
	shopRepo repository.ShopRepository,
	serviceRepo repository.ServiceRepository,
	messageService MessageService,
) AppointmentService {
	return &appointmentService{
		appointmentRepo: appointmentRepo,
		barberRepo:      barberRepo,
		barberShopRepo:  barberShopRepo,
		shopRepo:        shopRepo,
		serviceRepo:     serviceRepo,
		messageService:  messageService,
	}
}

func (s *appointmentService) GetPastAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error) {
	appointments, total, err := s.appointmentRepo.GetPastByUserID(userID, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return appointments, pagination, nil
}

func (s *appointmentService) GetUpcomingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error) {
	appointments, total, err := s.appointmentRepo.GetUpcomingByUserID(userID, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return appointments, pagination, nil
}

func (s *appointmentService) GetAppointmentByID(id uint) (*models.Appointment, error) {
	appointment, err := s.appointmentRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("appointment not found")
	}
	return appointment, nil
}

func (s *appointmentService) CreateAppointment(userID uint, barberID uint, serviceIDs []uint, appointmentTime time.Time, notes *string) (*models.Appointment, error) {
	// Berber kontrolü
	barber, err := s.barberRepo.GetByID(barberID)
	if err != nil {
		return nil, errors.New("barber not found")
	}

	// Barber isSystemOff kontrolü
	if barber.IsSystemOff {
		return nil, errors.New("barber is currently not accepting appointments")
	}

	// Barber'ın bağlı olduğu shop'ların isSystemOff kontrolü
	barberShops, err := s.barberShopRepo.GetByBarberID(barberID)
	if err == nil {
		for _, bs := range barberShops {
			if bs.Status == models.BarberShopStatusApproved {
				shop, shopErr := s.shopRepo.GetByID(bs.ShopID)
				if shopErr == nil && shop.IsSystemOff {
					return nil, errors.New("the shop this barber belongs to is currently not accepting appointments")
				}
			}
		}
	}

	// UTC'ye çevir
	appointmentTimeUTC := appointmentTime.UTC()

	// Geçmiş tarih kontrolü
	if appointmentTimeUTC.Before(time.Now().UTC()) {
		return nil, errors.New("cannot create appointment in the past")
	}

	// Distributed lock key oluştur
	lockKey := fmt.Sprintf("appointment:lock:%d:%s", barberID, appointmentTimeUTC.Format(time.RFC3339))
	lockTTL := 30 * time.Second

	// Distributed lock ile race condition'ı engelle
	var appointment *models.Appointment
	err = utils.WithLock(lockKey, lockTTL, func() error {
		// Lock içinde tekrar müsaitlik kontrolü (double-check)
		// Aynı saat için randevu var mı kontrol et (±1 dakika tolerans)
		startCheck := appointmentTimeUTC.Add(-1 * time.Minute)
		endCheck := appointmentTimeUTC.Add(1 * time.Minute)

		existingAppointments, err := s.appointmentRepo.GetByBarberIDAndDateRange(barberID, startCheck, endCheck)
		if err == nil {
			for _, apt := range existingAppointments {
				if apt.Status != models.StatusRejected {
					return errors.New("this time slot is already booked")
				}
			}
		}

		// Servisleri kontrol et
		var services []models.Service
		for _, serviceID := range serviceIDs {
			service, err := s.serviceRepo.GetByID(serviceID)
			if err != nil {
				return errors.New("service not found")
			}
			services = append(services, *service)
		}

		appointment = &models.Appointment{
			UserID:          userID,
			BarberID:        barberID,
			AppointmentTime: appointmentTimeUTC,
			Status:          models.StatusPending,
			Notes:           notes,
			Services:        services,
		}

		if err := s.appointmentRepo.Create(appointment); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// RabbitMQ'ya mesaj gönder
	s.messageService.PublishAppointmentCreated(appointment.ID, userID, barberID)

	// İlişkileri yükle
	return s.appointmentRepo.GetByID(appointment.ID)
}

func (s *appointmentService) CancelAppointment(id, userID uint) error {
	appointment, err := s.appointmentRepo.GetByID(id)
	if err != nil {
		return errors.New("appointment not found")
	}

	if appointment.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.appointmentRepo.Delete(id); err != nil {
		return err
	}

	// RabbitMQ'ya mesaj gönder
	s.messageService.PublishAppointmentCancelled(id, userID)

	return nil
}

// getBarberByUserID helper function to get barber by user ID
func (s *appointmentService) getBarberByUserID(userID uint) (*models.Barber, error) {
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("barber profile not found")
	}
	return barber, nil
}

// GetBarberPendingAppointments returns pending appointments for the barber
func (s *appointmentService) GetBarberPendingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error) {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	appointments, total, err := s.appointmentRepo.GetPendingByBarberID(barber.ID, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return appointments, pagination, nil
}

// GetBarberUpcomingAppointments returns approved upcoming appointments for the barber
func (s *appointmentService) GetBarberUpcomingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error) {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	appointments, total, err := s.appointmentRepo.GetUpcomingByBarberID(barber.ID, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return appointments, pagination, nil
}

// GetBarberPastAppointments returns past appointments for the barber
func (s *appointmentService) GetBarberPastAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error) {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	appointments, total, err := s.appointmentRepo.GetPastByBarberID(barber.ID, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return appointments, pagination, nil
}

// ApproveAppointment approves a pending appointment
func (s *appointmentService) ApproveAppointment(appointmentID, userID uint) error {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return err
	}

	appointment, err := s.appointmentRepo.GetByID(appointmentID)
	if err != nil {
		return errors.New("appointment not found")
	}

	// Verify this appointment belongs to this barber
	if appointment.BarberID != barber.ID {
		return errors.New("unauthorized")
	}

	// Only pending appointments can be approved
	if appointment.Status != models.StatusPending {
		return errors.New("only pending appointments can be approved")
	}

	appointment.Status = models.StatusApproved
	if err := s.appointmentRepo.Update(appointment); err != nil {
		return err
	}

	// Publish event
	s.messageService.PublishAppointmentApproved(appointmentID, appointment.UserID, barber.ID)

	return nil
}

// RejectAppointment rejects an appointment
func (s *appointmentService) RejectAppointment(appointmentID, userID uint, reason *string) error {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return err
	}

	appointment, err := s.appointmentRepo.GetByID(appointmentID)
	if err != nil {
		return errors.New("appointment not found")
	}

	// Verify this appointment belongs to this barber
	if appointment.BarberID != barber.ID {
		return errors.New("unauthorized")
	}

	// Only pending or approved appointments can be rejected
	if appointment.Status != models.StatusPending && appointment.Status != models.StatusApproved {
		return errors.New("only pending or approved appointments can be rejected")
	}

	appointment.Status = models.StatusRejected
	if reason != nil {
		appointment.Notes = reason
	}
	if err := s.appointmentRepo.Update(appointment); err != nil {
		return err
	}

	// Publish event
	s.messageService.PublishAppointmentRejected(appointmentID, appointment.UserID, barber.ID)

	return nil
}

// CompleteAppointment marks an appointment as completed
func (s *appointmentService) CompleteAppointment(appointmentID, userID uint) error {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return err
	}

	appointment, err := s.appointmentRepo.GetByID(appointmentID)
	if err != nil {
		return errors.New("appointment not found")
	}

	// Verify this appointment belongs to this barber
	if appointment.BarberID != barber.ID {
		return errors.New("unauthorized")
	}

	// Only approved appointments can be completed
	if appointment.Status != models.StatusApproved {
		return errors.New("only approved appointments can be completed")
	}

	// Appointment time must be in the past
	if appointment.AppointmentTime.After(time.Now().UTC()) {
		return errors.New("cannot complete an appointment before its scheduled time")
	}

	appointment.Status = models.StatusCompleted
	if err := s.appointmentRepo.Update(appointment); err != nil {
		return err
	}

	// Publish event
	s.messageService.PublishAppointmentCompleted(appointmentID, appointment.UserID, barber.ID)

	return nil
}

// GetBarberAllAppointments returns all appointments for the barber with optional date filtering
// Returns all statuses: pending, approved, rejected, completed
func (s *appointmentService) GetBarberAllAppointments(userID uint, startDate, endDate *time.Time, page, limit int) ([]models.Appointment, utils.Pagination, error) {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	appointments, total, err := s.appointmentRepo.GetAllByBarberID(barber.ID, startDate, endDate, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return appointments, pagination, nil
}

// GetBarberDailyAppointments returns all appointments for the barber on a specific date
// Returns all statuses: pending, approved, rejected, completed
func (s *appointmentService) GetBarberDailyAppointments(userID uint, date time.Time) ([]models.Appointment, error) {
	barber, err := s.getBarberByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.appointmentRepo.GetDailyByBarberID(barber.ID, date)
}
