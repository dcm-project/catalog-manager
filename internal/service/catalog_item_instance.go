package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/placement"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/google/uuid"
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
	Rehydrate(ctx context.Context, id string) (*v1alpha1.CatalogItemInstance, error)
}

type catalogItemInstanceService struct {
	store       store.Store
	specBuilder *specBuilder
	pmClient    placement.Client
	logger      *slog.Logger
}

// newCatalogItemInstanceService creates a new CatalogItemInstanceService instance.
// pmClient must not be nil.
func newCatalogItemInstanceService(store store.Store, pmClient placement.Client, logger *slog.Logger) (CatalogItemInstanceService, error) {
	if pmClient == nil {
		return nil, fmt.Errorf("pmClient must not be nil")
	}
	return &catalogItemInstanceService{
		store:       store,
		specBuilder: newSpecBuilder(store),
		pmClient:    pmClient,
		logger:      logger,
	}, nil
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
	// Generate IDs
	id := getOrGenerateID(req.ID)
	resourceID := uuid.New().String()
	// Generate path
	path := fmt.Sprintf("catalog-item-instances/%s", id)

	// Build resource spec (resolves reference chain and validates user_values)
	resourceSpec, err := s.specBuilder.BuildResourceSpec(ctx, req.Spec.CatalogItemId, req.Spec.UserValues)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to build resource spec",
			"id", id,
			"catalog_item_id", req.Spec.CatalogItemId,
			"error", err,
		)
		return nil, err
	}

	// DB first — fail fast on constraint violations (ID conflict, FK violation)
	storeModel := catalogItemInstanceToStoreModel(id, resourceID, path, req)
	createdModel, err := s.store.CatalogItemInstance().Create(ctx, storeModel)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create catalog item instance in store", "id", id, "error", err)
		return nil, mapCatalogItemInstanceStoreError(err)
	}

	// Call Placement Manager — only after DB validation passes
	s.logger.DebugContext(ctx, "Calling placement manager to create resource", "id", id)
	_, err = s.pmClient.CreateResource(ctx, placement.CreateResourceRequest{
		CatalogItemInstanceID: id,
		Spec:                  resourceSpec,
	}, resourceID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Placement manager create failed, rolling back",
			"id", id,
			"error", err,
		)
		// Rollback: delete DB record
		_ = s.store.CatalogItemInstance().Delete(ctx, id)
		return nil, mapPlacementError(err, ErrPlacementManagerCreateFailed)
	}

	s.logger.InfoContext(ctx, "Catalog item instance created", "id", id, "catalog_item_id", req.Spec.CatalogItemId)
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

// Rehydrate rehydrates a catalog item instance by generating a new resource ID
// and delegating to the Placement Manager
func (s *catalogItemInstanceService) Rehydrate(ctx context.Context, id string) (*v1alpha1.CatalogItemInstance, error) {
	// Look up existing instance
	instance, err := s.store.CatalogItemInstance().Get(ctx, id)
	if err != nil {
		return nil, mapCatalogItemInstanceStoreError(err)
	}

	// Generate new resource ID
	newResourceID := uuid.New().String()

	// Call Placement Manager rehydrate
	s.logger.DebugContext(ctx, "Calling placement manager to rehydrate resource",
		"id", id,
		"old_resource_id", instance.ResourceID,
		"new_resource_id", newResourceID,
	)
	_, err = s.pmClient.RehydrateResource(ctx, instance.ResourceID, newResourceID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Placement manager rehydrate failed",
			"id", id,
			"error", err,
		)
		return nil, mapPlacementError(err, ErrPlacementManagerRehydrateFailed)
	}

	// Update resource_id in DB
	updatedModel, err := s.store.CatalogItemInstance().UpdateResourceID(ctx, id, newResourceID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to update resource ID in store",
			"id", id,
			"error", err,
		)
		return nil, err
	}

	s.logger.InfoContext(ctx, "Catalog item instance rehydrated",
		"id", id,
		"new_resource_id", newResourceID,
	)

	apiType := catalogItemInstanceToAPIType(updatedModel)
	return &apiType, nil
}

// Delete deletes a catalog item instance by ID
func (s *catalogItemInstanceService) Delete(ctx context.Context, id string) error {
	// Fetch instance for 404 handling and to get the resource ID
	instance, err := s.store.CatalogItemInstance().Get(ctx, id)
	if err != nil {
		return mapCatalogItemInstanceStoreError(err)
	}

	// Delete PM resource using the stored resource ID
	if instance.ResourceID != "" {
		s.logger.DebugContext(ctx, "Calling placement manager to delete resource", "id", id, "resource_id", instance.ResourceID)
		if err := s.pmClient.DeleteResource(ctx, instance.ResourceID); err != nil {
			s.logger.ErrorContext(ctx, "Placement manager delete failed", "id", id, "error", err)
			return fmt.Errorf("%w: %s", ErrPlacementManagerDeleteFailed, err.Error())
		}
	}

	// Delete local record
	err = s.store.CatalogItemInstance().Delete(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete catalog item instance from store", "id", id, "error", err)
		return mapCatalogItemInstanceStoreError(err)
	}

	s.logger.InfoContext(ctx, "Catalog item instance deleted", "id", id)
	return nil
}

// mapPlacementError inspects the error from the placement client and maps
// known HTTP status codes (406, 422) to specific sentinel errors. For
// unrecognised codes or non-PlacementError errors, the genericSentinel is used.
func mapPlacementError(err error, genericSentinel error) error {
	var pmErr *placement.PlacementError
	if errors.As(err, &pmErr) {
		switch pmErr.StatusCode {
		case http.StatusNotAcceptable:
			return fmt.Errorf("%w: %s", ErrPlacementManagerPolicyRejected, pmErr.Error())
		case http.StatusUnprocessableEntity:
			return fmt.Errorf("%w: %s", ErrPlacementManagerProviderError, pmErr.Error())
		}
	}
	return fmt.Errorf("%w: %s", genericSentinel, err.Error())
}
