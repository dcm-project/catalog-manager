package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
)

func (h *Handler) ListCatalogItems(ctx context.Context, request server.ListCatalogItemsRequestObject) (server.ListCatalogItemsResponseObject, error) {
	detail := "endpoint not implemented"
	return server.ListCatalogItems500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) CreateCatalogItem(ctx context.Context, request server.CreateCatalogItemRequestObject) (server.CreateCatalogItemResponseObject, error) {
	detail := "endpoint not implemented"
	return server.CreateCatalogItem500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) GetCatalogItem(ctx context.Context, request server.GetCatalogItemRequestObject) (server.GetCatalogItemResponseObject, error) {
	detail := "endpoint not implemented"
	return server.GetCatalogItem500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) UpdateCatalogItem(ctx context.Context, request server.UpdateCatalogItemRequestObject) (server.UpdateCatalogItemResponseObject, error) {
	detail := "endpoint not implemented"
	return server.UpdateCatalogItem500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) DeleteCatalogItem(ctx context.Context, request server.DeleteCatalogItemRequestObject) (server.DeleteCatalogItemResponseObject, error) {
	detail := "endpoint not implemented"
	return server.DeleteCatalogItem500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}
