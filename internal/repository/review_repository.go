package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *models.Review) error
	GetByID(id uint) (*models.Review, error)
	GetByTypeAndTargetID(reviewType models.ReviewType, targetID uint, page, limit int) ([]models.Review, int64, float64, error)
	GetAll(page, limit int) ([]models.Review, int64, error)
	Update(review *models.Review) error
	Delete(id uint) error
	GetAverageRating(reviewType models.ReviewType, targetID uint) (float64, int, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository() ReviewRepository {
	return &reviewRepository{
		db: database.DB,
	}
}

func (r *reviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) GetByID(id uint) (*models.Review, error) {
	var review models.Review
	err := r.db.Preload("User").Preload("Appointment").First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) GetByTypeAndTargetID(reviewType models.ReviewType, targetID uint, page, limit int) ([]models.Review, int64, float64, error) {
	var reviews []models.Review
	var total int64
	var avgRating float64

	offset := (page - 1) * limit

	query := r.db.Model(&models.Review{}).
		Where("target_type = ? AND target_id = ?", string(reviewType), targetID).
		Preload("User").Preload("Appointment")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// Ortalama rating hesapla
	r.db.Model(&models.Review{}).
		Where("target_type = ? AND target_id = ?", string(reviewType), targetID).
		Select("COALESCE(AVG(rating), 0)").Scan(&avgRating)

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&reviews).Error
	return reviews, total, avgRating, err
}

func (r *reviewRepository) GetAll(page, limit int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	offset := (page - 1) * limit

	query := r.db.Model(&models.Review{}).
		Preload("User").Preload("Appointment")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&reviews).Error
	return reviews, total, err
}

func (r *reviewRepository) Update(review *models.Review) error {
	return r.db.Save(review).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&models.Review{}, id).Error
}

func (r *reviewRepository) GetAverageRating(reviewType models.ReviewType, targetID uint) (float64, int, error) {
	type Result struct {
		AvgRating float64
		Count     int64
	}
	var result Result

	err := r.db.Model(&models.Review{}).
		Where("target_type = ? AND target_id = ?", string(reviewType), targetID).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Scan(&result).Error

	return result.AvgRating, int(result.Count), err
}
