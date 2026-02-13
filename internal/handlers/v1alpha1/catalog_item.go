package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/service"
)

const (
	supportedAPIVersion = "v1alpha1"
)

func (h *Handler) ListCatalogItems(ctx context.Context, request server.ListCatalogItemsRequestObject) (server.ListCatalogItemsResponseObject, error) {
	// Build service request from HTTP params
	opts := service.CatalogItemListOptions{
		PageToken:   request.Params.PageToken,
		MaxPageSize: request.Params.MaxPageSize,
		ServiceType: request.Params.ServiceType,
	}

	// Call service layer
	result, err := h.service.CatalogItem().List(ctx, opts)
	if err != nil {
		return server.ListCatalogItems500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: stringPtr(err.Error()),
			},
		}, nil
	}

	// Return HTTP response
	response := server.ListCatalogItems200JSONResponse(v1alpha1.CatalogItemList{
		Results: result.CatalogItems,
	})
	if result.NextPageToken != nil {
		response.NextPageToken = *result.NextPageToken
	}
	return response, nil
}

func (h *Handler) CreateCatalogItem(ctx context.Context, request server.CreateCatalogItemRequestObject) (server.CreateCatalogItemResponseObject, error) {
	// Build service request from HTTP params
	req, err := validateAndGetbuildCreateCatalogItemRequest(request)
	if err != nil {
		return server.CreateCatalogItem400JSONResponse(v1alpha1.Error{
			Type:   v1alpha1.INVALIDARGUMENT,
			Status: 400,
			Title:  "Bad Request",
			Detail: stringPtr(err.Error()),
		}), nil
	}

	// Call service layer
	result, err := h.service.CatalogItem().Create(ctx, req)
	if err != nil {
		return mapCreateCatalogItemErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.CreateCatalogItem201JSONResponse(*result), nil
}

func validateAndGetbuildCreateCatalogItemRequest(request server.CreateCatalogItemRequestObject) (*service.CreateCatalogItemRequest, error) {
	if request.Body.ApiVersion == nil || *request.Body.ApiVersion != supportedAPIVersion {
		return nil, ErrInvalidAPIVersion
	}
	if request.Body.DisplayName == nil {
		return nil, ErrInvalidDisplayName
	}
	if request.Body.Spec == nil {
		return nil, ErrEmptySpec
	}
	if request.Body.Spec.ServiceType == nil {
		return nil, ErrInvalidServiceType
	}
	if request.Body.Spec.Fields == nil {
		return nil, ErrEmptyFields
	}
	return &service.CreateCatalogItemRequest{
		ID:          request.Params.Id,
		ApiVersion:  *request.Body.ApiVersion,
		DisplayName: *request.Body.DisplayName,
		Spec:        *request.Body.Spec,
	}, nil
}

func (h *Handler) GetCatalogItem(ctx context.Context, request server.GetCatalogItemRequestObject) (server.GetCatalogItemResponseObject, error) {
	// Call service layer
	result, err := h.service.CatalogItem().Get(ctx, request.CatalogItemId)
	if err != nil {
		return mapGetCatalogItemErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.GetCatalogItem200JSONResponse(*result), nil
}

func (h *Handler) UpdateCatalogItem(ctx context.Context, request server.UpdateCatalogItemRequestObject) (server.UpdateCatalogItemResponseObject, error) {
	// Body is already a CatalogItem (partial update via JSON merge patch)
	// Build update request from provided fields
	updateReq := &service.UpdateCatalogItemRequest{
		DisplayName: request.Body.DisplayName,
		Spec:        request.Body.Spec,
	}

	// Call service layer
	result, err := h.service.CatalogItem().Update(ctx, request.CatalogItemId, updateReq)
	if err != nil {
		return mapUpdateCatalogItemErrorToHTTP(err), nil
	}

	// Return HTTP response
	return server.UpdateCatalogItem200JSONResponse(*result), nil
}

func (h *Handler) DeleteCatalogItem(ctx context.Context, request server.DeleteCatalogItemRequestObject) (server.DeleteCatalogItemResponseObject, error) {
	// Call service layer
	err := h.service.CatalogItem().Delete(ctx, request.CatalogItemId)
	if err != nil {
		return mapDeleteCatalogItemErrorToHTTP(err), nil
	}

	// Return HTTP 204 No Content response
	return server.DeleteCatalogItem204Response{}, nil
}
