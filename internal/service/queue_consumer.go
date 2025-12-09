package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/pkg/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
)

// StartQueueConsumer Queue consumer'ı başlatır (Elasticsearch indexing ve diğer event'ler için)
func StartQueueConsumer(
	barberRepo repository.BarberRepository,
	shopRepo repository.ShopRepository,
	searchService SearchService,
) {
	if rabbitmq.Channel == nil {
		log.Fatal("RabbitMQ channel is not available")
	}

	// Barber created queue
	if err := rabbitmq.Consume("barber.created", func(body []byte) error {
		var msg BarberEventMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal barber message: %w", err)
		}

		log.Printf("Barber update event received for barber %d (created)", msg.BarberID)

		barber, err := barberRepo.GetByID(msg.BarberID)
		if err != nil {
			return fmt.Errorf("barber not found: %w", err)
		}

		if err := searchService.IndexBarber(barber); err != nil {
			log.Printf("Error indexing barber %d: %v", msg.BarberID, err)
			return err
		}

		log.Printf("Successfully indexed barber %d in Elasticsearch (created)", msg.BarberID)
		return nil
	}); err != nil {
		log.Printf("Failed to start barber.created consumer: %v", err)
	}

	// Barber updated queue
	if err := rabbitmq.Consume("barber.updated", func(body []byte) error {
		var msg BarberEventMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal barber message: %w", err)
		}

		log.Printf("Barber update event received for barber %d (updated)", msg.BarberID)

		barber, err := barberRepo.GetByID(msg.BarberID)
		if err != nil {
			return fmt.Errorf("barber not found: %w", err)
		}

		if err := searchService.IndexBarber(barber); err != nil {
			log.Printf("Error indexing barber %d: %v", msg.BarberID, err)
			return err
		}

		log.Printf("Successfully indexed barber %d in Elasticsearch (updated)", msg.BarberID)
		return nil
	}); err != nil {
		log.Printf("Failed to start barber.updated consumer: %v", err)
	}

	// Shop created queue
	if err := rabbitmq.Consume("shop.created", func(body []byte) error {
		var msg ShopEventMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal shop message: %w", err)
		}

		log.Printf("Shop update event received for shop %d (created)", msg.ShopID)

		shop, err := shopRepo.GetByID(msg.ShopID)
		if err != nil {
			return fmt.Errorf("shop not found: %w", err)
		}

		// Sadece confirmed shop'ları indexle
		if shop.Status == models.ShopStatusConfirmed {
			if err := searchService.IndexShop(shop); err != nil {
				log.Printf("Error indexing shop %d: %v", msg.ShopID, err)
				return err
			}
			log.Printf("Successfully indexed shop %d in Elasticsearch (created)", msg.ShopID)
		} else {
			log.Printf("Shop %d not indexed (status: %s, only confirmed shops are indexed)", msg.ShopID, shop.Status)
		}

		return nil
	}); err != nil {
		log.Printf("Failed to start shop.created consumer: %v", err)
	}

	// Shop updated queue
	if err := rabbitmq.Consume("shop.updated", func(body []byte) error {
		var msg ShopEventMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal shop message: %w", err)
		}

		log.Printf("Shop update event received for shop %d (updated)", msg.ShopID)

		shop, err := shopRepo.GetByID(msg.ShopID)
		if err != nil {
			return fmt.Errorf("shop not found: %w", err)
		}

		// Sadece confirmed shop'ları indexle
		if shop.Status == models.ShopStatusConfirmed {
			if err := searchService.IndexShop(shop); err != nil {
				log.Printf("Error indexing shop %d: %v", msg.ShopID, err)
				return err
			}
			log.Printf("Successfully indexed shop %d in Elasticsearch (updated)", msg.ShopID)
		} else {
			log.Printf("Shop %d not indexed (status: %s, only confirmed shops are indexed)", msg.ShopID, shop.Status)
		}

		return nil
	}); err != nil {
		log.Printf("Failed to start shop.updated consumer: %v", err)
	}

	log.Println("Queue consumer started (barber.created, barber.updated, shop.created, shop.updated)")

	// Buraya diğer queue consumer'lar eklenebilir
	// Örnek: notifications, emails, analytics vb.
}
