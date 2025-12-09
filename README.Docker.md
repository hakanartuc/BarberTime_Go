# Docker Compose Kullanım Kılavuzu

Bu proje Docker Compose ile çalıştırılabilir. Üç farklı compose dosyası bulunmaktadır:

## Dosyalar

- `docker-compose.yml` - Sadece servisler (PostgreSQL, Redis, Elasticsearch, RabbitMQ)
- `docker-compose.dev.yml` - Development ortamı (servisler + uygulama)
- `docker-compose.prod.yml` - Production ortamı (servisler + uygulama)

## Hızlı Başlangıç

### Sadece Servisleri Başlatma

```bash
docker-compose up -d
```

Bu komut şunları başlatır:
- PostgreSQL (port 5432)
- Redis (port 6379)
- Elasticsearch (port 9200)
- RabbitMQ (port 5672, Management UI: 15672)

### Development Ortamı

```bash
docker-compose -f docker-compose.dev.yml up -d
```

Bu komut tüm servisleri + Go uygulamasını başlatır.

### Production Ortamı

```bash
docker-compose -f docker-compose.prod.yml up -d
```

**Not:** Production için `.env` dosyasını oluşturup gerekli değişkenleri ayarlayın.

## Servis Erişim Bilgileri

### PostgreSQL
- Host: `localhost` (dışarıdan) veya `postgres` (container içinden)
- Port: `5432`
- User: `postgres`
- Password: `postgres`
- Database: `barbertime`

### Redis
- Host: `localhost` (dışarıdan) veya `redis` (container içinden)
- Port: `6379`
- Password: (development'ta yok, production'da `.env`'den)

### Elasticsearch
- URL: `http://localhost:9200` (dışarıdan) veya `http://elasticsearch:9200` (container içinden)
- Security: Devre dışı (development)

### RabbitMQ
- AMQP: `amqp://guest:guest@localhost:5672/` (dışarıdan)
- Management UI: `http://localhost:15672` (guest/guest)
- Container içinden: `amqp://guest:guest@rabbitmq:5672/`

## Yönetim Komutları

### Servisleri Durdurma
```bash
docker-compose down
```

### Servisleri Durdurma (Volumes'ları da sil)
```bash
docker-compose down -v
```

### Logları Görüntüleme
```bash
docker-compose logs -f
```

### Belirli Bir Servisin Loglarını Görüntüleme
```bash
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f elasticsearch
docker-compose logs -f rabbitmq
```

### Servisleri Yeniden Başlatma
```bash
docker-compose restart
```

### Belirli Bir Servisi Yeniden Başlatma
```bash
docker-compose restart postgres
```

## Veritabanı İşlemleri

### PostgreSQL'e Bağlanma
```bash
docker-compose exec postgres psql -U postgres -d barbertime
```

### Redis CLI'ye Bağlanma
```bash
docker-compose exec redis redis-cli
```

### Elasticsearch Health Check
```bash
curl http://localhost:9200/_cluster/health
```

## Volume Yönetimi

### Volume'ları Listeleme
```bash
docker volume ls | grep barbertime
```

### Volume'u Silme
```bash
docker volume rm barbertime_postgres_data
```

## Troubleshooting

### Servisler Başlamıyor
```bash
# Tüm servislerin durumunu kontrol et
docker-compose ps

# Health check'leri kontrol et
docker-compose ps --format "table {{.Name}}\t{{.Status}}"
```

### Port Çakışması
Eğer portlar zaten kullanılıyorsa, `docker-compose.yml` dosyasındaki port numaralarını değiştirin.

### Elasticsearch Memory Hatası
Development ortamında Elasticsearch için memory limit'i düşürülmüştür (256MB). Production'da 1GB kullanılmaktadır.

### Veritabanı Bağlantı Hatası
Uygulama container'ından PostgreSQL'e bağlanırken host adını `postgres` olarak kullanın (localhost değil).

## Production Deployment

1. `.env` dosyasını oluşturun ve güvenli değerler girin
2. `docker-compose.prod.yml` dosyasını kullanın
3. SSL/TLS için reverse proxy (nginx) ekleyin
4. Volume'ları persistent storage'a bağlayın
5. Backup stratejisi oluşturun

## Notlar

- Development ortamında tüm servisler varsayılan şifrelerle çalışır
- Production'da mutlaka güçlü şifreler kullanın
- Elasticsearch data volume'u büyük olabilir, disk alanını kontrol edin
- RabbitMQ management UI'ya erişim için port 15672'yi açın

