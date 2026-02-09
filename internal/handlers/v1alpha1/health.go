package v1alpha1

import (
	"context"
	"fmt"

	"github.com/dcm-project/catalog-manager/internal/api/server"
)

func (h *Handler) GetHealth(ctx context.Context, request server.GetHealthRequestObject) (server.GetHealthResponseObject, error) {
	status := "healthy"
	path := fmt.Sprintf("%shealth", apiPrefix)
	return server.GetHealth200JSONResponse{
		Status: status,
		Path:   &path,
	}, nil
}
