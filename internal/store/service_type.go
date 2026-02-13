package store

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"github.com/dcm-project/catalog-manager/internal/store/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrServiceTypeNotFound is returned when a service type is not found
	ErrServiceTypeNotFound = errors.New("service type not found")
	// ErrServiceTypeIDTaken is returned when a service type ID is already taken
	ErrServiceTypeIDTaken = errors.New("service type ID already exists")
	// ErrServiceTypeServiceTypeTaken is returned when a service type service type is already taken
	ErrServiceTypeServiceTypeTaken = errors.New("service type service type already exists")
)

// PolicyListOptions contains options for listing policies.
type ServiceTypeListOptions struct {
	PageToken *string
	PageSize  int
}

// PolicyListResult contains the result of a List operation.
type ServiceTypeListResult struct {
	ServiceTypes  model.ServiceTypeList
	NextPageToken *string
}

// ServiceTypeStore defines operations for ServiceType resources
type ServiceTypeStore interface {
	List(ctx context.Context, opts *ServiceTypeListOptions) (*ServiceTypeListResult, error)
	Create(ctx context.Context, serviceType model.ServiceType) (*model.ServiceType, error)
	Get(ctx context.Context, id string) (*model.ServiceType, error)
}

type serviceTypeStore struct {
	db *gorm.DB
}

// NewServiceTypeStore creates a new ServiceType store
func NewServiceTypeStore(db *gorm.DB) ServiceTypeStore {
	return &serviceTypeStore{db: db}
}

// List returns a paginated list of service types
func (s *serviceTypeStore) List(ctx context.Context, opts *ServiceTypeListOptions) (*ServiceTypeListResult, error) {
	var serviceTypes model.ServiceTypeList
	query := s.db.WithContext(ctx)

	// Default page size
	pageSize := 50
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
// by querying the DB to see which constraint would be violated (ID, display_name+policy_type, or priority+policy_type).
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
