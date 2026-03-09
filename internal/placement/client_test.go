package placement_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pmclient "github.com/dcm-project/placement-manager/pkg/client"

	"github.com/dcm-project/catalog-manager/internal/placement"
)

func newTestClient(serverURL string) placement.Client {
	client, err := placement.NewClient(serverURL, pmclient.WithHTTPClient(http.DefaultClient))
	Expect(err).ToNot(HaveOccurred())
	return client
}

var _ = Describe("Placement Client", func() {
	var (
		ctx    context.Context
		server *httptest.Server
		client placement.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("CreateResource", func() {
		Context("when the server returns success", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal(http.MethodPost))
					Expect(r.URL.Query().Get("id")).To(Equal("my-resource"))

					var body map[string]any
					err := json.NewDecoder(r.Body).Decode(&body)
					Expect(err).ToNot(HaveOccurred())
					Expect(body["catalog_item_instance_id"]).To(Equal("instance-123"))

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":                       "pm-resource-id",
						"path":                     "resources/pm-resource-id",
						"catalog_item_instance_id": "instance-123",
						"spec":                     map[string]any{"vcpu": map[string]any{"count": float64(4)}},
					})
				}))
				client = newTestClient(server.URL)
			})

			It("returns the created resource", func() {
				resource, err := client.CreateResource(ctx, placement.CreateResourceRequest{
					CatalogItemInstanceID: "instance-123",
					Spec:                  map[string]any{"vcpu": map[string]any{"count": float64(4)}},
				}, "my-resource")

				Expect(err).ToNot(HaveOccurred())
				Expect(resource.ID).To(Equal("pm-resource-id"))
				Expect(resource.Path).To(Equal("resources/pm-resource-id"))
			})
		})

		Context("when no ID is provided", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Query().Get("id")).To(BeEmpty())

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"id":                       "auto-generated-id",
						"path":                     "resources/auto-generated-id",
						"catalog_item_instance_id": "instance-456",
						"spec":                     map[string]any{},
					})
				}))
				client = newTestClient(server.URL)
			})

			It("sends no id query param and returns the resource", func() {
				resource, err := client.CreateResource(ctx, placement.CreateResourceRequest{
					CatalogItemInstanceID: "instance-456",
					Spec:                  map[string]any{},
				}, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(resource.ID).To(Equal("auto-generated-id"))
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"title": "internal error", "type": "internal"}`))
				}))
				client = newTestClient(server.URL)
			})

			It("returns an error and nil resource", func() {
				resource, err := client.CreateResource(ctx, placement.CreateResourceRequest{
					CatalogItemInstanceID: "instance-789",
					Spec:                  map[string]any{},
				}, "")

				Expect(err).To(HaveOccurred())
				Expect(resource).To(BeNil())
			})
		})
	})

	Describe("DeleteResource", func() {
		Context("when the server returns success", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal(http.MethodDelete))
					w.WriteHeader(http.StatusNoContent)
				}))
				client = newTestClient(server.URL)
			})

			It("succeeds without error", func() {
				err := client.DeleteResource(ctx, "pm-resource-id")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the resource is not found", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"title": "not found", "type": "not_found"}`))
				}))
				client = newTestClient(server.URL)
			})

			It("returns an error", func() {
				err := client.DeleteResource(ctx, "nonexistent")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"title": "internal error", "type": "internal"}`))
				}))
				client = newTestClient(server.URL)
			})

			It("returns an error", func() {
				err := client.DeleteResource(ctx, "some-id")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
