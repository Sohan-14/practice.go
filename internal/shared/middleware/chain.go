package middleware

import (
	"net/http"
)

// Chain applies multiple middleware to a handler in order.
// The first middleware in the slice is the outermost layer.
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply in reverse order so the first middleware is outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
