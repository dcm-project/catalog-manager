package store

import (
	"context"
	"errors"
	"log"

	"github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

// SeedServiceTypesIfEmpty ensures required service types exist for catalog item seeds.
func (s *serviceTypeStore) SeedServiceTypesIfEmpty(ctx context.Context) error {
	required := []model.ServiceType{
		{
			ID:          "three_tier_app_demo",
			ApiVersion:  "v1alpha1",
			ServiceType: "three_tier_app_demo",
			Spec:        map[string]any{},
			Path:        "service-types/three_tier_app_demo",
		},
	}
	for _, st := range required {
		_, err := s.Get(ctx, st.ID)
		if err == nil {
			continue
		}
		if !errors.Is(err, ErrServiceTypeNotFound) {
			return err
		}
		if _, err := s.Create(ctx, st); err != nil {
			return err
		}
		log.Printf("Seeded service type %s", st.ServiceType)
	}
	return nil
}

// defaultCatalogItems returns the initial catalog items to seed when the DB is empty.
func defaultCatalogItems() []v1alpha1.CatalogItem {
	editableTrue := true
	editableFalse := false
	dbEngineDisplay := "Database engine"
	path := "catalog-items/pet-clinic"
	uid := "pet-clinic"
	fields := []v1alpha1.FieldConfiguration{
		{Path: "database.engine", Default: "postgres", DisplayName: &dbEngineDisplay, Editable: &editableTrue,
			ValidationSchema: &map[string]interface{}{"type": "string", "enum": []string{"mysql", "postgres"}}},
		{Path: "database.version", Default: "18", DisplayName: strPtr("Database version"), Editable: &editableTrue,
			DependsOn: &v1alpha1.FieldConfigurationDependsOn{
				Path: "database.engine",
				Mapping: map[string]interface{}{
					"postgres": []interface{}{"16", "17", "18"},
					"mysql":   []interface{}{"8.0", "8.4"},
				},
			}},
		{Path: "database.image", DisplayName: strPtr("Database image"), Editable: &editableFalse,
			DependsOn: &v1alpha1.FieldConfigurationDependsOn{
				Path: "database.version",
				Mapping: map[string]interface{}{"16": "postgres:16", "17": "postgres:17", "18": "postgres:18", "8.0": "mysql:8.0", "8.4": "mysql:8.4"},
			}},
		{Path: "app.image", Default: "docker.io/springcommunity/spring-framework-petclinic:6.1.2", DisplayName: strPtr("App image"), Editable: &editableFalse},
		{Path: "web.image", Default: "docker.io/library/nginx:alpine", DisplayName: strPtr("Web image"), Editable: &editableFalse},
	}
	return []v1alpha1.CatalogItem{
		{
			ApiVersion:  strPtr("v1alpha1"),
			DisplayName: strPtr("Pet Clinic"),
			Path:        &path,
			Uid:         &uid,
			Spec: &v1alpha1.CatalogItemSpec{
				ServiceType: strPtr("three_tier_app_demo"),
				Fields:      &fields,
			},
		},
	}
}

// SeedIfEmpty inserts default catalog items if the table has no rows.
func (s *catalogItemStore) SeedIfEmpty(ctx context.Context) error {
	items := defaultCatalogItems()
	var n int64
	if err := s.db.WithContext(ctx).Model(&model.CatalogItem{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, item := range items {
		path := ""
		if item.Path != nil {
			path = *item.Path
		}
		id := ""
		if item.Uid != nil {
			id = *item.Uid
		}
		apiVersion := ""
		if item.ApiVersion != nil {
			apiVersion = *item.ApiVersion
		}
		displayName := ""
		if item.DisplayName != nil {
			displayName = *item.DisplayName
		}
		serviceType := ""
		if item.Spec != nil && item.Spec.ServiceType != nil {
			serviceType = *item.Spec.ServiceType
		}
		var fields []v1alpha1.FieldConfiguration
		if item.Spec != nil && item.Spec.Fields != nil {
			fields = *item.Spec.Fields
		}
		m := model.CatalogItem{
			ID:          id,
			ApiVersion:  apiVersion,
			DisplayName: displayName,
			Path:        path,
			Spec: model.CatalogItemSpec{
				ServiceType: serviceType,
				Fields:      convertFieldConfig(fields),
			},
		}
		if _, err := s.Create(ctx, m); err != nil {
			return err
		}
	}
	log.Printf("Seeded %d default catalog item(s)", len(items))
	return nil
}

func strPtr(s string) *string { return &s }

func convertFieldConfig(f []v1alpha1.FieldConfiguration) []model.FieldConfiguration {
	if f == nil {
		return nil
	}
	out := make([]model.FieldConfiguration, len(f))
	for i := range f {
		var displayName string
		if f[i].DisplayName != nil {
			displayName = *f[i].DisplayName
		}
		var vs map[string]any
		if f[i].ValidationSchema != nil {
			vs = make(map[string]any)
			for k, v := range *f[i].ValidationSchema {
				vs[k] = v
			}
		}
		var dep *model.DependsOn
		if f[i].DependsOn != nil {
			m := make(map[string]any)
			for k, v := range f[i].DependsOn.Mapping {
				m[k] = v
			}
			dep = &model.DependsOn{Path: f[i].DependsOn.Path, Mapping: m}
		}
		out[i] = model.FieldConfiguration{
			Path:             f[i].Path,
			DisplayName:      displayName,
			Editable:         f[i].Editable != nil && *f[i].Editable,
			Default:          f[i].Default,
			ValidationSchema: vs,
			DependsOn:        dep,
		}
	}
	return out
}
