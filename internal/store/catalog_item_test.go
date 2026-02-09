package store_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

var _ = Describe("CatalogItem Store", func() {
	var (
		db               *gorm.DB
		catalogItemStore store.CatalogItemStore
	)

	BeforeEach(func() {
		// Create in-memory SQLite database
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Discard,
		})
		Expect(err).ToNot(HaveOccurred())

		// Auto-migrate
		err = db.AutoMigrate(&model.CatalogItem{})
		Expect(err).ToNot(HaveOccurred())

		catalogItemStore = store.NewCatalogItemStore(db)
	})

	AfterEach(func() {
		sqlDB, err := db.DB()
		Expect(err).ToNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("Create", func() {
		It("should create a new catalog item successfully", func() {
			ci := &model.CatalogItem{
				ID:          "small-vm",
				ApiVersion:  "v1alpha1",
				DisplayName: "Small VM",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields: []model.FieldConfiguration{
						{
							Path:     "spec.vcpu.count",
							Editable: false,
							Default:  2,
						},
					},
				},
				Path: "catalog-items/small-vm",
			}

			_, err := catalogItemStore.Create(context.Background(), *ci)
			Expect(err).ToNot(HaveOccurred())

			// Verify it was created
			retrieved, err := catalogItemStore.Get(context.Background(), "small-vm")
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.ID).To(Equal("small-vm"))
			Expect(retrieved.DisplayName).To(Equal("Small VM"))
			Expect(retrieved.Spec.ServiceType).To(Equal("vm"))
			Expect(retrieved.SpecServiceType).To(Equal("vm"))
		})

		It("should return error when creating duplicate ID", func() {
			ci := &model.CatalogItem{
				ID:          "duplicate-ci",
				ApiVersion:  "v1alpha1",
				DisplayName: "Original",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields:      []model.FieldConfiguration{},
				},
				Path: "catalog-items/duplicate-ci",
			}

			_, err := catalogItemStore.Create(context.Background(), *ci)
			Expect(err).ToNot(HaveOccurred())

			// Try to create again with same ID
			ci2 := model.CatalogItem{
				ID:          "duplicate-ci",
				ApiVersion:  "v1alpha1",
				DisplayName: "Duplicate",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields:      []model.FieldConfiguration{},
				},
				Path: "catalog-items/duplicate-ci",
			}

			_, err = catalogItemStore.Create(context.Background(), ci2)
			Expect(err).To(Equal(store.ErrCatalogItemIDTaken))
		})
	})

	Describe("Get", func() {
		It("should retrieve an existing catalog item", func() {
			ci := &model.CatalogItem{
				ID:          "get-test-ci",
				ApiVersion:  "v1alpha1",
				DisplayName: "Test Item",
				Spec: model.CatalogItemSpec{
					ServiceType: "database",
					Fields: []model.FieldConfiguration{
						{Path: "spec.engine", Default: "postgres"},
					},
				},
				Path: "catalog-items/get-test-ci",
			}

			_, err := catalogItemStore.Create(context.Background(), *ci)
			Expect(err).ToNot(HaveOccurred())

			retrieved, err := catalogItemStore.Get(context.Background(), "get-test-ci")
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.ID).To(Equal("get-test-ci"))
			Expect(retrieved.Spec.ServiceType).To(Equal("database"))
		})

		It("should return error for non-existent catalog item", func() {
			_, err := catalogItemStore.Get(context.Background(), "non-existent")
			Expect(err).To(Equal(store.ErrCatalogItemNotFound))
		})
	})

	Describe("Update", func() {
		It("should update mutable fields successfully", func() {
			ci := &model.CatalogItem{
				ID:          "update-test",
				ApiVersion:  "v1alpha1",
				DisplayName: "Original Name",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields: []model.FieldConfiguration{
						{Path: "spec.vcpu.count", Default: 2},
					},
				},
				Path: "catalog-items/update-test",
			}

			created, err := catalogItemStore.Create(context.Background(), *ci)
			Expect(err).ToNot(HaveOccurred())
			ci = created

			// Update mutable fields
			ci.DisplayName = "Updated Name"
			ci.Spec.Fields = append(ci.Spec.Fields, model.FieldConfiguration{
				Path:    "spec.memory.size_gb",
				Default: 8,
			})

			err = catalogItemStore.Update(context.Background(), ci)
			Expect(err).ToNot(HaveOccurred())

			// Verify update
			retrieved, err := catalogItemStore.Get(context.Background(), "update-test")
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.DisplayName).To(Equal("Updated Name"))
			Expect(retrieved.Spec.Fields).To(HaveLen(2))
		})

		It("should return error when updating non-existent catalog item", func() {
			ci := &model.CatalogItem{
				ID:          "non-existent",
				DisplayName: "Updated",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields:      []model.FieldConfiguration{},
				},
			}

			err := catalogItemStore.Update(context.Background(), ci)
			Expect(err).To(Equal(store.ErrCatalogItemNotFound))
		})
	})

	Describe("Delete", func() {
		It("should delete an existing catalog item", func() {
			ci := &model.CatalogItem{
				ID:          "delete-test",
				ApiVersion:  "v1alpha1",
				DisplayName: "To Delete",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields:      []model.FieldConfiguration{},
				},
				Path: "catalog-items/delete-test",
			}

			_, err := catalogItemStore.Create(context.Background(), *ci)
			Expect(err).ToNot(HaveOccurred())

			err = catalogItemStore.Delete(context.Background(), "delete-test")
			Expect(err).ToNot(HaveOccurred())

			// Verify deletion
			_, err = catalogItemStore.Get(context.Background(), "delete-test")
			Expect(err).To(Equal(store.ErrCatalogItemNotFound))
		})

		It("should return error when deleting non-existent catalog item", func() {
			err := catalogItemStore.Delete(context.Background(), "non-existent")
			Expect(err).To(Equal(store.ErrCatalogItemNotFound))
		})
	})

	Describe("List", func() {
		It("should return empty list when no catalog items exist", func() {
			result, err := catalogItemStore.List(context.Background(), &store.CatalogItemListOptions{PageSize: 100})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.CatalogItems).To(BeEmpty())
			Expect(result.NextPageToken).To(Equal(""))
		})

		It("should list all catalog items", func() {
			// Create multiple catalog items
			for i := 1; i <= 3; i++ {
				ci := model.CatalogItem{
					ID:          fmt.Sprintf("ci-%d", i),
					ApiVersion:  "v1alpha1",
					DisplayName: fmt.Sprintf("Item %d", i),
					Spec: model.CatalogItemSpec{
						ServiceType: "vm",
						Fields:      []model.FieldConfiguration{},
					},
					Path: fmt.Sprintf("catalog-items/ci-%d", i),
				}
				time.Sleep(time.Millisecond)
				_, err := catalogItemStore.Create(context.Background(), ci)
				Expect(err).ToNot(HaveOccurred())
			}

			result, err := catalogItemStore.List(context.Background(), &store.CatalogItemListOptions{PageSize: 100})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.CatalogItems).To(HaveLen(3))
			Expect(result.NextPageToken).To(Equal(""))
		})

		It("should filter by service type", func() {
			// Create catalog items with different service types
			ci1 := model.CatalogItem{
				ID:          "vm-item",
				ApiVersion:  "v1alpha1",
				DisplayName: "VM Item",
				Spec: model.CatalogItemSpec{
					ServiceType: "vm",
					Fields:      []model.FieldConfiguration{},
				},
				Path: "catalog-items/vm-item",
			}
			_, err := catalogItemStore.Create(context.Background(), ci1)
			Expect(err).ToNot(HaveOccurred())

			ci2 := model.CatalogItem{
				ID:          "db-item",
				ApiVersion:  "v1alpha1",
				DisplayName: "DB Item",
				Spec: model.CatalogItemSpec{
					ServiceType: "database",
					Fields:      []model.FieldConfiguration{},
				},
				Path: "catalog-items/db-item",
			}
			_, err = catalogItemStore.Create(context.Background(), ci2)
			Expect(err).ToNot(HaveOccurred())

			// Filter for vm service type
			result, err := catalogItemStore.List(context.Background(), &store.CatalogItemListOptions{PageSize: 100, ServiceType: "vm"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.CatalogItems).To(HaveLen(1))
			Expect(result.CatalogItems[0].Spec.ServiceType).To(Equal("vm"))

			// Filter for database service type
			result, err = catalogItemStore.List(context.Background(), &store.CatalogItemListOptions{PageSize: 100, ServiceType: "database"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.CatalogItems).To(HaveLen(1))
			Expect(result.CatalogItems[0].Spec.ServiceType).To(Equal("database"))
		})

		It("should handle pagination correctly", func() {
			// Create 5 catalog items
			for i := 1; i <= 5; i++ {
				ci := model.CatalogItem{
					ID:          fmt.Sprintf("page-ci-%d", i),
					ApiVersion:  "v1alpha1",
					DisplayName: fmt.Sprintf("Item %d", i),
					Spec: model.CatalogItemSpec{
						ServiceType: "vm",
						Fields:      []model.FieldConfiguration{},
					},
					Path: fmt.Sprintf("catalog-items/page-ci-%d", i),
				}
				time.Sleep(time.Millisecond)
				_, err := catalogItemStore.Create(context.Background(), ci)
				Expect(err).ToNot(HaveOccurred())
			}

			// Get first page
			result, err := catalogItemStore.List(context.Background(), &store.CatalogItemListOptions{PageSize: 2})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.CatalogItems).To(HaveLen(2))
			Expect(result.NextPageToken).ToNot(Equal(""))

			// Get second page
			result2, err := catalogItemStore.List(context.Background(), &store.CatalogItemListOptions{
				PageToken: &result.NextPageToken,
				PageSize:  2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(result2.CatalogItems).To(HaveLen(2))
			Expect(result2.NextPageToken).ToNot(Equal(""))
		})
	})
})
