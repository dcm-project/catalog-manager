package v1alpha1

import (
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/service"
)

const (
	apiPrefix = "/api/v1alpha1/"
)

type Handler struct {
	service service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// Compile-time verification
var _ server.StrictServerInterface = (*Handler)(nil)

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
