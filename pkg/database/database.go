package database

import (
	"fmt"
	"log"

	"barbertime_go/internal/config"
	"barbertime_go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.AppConfig.DBHost,
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBName,
		config.AppConfig.DBPort,
		config.AppConfig.DBSSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Shop{},
		&models.Barber{},
		&models.Service{},
		&models.ShopService{},
		&models.Appointment{},
		&models.Review{},
		&models.Favorite{},
		&models.BarberShopRequest{},
		&models.BarberShop{},
		&models.PhoneVerificationCode{},
		&models.Photo{},
		&models.PhotoLike{},
		&models.ShopClosedDay{},
		&models.ShopWorkingHours{},
	)

	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// BarberShop için composite unique index (bir barber aynı shop'a birden fazla kez kayıt olamaz)
	DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_barber_shops_barber_shop 
		ON barber_shops(barber_id, shop_id) 
		WHERE deleted_at IS NULL;
	`)

	// Composite unique index for favorites
	DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_favorites_user_type_target 
		ON favorites(user_id, type, target_id) 
		WHERE deleted_at IS NULL;
	`)

	// ShopService için composite unique index (bir shop aynı service'e birden fazla kez sahip olamaz)
	DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_services_shop_service 
		ON shop_services(shop_id, service_id) 
		WHERE deleted_at IS NULL;
	`)

	// PhotoLike için composite unique index (bir kullanıcı aynı fotoğrafı birden fazla kez beğenemez)
	DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_photo_likes_user_photo 
		ON photo_likes(user_id, photo_id) 
		WHERE deleted_at IS NULL;
	`)

	log.Println("Database migration completed successfully")
	return nil
}
