// Package webhook provides Slack webhook signature verification and challenge handling.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	// SlackSignatureHeader is the header containing the HMAC signature.
	SlackSignatureHeader = "X-Slack-Signature"

	// SlackTimestampHeader is the header containing the Unix timestamp.
	SlackTimestampHeader = "X-Slack-Request-Timestamp"

	// signatureVersion is the Slack signing version prefix.
	signatureVersion = "v0"
)

// Verifier validates Slack request signatures using HMAC-SHA256.
// There is NO bypass path -- not for dev, not for test, not for debug, not ever.
type Verifier struct {
	signingSecret   []byte
	maxTimestampAge time.Duration
}

// NewVerifier creates a Verifier with the given signing secret.
// Default max timestamp age is 5 minutes.
func NewVerifier(signingSecret string) *Verifier {
	return &Verifier{
		signingSecret:   []byte(signingSecret),
		maxTimestampAge: 5 * time.Minute,
	}
}

// Verify checks the Slack signature and timestamp against the request body.
// Returns nil if valid, or an error describing the verification failure.
//
// SECURITY: Uses crypto/hmac.Equal for constant-time comparison.
// SECURITY: Rejects timestamps older than maxTimestampAge to prevent replay attacks.
// SECURITY: There is NO bypass path.
func (v *Verifier) Verify(header http.Header, body []byte) error {
	// Validate timestamp
	tsStr := header.Get(SlackTimestampHeader)
	if tsStr == "" {
		return fmt.Errorf("missing %s header", SlackTimestampHeader)
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s header: %w", SlackTimestampHeader, err)
	}

	age := math.Abs(float64(time.Now().Unix() - ts))
	if age > v.maxTimestampAge.Seconds() {
		return fmt.Errorf("timestamp expired: age %.0fs exceeds maximum %s", age, v.maxTimestampAge)
	}

	// Validate signature
	sigStr := header.Get(SlackSignatureHeader)
	if sigStr == "" {
		return fmt.Errorf("missing %s header", SlackSignatureHeader)
	}

	// Compute expected signature: HMAC-SHA256(secret, "v0:" + timestamp + ":" + body)
	sigBase := fmt.Sprintf("%s:%s:%s", signatureVersion, tsStr, string(body))
	mac := hmac.New(sha256.New, v.signingSecret)
	mac.Write([]byte(sigBase))
	expected := fmt.Sprintf("%s=%s", signatureVersion, hex.EncodeToString(mac.Sum(nil)))

	// SECURITY: Constant-time comparison to prevent timing attacks.
	// NEVER use == or bytes.Equal for HMAC comparison.
	if !hmac.Equal([]byte(expected), []byte(sigStr)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// Handler returns middleware that verifies Slack request signatures.
// Requests with invalid signatures receive a 401 Unauthorized response.
// The request body is read, verified, and then restored for downstream handlers.
//
// SECURITY: There is NO bypass path.
func (v *Verifier) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		if err := v.Verify(r.Header, body); err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Restore body for downstream handlers
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}
