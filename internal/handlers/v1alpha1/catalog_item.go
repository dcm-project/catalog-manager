package v1alpha1

import (
	"context"
	"errors"
	"log"

	v1alpha1 "github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

func (h *Handler) ListCatalogItems(ctx context.Context, request server.ListCatalogItemsRequestObject) (server.ListCatalogItemsResponseObject, error) {
	opts := &store.CatalogItemListOptions{}
	if request.Params.PageToken != nil {
		opts.PageToken = request.Params.PageToken
	}
	if request.Params.MaxPageSize != nil {
		opts.PageSize = int(*request.Params.MaxPageSize)
	}
	if request.Params.ServiceType != nil {
		opts.ServiceType = request.Params.ServiceType
	}

	result, err := h.store.CatalogItem().List(ctx, opts)
	if err != nil {
		log.Printf("ListCatalogItems failed: %v", err)
		detail := "An internal error occurred"
		return server.ListCatalogItems500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: &detail,
			},
		}, nil
	}

	nextPageToken := ""
	if result.NextPageToken != nil {
		nextPageToken = *result.NextPageToken
	}
	return server.ListCatalogItems200JSONResponse{
		NextPageToken: nextPageToken,
		Results:       modelToAPICatalogItems(result.CatalogItems),
	}, nil
}

func modelToAPICatalogItems(items model.CatalogItemList) []v1alpha1.CatalogItem {
	out := make([]v1alpha1.CatalogItem, len(items))
	for i := range items {
		out[i] = modelToAPICatalogItem(&items[i])
	}
	return out
}

func modelToAPICatalogItem(m *model.CatalogItem) v1alpha1.CatalogItem {
	path := m.Path
	uid := m.ID
	apiVersion := m.ApiVersion
	displayName := m.DisplayName
	serviceType := m.Spec.ServiceType
	fields := modelToAPIFieldConfigs(m.Spec.Fields)
	return v1alpha1.CatalogItem{
		ApiVersion:  &apiVersion,
		DisplayName: &displayName,
		Path:        &path,
		Uid:         &uid,
		CreateTime:  &m.CreateTime,
		UpdateTime:  &m.UpdateTime,
		Spec: &v1alpha1.CatalogItemSpec{
			ServiceType: &serviceType,
			Fields:      &fields,
		},
	}
}

func modelToAPIFieldConfigs(f []model.FieldConfiguration) []v1alpha1.FieldConfiguration {
	out := make([]v1alpha1.FieldConfiguration, len(f))
	for i := range f {
		var displayName *string
		if f[i].DisplayName != "" {
			displayName = &f[i].DisplayName
		}
		editable := f[i].Editable
		var vs *map[string]interface{}
		if len(f[i].ValidationSchema) > 0 {
			m := make(map[string]interface{})
			for k, v := range f[i].ValidationSchema {
				m[k] = v
			}
			vs = &m
		}
		var dep *v1alpha1.FieldConfigurationDependsOn
		if f[i].DependsOn != nil {
			m := make(map[string]interface{})
			for k, v := range f[i].DependsOn.Mapping {
				m[k] = v
			}
			dep = &v1alpha1.FieldConfigurationDependsOn{Path: f[i].DependsOn.Path, Mapping: m}
		}
		out[i] = v1alpha1.FieldConfiguration{
			Path:             f[i].Path,
			Default:          f[i].Default,
			DisplayName:      displayName,
			Editable:         &editable,
			ValidationSchema: vs,
			DependsOn:        dep,
		}
	}
	return out
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
	item, err := h.store.CatalogItem().Get(ctx, string(request.CatalogItemId))
	if err != nil {
		if errors.Is(err, store.ErrCatalogItemNotFound) {
			detail := "catalog item not found"
			return server.GetCatalogItem404JSONResponse{
				NotFoundJSONResponse: server.NotFoundJSONResponse{
					Type:   v1alpha1.NOTFOUND,
					Status: 404,
					Title:  "Not Found",
					Detail: &detail,
				},
			}, nil
		}
		log.Printf("GetCatalogItem failed: %v", err)
		detail := "An internal error occurred"
		return server.GetCatalogItem500JSONResponse{
			InternalServerErrorJSONResponse: server.InternalServerErrorJSONResponse{
				Type:   v1alpha1.INTERNAL,
				Status: 500,
				Title:  "Internal Server Error",
				Detail: &detail,
			},
		}, nil
	}
	return server.GetCatalogItem200JSONResponse(modelToAPICatalogItem(item)), nil
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
