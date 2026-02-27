package service

import (
	"errors"

	"github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

// catalogItemToStoreModel converts a CreateCatalogItemRequest to a store model
func catalogItemToStoreModel(id, path string, req *CreateCatalogItemRequest) model.CatalogItem {
	// Convert FieldConfiguration from API type to model type
	fields := make([]model.FieldConfiguration, len(*req.Spec.Fields))
	for i, f := range *req.Spec.Fields {
		fields[i] = model.FieldConfiguration{
			Path:        f.Path,
			Default:     f.Default,
			Editable:    f.Editable != nil && *f.Editable,
			DisplayName: "",
		}
		if f.DisplayName != nil {
			fields[i].DisplayName = *f.DisplayName
		}
		if f.ValidationSchema != nil {
			fields[i].ValidationSchema = *f.ValidationSchema
		}
	}

	storeModel := model.CatalogItem{
		ID:          id,
		ApiVersion:  req.ApiVersion,
		DisplayName: req.DisplayName,
		Spec: model.CatalogItemSpec{
			ServiceType: *req.Spec.ServiceType,
			Fields:      fields,
		},
		Path:            path,
		SpecServiceType: *req.Spec.ServiceType, // Indexed field for filtering
	}

	return storeModel
}

// catalogItemToAPIType converts a store model to an API type
func catalogItemToAPIType(m *model.CatalogItem) v1alpha1.CatalogItem {
	// Convert FieldConfiguration from model type to API type
	fields := make([]v1alpha1.FieldConfiguration, len(m.Spec.Fields))
	for i, f := range m.Spec.Fields {
		fields[i] = v1alpha1.FieldConfiguration{
			Path:    f.Path,
			Default: f.Default,
		}
		if f.DisplayName != "" {
			displayName := f.DisplayName
			fields[i].DisplayName = &displayName
		}
		if f.Editable {
			editable := true
			fields[i].Editable = &editable
		}
		if f.ValidationSchema != nil {
			validationSchema := f.ValidationSchema
			fields[i].ValidationSchema = &validationSchema
		}
	}

	apiType := v1alpha1.CatalogItem{
		ApiVersion:  &m.ApiVersion,
		DisplayName: &m.DisplayName,
		Spec: &v1alpha1.CatalogItemSpec{
			ServiceType: &m.Spec.ServiceType,
			Fields:      &fields,
		},
		Path:       &m.Path,
		Uid:        &m.ID,
		CreateTime: &m.CreateTime,
		UpdateTime: &m.UpdateTime,
	}

	return apiType
}

// mapCatalogItemStoreError converts store errors to service domain errors
func mapCatalogItemStoreError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, store.ErrCatalogItemNotFound):
		return ErrCatalogItemNotFound
	case errors.Is(err, store.ErrCatalogItemIDTaken):
		return ErrCatalogItemIDTaken
	case errors.Is(err, store.ErrCatalogItemHasInstances):
		return ErrCatalogItemHasInstances
	case errors.Is(err, store.ErrServiceTypeNotFound):
		return ErrServiceTypeNotFound
	default:
		return err
	}
}
