package service

import (
	"barbertime_go/pkg/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
)

type MessageService interface {
	PublishAppointmentCreated(appointmentID uint, userID uint, barberID uint) error
	PublishAppointmentCancelled(appointmentID uint, userID uint) error
	PublishReviewCreated(reviewID uint, targetType string, targetID uint) error
	PublishBarberCreated(barberID uint) error
	PublishBarberUpdated(barberID uint) error
	PublishShopCreated(shopID uint) error
	PublishShopUpdated(shopID uint) error
}

type messageService struct{}

func NewMessageService() MessageService {
	return &messageService{}
}

type AppointmentCreatedMessage struct {
	AppointmentID uint   `json:"appointmentId"`
	UserID        uint   `json:"userId"`
	BarberID      uint   `json:"barberId"`
	EventType     string `json:"eventType"`
}

type AppointmentCancelledMessage struct {
	AppointmentID uint   `json:"appointmentId"`
	UserID        uint   `json:"userId"`
	EventType     string `json:"eventType"`
}

type ReviewCreatedMessage struct {
	ReviewID   uint   `json:"reviewId"`
	TargetType string `json:"targetType"`
	TargetID   uint   `json:"targetId"`
	EventType  string `json:"eventType"`
}

type BarberEventMessage struct {
	BarberID  uint   `json:"barberId"`
	EventType string `json:"eventType"`
}

type ShopEventMessage struct {
	ShopID    uint   `json:"shopId"`
	EventType string `json:"eventType"`
}

func (s *messageService) PublishAppointmentCreated(appointmentID uint, userID uint, barberID uint) error {
	if rabbitmq.Channel == nil {
		return nil // RabbitMQ yoksa sessizce geç
	}

	message := AppointmentCreatedMessage{
		AppointmentID: appointmentID,
		UserID:        userID,
		BarberID:      barberID,
		EventType:     "appointment.created",
	}

	return rabbitmq.Publish("appointments", message)
}

func (s *messageService) PublishAppointmentCancelled(appointmentID uint, userID uint) error {
	if rabbitmq.Channel == nil {
		return nil
	}

	message := AppointmentCancelledMessage{
		AppointmentID: appointmentID,
		UserID:        userID,
		EventType:     "appointment.cancelled",
	}

	return rabbitmq.Publish("appointments", message)
}

func (s *messageService) PublishReviewCreated(reviewID uint, targetType string, targetID uint) error {
	if rabbitmq.Channel == nil {
		return nil
	}

	message := ReviewCreatedMessage{
		ReviewID:   reviewID,
		TargetType: targetType,
		TargetID:   targetID,
		EventType:  "review.created",
	}

	return rabbitmq.Publish("reviews", message)
}

func (s *messageService) PublishBarberCreated(barberID uint) error {
	if rabbitmq.Channel == nil {
		return nil
	}

	message := BarberEventMessage{
		BarberID:  barberID,
		EventType: "barber.created",
	}

	return rabbitmq.Publish("barber.created", message)
}

func (s *messageService) PublishBarberUpdated(barberID uint) error {
	if rabbitmq.Channel == nil {
		return nil
	}

	message := BarberEventMessage{
		BarberID:  barberID,
		EventType: "barber.updated",
	}

	return rabbitmq.Publish("barber.updated", message)
}

func (s *messageService) PublishShopCreated(shopID uint) error {
	if rabbitmq.Channel == nil {
		return nil
	}

	message := ShopEventMessage{
		ShopID:    shopID,
		EventType: "shop.created",
	}

	return rabbitmq.Publish("shop.created", message)
}

func (s *messageService) PublishShopUpdated(shopID uint) error {
	if rabbitmq.Channel == nil {
		return nil
	}

	message := ShopEventMessage{
		ShopID:    shopID,
		EventType: "shop.updated",
	}

	return rabbitmq.Publish("shop.updated", message)
}

// Consumer örnekleri (opsiyonel)
func StartAppointmentConsumer() {
	if rabbitmq.Channel == nil {
		return
	}

	rabbitmq.Consume("appointments", func(body []byte) error {
		var message AppointmentCreatedMessage
		if err := json.Unmarshal(body, &message); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		log.Printf("Processing appointment event: %+v", message)
		// Burada notification gönderme, email gönderme vb. işlemler yapılabilir
		return nil
	})
}
