package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type PhotoLikeRepository interface {
	Create(like *models.PhotoLike) error
	GetByPhotoID(photoID uint) ([]models.PhotoLike, error)
	GetByUserIDAndPhotoID(userID uint, photoID uint) (*models.PhotoLike, error)
	Delete(likeID uint) error
	DeleteByUserAndPhoto(userID uint, photoID uint) error
	IsLiked(userID uint, photoID uint) (bool, error)
	CountByPhotoID(photoID uint) (int64, error)
}

type photoLikeRepository struct {
	db *gorm.DB
}

func NewPhotoLikeRepository() PhotoLikeRepository {
	return &photoLikeRepository{
		db: database.DB,
	}
}

func (r *photoLikeRepository) Create(like *models.PhotoLike) error {
	return r.db.Create(like).Error
}

func (r *photoLikeRepository) GetByPhotoID(photoID uint) ([]models.PhotoLike, error) {
	var likes []models.PhotoLike
	err := r.db.Where("photo_id = ?", photoID).
		Preload("User").
		Order("created_at DESC").
		Find(&likes).Error
	return likes, err
}

func (r *photoLikeRepository) GetByUserIDAndPhotoID(userID uint, photoID uint) (*models.PhotoLike, error) {
	var like models.PhotoLike
	err := r.db.Where("user_id = ? AND photo_id = ?", userID, photoID).First(&like).Error
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *photoLikeRepository) Delete(likeID uint) error {
	return r.db.Delete(&models.PhotoLike{}, likeID).Error
}

func (r *photoLikeRepository) DeleteByUserAndPhoto(userID uint, photoID uint) error {
	return r.db.Where("user_id = ? AND photo_id = ?", userID, photoID).
		Delete(&models.PhotoLike{}).Error
}

func (r *photoLikeRepository) IsLiked(userID uint, photoID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.PhotoLike{}).
		Where("user_id = ? AND photo_id = ?", userID, photoID).
		Count(&count).Error
	return count > 0, err
}

func (r *photoLikeRepository) CountByPhotoID(photoID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.PhotoLike{}).
		Where("photo_id = ?", photoID).
		Count(&count).Error
	return count, err
}

