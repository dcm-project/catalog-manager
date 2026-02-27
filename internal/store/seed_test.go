package store_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

var _ = Describe("Seed", func() {
	var (
		db               *gorm.DB
		catalogItemStore store.CatalogItemStore
		serviceTypeStore store.ServiceTypeStore
	)

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Discard,
		})
		Expect(err).ToNot(HaveOccurred())

		err = db.Exec("PRAGMA foreign_keys = ON").Error
		Expect(err).ToNot(HaveOccurred())

		err = db.AutoMigrate(&model.ServiceType{}, &model.CatalogItem{})
		Expect(err).ToNot(HaveOccurred())

		catalogItemStore = store.NewCatalogItemStore(db)
		serviceTypeStore = store.NewServiceTypeStore(db)
	})

	AfterEach(func() {
		sqlDB, err := db.DB()
		Expect(err).ToNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("SeedIfEmpty", func() {
		It("seeds Pet Clinic catalog item when table is empty", func() {
			ctx := context.Background()

			// Create required service type for foreign key
			st := model.ServiceType{
				ID:          "three_tier_app_demo",
				ApiVersion:  "v1alpha1",
				ServiceType: "three_tier_app_demo",
				Spec:        map[string]any{},
				Path:        "service-types/three_tier_app_demo",
			}
			_, err := serviceTypeStore.Create(ctx, st)
			Expect(err).ToNot(HaveOccurred())

			err = catalogItemStore.SeedIfEmpty(ctx)
			Expect(err).ToNot(HaveOccurred())

			ci, err := catalogItemStore.Get(ctx, "pet-clinic")
			Expect(err).ToNot(HaveOccurred())
			Expect(ci.ID).To(Equal("pet-clinic"))
			Expect(ci.DisplayName).To(Equal("Pet Clinic"))
			Expect(ci.Path).To(Equal("catalog-items/pet-clinic"))
			Expect(ci.Spec.ServiceType).To(Equal("three_tier_app_demo"))
			Expect(ci.Spec.Fields).To(HaveLen(5))

			// Verify key field configs
			fieldPaths := make([]string, len(ci.Spec.Fields))
			for i, f := range ci.Spec.Fields {
				fieldPaths[i] = f.Path
			}
			Expect(fieldPaths).To(ContainElement("database.engine"))
			Expect(fieldPaths).To(ContainElement("database.version"))
			Expect(fieldPaths).To(ContainElement("database.image"))
			Expect(fieldPaths).To(ContainElement("app.image"))
			Expect(fieldPaths).To(ContainElement("web.image"))

			// Verify database.engine is editable with mysql/postgres enum
			var dbEngineField *model.FieldConfiguration
			for i := range ci.Spec.Fields {
				if ci.Spec.Fields[i].Path == "database.engine" {
					dbEngineField = &ci.Spec.Fields[i]
					break
				}
			}
			Expect(dbEngineField).ToNot(BeNil())
			Expect(dbEngineField.Editable).To(BeTrue())
			Expect(dbEngineField.ValidationSchema).To(HaveKey("enum"))
			Expect(dbEngineField.ValidationSchema["enum"]).To(ContainElement("mysql"))
			Expect(dbEngineField.ValidationSchema["enum"]).To(ContainElement("postgres"))

			// Verify database.version is editable with depends_on options per engine
			var dbVersionField *model.FieldConfiguration
			for i := range ci.Spec.Fields {
				if ci.Spec.Fields[i].Path == "database.version" {
					dbVersionField = &ci.Spec.Fields[i]
					break
				}
			}
			Expect(dbVersionField).ToNot(BeNil())
			Expect(dbVersionField.Editable).To(BeTrue())
			Expect(dbVersionField.DependsOn).ToNot(BeNil())
			Expect(dbVersionField.DependsOn.Path).To(Equal("database.engine"))
			Expect(dbVersionField.DependsOn.Mapping).To(HaveKey("postgres"))
			Expect(dbVersionField.DependsOn.Mapping["postgres"]).To(ContainElement("18"))
		})

		It("does not seed when catalog items already exist", func() {
			ctx := context.Background()

			createTestServiceType := func(id, serviceType string) {
				st := model.ServiceType{
					ID:          id,
					ApiVersion:  "v1alpha1",
					ServiceType: serviceType,
					Spec:        map[string]any{},
					Path:        fmt.Sprintf("service-types/%s", id),
				}
				_, err := serviceTypeStore.Create(ctx, st)
				Expect(err).ToNot(HaveOccurred())
			}
			createTestServiceType("three_tier_app_demo", "three_tier_app_demo")
			createTestServiceType("vm-st", "vm")

			// Create an existing catalog item
			ci := model.CatalogItem{
				ID:          "existing-item",
				ApiVersion:  "v1alpha1",
				DisplayName: "Existing",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields:      []model.FieldConfiguration{},
				},
				Path: "catalog-items/existing-item",
			}
			_, err := catalogItemStore.Create(ctx, ci)
			Expect(err).ToNot(HaveOccurred())

			err = catalogItemStore.SeedIfEmpty(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Pet Clinic should NOT have been seeded
			_, err = catalogItemStore.Get(ctx, "pet-clinic")
			Expect(err).To(Equal(store.ErrCatalogItemNotFound))
		})
	})
})
