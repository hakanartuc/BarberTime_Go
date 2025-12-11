package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"errors"
	"mime/multipart"
)

type PhotoService interface {
	UploadPhoto(targetType models.PhotoType, targetID uint, userID uint, file *multipart.FileHeader, caption string) (*models.Photo, error)
	GetPhotos(targetType models.PhotoType, targetID uint) ([]models.Photo, error)
	GetPhoto(photoID uint) (*models.Photo, error)
	DeletePhoto(photoID uint, userID uint) error
}

type photoService struct {
	photoRepo        repository.PhotoRepository
	barberRepo       repository.BarberRepository
	shopRepo         repository.ShopRepository
	barberShopRepo   repository.BarberShopRepository
	photoLikeRepo    repository.PhotoLikeRepository
}

func NewPhotoService(photoRepo repository.PhotoRepository, barberRepo repository.BarberRepository, shopRepo repository.ShopRepository, barberShopRepo repository.BarberShopRepository, photoLikeRepo repository.PhotoLikeRepository) PhotoService {
	return &photoService{
		photoRepo:      photoRepo,
		barberRepo:     barberRepo,
		shopRepo:       shopRepo,
		barberShopRepo: barberShopRepo,
		photoLikeRepo:  photoLikeRepo,
	}
}

func (s *photoService) UploadPhoto(targetType models.PhotoType, targetID uint, userID uint, file *multipart.FileHeader, caption string) (*models.Photo, error) {
	// Target'ın varlığını ve kullanıcının yetkisini kontrol et
	var hasPermission bool
	var err error

	if targetType == models.PhotoTypeBarber {
		barber, err := s.barberRepo.GetByID(targetID)
		if err != nil {
			return nil, errors.New("barber not found")
		}
		// Sadece barber'ın kendisi kendi profili için fotoğraf ekleyebilir
		if barber.UserID != userID {
			return nil, errors.New("you can only upload photos for your own barber profile")
		}
		hasPermission = true
	} else if targetType == models.PhotoTypeShop {
		shop, err := s.shopRepo.GetByID(targetID)
		if err != nil {
			return nil, errors.New("shop not found")
		}
		// Sadece shop owner'ları fotoğraf ekleyebilir
		hasPermission, err = s.barberShopRepo.IsOwner(shop.ID, userID)
		if err != nil {
			return nil, err
		}
		if !hasPermission {
			return nil, errors.New("only shop owners can upload photos for this shop")
		}
	} else {
		return nil, errors.New("invalid target type")
	}

	// Dosyayı kaydet
	subDir := "photos"
	if targetType == models.PhotoTypeBarber {
		subDir = "photos/barbers"
	} else {
		subDir = "photos/shops"
	}

	imageURL, err := utils.SaveUploadedFile(file, subDir)
	if err != nil {
		return nil, err
	}

	// Photo kaydı oluştur
	photo := &models.Photo{
		TargetID:   targetID,
		TargetType: targetType,
		ImageURL:   imageURL,
		Caption:    caption,
	}

	if err := s.photoRepo.Create(photo); err != nil {
		// Hata durumunda dosyayı sil
		utils.DeleteFile(imageURL)
		return nil, err
	}

	return photo, nil
}

func (s *photoService) GetPhotos(targetType models.PhotoType, targetID uint) ([]models.Photo, error) {
	photos, err := s.photoRepo.GetByTarget(targetType, targetID)
	if err != nil {
		return nil, err
	}

	// Her fotoğraf için beğeni sayısını preload ile yükle
	for i := range photos {
		likeCount, _ := s.photoLikeRepo.CountByPhotoID(photos[i].ID)
		// Like count bilgisi response'da kullanılabilir (şimdilik preload ile alınıyor)
		_ = likeCount
	}

	return photos, nil
}

func (s *photoService) GetPhoto(photoID uint) (*models.Photo, error) {
	return s.photoRepo.GetByID(photoID)
}

func (s *photoService) DeletePhoto(photoID uint, userID uint) error {
	photo, err := s.photoRepo.GetByID(photoID)
	if err != nil {
		return errors.New("photo not found")
	}

	// Yetki kontrolü
	var hasPermission bool

	if photo.TargetType == models.PhotoTypeBarber {
		barber, err := s.barberRepo.GetByID(photo.TargetID)
		if err != nil {
			return errors.New("barber not found")
		}
		// Sadece barber'ın kendisi kendi profilindeki fotoğrafları silebilir
		if barber.UserID != userID {
			return errors.New("you can only delete photos from your own barber profile")
		}
		hasPermission = true
	} else if photo.TargetType == models.PhotoTypeShop {
		// Sadece shop owner'ları fotoğraf silebilir
		hasPermission, err = s.barberShopRepo.IsOwner(photo.TargetID, userID)
		if err != nil {
			return err
		}
		if !hasPermission {
			return errors.New("only shop owners can delete photos from this shop")
		}
	} else {
		return errors.New("invalid photo target type")
	}

	// Dosyayı sil
	if err := utils.DeleteFile(photo.ImageURL); err != nil {
		// Log hatası ama devam et
	}

	// Veritabanından sil
	return s.photoRepo.Delete(photoID)
}

