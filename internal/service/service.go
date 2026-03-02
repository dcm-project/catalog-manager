package service

import (
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/google/uuid"
)

// Service is the main interface that aggregates all service interfaces
type Service interface {
	ServiceType() ServiceTypeService
	CatalogItem() CatalogItemService
}

// service is the implementation of the Service interface
type service struct {
	store              store.Store
	serviceTypeService ServiceTypeService
	catalogItemService CatalogItemService
}

// NewService creates a new Service instance
func NewService(store store.Store) Service {
	return &service{
		store:              store,
		serviceTypeService: newServiceTypeService(store),
		catalogItemService: newCatalogItemService(store),
	}
}

// ServiceType returns the ServiceTypeService
func (s *service) ServiceType() ServiceTypeService {
	return s.serviceTypeService
}

// CatalogItem returns the CatalogItemService
func (s *service) CatalogItem() CatalogItemService {
	return s.catalogItemService
}

func getOrGenerateID(id *string) string {
	if id != nil && *id != "" {
		return *id
	}

	// Generate UUID if not provided
	return uuid.New().String()
}
