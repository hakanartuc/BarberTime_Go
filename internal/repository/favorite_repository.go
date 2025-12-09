package repository

import (
	"barbertime_go/internal/models"
	"barbertime_go/pkg/database"

	"gorm.io/gorm"
)

type FavoriteRepository interface {
	Create(favorite *models.Favorite) error
	GetByUserID(userID uint, favoriteType *models.FavoriteType) ([]models.Favorite, error)
	GetByUserIDAndTarget(userID uint, favoriteType models.FavoriteType, targetID uint) (*models.Favorite, error)
	Delete(favoriteID uint) error
	DeleteByUserAndTarget(userID uint, favoriteType models.FavoriteType, targetID uint) error
	IsFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) (bool, error)
}

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository() FavoriteRepository {
	return &favoriteRepository{
		db: database.DB,
	}
}

func (r *favoriteRepository) Create(favorite *models.Favorite) error {
	return r.db.Create(favorite).Error
}

func (r *favoriteRepository) GetByUserID(userID uint, favoriteType *models.FavoriteType) ([]models.Favorite, error) {
	var favorites []models.Favorite
	query := r.db.Where("user_id = ?", userID)
	
	if favoriteType != nil {
		query = query.Where("type = ?", *favoriteType)
	}

	err := query.Find(&favorites).Error
	return favorites, err
}

func (r *favoriteRepository) GetByUserIDAndTarget(userID uint, favoriteType models.FavoriteType, targetID uint) (*models.Favorite, error) {
	var favorite models.Favorite
	err := r.db.Where("user_id = ? AND type = ? AND target_id = ?", userID, favoriteType, targetID).First(&favorite).Error
	if err != nil {
		return nil, err
	}
	return &favorite, nil
}

func (r *favoriteRepository) Delete(favoriteID uint) error {
	return r.db.Delete(&models.Favorite{}, favoriteID).Error
}

func (r *favoriteRepository) DeleteByUserAndTarget(userID uint, favoriteType models.FavoriteType, targetID uint) error {
	return r.db.Where("user_id = ? AND type = ? AND target_id = ?", userID, favoriteType, targetID).
		Delete(&models.Favorite{}).Error
}

func (r *favoriteRepository) IsFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Favorite{}).
		Where("user_id = ? AND type = ? AND target_id = ?", userID, favoriteType, targetID).
		Count(&count).Error
	return count > 0, err
}

