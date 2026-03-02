package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/service"
)

func (h *Handler) ListCatalogItemInstances(ctx context.Context, request server.ListCatalogItemInstancesRequestObject) (server.ListCatalogItemInstancesResponseObject, error) {
	// Build service request from HTTP params
	opts := service.CatalogItemInstanceListOptions{
		PageToken:     request.Params.PageToken,
		MaxPageSize:   request.Params.MaxPageSize,
		CatalogItemId: request.Params.CatalogItemId,
	}

	// Call service layer
	result, err := h.service.CatalogItemInstance().List(ctx, opts)
	if err != nil {
		return server.ListCatalogItemInstances500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}, nil
	}

	// Return HTTP response
	response := server.ListCatalogItemInstances200JSONResponse(v1alpha1.CatalogItemInstanceList{
		Results: result.CatalogItemInstances,
	})
	if result.NextPageToken != nil {
		response.NextPageToken = *result.NextPageToken
	}
	return response, nil
}

func (h *Handler) CreateCatalogItemInstance(ctx context.Context, request server.CreateCatalogItemInstanceRequestObject) (server.CreateCatalogItemInstanceResponseObject, error) {
	// Validate and build service request
	req, err := validateAndBuildCreateCatalogItemInstanceRequest(request)
	if err != nil {
		return server.CreateCatalogItemInstance400JSONResponse(v1alpha1.Error{
			Type:   v1alpha1.INVALIDARGUMENT,
			Status: 400,
			Title:  "Bad Request",
			Detail: stringPtr(err.Error()),
		}), nil
	}

	// Call service layer
	result, err := h.service.CatalogItemInstance().Create(ctx, req)
	if err != nil {
		return mapCreateCatalogItemInstanceErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.CreateCatalogItemInstance201JSONResponse(*result), nil
}

func validateAndBuildCreateCatalogItemInstanceRequest(request server.CreateCatalogItemInstanceRequestObject) (*service.CreateCatalogItemInstanceRequest, error) {
	if request.Body.ApiVersion != supportedAPIVersion {
		return nil, ErrInvalidCatalogItemInstanceAPIVersion
	}
	if request.Body.DisplayName == "" {
		return nil, ErrInvalidCatalogItemInstanceDisplayName
	}
	if request.Body.Spec.CatalogItemId == "" {
		return nil, ErrInvalidCatalogItemId
	}
	return &service.CreateCatalogItemInstanceRequest{
		ID:          request.Params.Id,
		ApiVersion:  request.Body.ApiVersion,
		DisplayName: request.Body.DisplayName,
		Spec:        request.Body.Spec,
	}, nil
}

func (h *Handler) GetCatalogItemInstance(ctx context.Context, request server.GetCatalogItemInstanceRequestObject) (server.GetCatalogItemInstanceResponseObject, error) {
	// Call service layer
	result, err := h.service.CatalogItemInstance().Get(ctx, request.CatalogItemInstanceId)
	if err != nil {
		return mapGetCatalogItemInstanceErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.GetCatalogItemInstance200JSONResponse(*result), nil
}

func (h *Handler) DeleteCatalogItemInstance(ctx context.Context, request server.DeleteCatalogItemInstanceRequestObject) (server.DeleteCatalogItemInstanceResponseObject, error) {
	// Call service layer
	err := h.service.CatalogItemInstance().Delete(ctx, request.CatalogItemInstanceId)
	if err != nil {
		return mapDeleteCatalogItemInstanceErrorToHTTP(err), nil
	}

	// Return HTTP 204 No Content response
	return server.DeleteCatalogItemInstance204Response{}, nil
}
