// Package hook provides primitives for secure hook execution and authentication.
package hook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"
)

// HookPayload represents the common structure for hook payloads with metadata.
type HookPayload struct {
	Event     string `json:"hook_event_name"`
	Signature string `json:"x_knossos_signature"`
}

// Verify verifies the HMAC-SHA256 signature of the raw body using the KNOSSOS_HOOK_SECRET.
// If KNOSSOS_HOOK_SECRET is unset, it allows the request (transition period/opt-in).
// If set, the signature MUST be present and valid.
func Verify(rawBody []byte, signature string) bool {
	secret := os.Getenv("KNOSSOS_HOOK_SECRET")
	if secret == "" {
		// Temporary: allow if no secret configured to avoid breaking existing setups
		return true
	}

	if signature == "" {
		return false
	}

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expectedMAC := mac.Sum(nil)

	return subtle.ConstantTimeCompare(sigBytes, expectedMAC) == 1
}

// Sign computes the HMAC-SHA256 signature for a raw payload using the KNOSSOS_HOOK_SECRET.
// Returns an empty string if the secret is not set.
func Sign(rawBody []byte) string {
	secret := os.Getenv("KNOSSOS_HOOK_SECRET")
	if secret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}
