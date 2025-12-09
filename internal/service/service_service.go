package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
)

type ServiceService interface {
	GetAllServices() ([]models.Service, error)
}

type serviceService struct {
	serviceRepo repository.ServiceRepository
}

func NewServiceService(serviceRepo repository.ServiceRepository) ServiceService {
	return &serviceService{
		serviceRepo: serviceRepo,
	}
}

func (s *serviceService) GetAllServices() ([]models.Service, error) {
	return s.serviceRepo.GetAll()
}
