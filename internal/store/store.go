package store

import (
	"context"

	"gorm.io/gorm"
)

// Store provides access to all resource stores
type Store interface {
	ServiceType() ServiceTypeStore
	CatalogItem() CatalogItemStore
	CatalogItemInstance() CatalogItemInstanceStore
	Seed(ctx context.Context) error
	Close() error
}

// DataStore implements the Store interface
type DataStore struct {
	db                  *gorm.DB
	serviceType         ServiceTypeStore
	catalogItem         CatalogItemStore
	catalogItemInstance CatalogItemInstanceStore
}

// NewStore creates a new DataStore
func NewStore(db *gorm.DB) Store {
	return &DataStore{
		db:                  db,
		serviceType:         NewServiceTypeStore(db),
		catalogItem:         NewCatalogItemStore(db),
		catalogItemInstance: NewCatalogItemInstanceStore(db),
	}
}

// ServiceType returns the ServiceType store
func (s *DataStore) ServiceType() ServiceTypeStore {
	return s.serviceType
}

// CatalogItem returns the CatalogItem store
func (s *DataStore) CatalogItem() CatalogItemStore {
	return s.catalogItem
}

// CatalogItemInstance returns the CatalogItemInstance store
func (s *DataStore) CatalogItemInstance() CatalogItemInstanceStore {
	return s.catalogItemInstance
}

// Seed ensures required service types and default catalog items exist.
func (s *DataStore) Seed(ctx context.Context) error {
	if err := s.ServiceType().SeedServiceTypesIfEmpty(ctx); err != nil {
		return err
	}
	return s.CatalogItem().SeedIfEmpty(ctx)
}

// Close closes the database connection
func (s *DataStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
