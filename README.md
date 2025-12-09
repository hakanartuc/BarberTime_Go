# BarberTime Go Backend API

Kuaför Saati projesi için Go ile geliştirilmiş RESTful backend API.

## Özellikler

- JWT tabanlı kimlik doğrulama
- Kullanıcı profil yönetimi
- Randevu yönetimi
- Berber ve salon yönetimi
- Favori işlemleri
- Yorum ve değerlendirme sistemi
- Konum bazlı arama
- Müsaitlik kontrolü

## Teknolojiler

- Go 1.21+
- Gin Web Framework
- GORM (PostgreSQL)
- JWT Authentication
- PostgreSQL

## Kurulum

1. PostgreSQL veritabanını başlatın
2. `.env.example` dosyasını `.env` olarak kopyalayın ve gerekli ayarları yapın
3. Bağımlılıkları yükleyin:
```bash
go mod download
```
4. Uygulamayı çalıştırın:
```bash
go run cmd/server/main.go
```

## API Dokümantasyonu

API endpoint'leri ve kullanım detayları için proje dokümantasyonuna bakın.

