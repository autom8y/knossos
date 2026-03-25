package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---- Group G: Health Checks (C6) ----

func TestIntegration_AllHealthChecksPass(t *testing.T) {
	// PT-09 C6: Register all 5 production health checks, all returning nil -> 200.
	c := NewChecker()
	c.Register("slack", func(_ context.Context) error { return nil })
	c.Register("reasoning", func(_ context.Context) error { return nil })
	c.Register("catalog", func(_ context.Context) error { return nil })
	c.Register("search_index", func(_ context.Context) error { return nil })
	c.Register("claude_api", func(_ context.Context) error { return nil })

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("PT-09 C6: expected status 200 when all checks pass, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got %q", resp.Status)
	}
	if len(resp.Failures) != 0 {
		t.Errorf("expected no failures, got %v", resp.Failures)
	}

	// Verify all 5 checks are reported.
	expectedChecks := []string{"slack", "reasoning", "catalog", "search_index", "claude_api"}
	for _, name := range expectedChecks {
		val, exists := resp.Checks[name]
		if !exists {
			t.Errorf("check %q missing from response", name)
			continue
		}
		if val != "ok" {
			t.Errorf("check %q expected 'ok', got %q", name, val)
		}
	}
}

func TestIntegration_HealthCheckFailure_Catalog(t *testing.T) {
	// PT-09 C6: catalog check fails -> 503.
	c := NewChecker()
	c.Register("slack", func(_ context.Context) error { return nil })
	c.Register("reasoning", func(_ context.Context) error { return nil })
	c.Register("catalog", func(_ context.Context) error {
		return fmt.Errorf("catalog sync failed: no repos found")
	})
	c.Register("search_index", func(_ context.Context) error { return nil })
	c.Register("claude_api", func(_ context.Context) error { return nil })

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 when catalog fails, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not_ready" {
		t.Errorf("expected status 'not_ready', got %q", resp.Status)
	}

	// Verify catalog is in failures list.
	found := false
	for _, f := range resp.Failures {
		if f == "catalog" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'catalog' in failures, got %v", resp.Failures)
	}

	// Verify catalog check reports the error.
	if resp.Checks["catalog"] != "catalog sync failed: no repos found" {
		t.Errorf("expected catalog error message, got %q", resp.Checks["catalog"])
	}

	// Verify passing checks still report OK.
	for _, name := range []string{"slack", "reasoning", "search_index", "claude_api"} {
		if resp.Checks[name] != "ok" {
			t.Errorf("passing check %q expected 'ok', got %q", name, resp.Checks[name])
		}
	}
}

func TestIntegration_HealthCheckFailure_ClaudeAPI(t *testing.T) {
	// PT-09 C6: claude_api check fails -> 503.
	c := NewChecker()
	c.Register("slack", func(_ context.Context) error { return nil })
	c.Register("reasoning", func(_ context.Context) error { return nil })
	c.Register("catalog", func(_ context.Context) error { return nil })
	c.Register("search_index", func(_ context.Context) error { return nil })
	c.Register("claude_api", func(_ context.Context) error {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	})

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 when claude_api fails, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not_ready" {
		t.Errorf("expected status 'not_ready', got %q", resp.Status)
	}

	// Verify claude_api is in failures.
	found := false
	for _, f := range resp.Failures {
		if f == "claude_api" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'claude_api' in failures, got %v", resp.Failures)
	}

	// Verify claude_api check reports the error.
	if resp.Checks["claude_api"] != "ANTHROPIC_API_KEY not set" {
		t.Errorf("expected claude_api error message, got %q", resp.Checks["claude_api"])
	}
}
