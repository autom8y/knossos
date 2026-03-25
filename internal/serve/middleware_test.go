package serve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPanicRecovery(t *testing.T) {
	handler := PanicRecovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestPanicRecovery_NoPanic(t *testing.T) {
	handler := PanicRecovery()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRequestID(t *testing.T) {
	var capturedID string
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check header
	headerID := rec.Header().Get("X-Request-Id")
	if headerID == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}

	// Check context value matches header
	if capturedID != headerID {
		t.Errorf("context ID %q does not match header ID %q", capturedID, headerID)
	}

	// Check UUID format (8-4-4-4-12 hex)
	parts := strings.Split(headerID, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 UUID parts, got %d: %s", len(parts), headerID)
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("invalid UUID format: %s", headerID)
	}
}

func TestRequestID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		id := rec.Header().Get("X-Request-Id")
		if ids[id] {
			t.Fatalf("duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	id := RequestIDFromContext(context.Background())
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func TestRequestLogger(t *testing.T) {
	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestChain(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}

	handler := Chain(mw1, mw2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("call %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestConcurrencyLimit_AllowsUnderLimit(t *testing.T) {
	handler := ConcurrencyLimit(2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestConcurrencyLimit_RejectsAtCapacity(t *testing.T) {
	// Create a limiter with capacity 1.
	blocked := make(chan struct{})
	proceed := make(chan struct{})

	handler := ConcurrencyLimit(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(blocked)   // Signal that the handler is executing.
		<-proceed        // Wait until told to finish.
		w.WriteHeader(http.StatusOK)
	}))

	// Start the first request (will block inside handler).
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest("GET", "/slow", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}()

	// Wait for first request to be inside the handler.
	<-blocked

	// Second request should be rejected (capacity = 1, slot occupied).
	req := httptest.NewRequest("GET", "/rejected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}

	// Unblock the first request.
	close(proceed)
	wg.Wait()
}

func TestConcurrencyLimit_ReleasesSlotAfterCompletion(t *testing.T) {
	var active atomic.Int32
	handler := ConcurrencyLimit(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active.Add(1)
		time.Sleep(5 * time.Millisecond)
		active.Add(-1)
		w.WriteHeader(http.StatusOK)
	}))

	// First request.
	req1 := httptest.NewRequest("GET", "/first", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rec1.Code)
	}

	// Second request after first completes should succeed (slot released).
	req2 := httptest.NewRequest("GET", "/second", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("second request: expected 200, got %d", rec2.Code)
	}
}

func TestStatusRecorder(t *testing.T) {
	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))

	req := httptest.NewRequest("GET", "/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
