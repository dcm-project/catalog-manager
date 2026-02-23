package v1alpha1_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dcm-project/catalog-manager/internal/api/server"
	v1alpha1 "github.com/dcm-project/catalog-manager/internal/handlers/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

var _ = Describe("Health Handler", func() {
	var (
		handler   *v1alpha1.Handler
		db        *gorm.DB
		dataStore store.Store
	)

	BeforeEach(func() {
		// Create in-memory SQLite database for testing
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Discard,
		})
		Expect(err).ToNot(HaveOccurred())

		// Auto-migrate
		err = db.AutoMigrate(
			&model.ServiceType{},
			&model.CatalogItem{},
			&model.CatalogItemInstance{},
		)
		Expect(err).ToNot(HaveOccurred())

		dataStore = store.NewStore(db)
		handler = v1alpha1.NewHandler(dataStore)
	})

	AfterEach(func() {
		dataStore.Close()
	})

	Describe("GetHealth", func() {
		It("should return healthy status", func() {
			request := server.GetHealthRequestObject{}
			response, err := handler.GetHealth(context.Background(), request)

			Expect(err).ToNot(HaveOccurred())
			Expect(response).To(BeAssignableToTypeOf(server.GetHealth200JSONResponse{}))

			healthResponse := response.(server.GetHealth200JSONResponse)
			Expect(healthResponse.Status).To(Equal("healthy"))
			Expect(healthResponse.Path).ToNot(BeNil())
			Expect(*healthResponse.Path).To(Equal("/api/v1alpha1/health"))
		})
	})
})
