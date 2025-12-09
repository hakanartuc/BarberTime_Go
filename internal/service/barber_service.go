package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"errors"
	"fmt"
	"time"
)

type BarberService interface {
	GetNearbyBarbers(latitude, longitude float64, page, limit int) ([]models.Barber, utils.Pagination, error)
	GetBarberByID(id uint) (*models.Barber, error)
	GetAvailability(barberID uint, startDate, endDate string) ([]AvailabilitySlot, error)
}

type AvailabilitySlot struct {
	Date           string     `json:"date"`
	Slots          []TimeSlot `json:"slots"`
	OccupancyRate  float64    `json:"occupancyRate"`
	TotalSlots     int        `json:"totalSlots"`
	AvailableSlots int        `json:"availableSlots"`
}

type TimeSlot struct {
	Time      string `json:"time"`
	Available bool   `json:"available"`
}

type barberService struct {
	barberRepo      repository.BarberRepository
	appointmentRepo repository.AppointmentRepository
	cacheService    CacheService
}

func NewBarberService(
	barberRepo repository.BarberRepository,
	appointmentRepo repository.AppointmentRepository,
	cacheService CacheService,
) BarberService {
	return &barberService{
		barberRepo:      barberRepo,
		appointmentRepo: appointmentRepo,
		cacheService:    cacheService,
	}
}

func (s *barberService) GetNearbyBarbers(latitude, longitude float64, page, limit int) ([]models.Barber, utils.Pagination, error) {
	barbers, total, err := s.barberRepo.GetNearby(latitude, longitude, nil, page, limit)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return barbers, pagination, nil
}

func (s *barberService) GetBarberByID(id uint) (*models.Barber, error) {
	// Önce cache'den kontrol et
	cacheKey := fmt.Sprintf("barber:%d", id)
	var barber models.Barber
	if err := s.cacheService.Get(cacheKey, &barber); err == nil {
		return &barber, nil
	}

	// Cache'de yoksa DB'den al
	barberPtr, err := s.barberRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("barber not found")
	}

	// Cache'e kaydet (1 saat TTL)
	s.cacheService.Set(cacheKey, barberPtr, 1*time.Hour)

	return barberPtr, nil
}

func (s *barberService) SetBarberCache(barberID uint, barber *models.Barber) {
	cacheKey := fmt.Sprintf("barber:%d", barberID)
	s.cacheService.Set(cacheKey, barber, 1*time.Hour)
}

func (s *barberService) DeleteBarberCache(barberID uint) {
	cacheKey := fmt.Sprintf("barber:%d", barberID)
	s.cacheService.Delete(cacheKey)
}

func (s *barberService) GetAvailability(barberID uint, startDate, endDate string) ([]AvailabilitySlot, error) {
	// Tarih aralığını parse et
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errors.New("invalid end date format")
	}

	if end.Before(start) {
		return nil, errors.New("end date must be after start date")
	}

	var availabilitySlots []AvailabilitySlot
	currentDate := start

	// Varsayılan çalışma saatleri (09:00 - 18:00, 30 dakika aralıklarla)
	timeSlots := generateTimeSlots("09:00", "18:00", 30)

	for !currentDate.After(end) {
		dateStr := currentDate.Format("2006-01-02")

		// Bu tarihteki randevuları getir
		appointments, _ := s.appointmentRepo.GetByBarberIDAndDate(barberID, dateStr)

		// Meşgul saatleri belirle
		busyTimes := make(map[string]bool)
		for _, apt := range appointments {
			if apt.Status != models.StatusRejected {
				busyTimes[apt.Time] = true
			}
		}

		// Slot'ları oluştur
		var slots []TimeSlot
		availableCount := 0
		for _, ts := range timeSlots {
			available := !busyTimes[ts] && !isPastDateTime(dateStr, ts)
			slots = append(slots, TimeSlot{
				Time:      ts,
				Available: available,
			})
			if available {
				availableCount++
			}
		}

		totalSlots := len(slots)
		occupancyRate := 0.0
		if totalSlots > 0 {
			occupancyRate = float64(totalSlots-availableCount) / float64(totalSlots) * 100
		}

		availabilitySlots = append(availabilitySlots, AvailabilitySlot{
			Date:           dateStr,
			Slots:          slots,
			OccupancyRate:  occupancyRate,
			TotalSlots:     totalSlots,
			AvailableSlots: availableCount,
		})

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return availabilitySlots, nil
}

func generateTimeSlots(start, end string, intervalMinutes int) []string {
	var slots []string
	startTime, _ := time.Parse("15:04", start)
	endTime, _ := time.Parse("15:04", end)

	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		slots = append(slots, current.Format("15:04"))
		current = current.Add(time.Duration(intervalMinutes) * time.Minute)
	}

	return slots
}

func isPastDateTime(date, timeStr string) bool {
	datetime, err := time.Parse("2006-01-02 15:04", date+" "+timeStr)
	if err != nil {
		return false
	}
	return datetime.Before(time.Now())
}
