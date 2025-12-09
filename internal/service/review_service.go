package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"errors"
)

type ReviewService interface {
	GetReviews(reviewType *models.ReviewType, targetID *uint, page, limit int) ([]models.Review, utils.Pagination, error)
	CreateReview(userID uint, reviewType models.ReviewType, targetID uint, rating int, comment string, appointmentID *uint) (*models.Review, error)
	DeleteReview(id, userID uint) error
}

type reviewService struct {
	reviewRepo     repository.ReviewRepository
	messageService MessageService
}

func NewReviewService(reviewRepo repository.ReviewRepository, messageService MessageService) ReviewService {
	return &reviewService{
		reviewRepo:     reviewRepo,
		messageService: messageService,
	}
}

func (s *reviewService) GetReviews(reviewType *models.ReviewType, targetID *uint, page, limit int) ([]models.Review, utils.Pagination, error) {
	var reviews []models.Review
	var total int64
	var err error

	if reviewType != nil && targetID != nil {
		reviews, total, _, err = s.reviewRepo.GetByTypeAndTargetID(*reviewType, *targetID, page, limit)
	} else {
		reviews, total, err = s.reviewRepo.GetAll(page, limit)
	}

	if err != nil {
		return nil, utils.Pagination{}, err
	}

	pagination := utils.NewPagination(page, limit, total)
	return reviews, pagination, nil
}

func (s *reviewService) CreateReview(userID uint, reviewType models.ReviewType, targetID uint, rating int, comment string, appointmentID *uint) (*models.Review, error) {
	// Rating validasyonu
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	review := &models.Review{
		TargetType:    string(reviewType), // GORM polymorphic için string olarak kaydet
		TargetID:      targetID,
		UserID:        userID,
		Rating:        rating,
		Comment:       comment,
		AppointmentID: appointmentID,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	// RabbitMQ'ya mesaj gönder
	s.messageService.PublishReviewCreated(review.ID, review.TargetType, review.TargetID)

	return s.reviewRepo.GetByID(review.ID)
}

func (s *reviewService) DeleteReview(id, userID uint) error {
	review, err := s.reviewRepo.GetByID(id)
	if err != nil {
		return errors.New("review not found")
	}

	if review.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.reviewRepo.Delete(id)
}
