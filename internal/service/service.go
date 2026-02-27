package service

import "github.com/dcm-project/catalog-manager/internal/store"

// Service is the main interface that aggregates all service interfaces
type Service interface {
	ServiceType() ServiceTypeService
}

// service is the implementation of the Service interface
type service struct {
	store              store.Store
	serviceTypeService ServiceTypeService
}

// NewService creates a new Service instance
func NewService(store store.Store) Service {
	return &service{
		store:              store,
		serviceTypeService: newServiceTypeService(store),
	}
}

// ServiceType returns the ServiceTypeService
func (s *service) ServiceType() ServiceTypeService {
	return s.serviceTypeService
}
