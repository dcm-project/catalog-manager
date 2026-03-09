package service

import (
	"context"

	"github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes/three_tier_app_demo"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

// Seed ensures required service types and default catalog items exist.
func (s *service) Seed(ctx context.Context) error {
	if err := s.store.ServiceType().SeedIfEmpty(ctx, defaultServiceTypes()); err != nil {
		return err
	}
	return s.store.CatalogItem().SeedIfEmpty(ctx, defaultCatalogItems())
}

func defaultServiceTypes() []model.ServiceType {
	emptyNetwork := &three_tier_app_demo.Network{}
	return []model.ServiceType{
		{
			ID:          "three_tier_app_demo",
			ApiVersion:  "v1alpha1",
			ServiceType: "three_tier_app_demo",
			Spec: map[string]any{
				"database": three_tier_app_demo.DatabaseTier{Engine: three_tier_app_demo.DefaultDatabaseEngine, Version: three_tier_app_demo.DefaultDatabaseVersion, Network: emptyNetwork},
				"app":      three_tier_app_demo.AppTier{Image: "", Network: emptyNetwork},
				"web":      three_tier_app_demo.WebTier{Image: "", Network: emptyNetwork},
			},
			Path: "service-types/three_tier_app_demo",
		},
	}
}

func defaultCatalogItems() []model.CatalogItem {
	return []model.CatalogItem{
		petClinicCatalogItem(),
	}
}

func petClinicCatalogItem() model.CatalogItem {
	return model.CatalogItem{
		ID:          "pet-clinic",
		ApiVersion:  "v1alpha1",
		DisplayName: "Pet Clinic",
		Path:        "catalog-items/pet-clinic",
		Spec: model.CatalogItemSpec{
			ServiceType: "three_tier_app_demo",
			Fields:      petClinicFields(),
		},
		SpecServiceType: "three_tier_app_demo",
	}
}

func petClinicFields() []model.FieldConfiguration {
	return []model.FieldConfiguration{
		fieldConfig("database.engine", "Database engine", true,
			three_tier_app_demo.DefaultDatabaseEngine,
			map[string]any{"type": "string", "enum": []any{"postgres", "mysql"}}, nil),
		fieldConfig("database.version", "Database version", true,
			three_tier_app_demo.DefaultDatabaseVersion,
			map[string]any{"type": "string"}, dependsOn("database.engine", map[string][]any{
				"postgres": {three_tier_app_demo.DefaultDatabaseVersion, "17"},
				"mysql":    {"8.4", "8.3", "8"},
			})),
		fieldConfig("app.image", "App image", false,
			"docker.io/springcommunity/spring-framework-petclinic:6.1.2", nil, nil),
		fieldConfig("web.image", "Web image", false,
			"docker.io/library/nginx:alpine", nil, nil),
	}
}

func fieldConfig(path, displayName string, editable bool, defaultVal any,
	validationSchema map[string]any, dependsOn *model.DependsOn,
) model.FieldConfiguration {
	return model.FieldConfiguration{
		Path:             path,
		DisplayName:      displayName,
		Editable:         editable,
		Default:          defaultVal,
		ValidationSchema: validationSchema,
		DependsOn:        dependsOn,
	}
}

func dependsOn(path string, allowedValues map[string][]any) *model.DependsOn {
	return &model.DependsOn{Path: path, AllowedValues: allowedValues}
}
