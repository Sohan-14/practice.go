package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// TLS
	TLSCert string
	TLSKey  string

	// Environment
	Environment string

	// Rate Limiting
	RateLimitCapacity   float64
	RateLimitRefillRate float64
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		// Server
		Addr:            getEnv("SERVER_ADDR", ":8080"),
		ReadTimeout:     getDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getDuration("WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),

		// TLS
		TLSCert: getEnv("TLS_CERT", ""),
		TLSKey:  getEnv("TLS_KEY", ""),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),

		// Rate Limiting
		RateLimitCapacity:   getFloat("RATE_LIMIT_CAPACITY", 10.0),
		RateLimitRefillRate: getFloat("RATE_LIMIT_REFILL_RATE", 2.0),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
