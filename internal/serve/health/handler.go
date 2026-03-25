package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const readinessTimeout = 5 * time.Second

// livenessResponse is the JSON body for /health.
type livenessResponse struct {
	Status string `json:"status"`
}

// readinessResponse is the JSON body for /ready.
type readinessResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks"`
	Failures []string          `json:"failures,omitempty"`
}

// Liveness handles GET /health. Always returns 200 if the process is running.
func (c *Checker) Liveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(livenessResponse{Status: "ok"})
}

// Readiness handles GET /ready. Runs all registered checks with a 5s timeout.
// Returns 200 if all pass, 503 if any fail. No registered checks = ready.
func (c *Checker) Readiness(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), readinessTimeout)
	defer cancel()

	results, failures := c.runChecks(ctx)

	w.Header().Set("Content-Type", "application/json")

	resp := readinessResponse{
		Checks: results,
	}

	if len(failures) > 0 {
		resp.Status = "not_ready"
		resp.Failures = failures
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		resp.Status = "ready"
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(resp)
}
