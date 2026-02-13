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

var _ = Describe("ServiceType Store", func() {
	var (
		db               *gorm.DB
		serviceTypeStore store.ServiceTypeStore
	)

	BeforeEach(func() {
		// Create in-memory SQLite database
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Discard,
		})
		Expect(err).ToNot(HaveOccurred())

		// Auto-migrate
		err = db.AutoMigrate(&model.ServiceType{})
		Expect(err).ToNot(HaveOccurred())

		serviceTypeStore = store.NewServiceTypeStore(db)
	})

	AfterEach(func() {
		sqlDB, err := db.DB()
		Expect(err).ToNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("Create", func() {
		It("should create a new service type successfully", func() {
			st := model.ServiceType{
				ID:          "test-vm",
				ApiVersion:  "v1alpha1",
				ServiceType: "vm",
				Metadata: model.Metadata{
					Labels: map[string]string{"env": "test"},
				},
				Spec: map[string]any{
					"vcpu": map[string]any{"count": 2},
				},
				Path: "service-types/test-vm",
			}

			_, err := serviceTypeStore.Create(context.Background(), st)
			Expect(err).ToNot(HaveOccurred())

			// Verify it was created
			retrieved, err := serviceTypeStore.Get(context.Background(), "test-vm")
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.ID).To(Equal("test-vm"))
			Expect(retrieved.ApiVersion).To(Equal("v1alpha1"))
			Expect(retrieved.ServiceType).To(Equal("vm"))
			Expect(retrieved.Metadata.Labels["env"]).To(Equal("test"))
		})

		It("should return error when creating duplicate ID", func() {
			st := model.ServiceType{
				ID:          "duplicate-id",
				ApiVersion:  "v1alpha1",
				ServiceType: "vm",
				Spec:        map[string]any{},
				Path:        "service-types/duplicate-id",
			}

			_, err := serviceTypeStore.Create(context.Background(), st)
			Expect(err).ToNot(HaveOccurred())

			// Try to create again with same ID
			st2 := model.ServiceType{
				ID:          "duplicate-id",
				ApiVersion:  "v1alpha1",
				ServiceType: "container",
				Spec:        map[string]any{},
				Path:        "service-types/duplicate-id",
			}

			_, err = serviceTypeStore.Create(context.Background(), st2)
			Expect(err).To(Equal(store.ErrServiceTypeIDTaken))
		})
		It("should return error when creating duplicate Service Type", func() {
			st := model.ServiceType{
				ID:          "duplicate-id",
				ApiVersion:  "v1alpha1",
				ServiceType: "vm",
				Spec:        map[string]any{},
				Path:        "service-types/duplicate-id",
			}

			_, err := serviceTypeStore.Create(context.Background(), st)
			Expect(err).ToNot(HaveOccurred())

			// Try to create again with same ID
			st2 := model.ServiceType{
				ID:          "duplicate-service-type",
				ApiVersion:  "v1alpha1",
				ServiceType: "vm",
				Spec:        map[string]any{},
				Path:        "service-types/duplicate-service-type",
			}

			_, err = serviceTypeStore.Create(context.Background(), st2)
			Expect(err).To(Equal(store.ErrServiceTypeServiceTypeTaken))
		})
	})

	Describe("Get", func() {
		It("should retrieve an existing service type", func() {
			st := model.ServiceType{
				ID:          "get-test",
				ApiVersion:  "v1alpha1",
				ServiceType: "database",
				Spec:        map[string]any{"engine": "postgres"},
				Path:        "service-types/get-test",
			}

			_, err := serviceTypeStore.Create(context.Background(), st)
			Expect(err).ToNot(HaveOccurred())

			retrieved, err := serviceTypeStore.Get(context.Background(), "get-test")
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved.ID).To(Equal("get-test"))
			Expect(retrieved.ServiceType).To(Equal("database"))
		})

		It("should return error for non-existent service type", func() {
			_, err := serviceTypeStore.Get(context.Background(), "non-existent")
			Expect(err).To(Equal(store.ErrServiceTypeNotFound))
		})
	})

	Describe("List", func() {
		It("should return empty list when no service types exist", func() {
			results, err := serviceTypeStore.List(context.Background(), &store.ServiceTypeListOptions{
				PageSize: 100,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(results.ServiceTypes).To(BeEmpty())
			Expect(results.NextPageToken).To(Equal(""))
		})

		It("should list all service types", func() {
			// Create multiple service types
			for i := 1; i <= 3; i++ {
				st := model.ServiceType{
					ID:          fmt.Sprintf("st-%d", i),
					ApiVersion:  "v1alpha1",
					ServiceType: fmt.Sprintf("vm-%d", i),
					Spec:        map[string]any{},
					Path:        fmt.Sprintf("service-types/st-%d", i),
				}
				// Add small delay to ensure different create times
				time.Sleep(time.Millisecond)
				_, err := serviceTypeStore.Create(context.Background(), st)
				Expect(err).ToNot(HaveOccurred())
			}

			results, err := serviceTypeStore.List(context.Background(), &store.ServiceTypeListOptions{
				PageSize: 100,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(results.ServiceTypes).To(HaveLen(3))
			Expect(results.NextPageToken).To(Equal(""))
		})

		It("should handle pagination correctly", func() {
			// Create 5 service types
			for i := 1; i <= 5; i++ {
				st := model.ServiceType{
					ID:          fmt.Sprintf("page-st-%d", i),
					ApiVersion:  "v1alpha1",
					ServiceType: fmt.Sprintf("vm-%d", i),
					Spec:        map[string]any{},
					Path:        fmt.Sprintf("service-types/page-st-%d", i),
				}
				time.Sleep(time.Millisecond)
				_, err := serviceTypeStore.Create(context.Background(), st)
				Expect(err).ToNot(HaveOccurred())
			}

			// Get first page with size 2
			results, err := serviceTypeStore.List(context.Background(), &store.ServiceTypeListOptions{
				PageSize: 2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(results.ServiceTypes).To(HaveLen(2))
			Expect(results.NextPageToken).ToNot(Equal(""))

			// Get second page
			results2, err := serviceTypeStore.List(context.Background(), &store.ServiceTypeListOptions{
				PageToken: &results.NextPageToken,
				PageSize:  2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(results2.ServiceTypes).To(HaveLen(2))
			Expect(results2.NextPageToken).ToNot(BeEmpty())

			// Get third page (should have 1 item)
			results3, err := serviceTypeStore.List(context.Background(), &store.ServiceTypeListOptions{
				PageToken: &results2.NextPageToken,
				PageSize:  2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(results3.ServiceTypes).To(HaveLen(1))
			Expect(results3.NextPageToken).To(BeEmpty())
		})
	})
})
