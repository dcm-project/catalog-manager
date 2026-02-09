package v1alpha1

import (
	"context"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
)

func (h *Handler) ListServiceTypes(ctx context.Context, request server.ListServiceTypesRequestObject) (server.ListServiceTypesResponseObject, error) {
	detail := "endpoint not implemented"
	return server.ListServiceTypes500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) CreateServiceType(ctx context.Context, request server.CreateServiceTypeRequestObject) (server.CreateServiceTypeResponseObject, error) {
	detail := "endpoint not implemented"
	return server.CreateServiceType500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}

func (h *Handler) GetServiceType(ctx context.Context, request server.GetServiceTypeRequestObject) (server.GetServiceTypeResponseObject, error) {
	detail := "endpoint not implemented"
	return server.GetServiceType500JSONResponse{
		InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
			Type:   v1alpha1.UNIMPLEMENTED,
			Status: 500,
			Title:  "Not Implemented",
			Detail: &detail,
		},
	}, nil
}
