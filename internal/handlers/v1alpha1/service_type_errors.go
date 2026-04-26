package v1alpha1

import (
	"errors"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/service"
)

// mapCreateServiceErrorToHTTP converts service domain errors to CreateServiceType HTTP responses
func mapCreateServiceErrorToHTTP(err error) server.CreateServiceTypeResponseObject {
	switch {
	case errors.Is(err, service.ErrInvalidServiceType):
		// Validation errors -> 400 Bad Request
		return server.CreateServiceType400JSONResponse(v1alpha1.Error{
			Type:   v1alpha1.INVALIDARGUMENT,
			Status: 400,
			Title:  "Bad Request",
			Detail: stringPtr(err.Error()),
		})
	case errors.Is(err, service.ErrServiceTypeIDTaken), errors.Is(err, service.ErrServiceTypeNameTaken):
		// Conflict errors -> 409 Conflict
		return server.CreateServiceType409JSONResponse{
			AlreadyExistsJSONResponse: server.AlreadyExistsJSONResponse{
				Type:   v1alpha1.ALREADYEXISTS,
				Status: 409,
				Title:  "Conflict",
				Detail: stringPtr(err.Error()),
			},
		}
	default:
		// Unknown errors -> 500 Internal Server Error
		return server.CreateServiceType500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}
	}
}

// mapUpdateServiceTypeErrorToHTTP converts service domain errors to UpdateServiceType HTTP responses
func mapUpdateServiceTypeErrorToHTTP(err error) server.UpdateServiceTypeResponseObject {
	switch {
	case errors.Is(err, service.ErrServiceTypeNotFound):
		return server.UpdateServiceType404JSONResponse{
			NotFoundJSONResponse: server.NotFoundJSONResponse{
				Type:   v1alpha1.NOTFOUND,
				Status: 404,
				Title:  "Not Found",
				Detail: stringPtr(err.Error()),
			},
		}
	default:
		return server.UpdateServiceType500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}
	}
}

// mapDeleteServiceTypeErrorToHTTP converts service domain errors to DeleteServiceType HTTP responses
func mapDeleteServiceTypeErrorToHTTP(err error) server.DeleteServiceTypeResponseObject {
	switch {
	case errors.Is(err, service.ErrServiceTypeNotFound):
		return server.DeleteServiceType404JSONResponse{
			NotFoundJSONResponse: server.NotFoundJSONResponse{
				Type:   v1alpha1.NOTFOUND,
				Status: 404,
				Title:  "Not Found",
				Detail: stringPtr(err.Error()),
			},
		}
	case errors.Is(err, service.ErrServiceTypeHasCatalogItems):
		return server.DeleteServiceType409JSONResponse{
			HasDependentsJSONResponse: server.HasDependentsJSONResponse{
				Type:   v1alpha1.FAILEDPRECONDITION,
				Status: 409,
				Title:  "Failed Precondition",
				Detail: stringPtr(err.Error()),
			},
		}
	default:
		return server.DeleteServiceType500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}
	}
}

// mapGetServiceErrorToHTTP converts service domain errors to GetServiceType HTTP responses
func mapGetServiceErrorToHTTP(err error) server.GetServiceTypeResponseObject {
	switch {
	case errors.Is(err, service.ErrServiceTypeNotFound):
		// Not found -> 404 Not Found
		return server.GetServiceType404JSONResponse{
			NotFoundJSONResponse: server.NotFoundJSONResponse{
				Type:   v1alpha1.NOTFOUND,
				Status: 404,
				Title:  "Not Found",
				Detail: stringPtr(err.Error()),
			},
		}
	default:
		// Unknown errors -> 500 Internal Server Error
		return server.GetServiceType500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}
	}
}
