package store

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dcm-project/catalog-manager/internal/store/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrServiceTypeNotFound is returned when a service type is not found
	ErrServiceTypeNotFound = errors.New("store: service type not found")
	// ErrServiceTypeIDTaken is returned when a service type ID is already taken
	ErrServiceTypeIDTaken = errors.New("store: service type ID already exists")
	// ErrServiceTypeServiceTypeTaken is returned when a service type service type is already taken
	ErrServiceTypeServiceTypeTaken = errors.New("store: service type service type already exists")
	// ErrServiceTypeHasCatalogItems is returned when attempting to delete a service type with existing catalog items
	ErrServiceTypeHasCatalogItems = errors.New("cannot delete service type with existing catalog items")
)

// ServiceTypeListOptions contains options for listing service types.
type ServiceTypeListOptions struct {
	PageToken *string
	PageSize  int
}

// ServiceTypeListResult contains the result of a List operation.
type ServiceTypeListResult struct {
	ServiceTypes  model.ServiceTypeList
	NextPageToken *string
}

// ServiceTypeStore defines operations for ServiceType resources
type ServiceTypeStore interface {
	List(ctx context.Context, opts *ServiceTypeListOptions) (*ServiceTypeListResult, error)
	Create(ctx context.Context, serviceType model.ServiceType) (*model.ServiceType, error)
	Get(ctx context.Context, id string) (*model.ServiceType, error)
	GetByServiceType(ctx context.Context, serviceType string) (*model.ServiceType, error)
	Update(ctx context.Context, serviceType *model.ServiceType) error
	Delete(ctx context.Context, id string) error
	SeedIfEmpty(ctx context.Context, items []model.ServiceType) error
}

type serviceTypeStore struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewServiceTypeStore creates a new ServiceType store
func NewServiceTypeStore(db *gorm.DB, logger *slog.Logger) ServiceTypeStore {
	return &serviceTypeStore{db: db, logger: logger}
}

// List returns a paginated list of service types
func (s *serviceTypeStore) List(ctx context.Context, opts *ServiceTypeListOptions) (*ServiceTypeListResult, error) {
	var serviceTypes model.ServiceTypeList
	query := s.db.WithContext(ctx)

	// Default max page size
	pageSize := 100
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}

	// Decode page token to get offset
	offset := 0
	if opts != nil && opts.PageToken != nil && *opts.PageToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(*opts.PageToken)
		if err == nil {
			if parsedOffset, err := strconv.Atoi(string(decoded)); err == nil {
				offset = parsedOffset
			}
		}
	}

	query = query.Order("service_type ASC").Limit(pageSize + 1).Offset(offset)

	if err := query.Find(&serviceTypes).Error; err != nil {
		return nil, err
	}

	// Generate next page token if there are more results
	result := &ServiceTypeListResult{
		ServiceTypes: serviceTypes,
	}

	if len(serviceTypes) > pageSize {
		// Trim to requested page size
		result.ServiceTypes = serviceTypes[:pageSize]
		// Encode next offset as page token
		nextOffset := offset + pageSize
		nextPageToken := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(nextOffset)))
		result.NextPageToken = &nextPageToken
	}

	return result, nil
}

func (s *serviceTypeStore) Create(ctx context.Context, serviceType model.ServiceType) (*model.ServiceType, error) {
	if err := s.db.WithContext(ctx).Clauses(clause.Returning{}).Select("*").Create(&serviceType).Error; err != nil {
		return nil, s.mapUniqueConstraintError(ctx, err, serviceType)
	}
	return &serviceType, nil
}

// mapUniqueConstraintError maps a DB unique constraint violation to a store sentinel error.
// by querying the DB to see which constraint would be violated (ID, service_type).
func (s *serviceTypeStore) mapUniqueConstraintError(ctx context.Context, err error, attempted model.ServiceType) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		// Raw driver error (e.g. tests without TranslateError)
		if !strings.Contains(strings.ToLower(err.Error()), "unique") &&
			!strings.Contains(err.Error(), "duplicate key") {
			return err
		}
	}

	checks := []struct {
		sentinel error
		query    *gorm.DB
	}{
		{ErrServiceTypeIDTaken, s.db.WithContext(ctx).Where("id = ?", attempted.ID).Limit(1)},
		{ErrServiceTypeServiceTypeTaken, s.db.WithContext(ctx).Where("service_type = ?", attempted.ServiceType).Limit(1)},
	}

	for _, c := range checks {
		var row model.ServiceType
		dberr := c.query.First(&row).Error
		if dberr == nil {
			return c.sentinel
		}
		if !errors.Is(dberr, gorm.ErrRecordNotFound) {
			return err
		}
	}

	return err
}

// Get retrieves a service type by ID
func (s *serviceTypeStore) Get(ctx context.Context, id string) (*model.ServiceType, error) {
	var serviceType model.ServiceType
	if err := s.db.WithContext(ctx).First(&serviceType, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceTypeNotFound
		}
		return nil, err
	}
	return &serviceType, nil
}

// GetByServiceType retrieves a service type by its service_type value
func (s *serviceTypeStore) GetByServiceType(ctx context.Context, serviceType string) (*model.ServiceType, error) {
	var st model.ServiceType
	if err := s.db.WithContext(ctx).Where("service_type = ?", serviceType).First(&st).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceTypeNotFound
		}
		return nil, err
	}
	return &st, nil
}

// Update updates a service type (only mutable fields)
func (s *serviceTypeStore) Update(ctx context.Context, serviceType *model.ServiceType) error {
	result := s.db.WithContext(ctx).Model(&model.ServiceType{}).
		Where("id = ?", serviceType.ID).
		Select("metadata", "spec").
		Updates(serviceType)

	if result.Error != nil {
		return s.mapUniqueConstraintError(ctx, result.Error, *serviceType)
	}
	if result.RowsAffected == 0 {
		return ErrServiceTypeNotFound
	}
	return nil
}

// Delete deletes a service type by ID
func (s *serviceTypeStore) Delete(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ServiceType{})
	if result.Error != nil {
		errStr := strings.ToLower(result.Error.Error())
		if strings.Contains(errStr, "foreign key") {
			return ErrServiceTypeHasCatalogItems
		}
		return fmt.Errorf("failed to delete service type: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrServiceTypeNotFound
	}
	return nil
}

// SeedIfEmpty inserts the given service types if the table has no rows.
// Uses a transaction to avoid races when multiple instances start concurrently.
func (s *serviceTypeStore) SeedIfEmpty(ctx context.Context, items []model.ServiceType) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&model.ServiceType{}).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return nil
		}
		var inserted int64
		for _, m := range items {
			result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&m)
			if err := result.Error; err != nil {
				return err
			}
			inserted += result.RowsAffected
		}
		if inserted > 0 {
			s.logger.InfoContext(ctx, "Seeded default service types", "count", inserted)
		}
		return nil
	})
}
