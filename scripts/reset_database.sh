#!/bin/bash

# PostgreSQL veritabanını sıfırlama script'i

DB_NAME="barbertime"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"

echo "Aktif bağlantıları sonlandırılıyor..."

# Aktif bağlantıları sonlandır
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
"

echo "Veritabanı siliniyor..."
# Veritabanını sil
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"

echo "Yeni veritabanı oluşturuluyor..."
# Yeni veritabanını oluştur
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"

echo "Veritabanı başarıyla sıfırlandı!"

