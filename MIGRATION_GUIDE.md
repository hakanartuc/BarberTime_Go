# Database Migration Rehberi

## Otomatik Migration (Geliştirme Ortamı)

Uygulama başlatıldığında GORM AutoMigrate otomatik olarak çalışır. Sadece uygulamayı çalıştırın:

```bash
go run cmd/server/main.go
```

veya Docker ile:

```bash
docker-compose -f docker-compose.dev.yml up
```

## Önemli Değişiklikler

### 1. Yeni Tablo: `barber_shops`
- Barber ve Shop arasında many-to-many ilişki
- Status: `pending`, `approved`, `rejected`

### 2. Barber Tablosu Değişiklikleri
- `shop_id` kolonu kaldırıldı
- `shop_approval_status` kolonu kaldırıldı
- `shops` (many-to-many) ilişkisi eklendi

### 3. Shop Tablosu Değişiklikleri
- `barbers` artık many-to-many ilişki

## Production Ortamında Migration

Eğer production veritabanında mevcut veri varsa, aşağıdaki adımları izleyin:

### Adım 1: Veritabanı Yedekleme

```bash
pg_dump -h localhost -U postgres -d barbertime > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Adım 2: Mevcut Barber-Shop İlişkilerini Migrate Et

Eğer `barbers` tablosunda `shop_id` ve `shop_approval_status` kolonları varsa, bu verileri `barber_shops` tablosuna taşımanız gerekir:

```sql
-- Mevcut barber-shop ilişkilerini barber_shops tablosuna taşı
INSERT INTO barber_shops (barber_id, shop_id, status, created_at, updated_at)
SELECT 
    id as barber_id,
    shop_id,
    CASE 
        WHEN shop_approval_status = 'approved' THEN 'approved'
        WHEN shop_approval_status = 'rejected' THEN 'rejected'
        ELSE 'pending'
    END as status,
    created_at,
    updated_at
FROM barbers
WHERE shop_id IS NOT NULL
ON CONFLICT (barber_id, shop_id) DO NOTHING;
```

### Adım 3: Eski Kolonları Kaldır (Dikkatli!)

```sql
-- Önce verilerin taşındığını doğrulayın
SELECT COUNT(*) FROM barber_shops;
SELECT COUNT(*) FROM barbers WHERE shop_id IS NOT NULL;

-- Eğer sayılar eşleşiyorsa, eski kolonları kaldırabilirsiniz
ALTER TABLE barbers DROP COLUMN IF EXISTS shop_id;
ALTER TABLE barbers DROP COLUMN IF EXISTS shop_approval_status;
```

### Adım 4: Uygulamayı Başlat

Uygulama başlatıldığında GORM otomatik olarak:
- Yeni `barber_shops` tablosunu oluşturacak
- Eksik kolonları ekleyecek
- Index'leri oluşturacak

## Manuel Migration Script

Eğer manuel migration yapmak isterseniz, aşağıdaki SQL script'i kullanabilirsiniz:

```sql
-- 1. Barber_shops tablosunu oluştur (GORM zaten yapacak ama manuel de yapabilirsiniz)
CREATE TABLE IF NOT EXISTS barber_shops (
    id SERIAL PRIMARY KEY,
    barber_id INTEGER NOT NULL,
    shop_id INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    approved_by INTEGER,
    rejected_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    FOREIGN KEY (barber_id) REFERENCES barbers(id),
    FOREIGN KEY (shop_id) REFERENCES shops(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
);

-- 2. Unique index oluştur
CREATE UNIQUE INDEX IF NOT EXISTS idx_barber_shops_barber_shop 
ON barber_shops(barber_id, shop_id) 
WHERE deleted_at IS NULL;

-- 3. Mevcut verileri taşı (eğer varsa)
INSERT INTO barber_shops (barber_id, shop_id, status, created_at, updated_at)
SELECT 
    id as barber_id,
    shop_id,
    CASE 
        WHEN shop_approval_status = 'approved' THEN 'approved'
        WHEN shop_approval_status = 'rejected' THEN 'rejected'
        ELSE 'pending'
    END as status,
    created_at,
    updated_at
FROM barbers
WHERE shop_id IS NOT NULL
ON CONFLICT (barber_id, shop_id) DO NOTHING;

-- 4. Eski kolonları kaldır (DİKKAT: Önce verilerin taşındığını doğrulayın!)
-- ALTER TABLE barbers DROP COLUMN IF EXISTS shop_id;
-- ALTER TABLE barbers DROP COLUMN IF EXISTS shop_approval_status;
```

## Kontrol Komutları

Migration sonrası kontrol için:

```sql
-- Barber_shops tablosunu kontrol et
SELECT COUNT(*) FROM barber_shops;

-- Status dağılımını kontrol et
SELECT status, COUNT(*) FROM barber_shops GROUP BY status;

-- Bir barber'ın shop'larını kontrol et
SELECT b.name, s.name, bs.status 
FROM barber_shops bs
JOIN barbers b ON bs.barber_id = b.id
JOIN shops s ON bs.shop_id = s.id
WHERE b.id = 1;
```

## Sorun Giderme

### Hata: "column shop_id does not exist"
- Bu normal, çünkü kolon kaldırıldı
- Eğer hala görüyorsanız, veritabanınız güncel değil demektir

### Hata: "duplicate key value violates unique constraint"
- Barber_shops tablosunda aynı barber-shop kombinasyonu zaten var
- `ON CONFLICT DO NOTHING` kullanarak çözülebilir

### Veri Kaybı Endişesi
- Önce mutlaka veritabanını yedekleyin
- Test ortamında önce deneyin
- Production'da maintenance window kullanın

