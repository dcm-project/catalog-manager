package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
)

func (h *Handler) ListCatalogItemInstances(ctx context.Context, request server.ListCatalogItemInstancesRequestObject) (server.ListCatalogItemInstancesResponseObject, error) {
	detail := "endpoint not implemented"
	return server.ListCatalogItemInstances500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) CreateCatalogItemInstance(ctx context.Context, request server.CreateCatalogItemInstanceRequestObject) (server.CreateCatalogItemInstanceResponseObject, error) {
	detail := "endpoint not implemented"
	return server.CreateCatalogItemInstance500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) GetCatalogItemInstance(ctx context.Context, request server.GetCatalogItemInstanceRequestObject) (server.GetCatalogItemInstanceResponseObject, error) {
	detail := "endpoint not implemented"
	return server.GetCatalogItemInstance500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) DeleteCatalogItemInstance(ctx context.Context, request server.DeleteCatalogItemInstanceRequestObject) (server.DeleteCatalogItemInstanceResponseObject, error) {
	detail := "endpoint not implemented"
	return server.DeleteCatalogItemInstance500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}
