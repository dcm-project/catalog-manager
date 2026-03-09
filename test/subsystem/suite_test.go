//go:build subsystem
package subsystem_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dcm-project/catalog-manager/pkg/client"
)

var (
	apiClient      *client.ClientWithResponses
	wireMockURL    string
	httpClient     = &http.Client{Timeout: 10 * time.Second}
)

func TestSubsystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Subsystem Suite")
}

var _ = BeforeSuite(func() {
	serviceURL := envOrDefault("CATALOG_MANAGER_URL", "http://localhost:28080")
	wireMockURL = envOrDefault("PLACEMENT_MANAGER_URL", "http://localhost:28081")

	var err error
	apiClient, err = client.NewClientWithResponses(serviceURL + "/api/v1alpha1")
	Expect(err).NotTo(HaveOccurred())

	// Wait for the service to be healthy
	Eventually(func() error {
		resp, err := apiClient.GetHealthWithResponse(context.Background())
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("health check returned %d", resp.StatusCode())
		}
		return nil
	}).WithTimeout(60 * time.Second).WithPolling(2 * time.Second).Should(Succeed())
})

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
