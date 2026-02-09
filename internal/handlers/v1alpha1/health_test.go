package v1alpha1_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dcm-project/catalog-manager/internal/api/server"
	v1alpha1 "github.com/dcm-project/catalog-manager/internal/handlers/v1alpha1"
)

var _ = Describe("Health Handler", func() {
	var handler *v1alpha1.Handler

	BeforeEach(func() {
		handler = v1alpha1.NewHandler()
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
