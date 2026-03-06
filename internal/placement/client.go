// Package placement provides a client for the Placement Manager service.
package placement

import (
	"context"
	"fmt"

	pmv1alpha1 "github.com/dcm-project/placement-manager/api/v1alpha1"
	pmclient "github.com/dcm-project/placement-manager/pkg/client"
)

// CreateResourceRequest is the request body for creating a resource in the Placement Manager
type CreateResourceRequest struct {
	CatalogItemInstanceID string         `json:"catalog_item_instance_id"`
	Spec                  map[string]any `json:"spec"`
}

// Resource is the response from the Placement Manager
type Resource struct {
	ID   string         `json:"id"`
	Path string         `json:"path"`
	Spec map[string]any `json:"spec"`
}

// Client defines the interface for interacting with the Placement Manager
type Client interface {
	CreateResource(ctx context.Context, req CreateResourceRequest, id string) (*Resource, error)
	DeleteResource(ctx context.Context, resourceID string) error
}

type client struct {
	pm *pmclient.ClientWithResponses
}

// NewClient creates a new Placement Manager client
func NewClient(baseURL string, opts ...pmclient.ClientOption) (Client, error) {
	pm, err := pmclient.NewClientWithResponses(baseURL+"/api/v1alpha1", opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create placement manager client: %w", err)
	}
	return &client{pm: pm}, nil
}

// CreateResource creates a resource in the Placement Manager
func (c *client) CreateResource(ctx context.Context, req CreateResourceRequest, id string) (*Resource, error) {
	params := &pmv1alpha1.CreateResourceParams{}
	if id != "" {
		params.Id = &id
	}

	body := pmv1alpha1.Resource{
		CatalogItemInstanceId: req.CatalogItemInstanceID,
		Spec:                  req.Spec,
	}

	resp, err := c.pm.CreateResourceWithResponse(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("failed to call placement manager: %w", err)
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("placement manager returned status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	return mapResource(resp.JSON201), nil
}

// DeleteResource deletes a resource from the Placement Manager
func (c *client) DeleteResource(ctx context.Context, resourceID string) error {
	resp, err := c.pm.DeleteResourceWithResponse(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("failed to call placement manager: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return fmt.Errorf("placement manager returned status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	return nil
}

func mapResource(r *pmv1alpha1.Resource) *Resource {
	res := &Resource{
		Spec: r.Spec,
	}
	if r.Id != nil {
		res.ID = *r.Id
	}
	if r.Path != nil {
		res.Path = *r.Path
	}
	return res
}
