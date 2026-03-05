package service

import "errors"

// Domain errors for the service layer
var (
	// ErrInvalidServiceType indicates the service type is not one of the allowed values (vm, container, cluster, database)
	ErrInvalidServiceType = errors.New("invalid service type: must be one of: vm, container, cluster, database")

	// ErrServiceTypeIDTaken indicates a service type with the given ID already exists
	ErrServiceTypeIDTaken = errors.New("service type ID already exists")

	// ErrServiceTypeNameTaken indicates a service type with the given service_type value already exists
	ErrServiceTypeNameTaken = errors.New("service type name already taken")

	// ErrServiceTypeNotFound indicates the requested service type does not exist
	ErrServiceTypeNotFound = errors.New("service type not found")

	// ErrCatalogItemNotFound indicates the requested catalog item does not exist
	ErrCatalogItemNotFound = errors.New("catalog item not found")

	// ErrCatalogItemIDTaken indicates a catalog item with the given ID already exists
	ErrCatalogItemIDTaken = errors.New("catalog item ID already exists")

	// ErrCatalogItemHasInstances indicates a catalog item has existing instances
	ErrCatalogItemHasInstances = errors.New("catalog item has existing instances")

	// ErrImmutableFieldUpdate indicates an attempt to change api_version or spec.service_type
	ErrImmutableFieldUpdate = errors.New("cannot update immutable fields: api_version and spec.service_type are immutable")

	// ErrCatalogItemInstanceNotFound indicates the requested catalog item instance does not exist
	ErrCatalogItemInstanceNotFound = errors.New("catalog item instance not found")

	// ErrCatalogItemInstanceIDTaken indicates a catalog item instance with the given ID already exists
	ErrCatalogItemInstanceIDTaken = errors.New("catalog item instance ID already exists")

	// ErrCatalogItemNotFoundForInstance indicates the catalog item referenced by the instance does not exist
	ErrCatalogItemNotFoundForInstance = errors.New("referenced catalog item does not exist")

	// ErrUserValuePathNotFound indicates a user_value path does not match any CatalogItem field
	ErrUserValuePathNotFound = errors.New("user value path not found in catalog item fields")

	// ErrUserValueNotEditable indicates the field at the given path is not editable
	ErrUserValueNotEditable = errors.New("field is not editable")

	// ErrUserValueValidationFailed indicates the user value failed validation against the field's validation_schema
	ErrUserValueValidationFailed = errors.New("user value validation failed")

	// ErrPlacementManagerCreateFailed indicates the Placement Manager failed to create a resource
	ErrPlacementManagerCreateFailed = errors.New("placement manager create resource failed")

	// ErrPlacementManagerDeleteFailed indicates the Placement Manager failed to delete a resource
	ErrPlacementManagerDeleteFailed = errors.New("placement manager delete resource failed")
)
