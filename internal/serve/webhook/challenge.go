package webhook

import (
	"encoding/json"
	"net/http"
)

// challengeRequest is the JSON body Slack sends for URL verification.
type challengeRequest struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
}

// challengeResponse is the JSON body returned for URL verification.
type challengeResponse struct {
	Challenge string `json:"challenge"`
}

// HandleChallenge checks if the request body is a Slack URL verification challenge.
// If it is, it writes the challenge response and returns true.
// If it is not a challenge, it returns false and the caller should continue processing.
//
// This is called after signature verification -- the challenge body has already
// been authenticated.
func HandleChallenge(w http.ResponseWriter, _ *http.Request, body []byte) bool {
	var req challengeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}

	if req.Type != "url_verification" || req.Challenge == "" {
		return false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(challengeResponse{Challenge: req.Challenge})
	return true
}
