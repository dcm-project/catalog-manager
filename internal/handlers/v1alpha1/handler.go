package v1alpha1

import (
	"github.com/dcm-project/catalog-manager/internal/api/server"
)

const (
	apiPrefix = "/api/v1alpha1/"
)

type Handler struct {
	// Future: storage layer will be injected here
}

func NewHandler() *Handler {
	return &Handler{}
}

// Compile-time verification
var _ server.StrictServerInterface = (*Handler)(nil)
