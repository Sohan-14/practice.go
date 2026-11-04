package catalog

import (
	"net/http"

	"github.com/sohan-14/practice/internal/catalog/api"
	"github.com/sohan-14/practice/internal/catalog/app"
	"github.com/sohan-14/practice/internal/catalog/store"
)

// Module is the public surface of the catalog module.
// Other modules interact with catalog ONLY through this interface.
type Module struct {
	handler *api.Handler
}

// NewModule creates and wires the catalog module.
func NewModule() *Module {
	// Wire dependencies (dependency injection)
	repo := store.NewMemoryRepository()
	service := app.NewProductService(repo)
	handler := api.NewHandler(service)

	return &Module{
		handler: handler,
	}
}

// RegisterRoutes registers the catalog routes on the given mux.
func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}
