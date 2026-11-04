package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sohan-14/practice/internal/billing"
	"github.com/sohan-14/practice/internal/catalog"
	"github.com/sohan-14/practice/internal/platform/config"
	"github.com/sohan-14/practice/internal/platform/httpx"
	"github.com/sohan-14/practice/internal/platform/observability"
	"github.com/sohan-14/practice/internal/shared/middleware"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Initialize structured logger
	logger := observability.NewLogger(cfg.Environment)

	// 3. Create the rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitCapacity, cfg.RateLimitRefillRate)
	rateLimiter.StartCleanup(5*time.Minute, 1*time.Hour)

	// 4. Create the mux
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Register modules
	catalogModule := catalog.NewModule()
	catalogModule.RegisterRoutes(mux)

	billingModule := billing.NewModule(logger)
	billingModule.RegisterRoutes(mux)

	// 5. Apply middleware chain
	handler := middleware.Chain(
		mux,
		middleware.Recovery(logger),
		middleware.RateLimit(rateLimiter, logger),
		middleware.Logging(logger),
	)

	// 6. Create HTTP server
	server := httpx.NewServer(cfg, handler, logger)

	// 7. Start server
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", cfg.Addr)

		var err error
		if cfg.TLSCert != "" && cfg.TLSKey != "" {
			err = server.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// 8. Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	case sig := <-sigChan:
		logger.Info("received shutdown signal", "signal", sig)
	}

	// 9. Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("graceful shutdown complete")
}
