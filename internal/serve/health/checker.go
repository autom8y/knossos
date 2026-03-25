// Package health provides health check endpoints for the HTTP server.
package health

import (
	"context"
	"sync"
)

// CheckFunc is a function that performs a health check.
// It returns nil if healthy, or an error describing the failure.
type CheckFunc func(ctx context.Context) error

// Checker manages named health checks and exposes liveness/readiness endpoints.
type Checker struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// NewChecker creates a new Checker with no registered checks.
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]CheckFunc),
	}
}

// Register adds a named health check. Overwrites any existing check with the same name.
func (c *Checker) Register(name string, check CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// runChecks executes all registered checks with the given context.
// Returns a map of check names to results (nil for pass, error message for fail)
// and a slice of failure names.
func (c *Checker) runChecks(ctx context.Context) (map[string]string, []string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]string, len(c.checks))
	var failures []string

	for name, check := range c.checks {
		if err := check(ctx); err != nil {
			results[name] = err.Error()
			failures = append(failures, name)
		} else {
			results[name] = "ok"
		}
	}

	return results, failures
}
