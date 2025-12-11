package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type PhotoRepository interface {
	Create(photo *models.Photo) error
	GetByID(id uint) (*models.Photo, error)
	GetByTarget(targetType models.PhotoType, targetID uint) ([]models.Photo, error)
	Delete(id uint) error
	CountByTarget(targetType models.PhotoType, targetID uint) (int64, error)
}

type photoRepository struct {
	db *gorm.DB
}

func NewPhotoRepository() PhotoRepository {
	return &photoRepository{
		db: database.DB,
	}
}

func (r *photoRepository) Create(photo *models.Photo) error {
	return r.db.Create(photo).Error
}

func (r *photoRepository) GetByID(id uint) (*models.Photo, error) {
	var photo models.Photo
	err := r.db.Preload("Likes").First(&photo, id).Error
	if err != nil {
		return nil, err
	}
	return &photo, nil
}

func (r *photoRepository) GetByTarget(targetType models.PhotoType, targetID uint) ([]models.Photo, error) {
	var photos []models.Photo
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Preload("Likes").
		Order("created_at DESC").
		Find(&photos).Error
	return photos, err
}

func (r *photoRepository) Delete(id uint) error {
	return r.db.Delete(&models.Photo{}, id).Error
}

func (r *photoRepository) CountByTarget(targetType models.PhotoType, targetID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Photo{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&count).Error
	return count, err
}

