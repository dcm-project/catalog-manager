package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/service"
)

func (h *Handler) ListServiceTypes(ctx context.Context, request server.ListServiceTypesRequestObject) (server.ListServiceTypesResponseObject, error) {
	// Build service request from HTTP params
	opts := &service.ServiceTypeListOptions{
		PageToken:   request.Params.PageToken,
		MaxPageSize: request.Params.MaxPageSize,
	}

	// Call service layer
	result, err := h.service.ServiceType().List(ctx, opts)
	if err != nil {
		return mapListServiceErrorToHTTP(err), nil
	}

	// Return HTTP response
	response := server.ListServiceTypes200JSONResponse(v1alpha1.ServiceTypeList{
		Results: result.ServiceTypes,
	})
	if result.NextPageToken != nil {
		response.NextPageToken = *result.NextPageToken
	}

	return response, nil
}

func (h *Handler) CreateServiceType(ctx context.Context, request server.CreateServiceTypeRequestObject) (server.CreateServiceTypeResponseObject, error) {
	// Build service request from HTTP params
	req := &service.CreateServiceTypeRequest{
		ID:          request.Params.Id,
		ApiVersion:  request.Body.ApiVersion,
		ServiceType: request.Body.ServiceType,
		Metadata:    request.Body.Metadata,
		Spec:        request.Body.Spec,
	}

	// Call service layer
	result, err := h.service.ServiceType().Create(ctx, req)
	if err != nil {
		return mapCreateServiceErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.CreateServiceType201JSONResponse(*result), nil
}

func (h *Handler) GetServiceType(ctx context.Context, request server.GetServiceTypeRequestObject) (server.GetServiceTypeResponseObject, error) {
	// Call service layer
	result, err := h.service.ServiceType().Get(ctx, request.ServiceTypeId)
	if err != nil {
		return mapGetServiceErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.GetServiceType200JSONResponse(*result), nil
}
