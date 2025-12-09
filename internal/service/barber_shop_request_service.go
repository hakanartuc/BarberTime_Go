package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
	"fmt"
	"time"
)

type BarberShopRequestService interface {
	CreateRequest(barberID, shopID uint, message string) (*models.BarberShopRequest, error)
	GetRequestsByShop(shopID, userID uint) ([]models.BarberShopRequest, error)
	GetRequestsByBarber(barberID uint) ([]models.BarberShopRequest, error)
	ApproveRequest(requestID, shopID, userID uint) error
	RejectRequest(requestID, shopID, userID uint, reason string) error
}

type barberShopRequestService struct {
	requestRepo    repository.BarberShopRequestRepository
	barberRepo     repository.BarberRepository
	shopRepo       repository.ShopRepository
	barberShopRepo repository.BarberShopRepository
	cacheService   CacheService
	messageService MessageService
}

func NewBarberShopRequestService(
	requestRepo repository.BarberShopRequestRepository,
	barberRepo repository.BarberRepository,
	shopRepo repository.ShopRepository,
	barberShopRepo repository.BarberShopRepository,
	cacheService CacheService,
	messageService MessageService,
) BarberShopRequestService {
	return &barberShopRequestService{
		requestRepo:    requestRepo,
		barberRepo:     barberRepo,
		shopRepo:       shopRepo,
		barberShopRepo: barberShopRepo,
		cacheService:   cacheService,
		messageService: messageService,
	}
}

func (s *barberShopRequestService) CreateRequest(barberID, shopID uint, message string) (*models.BarberShopRequest, error) {
	// Barber kontrolü
	_, err := s.barberRepo.GetByID(barberID)
	if err != nil {
		return nil, errors.New("barber not found")
	}

	// Shop kontrolü
	_, err = s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	// Zaten bu shop'a başvurmuş mu kontrol et
	existingRequests, _ := s.requestRepo.GetByBarberID(barberID)
	for _, req := range existingRequests {
		if req.ShopID == shopID && req.Status == models.RequestStatusPending {
			return nil, errors.New("already has a pending request for this shop")
		}
	}

	// Zaten bu shop'ta mı kontrol et (approved status ile)
	existingBarberShop, _ := s.barberShopRepo.GetByBarberAndShop(barberID, shopID)
	if existingBarberShop != nil && existingBarberShop.Status == models.BarberShopStatusApproved {
		return nil, errors.New("barber is already in this shop")
	}

	request := &models.BarberShopRequest{
		BarberID: barberID,
		ShopID:   shopID,
		Status:   models.RequestStatusPending,
		Message:  message,
	}

	if err := s.requestRepo.Create(request); err != nil {
		return nil, err
	}

	return s.requestRepo.GetByID(request.ID)
}

func (s *barberShopRequestService) GetRequestsByShop(shopID, userID uint) ([]models.BarberShopRequest, error) {
	// Owner kontrolü - barberShopRepo üzerinden
	// Önce user'ın barber olup olmadığını kontrol et
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("user must be registered as barber")
	}

	// BarberShop üzerinden owner kontrolü
	barberShop, err := s.barberShopRepo.GetByBarberAndShop(barber.ID, shopID)
	if err != nil || !barberShop.IsOwner {
		return nil, errors.New("unauthorized: only shop owners can view requests")
	}

	return s.requestRepo.GetPendingByShopID(shopID)
}

func (s *barberShopRequestService) GetRequestsByBarber(barberID uint) ([]models.BarberShopRequest, error) {
	return s.requestRepo.GetByBarberID(barberID)
}

func (s *barberShopRequestService) ApproveRequest(requestID, shopID, userID uint) error {
	// Owner kontrolü - barberShopRepo üzerinden
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return errors.New("user must be registered as barber")
	}

	ownerBarberShop, err := s.barberShopRepo.GetByBarberAndShop(barber.ID, shopID)
	if err != nil || !ownerBarberShop.IsOwner {
		return errors.New("unauthorized: only shop owners can approve requests")
	}

	request, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		return errors.New("request not found")
	}

	if request.ShopID != shopID {
		return errors.New("request does not belong to this shop")
	}

	if request.Status != models.RequestStatusPending {
		return errors.New("request is not pending")
	}

	// BarberShop kaydı oluştur veya güncelle
	barberShopRecord, err := s.barberShopRepo.GetByBarberAndShop(request.BarberID, shopID)
	if err != nil {
		// Yeni kayıt oluştur
		barberShopRecord = &models.BarberShop{
			BarberID:   request.BarberID,
			ShopID:     shopID,
			Status:     models.BarberShopStatusApproved,
			IsOwner:    false, // Başvuru onaylandığında owner olmaz, sadece barber olur
			ApprovedBy: &userID,
		}
		if err := s.barberShopRepo.Create(barberShopRecord); err != nil {
			return err
		}
	} else {
		// Mevcut kaydı güncelle
		barberShopRecord.Status = models.BarberShopStatusApproved
		barberShopRecord.ApprovedBy = &userID
		if err := s.barberShopRepo.Update(barberShopRecord); err != nil {
			return err
		}
	}

	// Request'i onayla
	request.Status = models.RequestStatusApproved
	request.ApprovedBy = &userID
	if err := s.requestRepo.Update(request); err != nil {
		return err
	}

	// Barber shop'a eklendiğinde cache'i güncelle ve Elasticsearch indexing mesajı gönder
	barberForIndex, err := s.barberRepo.GetByID(request.BarberID)
	if err == nil {
		// Cache'i güncelle
		cacheKey := fmt.Sprintf("barber:%d", barberForIndex.ID)
		s.cacheService.Set(cacheKey, barberForIndex, 1*time.Hour)
		// RabbitMQ ile event mesajı gönder
		s.messageService.PublishBarberUpdated(barberForIndex.ID)
	}

	return nil
}

func (s *barberShopRequestService) RejectRequest(requestID, shopID, userID uint, reason string) error {
	// Owner kontrolü - barberShopRepo üzerinden
	barber, err := s.barberRepo.GetByUserID(userID)
	if err != nil {
		return errors.New("user must be registered as barber")
	}

	barberShop, err := s.barberShopRepo.GetByBarberAndShop(barber.ID, shopID)
	if err != nil || !barberShop.IsOwner {
		return errors.New("unauthorized: only shop owners can reject requests")
	}

	request, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		return errors.New("request not found")
	}

	if request.ShopID != shopID {
		return errors.New("request does not belong to this shop")
	}

	if request.Status != models.RequestStatusPending {
		return errors.New("request is not pending")
	}

	request.Status = models.RequestStatusRejected
	request.RejectedReason = reason
	return s.requestRepo.Update(request)
}
