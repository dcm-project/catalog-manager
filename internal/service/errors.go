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
)
