package observability

import (
	"log/slog"
	"os"
)

// NewLogger creates a structured logger based on the environment.
func NewLogger(environment string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if environment == "production" {
		// JSON output for production (machine-readable)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text output for development (human-readable)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
