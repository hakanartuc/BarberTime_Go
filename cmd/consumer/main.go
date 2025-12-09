package main

import (
	"log"

	"barbertime_go/internal/config"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/service"
	"barbertime_go/pkg/database"
	"barbertime_go/pkg/elasticsearch"
	"barbertime_go/pkg/rabbitmq"
)

func main() {
	// Config yükle
	config.LoadConfig()

	// Veritabanı bağlantısı
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Elasticsearch bağlantısı (zorunlu - indexing için)
	if err := elasticsearch.Connect(); err != nil {
		log.Fatal("Failed to connect to Elasticsearch:", err)
	}

	// RabbitMQ bağlantısı (zorunlu)
	if err := rabbitmq.Connect(); err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitmq.Close()

	// Repository'leri oluştur
	barberRepo := repository.NewBarberRepository()
	shopRepo := repository.NewShopRepository()
	shopServiceRepo := repository.NewShopServiceRepository()

	// Service'leri oluştur
	searchService := service.NewSearchService(barberRepo, shopRepo, shopServiceRepo)

	// Queue consumer'ı başlat
	log.Println("Starting queue consumer...")
	service.StartQueueConsumer(barberRepo, shopRepo, searchService)

	// Uygulama çalışmaya devam et
	select {}
}
