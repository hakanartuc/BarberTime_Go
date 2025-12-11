package main

import (
	"log"
	"os"

	"barbertime_go/internal/config"
	"barbertime_go/internal/handler"
	"barbertime_go/internal/middleware"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/service"
	"barbertime_go/pkg/database"
	"barbertime_go/pkg/elasticsearch"
	"barbertime_go/pkg/rabbitmq"
	"barbertime_go/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// Config yükle
	config.LoadConfig()

	// Gin mode ayarla
	gin.SetMode(config.AppConfig.GinMode)

	// Veritabanı bağlantısı
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Elasticsearch bağlantısı (opsiyonel)
	if err := elasticsearch.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to Elasticsearch: %v (continuing without search)", err)
	}

	// Redis bağlantısı (opsiyonel)
	if err := redis.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v (continuing without cache)", err)
	}

	// RabbitMQ bağlantısı (opsiyonel)
	if err := rabbitmq.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v (continuing without messaging)", err)
	}
	defer rabbitmq.Close()

	// Upload dizinini oluştur
	if err := os.MkdirAll(config.AppConfig.UploadDir+"/avatars", 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}
	if err := os.MkdirAll(config.AppConfig.UploadDir+"/photos/barbers", 0755); err != nil {
		log.Fatal("Failed to create photos/barbers directory:", err)
	}
	if err := os.MkdirAll(config.AppConfig.UploadDir+"/photos/shops", 0755); err != nil {
		log.Fatal("Failed to create photos/shops directory:", err)
	}

	// Router oluştur
	r := gin.Default()

	// Middleware'leri ekle
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ErrorMiddleware())

	// Static dosyalar için route (upload edilen dosyalar için)
	r.Static("/uploads", config.AppConfig.UploadDir)

	// Repository'leri oluştur
	userRepo := repository.NewUserRepository()
	barberRepo := repository.NewBarberRepository()
	shopRepo := repository.NewShopRepository()
	appointmentRepo := repository.NewAppointmentRepository()
	serviceRepo := repository.NewServiceRepository()
	shopServiceRepo := repository.NewShopServiceRepository()
	reviewRepo := repository.NewReviewRepository()
	favoriteRepo := repository.NewFavoriteRepository()
	barberShopRequestRepo := repository.NewBarberShopRequestRepository()
	barberShopRepo := repository.NewBarberShopRepository()
	phoneVerificationCodeRepo := repository.NewPhoneVerificationCodeRepository()
	photoRepo := repository.NewPhotoRepository()
	photoLikeRepo := repository.NewPhotoLikeRepository()

	// Message ve Cache servisleri
	messageService := service.NewMessageService()
	cacheService := service.NewCacheService()

	// Phone Verification Service
	phoneVerificationService := service.NewPhoneVerificationService(phoneVerificationCodeRepo, userRepo)

	// Service'leri oluştur
	authService := service.NewAuthService(userRepo, phoneVerificationService)
	userService := service.NewUserService(userRepo)
	barberService := service.NewBarberService(barberRepo, appointmentRepo, cacheService)
	shopService := service.NewShopService(shopRepo, barberRepo, reviewRepo, cacheService)
	appointmentService := service.NewAppointmentService(appointmentRepo, barberRepo, serviceRepo, messageService)
	reviewService := service.NewReviewService(reviewRepo, messageService)
	favoriteService := service.NewFavoriteService(favoriteRepo)
	serviceService := service.NewServiceService(serviceRepo)
	shopServiceService := service.NewShopServiceService(shopServiceRepo, serviceRepo)
	searchService := service.NewSearchService(barberRepo, shopRepo, shopServiceRepo)
	barberProfileService := service.NewBarberProfileService(barberRepo, userRepo, cacheService, messageService)
	shopProfileService := service.NewShopProfileService(shopRepo, userRepo, barberRepo, barberShopRepo, shopServiceRepo, serviceRepo, cacheService, messageService)
	barberShopRequestService := service.NewBarberShopRequestService(barberShopRequestRepo, barberRepo, shopRepo, barberShopRepo, cacheService, messageService)
	adminService := service.NewAdminService(shopRepo, serviceRepo, cacheService, messageService)
	photoService := service.NewPhotoService(photoRepo, barberRepo, shopRepo, barberShopRepo, photoLikeRepo)
	photoLikeService := service.NewPhotoLikeService(photoLikeRepo, photoRepo)

	// Cache service'i kullanmak için (örnek: barber ve shop servislerine cache eklenebilir)
	_ = cacheService

	// Handler'ları oluştur
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	barberHandler := handler.NewBarberHandler(barberService)
	shopHandler := handler.NewShopHandler(shopService)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	serviceHandler := handler.NewServiceHandler(serviceService, shopServiceService)
	searchHandler := handler.NewSearchHandler(searchService)
	barberProfileHandler := handler.NewBarberProfileHandler(barberProfileService)
	shopProfileHandler := handler.NewShopProfileHandler(shopProfileService)
	barberShopRequestHandler := handler.NewBarberShopRequestHandler(barberShopRequestService, barberProfileService)
	adminHandler := handler.NewAdminHandler(adminService)
	healthHandler := handler.NewHealthHandler()
	photoHandler := handler.NewPhotoHandler(photoService)
	photoLikeHandler := handler.NewPhotoLikeHandler(photoLikeService)

	// Health check endpoint (public, auth gerekmez)
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/api/health", healthHandler.HealthCheck)

	// Public routes (Auth gerekmez)
	api := r.Group("/api")
	{
		api.POST("/auth/token", authHandler.GetToken)
		api.POST("/auth/verify-phone", authHandler.VerifyPhone)
	}

	// Protected routes (Auth gerekir)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		user := protected.Group("/user")
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PATCH("/profile", userHandler.UpdateProfile)
			user.PATCH("/avatar", userHandler.UpdateAvatar)
			user.DELETE("/avatar", userHandler.DeleteAvatar)
		}

		// Appointment routes
		appointments := protected.Group("/appointments")
		{
			appointments.GET("/past", appointmentHandler.GetPastAppointments)
			appointments.GET("/upcoming", appointmentHandler.GetUpcomingAppointments)
			appointments.GET("/:id", appointmentHandler.GetAppointment)
			appointments.POST("", appointmentHandler.CreateAppointment)
			appointments.DELETE("/:id", appointmentHandler.DeleteAppointment)
		}

		// Barber routes
		barbers := protected.Group("/barbers")
		{
			barbers.GET("/nearby", barberHandler.GetNearbyBarbers)
			barbers.GET("/:id", barberHandler.GetBarber)
			barbers.GET("/:id/availability", barberHandler.GetAvailability)
		}

		// Barber Profile routes
		barberProfile := protected.Group("/barber")
		{
			barberProfile.POST("/register", barberProfileHandler.RegisterAsBarber)
			barberProfile.GET("/profile", barberProfileHandler.GetBarberProfile)
			barberProfile.PATCH("/profile", barberProfileHandler.UpdateBarberProfile)
		}

		// Shop routes
		shops := protected.Group("/shops")
		{
			shops.GET("/nearby", shopHandler.GetNearbyShops)
			shops.GET("/:id", shopHandler.GetShop)
			shops.GET("/:id/barbers", shopHandler.GetShopBarbers)
			shops.GET("/:id/reviews", shopHandler.GetShopReviews)
		}

		// Shop Profile routes
		shopProfile := protected.Group("/shop")
		{
			shopProfile.POST("", shopProfileHandler.CreateShop)
			shopProfile.GET("/profile/:id", shopProfileHandler.GetShopProfile)
			shopProfile.PATCH("/profile/:id", shopProfileHandler.UpdateShopProfile)
			shopProfile.GET("/:id/owners", shopProfileHandler.GetShopOwners)
			shopProfile.POST("/:id/owners", shopProfileHandler.AddShopOwner)
			shopProfile.DELETE("/:id/owners/:ownerId", shopProfileHandler.RemoveShopOwner)
			shopProfile.POST("/:id/services", shopProfileHandler.CreateShopService)
			shopProfile.PATCH("/:id/services/:serviceId", shopProfileHandler.UpdateShopService)
			shopProfile.DELETE("/:id/services/:serviceId", shopProfileHandler.DeleteShopService)
		}

		// Barber Shop Request routes
		barberShopRequests := protected.Group("/shops")
		{
			barberShopRequests.POST("/:id/barber-requests", barberShopRequestHandler.CreateRequest)
			barberShopRequests.GET("/:id/barber-requests", barberShopRequestHandler.GetRequestsByShop)
			barberShopRequests.POST("/:id/barber-requests/:requestId/approve", barberShopRequestHandler.ApproveRequest)
			barberShopRequests.POST("/:id/barber-requests/:requestId/reject", barberShopRequestHandler.RejectRequest)
		}

		// Barber's own requests
		protected.GET("/barber/requests", barberShopRequestHandler.GetRequestsByBarber)

		// Favorite routes
		favorites := protected.Group("/favorites")
		{
			favorites.GET("", favoriteHandler.GetFavorites)
			favorites.GET("/shops", favoriteHandler.GetFavoriteShops)
			favorites.GET("/barbers", favoriteHandler.GetFavoriteBarbers)
			favorites.POST("/shops/:id", favoriteHandler.AddFavoriteShop)
			favorites.DELETE("/shops/:id", favoriteHandler.RemoveFavoriteShop)
			favorites.POST("/barbers/:id", favoriteHandler.AddFavoriteBarber)
			favorites.DELETE("/barbers/:id", favoriteHandler.RemoveFavoriteBarber)
		}

		// Review routes
		reviews := protected.Group("/reviews")
		{
			reviews.GET("", reviewHandler.GetReviews)
			reviews.POST("", reviewHandler.CreateReview)
			reviews.DELETE("/:id", reviewHandler.DeleteReview)
		}

		// Service routes
		services := protected.Group("/service")
		{
			services.GET("/shop/:id", serviceHandler.GetShopServices)
		}

		// Services list (tüm servisleri listele - admin olmayanlar da görebilir)
		protected.GET("/services", serviceHandler.GetAllServices)

		// Search route
		protected.GET("/search", searchHandler.Search)

		// Photo routes
		photos := protected.Group("/photos")
		{
			// İki parametreli route'lar (type ve id)
			photos.POST("/:type/:id", photoHandler.UploadPhoto)        // type: barber veya shop
			photos.GET("/:type/:id", photoHandler.GetPhotos)            // type: barber veya shop
		}

		// Photo (tekil) routes - çakışmayı önlemek için ayrı path
		photo := protected.Group("/photo")
		{
			photo.GET("/:id", photoHandler.GetPhoto)                   // Tek fotoğraf detayı
			photo.DELETE("/:id", photoHandler.DeletePhoto)
		}

		// Photo Like routes
		photoLikes := protected.Group("/photo")
		{
			photoLikes.POST("/:id/like", photoLikeHandler.LikePhoto)
			photoLikes.DELETE("/:id/like", photoLikeHandler.UnlikePhoto)
			photoLikes.GET("/:id/likes", photoLikeHandler.GetPhotoLikes)
			photoLikes.GET("/:id/likes/count", photoLikeHandler.GetPhotoLikeCount)
			photoLikes.GET("/:id/liked", photoLikeHandler.IsLiked)
		}
	}

	// Admin routes (Auth + Admin gerekir)
	admin := api.Group("")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/admin/shops/pending", adminHandler.GetPendingShops)
		admin.POST("/admin/shops/:id/approve", adminHandler.ApproveShop)
		admin.POST("/admin/shops/:id/reject", adminHandler.RejectShop)
		admin.POST("/admin/services", adminHandler.CreateService)
		admin.PATCH("/admin/services/:id", adminHandler.UpdateService)
		admin.DELETE("/admin/services/:id", adminHandler.DeleteService)
	}

	// Sunucuyu başlat
	port := config.AppConfig.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
