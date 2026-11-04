package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/sohan-14/practice/internal/billing/app"
)

// Handler handles HTTP requests for the billing module.
type Handler struct {
	service *app.InvoiceService
}

// NewHandler creates a new billing HTTP handler.
func NewHandler(service *app.InvoiceService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the billing routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/invoices", h.createInvoice)
	mux.HandleFunc("GET /api/invoices", h.listInvoices)
	mux.HandleFunc("GET /api/invoices/{id}", h.getInvoice)
	mux.HandleFunc("POST /api/invoices/{id}/issue", h.issueInvoice)
	mux.HandleFunc("POST /api/invoices/{id}/pay", h.payInvoice)
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	var input app.CreateInvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return
	}

	invoice, err := h.service.CreateInvoice(r.Context(), input)
	if err != nil {
		if errors.Is(err, app.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create invoice")
		return
	}

	writeJSON(w, http.StatusCreated, invoice)
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	invoices, err := h.service.ListInvoices(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list invoices")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"invoices": invoices,
		"count":    len(invoices),
	})
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	invoice, err := h.service.GetInvoice(r.Context(), id)
	if err != nil {
		if errors.Is(err, app.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get invoice")
		return
	}

	writeJSON(w, http.StatusOK, invoice)
}

func (h *Handler) issueInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	invoice, err := h.service.IssueInvoice(r.Context(), id)
	if err != nil {
		if errors.Is(err, app.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if errors.Is(err, app.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid_transition", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to issue invoice")
		return
	}

	writeJSON(w, http.StatusOK, invoice)
}

func (h *Handler) payInvoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	invoice, err := h.service.PayInvoice(r.Context(), id)
	if err != nil {
		if errors.Is(err, app.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if errors.Is(err, app.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid_transition", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to pay invoice")
		return
	}

	writeJSON(w, http.StatusOK, invoice)
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}
