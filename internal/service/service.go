// Package service implements the business logic for catalog management.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dcm-project/catalog-manager/internal/config"
	"github.com/dcm-project/catalog-manager/internal/placement"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/google/uuid"
)

// Service is the main interface that aggregates all service interfaces
type Service interface {
	ServiceType() ServiceTypeService
	CatalogItem() CatalogItemService
	CatalogItemInstance() CatalogItemInstanceService
	Seed(ctx context.Context) error
}

// service is the implementation of the Service interface
type service struct {
	store                      store.Store
	logger                     *slog.Logger
	petClinicConfig            config.PetClinicConfig
	serviceTypeService         ServiceTypeService
	catalogItemService         CatalogItemService
	catalogItemInstanceService CatalogItemInstanceService
}

// NewService creates a new Service instance
func NewService(store store.Store, pmClient placement.Client, cfg config.PetClinicConfig, logger *slog.Logger) (Service, error) {
	svcLogger := logger.With("component", "service")
	catalogItemInstanceSvc, err := newCatalogItemInstanceService(store, pmClient, svcLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog item instance service: %w", err)
	}
	return &service{
		store:                      store,
		logger:                     svcLogger,
		petClinicConfig:            cfg,
		serviceTypeService:         newServiceTypeService(store, svcLogger),
		catalogItemService:         newCatalogItemService(store, svcLogger),
		catalogItemInstanceService: catalogItemInstanceSvc,
	}, nil
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
