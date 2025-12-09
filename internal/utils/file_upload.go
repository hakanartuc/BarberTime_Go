package utils

import (
	"barbertime_go/internal/config"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveUploadedFile(file *multipart.FileHeader, subDir string) (string, error) {
	// Dosya boyutunu kontrol et
	if file.Size > config.AppConfig.MaxUploadSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.AppConfig.MaxUploadSize)
	}

	// Dosya uzantısını kontrol et
	ext := filepath.Ext(file.Filename)
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	ext = strings.ToLower(ext)
	
	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return "", fmt.Errorf("file type not allowed. Allowed types: %v", allowedExts)
	}

	// Upload dizinini oluştur
	uploadPath := filepath.Join(config.AppConfig.UploadDir, subDir)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Benzersiz dosya adı oluştur
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random filename: %w", err)
	}
	
	filename := fmt.Sprintf("%s_%d%s", hex.EncodeToString(randomBytes), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)

	// Dosyayı aç
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Hedef dosyayı oluştur
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Dosyayı kopyala
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// URL döndür (relative path)
	url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)
	return url, nil
}

func DeleteFile(filePath string) error {
	// URL'den dosya yoluna çevir
	if strings.HasPrefix(filePath, "/uploads/") {
		filePath = filepath.Join(".", filePath)
	}
	
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

func GetFileURL(filePath string) string {
	if filePath == "" {
		return ""
	}
	
	// Eğer zaten URL formatındaysa olduğu gibi döndür
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") || strings.HasPrefix(filePath, "/") {
		return filePath
	}
	
	// Relative path ise / ile başlat
	return "/" + filePath
}

