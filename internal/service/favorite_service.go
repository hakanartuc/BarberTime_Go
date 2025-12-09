package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
)

type FavoriteService interface {
	GetFavorites(userID uint, favoriteType *models.FavoriteType) ([]models.Favorite, error)
	AddFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) error
	RemoveFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) error
	IsFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) (bool, error)
}

type favoriteService struct {
	favoriteRepo repository.FavoriteRepository
}

func NewFavoriteService(favoriteRepo repository.FavoriteRepository) FavoriteService {
	return &favoriteService{
		favoriteRepo: favoriteRepo,
	}
}

func (s *favoriteService) GetFavorites(userID uint, favoriteType *models.FavoriteType) ([]models.Favorite, error) {
	return s.favoriteRepo.GetByUserID(userID, favoriteType)
}

func (s *favoriteService) AddFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) error {
	// Zaten favori mi kontrol et
	isFav, err := s.favoriteRepo.IsFavorite(userID, favoriteType, targetID)
	if err != nil {
		return err
	}
	if isFav {
		return errors.New("already in favorites")
	}

	favorite := &models.Favorite{
		UserID:   userID,
		Type:     favoriteType,
		TargetID: targetID,
	}

	return s.favoriteRepo.Create(favorite)
}

func (s *favoriteService) RemoveFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) error {
	return s.favoriteRepo.DeleteByUserAndTarget(userID, favoriteType, targetID)
}

func (s *favoriteService) IsFavorite(userID uint, favoriteType models.FavoriteType, targetID uint) (bool, error) {
	return s.favoriteRepo.IsFavorite(userID, favoriteType, targetID)
}

