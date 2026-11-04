package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/sohan-14/practice/internal/catalog/app"
)

// Handler handles HTTP requests for the catalog module.
type Handler struct {
	service *app.ProductService
}

// NewHandler creates a new catalog HTTP handler.
func NewHandler(service *app.ProductService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the catalog routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/products", h.createProduct)
	mux.HandleFunc("GET /api/products", h.listProducts)
	mux.HandleFunc("GET /api/products/{id}", h.getProduct)
	mux.HandleFunc("PUT /api/products/{id}", h.updateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", h.deleteProduct)
}

// createProduct handles POST /api/products
func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var input app.CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), input)
	if err != nil {
		if errors.Is(err, app.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create product")
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

// listProducts handles GET /api/products
func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	products, err := h.service.ListProducts(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list products")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"products": products,
		"count":    len(products),
	})
}

// getProduct handles GET /api/products/{id}
func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, app.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get product")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

// updateProduct handles PUT /api/products/{id}
func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input app.UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body is not valid JSON")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, app.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if errors.Is(err, app.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update product")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

// deleteProduct handles DELETE /api/products/{id}
func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		if errors.Is(err, app.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
