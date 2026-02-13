package store

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dcm-project/catalog-manager/internal/store/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrCatalogItemNotFound is returned when a catalog item is not found
	ErrCatalogItemNotFound = errors.New("catalog item not found")
	// ErrCatalogItemIDTaken is returned when a catalog item ID is already taken
	ErrCatalogItemIDTaken = errors.New("catalog item ID already exists")
	// ErrCatalogItemHasInstances is returned when attempting to delete a catalog item with existing instances
	ErrCatalogItemHasInstances = errors.New("cannot delete catalog item with existing instances")
)

// CatalogItemListOptions contains options for listing catalog items
type CatalogItemListOptions struct {
	PageToken   *string
	PageSize    int
	ServiceType string
}

// CatalogItemListResult contains the result of a List operation
type CatalogItemListResult struct {
	CatalogItems  model.CatalogItemList
	NextPageToken string
}

// CatalogItemStore defines operations for CatalogItem resources
type CatalogItemStore interface {
	List(ctx context.Context, opts *CatalogItemListOptions) (*CatalogItemListResult, error)
	Create(ctx context.Context, catalogItem model.CatalogItem) (*model.CatalogItem, error)
	Get(ctx context.Context, id string) (*model.CatalogItem, error)
	Update(ctx context.Context, catalogItem *model.CatalogItem) error
	Delete(ctx context.Context, id string) error
}

type catalogItemStore struct {
	db *gorm.DB
}

// NewCatalogItemStore creates a new CatalogItem store
func NewCatalogItemStore(db *gorm.DB) CatalogItemStore {
	return &catalogItemStore{db: db}
}

// List returns a paginated list of catalog items
func (s *catalogItemStore) List(ctx context.Context, opts *CatalogItemListOptions) (*CatalogItemListResult, error) {
	var catalogItems model.CatalogItemList
	query := s.db.WithContext(ctx)

	// Default and max page size
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

	query = query.Order("id ASC").Limit(pageSize + 1).Offset(offset)
	if opts != nil && opts.ServiceType != "" {
		query = query.Where("spec_service_type = ?", opts.ServiceType)
	}

	if err := query.Find(&catalogItems).Error; err != nil {
		return nil, err
	}

	result := &CatalogItemListResult{
		CatalogItems: catalogItems,
	}
	if len(catalogItems) > pageSize {
		result.CatalogItems = catalogItems[:pageSize]
		nextOffset := offset + pageSize
		result.NextPageToken = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(nextOffset)))
	}
	return result, nil
}

// Create creates a new catalog item
func (s *catalogItemStore) Create(ctx context.Context, catalogItem model.CatalogItem) (*model.CatalogItem, error) {
	catalogItem.SpecServiceType = catalogItem.Spec.ServiceType
	if err := s.db.WithContext(ctx).Clauses(clause.Returning{}).Create(&catalogItem).Error; err != nil {
		return nil, s.mapConstraintError(ctx, err, catalogItem)
	}
	return &catalogItem, nil
}

// mapConstraintError maps a DB constraint violation to a store sentinel error
func (s *catalogItemStore) mapConstraintError(ctx context.Context, err error, attempted model.CatalogItem) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check for foreign key violation first (before checking for generic constraint failed)
	if strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "violates foreign key constraint") {
		// Verify which constraint failed by checking if service type exists
		var st model.ServiceType
		if err := s.db.WithContext(ctx).Where("service_type = ?", attempted.SpecServiceType).First(&st).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrServiceTypeNotFound
			}
		}
		return err
	}

	// Handle unique constraint violations
	if errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(errStr, "unique") ||
		strings.Contains(err.Error(), "duplicate key") {
		var row model.CatalogItem
		dberr := s.db.WithContext(ctx).Where("id = ?", attempted.ID).Limit(1).First(&row).Error
		if dberr == nil {
			return ErrCatalogItemIDTaken
		}
		if !errors.Is(dberr, gorm.ErrRecordNotFound) {
			return err
		}
	}

	return err
}

// Get retrieves a catalog item by ID
func (s *catalogItemStore) Get(ctx context.Context, id string) (*model.CatalogItem, error) {
	var catalogItem model.CatalogItem
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&catalogItem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCatalogItemNotFound
		}
		return nil, fmt.Errorf("failed to get catalog item: %w", err)
	}
	return &catalogItem, nil
}

// Update updates a catalog item (only mutable fields)
func (s *catalogItemStore) Update(ctx context.Context, catalogItem *model.CatalogItem) error {
	// Extract service type from spec for denormalized field
	catalogItem.SpecServiceType = catalogItem.Spec.ServiceType

	result := s.db.WithContext(ctx).Model(&model.CatalogItem{}).
		Where("id = ?", catalogItem.ID).
		Select("display_name", "spec", "spec_service_type").
		Updates(catalogItem)

	if result.Error != nil {
		// Check for foreign key violation
		errStr := strings.ToLower(result.Error.Error())
		if strings.Contains(errStr, "foreign key") ||
			strings.Contains(errStr, "violates foreign key constraint") ||
			strings.Contains(errStr, "constraint failed: foreign key") {
			return ErrServiceTypeNotFound
		}
		return fmt.Errorf("failed to update catalog item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCatalogItemNotFound
	}
	return nil
}

// Delete deletes a catalog item by ID
func (s *catalogItemStore) Delete(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.CatalogItem{})
	if result.Error != nil {
		// Check for foreign key violation (instances exist)
		errStr := strings.ToLower(result.Error.Error())
		if strings.Contains(errStr, "foreign key") ||
			strings.Contains(errStr, "violates foreign key constraint") ||
			strings.Contains(errStr, "constraint failed: foreign key") {
			return ErrCatalogItemHasInstances
		}
		return fmt.Errorf("failed to delete catalog item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrCatalogItemNotFound
	}
	return nil
}
