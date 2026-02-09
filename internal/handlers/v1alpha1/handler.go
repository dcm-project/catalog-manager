package v1alpha1

import (
	"github.com/dcm-project/catalog-manager/internal/api/server"
	"github.com/dcm-project/catalog-manager/internal/store"
)

const (
	apiPrefix = "/api/v1alpha1/"
)

type Handler struct {
	store store.Store
}

func NewHandler(store store.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// Compile-time verification
var _ server.StrictServerInterface = (*Handler)(nil)
