package billing

import (
	"log/slog"
	"net/http"

	"github.com/sohan-14/practice/internal/billing/api"
	"github.com/sohan-14/practice/internal/billing/app"
	"github.com/sohan-14/practice/internal/billing/store"
)

// Module is the public surface of the billing module.
type Module struct {
	handler *api.Handler
	service *app.InvoiceService
}

// NewModule creates and wires the billing module.
func NewModule(logger *slog.Logger) *Module {
	// Wire dependencies
	repo := store.NewMemoryRepository()
	service := app.NewInvoiceService(repo, logger)
	handler := api.NewHandler(service)

	return &Module{
		handler: handler,
		service: service,
	}
}

// RegisterRoutes registers the billing routes on the given mux.
func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}
