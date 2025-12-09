-- Veritabanını silmek için önce aktif bağlantıları sonlandır

-- 1. Aktif bağlantıları sonlandır
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = 'barbertime'
  AND pid <> pg_backend_pid();

-- 2. Veritabanını sil
DROP DATABASE IF EXISTS barbertime;

-- 3. Yeni veritabanını oluştur (opsiyonel)
CREATE DATABASE barbertime;

