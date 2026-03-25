package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testSecret = "8f742231b10e8888abcd99yyyzzz85a5"

// computeSignature generates a valid Slack signature for testing.
func computeSignature(secret, timestamp string, body []byte) string {
	sigBase := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigBase))
	return fmt.Sprintf("v0=%s", hex.EncodeToString(mac.Sum(nil)))
}

func validHeaders(body []byte) http.Header {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := computeSignature(testSecret, ts, body)
	h := http.Header{}
	h.Set(SlackTimestampHeader, ts)
	h.Set(SlackSignatureHeader, sig)
	return h
}

func TestVerify_ValidSignature(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{"type":"event_callback","event":{"type":"app_mention"}}`)
	headers := validHeaders(body)

	if err := v.Verify(headers, body); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{"type":"event_callback"}`)

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	h := http.Header{}
	h.Set(SlackTimestampHeader, ts)
	h.Set(SlackSignatureHeader, "v0=0000000000000000000000000000000000000000000000000000000000000000")

	err := v.Verify(h, body)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if !strings.Contains(err.Error(), "signature mismatch") {
		t.Errorf("expected 'signature mismatch' error, got: %v", err)
	}
}

func TestVerify_MissingSignatureHeader(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{}`)

	h := http.Header{}
	h.Set(SlackTimestampHeader, strconv.FormatInt(time.Now().Unix(), 10))

	err := v.Verify(h, body)
	if err == nil {
		t.Fatal("expected error for missing signature header")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("expected 'missing' in error, got: %v", err)
	}
}

func TestVerify_MissingTimestampHeader(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{}`)

	h := http.Header{}
	h.Set(SlackSignatureHeader, "v0=abc")

	err := v.Verify(h, body)
	if err == nil {
		t.Fatal("expected error for missing timestamp header")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("expected 'missing' in error, got: %v", err)
	}
}

func TestVerify_ExpiredTimestamp(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{}`)

	// Timestamp 10 minutes ago
	ts := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	sig := computeSignature(testSecret, ts, body)
	h := http.Header{}
	h.Set(SlackTimestampHeader, ts)
	h.Set(SlackSignatureHeader, sig)

	err := v.Verify(h, body)
	if err == nil {
		t.Fatal("expected error for expired timestamp")
	}
	if !strings.Contains(err.Error(), "timestamp expired") {
		t.Errorf("expected 'timestamp expired' error, got: %v", err)
	}
}

func TestVerify_FutureTimestamp(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{}`)

	// Timestamp 10 minutes in the future
	ts := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	sig := computeSignature(testSecret, ts, body)
	h := http.Header{}
	h.Set(SlackTimestampHeader, ts)
	h.Set(SlackSignatureHeader, sig)

	err := v.Verify(h, body)
	if err == nil {
		t.Fatal("expected error for future timestamp")
	}
	if !strings.Contains(err.Error(), "timestamp expired") {
		t.Errorf("expected 'timestamp expired' error, got: %v", err)
	}
}

func TestVerify_InvalidTimestampFormat(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{}`)

	h := http.Header{}
	h.Set(SlackTimestampHeader, "not-a-number")
	h.Set(SlackSignatureHeader, "v0=abc")

	err := v.Verify(h, body)
	if err == nil {
		t.Fatal("expected error for invalid timestamp format")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected 'invalid' in error, got: %v", err)
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	v := NewVerifier("wrong-secret")
	body := []byte(`{"type":"event_callback"}`)
	headers := validHeaders(body) // signed with testSecret

	err := v.Verify(headers, body)
	if err == nil {
		t.Fatal("expected error for wrong signing secret")
	}
}

func TestVerify_TamperedBody(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{"type":"event_callback"}`)
	headers := validHeaders(body)

	// Tamper with body after signing
	tamperedBody := []byte(`{"type":"event_callback","injected":"true"}`)

	err := v.Verify(headers, tamperedBody)
	if err == nil {
		t.Fatal("expected error for tampered body")
	}
}

func TestVerify_EmptyBody(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte("")
	headers := validHeaders(body)

	// Empty body with valid signature should pass
	if err := v.Verify(headers, body); err != nil {
		t.Errorf("expected nil error for empty body with valid signature, got: %v", err)
	}
}

func TestHandler_ValidRequest(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{"type":"event_callback","event":{"type":"app_mention"}}`)

	var receivedBody []byte
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	handler := v.Handler(inner)

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := computeSignature(testSecret, ts, body)

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader(body))
	req.Header.Set(SlackTimestampHeader, ts)
	req.Header.Set(SlackSignatureHeader, sig)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Verify body was restored for downstream handler
	if string(receivedBody) != string(body) {
		t.Errorf("body not restored: expected %q, got %q", string(body), string(receivedBody))
	}
}

func TestHandler_InvalidSignature_401(t *testing.T) {
	v := NewVerifier(testSecret)
	body := []byte(`{"type":"event_callback"}`)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid signature")
	})

	handler := v.Handler(inner)

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader(body))
	req.Header.Set(SlackTimestampHeader, strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set(SlackSignatureHeader, "v0=invalid")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestHandler_MissingHeaders_401(t *testing.T) {
	v := NewVerifier(testSecret)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for missing headers")
	})

	handler := v.Handler(inner)

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// TestNoBypassPath verifies there is no configuration, environment variable,
// or parameter that can disable signature verification.
func TestNoBypassPath(t *testing.T) {
	// The Verifier type has only two fields: signingSecret and maxTimestampAge.
	// Neither provides a bypass mechanism. This test documents that invariant.
	v := NewVerifier(testSecret)

	// Even with a zero-length secret, verification should reject unsigned requests
	emptyV := NewVerifier("")
	body := []byte(`{}`)

	h := http.Header{}
	h.Set(SlackTimestampHeader, strconv.FormatInt(time.Now().Unix(), 10))
	// No signature header
	if err := v.Verify(h, body); err == nil {
		t.Error("standard verifier must reject requests without signature")
	}
	if err := emptyV.Verify(h, body); err == nil {
		t.Error("empty-secret verifier must reject requests without signature")
	}
}

func TestHandleChallenge(t *testing.T) {
	body, _ := json.Marshal(challengeRequest{
		Type:      "url_verification",
		Challenge: "test_challenge_token",
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handled := HandleChallenge(rec, req, body)
	if !handled {
		t.Fatal("expected HandleChallenge to return true for url_verification")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp challengeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Challenge != "test_challenge_token" {
		t.Errorf("expected challenge 'test_challenge_token', got %q", resp.Challenge)
	}
}

func TestHandleChallenge_NotChallenge(t *testing.T) {
	body := []byte(`{"type":"event_callback","event":{"type":"app_mention"}}`)
	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handled := HandleChallenge(rec, req, body)
	if handled {
		t.Error("expected HandleChallenge to return false for non-challenge request")
	}
}

func TestHandleChallenge_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handled := HandleChallenge(rec, req, body)
	if handled {
		t.Error("expected HandleChallenge to return false for invalid JSON")
	}
}

func TestHandleChallenge_EmptyChallenge(t *testing.T) {
	body, _ := json.Marshal(challengeRequest{
		Type:      "url_verification",
		Challenge: "",
	})

	req := httptest.NewRequest("POST", "/slack/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handled := HandleChallenge(rec, req, body)
	if handled {
		t.Error("expected HandleChallenge to return false for empty challenge")
	}
}
