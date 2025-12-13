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
	GetPastAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetUpcomingAppointments(userID uint, page, limit int) ([]models.Appointment, utils.Pagination, error)
	GetAppointmentByID(id uint) (*models.Appointment, error)
	CreateAppointment(userID uint, barberID uint, serviceIDs []uint, appointmentTime time.Time, notes *string) (*models.Appointment, error)
	CancelAppointment(id, userID uint) error
}

type appointmentService struct {
	appointmentRepo repository.AppointmentRepository
	barberRepo      repository.BarberRepository
	serviceRepo     repository.ServiceRepository
	messageService  MessageService
}

func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	barberRepo repository.BarberRepository,
	serviceRepo repository.ServiceRepository,
	messageService MessageService,
) AppointmentService {
	return &appointmentService{
		appointmentRepo: appointmentRepo,
		barberRepo:      barberRepo,
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
	_, err := s.barberRepo.GetByID(barberID)
	if err != nil {
		return nil, errors.New("barber not found")
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
