package service

import (
	"github.com/dcm-project/catalog-manager/internal/placement"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/google/uuid"
)

// Service is the main interface that aggregates all service interfaces
type Service interface {
	ServiceType() ServiceTypeService
	CatalogItem() CatalogItemService
	CatalogItemInstance() CatalogItemInstanceService
}

// service is the implementation of the Service interface
type service struct {
	store                      store.Store
	serviceTypeService         ServiceTypeService
	catalogItemService         CatalogItemService
	catalogItemInstanceService CatalogItemInstanceService
}

// NewService creates a new Service instance
func NewService(store store.Store, pmClient placement.Client) Service {
	return &service{
		store:                      store,
		serviceTypeService:         newServiceTypeService(store),
		catalogItemService:         newCatalogItemService(store),
		catalogItemInstanceService: newCatalogItemInstanceService(store, pmClient),
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

// CatalogItemInstance returns the CatalogItemInstanceService
func (s *service) CatalogItemInstance() CatalogItemInstanceService {
	return s.catalogItemInstanceService
}

func getOrGenerateID(id *string) string {
	if id != nil && *id != "" {
		return *id
	}

	// Generate UUID if not provided
	return uuid.New().String()
}
