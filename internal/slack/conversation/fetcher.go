package conversation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// slackThreadFetcher implements SlackThreadFetcher via raw HTTP calls to
// the Slack conversations.replies API. Same request pattern as
// SlackClient.rawAPICall (internal/slack/client.go:111-150).
type slackThreadFetcher struct {
	botToken   string
	apiBaseURL string
}

// NewSlackThreadFetcher creates a fetcher that calls the production Slack API.
func NewSlackThreadFetcher(botToken string) SlackThreadFetcher {
	return &slackThreadFetcher{
		botToken:   botToken,
		apiBaseURL: "https://slack.com/api/",
	}
}

// slackMessage is the minimal Slack message shape for conversations.replies.
type slackMessage struct {
	User    string `json:"user"`
	BotID   string `json:"bot_id"`
	Text    string `json:"text"`
	TS      string `json:"ts"`
	SubType string `json:"subtype"`
}

// FetchThreadMessages retrieves messages from a Slack thread via
// conversations.replies. It filters out the parent message, bot messages,
// and messages with subtypes (same filters as handler.go:484-499).
//
// Fail-open: errors return (nil, error). The caller (resurrectThread) handles
// fallback via the RESURRECTING timeout.
func (f *slackThreadFetcher) FetchThreadMessages(ctx context.Context, channelID string, threadTS string, limit int) ([]ThreadMessage, error) {
	payload, err := json.Marshal(map[string]any{
		"channel":   channelID,
		"ts":        threadTS,
		"limit":     limit,
		"inclusive": true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal conversations.replies payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.apiBaseURL+"conversations.replies", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create conversations.replies request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.botToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("conversations.replies request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("conversations.replies returned HTTP %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read conversations.replies response: %w", err)
	}

	var result struct {
		OK       bool           `json:"ok"`
		Error    string         `json:"error,omitempty"`
		Messages []slackMessage `json:"messages"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse conversations.replies response: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("slack conversations.replies: %s", result.Error)
	}

	messages := make([]ThreadMessage, 0, len(result.Messages))
	for _, msg := range result.Messages {
		// Filter: parent message (Slack always includes it in replies).
		if msg.TS == threadTS {
			continue
		}
		// Filter: bot messages (same as handler.go:485).
		if msg.BotID != "" {
			continue
		}
		// Filter: messages with subtypes (same as handler.go:495).
		if msg.SubType != "" {
			continue
		}

		// Heuristic: messages without a User field that passed the BotID and SubType
		// filters above are classified as assistant responses. This is safe because
		// subtypeless system messages always carry a subtype field and are filtered out.
		role := "assistant"
		if msg.User != "" {
			role = "user"
		}

		ts, parseErr := parseSlackTimestamp(msg.TS)
		if parseErr != nil {
			// Non-fatal: use zero time rather than failing the whole fetch.
			ts = time.Time{}
		}

		messages = append(messages, ThreadMessage{
			Role:      role,
			Content:   msg.Text,
			Timestamp: ts,
		})
	}

	return messages, nil
}

// parseSlackTimestamp converts a Slack ts string (e.g. "1716000000.123456")
// to a time.Time. The integer part is seconds, the fractional part is
// sub-second precision.
func parseSlackTimestamp(ts string) (time.Time, error) {
	f, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse slack timestamp %q: %w", ts, err)
	}
	sec := int64(f)
	nsec := int64((f - float64(sec)) * float64(time.Second))
	// Guard against negative nsec from floating-point rounding.
	if nsec < 0 {
		nsec = 0
	}
	if nsec >= int64(time.Second) {
		nsec = int64(time.Second) - 1
	}
	return time.Unix(sec, nsec), nil
}

// init-time assertion: slackThreadFetcher satisfies the interface.
var _ SlackThreadFetcher = (*slackThreadFetcher)(nil)

