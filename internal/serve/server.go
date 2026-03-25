package serve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/autom8y/knossos/internal/serve/health"
)

// Option configures a Server.
type Option func(*Server)

// WithHealthChecker sets the health checker for the server.
func WithHealthChecker(checker *health.Checker) Option {
	return func(s *Server) {
		s.health = checker
	}
}

// Server is the HTTP server for ari serve.
type Server struct {
	httpServer  *http.Server
	mux         *http.ServeMux
	config      ServerConfig
	health      *health.Checker
	middlewares []Middleware
}

// New creates a new Server with the given config and options.
func New(cfg ServerConfig, opts ...Option) *Server {
	mux := http.NewServeMux()
	s := &Server{
		mux:    mux,
		config: cfg,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.health == nil {
		s.health = health.NewChecker()
	}

	return s
}

// RegisterHandler registers an HTTP handler on the server's mux.
// Pattern follows Go 1.22+ enhanced ServeMux format: "METHOD /path".
func (s *Server) RegisterHandler(method, pattern string, handler http.Handler) {
	fullPattern := fmt.Sprintf("%s %s", method, pattern)
	s.mux.Handle(fullPattern, handler)
}

// Use appends middlewares to the server's middleware chain.
func (s *Server) Use(mw ...Middleware) {
	s.middlewares = append(s.middlewares, mw...)
}

// Health returns the server's health checker for registering checks.
func (s *Server) Health() *health.Checker {
	return s.health
}

// Start registers routes, applies middleware, and starts the HTTP server.
// It blocks until a shutdown signal (SIGTERM/SIGINT) is received or the
// context is cancelled, then performs graceful shutdown with the configured
// drain timeout.
func (s *Server) Start(ctx context.Context) error {
	// Register health endpoints
	s.mux.HandleFunc("GET /health", s.health.Liveness)
	s.mux.HandleFunc("GET /ready", s.health.Readiness)

	// Build handler chain: built-in middlewares first, then user middlewares
	builtIn := []Middleware{
		PanicRecovery(),
		RequestID(),
		RequestLogger(),
	}
	allMiddleware := append(builtIn, s.middlewares...)
	s.httpServer.Handler = Chain(allMiddleware...)(s.mux)

	// Listen for OS signals
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", s.config.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.DrainTimeout)
	defer cancel()
	return s.httpServer.Shutdown(shutdownCtx)
}

// Shutdown initiates graceful shutdown of the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
