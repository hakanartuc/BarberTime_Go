# Postman Collection Kullanım Kılavuzu

Bu proje için hazırlanmış Postman collection dosyası tüm API endpoint'lerini içerir.

## Kurulum

1. Postman uygulamasını açın
2. **Import** butonuna tıklayın
3. `BarberTime_API.postman_collection.json` dosyasını seçin
4. Collection'ı import edin

## Environment Variables

Collection'da iki environment variable tanımlıdır:

- `baseUrl`: API base URL'i (varsayılan: `http://localhost:8080`)
- `accessToken`: JWT token (otomatik olarak auth endpoint'inden alınır)

### Environment Oluşturma

1. Postman'de sağ üst köşeden **Environments** sekmesine gidin
2. **+** butonuna tıklayın
3. Yeni environment oluşturun:
   - Name: `BarberTime Local`
   - Variables:
     - `baseUrl`: `http://localhost:8080`
     - `accessToken`: (boş bırakın, otomatik doldurulacak)

## Kullanım

### 1. Authentication

Önce token almanız gerekir:

1. **Authentication > Get Token** request'ini açın
2. Body'de telefon numarası ve device ID'yi güncelleyin
3. **Send** butonuna tıklayın
4. Token otomatik olarak `accessToken` variable'ına kaydedilir

### 2. Protected Endpoints

Token aldıktan sonra tüm protected endpoint'ler kullanılabilir. Her request'te Authorization header otomatik olarak eklenir.

## Endpoint Kategorileri

### Authentication
- `POST /api/auth/token` - Token alma

### User
- `GET /api/user/profile` - Profil bilgileri
- `PATCH /api/user/profile` - Profil güncelleme
- `PATCH /api/user/avatar` - Avatar yükleme (multipart/form-data)
- `DELETE /api/user/avatar` - Avatar silme

### Appointments
- `GET /api/appointments/past` - Geçmiş randevular
- `GET /api/appointments/upcoming` - Gelecek randevular
- `GET /api/appointments/:id` - Randevu detayı
- `POST /api/appointments` - Yeni randevu oluştur
- `DELETE /api/appointments/:id` - Randevu iptal

### Barbers
- `GET /api/barbers/nearby` - Yakındaki berberler
- `GET /api/barbers/:id` - Berber detayı
- `GET /api/barbers/:id/availability` - Müsaitlik durumu

### Shops
- `GET /api/shops/nearby` - Yakındaki salonlar
- `GET /api/shops/:id` - Salon detayı
- `GET /api/shops/:id/barbers` - Salondaki berberler
- `GET /api/shops/:id/reviews` - Salon yorumları

### Favorites
- `GET /api/favorites` - Tüm favoriler
- `GET /api/favorites/shops` - Favori salonlar
- `GET /api/favorites/barbers` - Favori berberler
- `POST /api/favorites/shops/:id` - Salonu favorilere ekle
- `DELETE /api/favorites/shops/:id` - Salonu favorilerden çıkar
- `POST /api/favorites/barbers/:id` - Berberi favorilere ekle
- `DELETE /api/favorites/barbers/:id` - Berberi favorilerden çıkar

### Reviews
- `GET /api/reviews` - Yorumları listele
- `POST /api/reviews` - Yeni yorum ekle
- `DELETE /api/reviews/:id` - Yorum sil

### Services
- `GET /api/service/barber/:id` - Berber hizmetleri
- `GET /api/service/shop/:id` - Salon hizmetleri

### Search
- `GET /api/search` - Genel arama

## Örnek Request'ler

### Randevu Oluşturma

```json
POST /api/appointments
{
    "barberId": 1,
    "serviceIds": [1, 2],
    "date": "2025-12-15",
    "time": "14:00",
    "notes": "Kısa saç kesimi"
}
```

### Yorum Ekleme

```json
POST /api/reviews
{
    "type": "shop",
    "targetId": 1,
    "rating": 5,
    "comment": "Harika bir salon, çok memnun kaldım!",
    "appointmentId": 1
}
```

### Yakındaki Berberleri Bulma

```
GET /api/barbers/nearby?latitude=41.0082&longitude=28.9784&page=1&limit=10
```

## Notlar

- Tüm protected endpoint'ler `Authorization: Bearer <token>` header'ı gerektirir
- Token otomatik olarak `Get Token` request'inden sonra kaydedilir
- Query parametreleri collection'da örnek değerlerle gelir, ihtiyacınıza göre güncelleyin
- ID'ler (örnek: `/barbers/1`) collection'da örnek değerlerdir, gerçek ID'lerle değiştirin

## Troubleshooting

### Token Alınamıyor
- Telefon numarası formatını kontrol edin
- Device ID'nin benzersiz olduğundan emin olun
- Server'ın çalıştığından emin olun

### 401 Unauthorized
- Token'ın geçerli olduğundan emin olun
- Environment variable'da `accessToken`'ın doğru set edildiğini kontrol edin
- Token'ın süresi dolmuş olabilir, yeni token alın

### 404 Not Found
- Base URL'in doğru olduğundan emin olun
- Endpoint path'ini kontrol edin
- Server'ın çalıştığından emin olun

