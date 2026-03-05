package service

import (
	"context"
	"fmt"

	"github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/placement"
	"github.com/dcm-project/catalog-manager/internal/store"
)

// CreateCatalogItemInstanceRequest contains the parameters for creating a catalog item instance
type CreateCatalogItemInstanceRequest struct {
	ID          *string                          // Optional user-specified ID
	ApiVersion  string                           // e.g., "v1alpha1"
	DisplayName string                           // Required, max 63 chars
	Spec        v1alpha1.CatalogItemInstanceSpec // Required, contains catalog_item_id and user_values
}

// CatalogItemInstanceListOptions contains options for listing catalog item instances
type CatalogItemInstanceListOptions struct {
	PageToken     *string
	MaxPageSize   *int32
	CatalogItemId *string // Filter by catalog_item_id
}

// CatalogItemInstanceListResult contains the result of a List operation
type CatalogItemInstanceListResult struct {
	CatalogItemInstances []v1alpha1.CatalogItemInstance
	NextPageToken        *string
}

// CatalogItemInstanceService defines the business logic for CatalogItemInstance operations
type CatalogItemInstanceService interface {
	List(ctx context.Context, opts CatalogItemInstanceListOptions) (*CatalogItemInstanceListResult, error)
	Create(ctx context.Context, req *CreateCatalogItemInstanceRequest) (*v1alpha1.CatalogItemInstance, error)
	Get(ctx context.Context, id string) (*v1alpha1.CatalogItemInstance, error)
	Delete(ctx context.Context, id string) error
}

type catalogItemInstanceService struct {
	store       store.Store
	specBuilder *specBuilder
	pmClient    placement.Client
}

// newCatalogItemInstanceService creates a new CatalogItemInstanceService instance
func newCatalogItemInstanceService(store store.Store, pmClient placement.Client) CatalogItemInstanceService {
	return &catalogItemInstanceService{
		store:       store,
		specBuilder: newSpecBuilder(store),
		pmClient:    pmClient,
	}
}

// List returns a paginated list of catalog item instances
func (s *catalogItemInstanceService) List(ctx context.Context, opts CatalogItemInstanceListOptions) (*CatalogItemInstanceListResult, error) {
	// Convert service options to store options
	storeOpts := &store.CatalogItemInstanceListOptions{
		PageToken:     opts.PageToken,
		CatalogItemId: opts.CatalogItemId,
	}
	if opts.MaxPageSize != nil {
		storeOpts.PageSize = int(*opts.MaxPageSize)
	}

	// Call store layer
	storeResult, err := s.store.CatalogItemInstance().List(ctx, storeOpts)
	if err != nil {
		return nil, err
	}

	// Convert store models to API types
	apiTypes := make([]v1alpha1.CatalogItemInstance, len(storeResult.CatalogItemInstances))
	for i, storeModel := range storeResult.CatalogItemInstances {
		apiTypes[i] = catalogItemInstanceToAPIType(&storeModel)
	}

	return &CatalogItemInstanceListResult{
		CatalogItemInstances: apiTypes,
		NextPageToken:        storeResult.NextPageToken,
	}, nil
}

// Create creates a new catalog item instance
func (s *catalogItemInstanceService) Create(ctx context.Context, req *CreateCatalogItemInstanceRequest) (*v1alpha1.CatalogItemInstance, error) {
	// Generate ID
	id := getOrGenerateID(req.ID)
	// Generate path
	path := fmt.Sprintf("catalog-item-instances/%s", id)

	// Build resource spec (resolves reference chain and validates user_values)
	resourceSpec, err := s.specBuilder.BuildResourceSpec(ctx, req.Spec.CatalogItemId, req.Spec.UserValues)
	if err != nil {
		return nil, err
	}

	// DB first — fail fast on constraint violations (ID conflict, FK violation)
	storeModel := catalogItemInstanceToStoreModel(id, path, req)
	createdModel, err := s.store.CatalogItemInstance().Create(ctx, storeModel)
	if err != nil {
		return nil, mapCatalogItemInstanceStoreError(err)
	}

	// Call Placement Manager — only after DB validation passes
	if s.pmClient != nil {
		_, err := s.pmClient.CreateResource(ctx, placement.CreateResourceRequest{
			CatalogItemInstanceID: id,
			Spec:                  resourceSpec,
		}, id)
		if err != nil {
			// Rollback: delete DB record
			_ = s.store.CatalogItemInstance().Delete(ctx, id)
			return nil, fmt.Errorf("%w: %s", ErrPlacementManagerCreateFailed, err.Error())
		}
	}

	// Convert result back to API type
	apiType := catalogItemInstanceToAPIType(createdModel)
	return &apiType, nil
}

// Get retrieves a catalog item instance by ID
func (s *catalogItemInstanceService) Get(ctx context.Context, id string) (*v1alpha1.CatalogItemInstance, error) {
	// Call store layer
	storeModel, err := s.store.CatalogItemInstance().Get(ctx, id)
	if err != nil {
		return nil, mapCatalogItemInstanceStoreError(err)
	}

	// Convert to API type
	apiType := catalogItemInstanceToAPIType(storeModel)
	return &apiType, nil
}

// Delete deletes a catalog item instance by ID
func (s *catalogItemInstanceService) Delete(ctx context.Context, id string) error {
	// Fetch instance for 404 handling
	_, err := s.store.CatalogItemInstance().Get(ctx, id)
	if err != nil {
		return mapCatalogItemInstanceStoreError(err)
	}

	// Delete PM resource (PM resource ID is the catalog item instance id)
	if s.pmClient != nil {
		if err := s.pmClient.DeleteResource(ctx, id); err != nil {
			return fmt.Errorf("%w: %s", ErrPlacementManagerDeleteFailed, err.Error())
		}
	}

	// Delete local record
	err = s.store.CatalogItemInstance().Delete(ctx, id)
	if err != nil {
		return mapCatalogItemInstanceStoreError(err)
	}
	return nil
}
