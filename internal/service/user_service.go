package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"errors"
	"mime/multipart"
)

type UserService interface {
	GetProfile(userID uint) (*models.User, error)
	UpdateProfile(userID uint, name *string, avatar *string) (*models.User, error)
	UpdateAvatar(userID uint, file *multipart.FileHeader) (string, error)
	DeleteAvatar(userID uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetProfile(userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) UpdateProfile(userID uint, name *string, avatar *string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if name != nil {
		user.Name = *name
	}
	if avatar != nil {
		user.Avatar = avatar
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateAvatar(userID uint, file *multipart.FileHeader) (string, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Eski avatar'ı sil
	if user.Avatar != nil && *user.Avatar != "" {
		utils.DeleteFile(*user.Avatar)
	}

	// Yeni dosyayı kaydet
	avatarURL, err := utils.SaveUploadedFile(file, "avatars")
	if err != nil {
		return "", err
	}

	user.Avatar = &avatarURL
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return avatarURL, nil
}

func (s *userService) DeleteAvatar(userID uint) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Avatar != nil && *user.Avatar != "" {
		utils.DeleteFile(*user.Avatar)
		user.Avatar = nil
		if err := s.userRepo.Update(user); err != nil {
			return err
		}
	}

	return nil
}

