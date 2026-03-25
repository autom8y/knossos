package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLiveness_AlwaysOK(t *testing.T) {
	c := NewChecker()
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	c.Liveness(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp livenessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestReadiness_NoChecks_Ready(t *testing.T) {
	c := NewChecker()
	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got %q", resp.Status)
	}
}

func TestReadiness_AllPass(t *testing.T) {
	c := NewChecker()
	c.Register("db", func(_ context.Context) error { return nil })
	c.Register("cache", func(_ context.Context) error { return nil })

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got %q", resp.Status)
	}
	if resp.Checks["db"] != "ok" {
		t.Errorf("expected db check 'ok', got %q", resp.Checks["db"])
	}
	if resp.Checks["cache"] != "ok" {
		t.Errorf("expected cache check 'ok', got %q", resp.Checks["cache"])
	}
}

func TestReadiness_OneFailure_503(t *testing.T) {
	c := NewChecker()
	c.Register("db", func(_ context.Context) error { return nil })
	c.Register("index", func(_ context.Context) error { return fmt.Errorf("index unavailable") })

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not_ready" {
		t.Errorf("expected status 'not_ready', got %q", resp.Status)
	}
	if len(resp.Failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(resp.Failures))
	}
	if resp.Failures[0] != "index" {
		t.Errorf("expected failure 'index', got %q", resp.Failures[0])
	}
	if resp.Checks["index"] != "index unavailable" {
		t.Errorf("expected index check error message, got %q", resp.Checks["index"])
	}
}

func TestReadiness_AllFail_503(t *testing.T) {
	c := NewChecker()
	c.Register("db", func(_ context.Context) error { return fmt.Errorf("connection refused") })
	c.Register("index", func(_ context.Context) error { return fmt.Errorf("not initialized") })

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	var resp readinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not_ready" {
		t.Errorf("expected status 'not_ready', got %q", resp.Status)
	}
	if len(resp.Failures) != 2 {
		t.Errorf("expected 2 failures, got %d", len(resp.Failures))
	}
}

func TestReadiness_ContentType(t *testing.T) {
	c := NewChecker()
	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestLiveness_ContentType(t *testing.T) {
	c := NewChecker()
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	c.Liveness(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestRegister_Overwrite(t *testing.T) {
	c := NewChecker()
	c.Register("db", func(_ context.Context) error { return fmt.Errorf("fail") })
	c.Register("db", func(_ context.Context) error { return nil })

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	c.Readiness(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 after overwrite, got %d", rec.Code)
	}
}
