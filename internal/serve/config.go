// Package serve provides the HTTP server for ari serve.
package serve

import "time"

// ServerConfig holds configuration for the HTTP server.
type ServerConfig struct {
	Port         int
	DrainTimeout time.Duration // default 30s
	ReadTimeout  time.Duration // default 5s
	WriteTimeout time.Duration // default 30s
	IdleTimeout  time.Duration // default 120s
}

// DefaultConfig returns a ServerConfig with production defaults.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Port:         8080,
		DrainTimeout: 30 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}
