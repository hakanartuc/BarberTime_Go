# PowerShell script for Windows - Veritabanını sıfırlama

$DB_NAME = "barbertime"
$DB_USER = "postgres"
$DB_HOST = "localhost"
$DB_PORT = "5432"

Write-Host "Aktif bağlantıları sonlandırılıyor..." -ForegroundColor Yellow

# Aktif bağlantıları sonlandır
$terminateQuery = @"
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
"@

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c $terminateQuery

Write-Host "Veritabanı siliniyor..." -ForegroundColor Yellow
# Veritabanını sil
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"

Write-Host "Yeni veritabanı oluşturuluyor..." -ForegroundColor Yellow
# Yeni veritabanını oluştur
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"

Write-Host "Veritabanı başarıyla sıfırlandı!" -ForegroundColor Green

