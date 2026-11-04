package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// contextKey is a custom type to avoid collisions with other packages.
type contextKey string

const requestIDKey contextKey = "request_id"

// NewRequestID generates a unique request ID using crypto/rand.
func NewRequestID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// ContextWithRequestID returns a new context with the request ID attached.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
