package httpx

import (
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/sohan-14/practice/internal/platform/config"
)

// NewServer creates an HTTP server with production-grade settings.
func NewServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) *http.Server {
	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Add TLS config if certificates are provided
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	return server
}
