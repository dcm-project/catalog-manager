package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/service"
)

func (h *Handler) ListServiceTypes(ctx context.Context, request server.ListServiceTypesRequestObject) (server.ListServiceTypesResponseObject, error) {
	h.logger.DebugContext(ctx, "Listing service types")

	// Build service request from HTTP params
	opts := &service.ServiceTypeListOptions{
		PageToken:   request.Params.PageToken,
		MaxPageSize: request.Params.MaxPageSize,
	}

	// Call service layer
	result, err := h.service.ServiceType().List(ctx, opts)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list service types", "error", err)
		return server.ListServiceTypes500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}, nil
	}

	h.logger.DebugContext(ctx, "Listed service types", "count", len(result.ServiceTypes))

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
	h.logger.InfoContext(ctx, "Creating service type",
		"id", request.Params.Id,
		"service_type", request.Body.ServiceType,
	)

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
		h.logServiceError(ctx, "Failed to create service type", err, "service_type", request.Body.ServiceType)
		return mapCreateServiceErrorToHTTP(err), nil
	}

	h.logger.InfoContext(ctx, "Created service type", "service_type", result.ServiceType)

	// Return HTTP response
	return server.CreateServiceType201JSONResponse(*result), nil
}

func (h *Handler) GetServiceType(ctx context.Context, request server.GetServiceTypeRequestObject) (server.GetServiceTypeResponseObject, error) {
	h.logger.DebugContext(ctx, "Getting service type", "id", request.ServiceTypeId)

	// Call service layer
	result, err := h.service.ServiceType().Get(ctx, request.ServiceTypeId)
	if err != nil {
		h.logServiceError(ctx, "Failed to get service type", err, "id", request.ServiceTypeId)
		return mapGetServiceErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.GetServiceType200JSONResponse(*result), nil
}

func (h *Handler) UpdateServiceType(ctx context.Context, request server.UpdateServiceTypeRequestObject) (server.UpdateServiceTypeResponseObject, error) {
	h.logger.InfoContext(ctx, "Updating service type", "id", request.ServiceTypeId)

	updateReq := &service.UpdateServiceTypeRequest{
		Metadata: request.Body.Metadata,
		Spec:     request.Body.Spec,
	}

	result, err := h.service.ServiceType().Update(ctx, request.ServiceTypeId, updateReq)
	if err != nil {
		h.logServiceError(ctx, "Failed to update service type", err, "id", request.ServiceTypeId)
		return mapUpdateServiceTypeErrorToHTTP(err), nil
	}

	h.logger.InfoContext(ctx, "Updated service type", "id", request.ServiceTypeId)
	return server.UpdateServiceType200JSONResponse(*result), nil
}

func (h *Handler) DeleteServiceType(ctx context.Context, request server.DeleteServiceTypeRequestObject) (server.DeleteServiceTypeResponseObject, error) {
	h.logger.InfoContext(ctx, "Deleting service type", "id", request.ServiceTypeId)

	err := h.service.ServiceType().Delete(ctx, request.ServiceTypeId)
	if err != nil {
		h.logServiceError(ctx, "Failed to delete service type", err, "id", request.ServiceTypeId)
		return mapDeleteServiceTypeErrorToHTTP(err), nil
	}

	h.logger.InfoContext(ctx, "Deleted service type", "id", request.ServiceTypeId)
	return server.DeleteServiceType204Response{}, nil
}
