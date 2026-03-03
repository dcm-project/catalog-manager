package service_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dcm-project/catalog-manager/internal/service"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

var _ = Describe("Seed", func() {
	var (
		db        *gorm.DB
		dataStore store.Store
		svc       service.Service
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

		dataStore = store.NewStore(db)
		svc = service.NewService(dataStore, nil)
	})

	AfterEach(func() {
		sqlDB, err := db.DB()
		Expect(err).ToNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("Seed", func() {
		Describe("Pet Clinic", func() {
			It("is idempotent when called multiple times", func() {
				ctx := context.Background()

				err := svc.Seed(ctx)
				Expect(err).ToNot(HaveOccurred())

				err = svc.Seed(ctx)
				Expect(err).ToNot(HaveOccurred())

				var count int64
				err = db.Model(&model.CatalogItem{}).Where("id = ?", "pet-clinic").Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(int64(1)))
			})

			It("seeds when table is empty", func() {
				ctx := context.Background()

				err := svc.Seed(ctx)
				Expect(err).ToNot(HaveOccurred())

				ci, err := dataStore.CatalogItem().Get(ctx, "pet-clinic")
				Expect(err).ToNot(HaveOccurred())
				Expect(ci.ID).To(Equal("pet-clinic"))
				Expect(ci.DisplayName).To(Equal("Pet Clinic"))
				Expect(ci.Path).To(Equal("catalog-items/pet-clinic"))
				Expect(ci.Spec.ServiceType).To(Equal("three_tier_app_demo"))
				Expect(ci.Spec.Fields).To(HaveLen(3))

				// Verify key field configs
				fieldPaths := make([]string, len(ci.Spec.Fields))
				for i, f := range ci.Spec.Fields {
					fieldPaths[i] = f.Path
				}
				Expect(fieldPaths).To(ContainElement("database.image"))
				Expect(fieldPaths).To(ContainElement("app.image"))
				Expect(fieldPaths).To(ContainElement("web.image"))

				// Verify database.image is editable
				dbImageField := findFieldByPath(ci.Spec.Fields, "database.image")
				Expect(dbImageField).ToNot(BeNil())
				Expect(dbImageField.Editable).To(BeTrue())
				Expect(dbImageField.Default).To(Equal("quay.io/myorg/postgres:15"))

				// Verify app.image and web.image fixed defaults
				appImageField := findFieldByPath(ci.Spec.Fields, "app.image")
				Expect(appImageField).ToNot(BeNil())
				Expect(appImageField.Default).To(Equal("docker.io/springcommunity/spring-framework-petclinic:6.1.2"))
				Expect(appImageField.Editable).To(BeFalse())

				webImageField := findFieldByPath(ci.Spec.Fields, "web.image")
				Expect(webImageField).ToNot(BeNil())
				Expect(webImageField.Default).To(Equal("docker.io/library/nginx:alpine"))
				Expect(webImageField.Editable).To(BeFalse())
			})
		})

		Describe("when catalog items exist", func() {
			It("does not seed", func() {
				ctx := context.Background()

				createTestServiceType := func(id, serviceType string) {
					st := model.ServiceType{
						ID:          id,
						ApiVersion:  "v1alpha1",
						ServiceType: serviceType,
						Spec:        map[string]any{},
						Path:        fmt.Sprintf("service-types/%s", id),
					}
					_, err := dataStore.ServiceType().Create(ctx, st)
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
				_, err := dataStore.CatalogItem().Create(ctx, ci)
				Expect(err).ToNot(HaveOccurred())

				err = svc.Seed(ctx)
				Expect(err).ToNot(HaveOccurred())

				// Default seed items should NOT have been added
				_, err = dataStore.CatalogItem().Get(ctx, "pet-clinic")
				Expect(err).To(Equal(store.ErrCatalogItemNotFound))
			})
		})
	})
})

func findFieldByPath(fields []model.FieldConfiguration, path string) *model.FieldConfiguration {
	for i := range fields {
		if fields[i].Path == path {
			return &fields[i]
		}
	}
	return nil
}
