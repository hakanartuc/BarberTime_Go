package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"errors"
)

type PhotoLikeService interface {
	LikePhoto(userID uint, photoID uint) error
	UnlikePhoto(userID uint, photoID uint) error
	IsLiked(userID uint, photoID uint) (bool, error)
	GetPhotoLikes(photoID uint) ([]models.PhotoLike, error)
	GetPhotoLikeCount(photoID uint) (int64, error)
}

type photoLikeService struct {
	photoLikeRepo repository.PhotoLikeRepository
	photoRepo     repository.PhotoRepository
}

func NewPhotoLikeService(photoLikeRepo repository.PhotoLikeRepository, photoRepo repository.PhotoRepository) PhotoLikeService {
	return &photoLikeService{
		photoLikeRepo: photoLikeRepo,
		photoRepo:     photoRepo,
	}
}

func (s *photoLikeService) LikePhoto(userID uint, photoID uint) error {
	// Fotoğrafın varlığını kontrol et
	_, err := s.photoRepo.GetByID(photoID)
	if err != nil {
		return errors.New("photo not found")
	}

	// Zaten beğenilmiş mi kontrol et
	isLiked, err := s.photoLikeRepo.IsLiked(userID, photoID)
	if err != nil {
		return err
	}
	if isLiked {
		return errors.New("photo already liked")
	}

	like := &models.PhotoLike{
		PhotoID: photoID,
		UserID:  userID,
	}

	return s.photoLikeRepo.Create(like)
}

func (s *photoLikeService) UnlikePhoto(userID uint, photoID uint) error {
	// Beğeni var mı kontrol et
	isLiked, err := s.photoLikeRepo.IsLiked(userID, photoID)
	if err != nil {
		return err
	}
	if !isLiked {
		return errors.New("photo not liked")
	}

	return s.photoLikeRepo.DeleteByUserAndPhoto(userID, photoID)
}

func (s *photoLikeService) IsLiked(userID uint, photoID uint) (bool, error) {
	return s.photoLikeRepo.IsLiked(userID, photoID)
}

func (s *photoLikeService) GetPhotoLikes(photoID uint) ([]models.PhotoLike, error) {
	return s.photoLikeRepo.GetByPhotoID(photoID)
}

func (s *photoLikeService) GetPhotoLikeCount(photoID uint) (int64, error) {
	return s.photoLikeRepo.CountByPhotoID(photoID)
}

