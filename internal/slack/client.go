package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	slackapi "github.com/slack-go/slack"
)

// SlackAPI abstracts the subset of the slack-go client used by Clew.
// Defined as an interface for testability.
type SlackAPI interface {
	PostMessage(channelID string, options ...slackapi.MsgOption) (string, string, error)
	UpdateMessage(channelID, timestamp string, options ...slackapi.MsgOption) (string, string, string, error)
	AuthTest() (*slackapi.AuthTestResponse, error)
}

// SlackClient wraps SlackAPI with Clew-specific convenience methods.
type SlackClient struct {
	api            SlackAPI
	botToken       string
	rawAPIBaseURL  string // Override for tests; empty uses https://slack.com/api/.
}

// NewSlackClient creates a SlackClient with the given bot token.
// Uses the real slack-go client for production.
func NewSlackClient(botToken string) *SlackClient {
	return &SlackClient{
		api:      slackapi.New(botToken),
		botToken: botToken,
	}
}

// NewSlackClientWithAPI creates a SlackClient with an injected API implementation.
// Used in tests.
func NewSlackClientWithAPI(api SlackAPI, botToken string) *SlackClient {
	return &SlackClient{
		api:      api,
		botToken: botToken,
	}
}

// SendBlocks posts a message with Block Kit blocks to a channel thread.
func (c *SlackClient) SendBlocks(channelID, threadTS string, blocks []slackapi.Block) error {
	opts := []slackapi.MsgOption{
		slackapi.MsgOptionBlocks(blocks...),
	}
	if threadTS != "" {
		opts = append(opts, slackapi.MsgOptionTS(threadTS))
	}
	_, _, err := c.api.PostMessage(channelID, opts...)
	if err != nil {
		return fmt.Errorf("send blocks to %s: %w", channelID, err)
	}
	return nil
}

// SetStatus sets the assistant thread status using the Slack assistant.threads.setStatus API.
// This is a raw HTTP call because slack-go may not support assistant thread APIs natively.
func (c *SlackClient) SetStatus(channelID, threadTS, emoji, statusText string) error {
	payload := map[string]any{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"status":     statusText,
	}
	return c.rawAPICall("assistant.threads.setStatus", payload)
}

// SetSuggestedPrompts sets the suggested prompts for an assistant thread.
// Uses raw HTTP since slack-go may not support this API.
func (c *SlackClient) SetSuggestedPrompts(channelID, threadTS string, prompts []string) error {
	// Build prompt objects per Slack API spec.
	promptObjects := make([]map[string]string, len(prompts))
	for i, p := range prompts {
		promptObjects[i] = map[string]string{
			"title":   p,
			"message": p,
		}
	}
	payload := map[string]any{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"prompts":    promptObjects,
	}
	return c.rawAPICall("assistant.threads.setSuggestedPrompts", payload)
}

// SetTitle sets the title of an assistant thread.
// Uses raw HTTP since slack-go may not support this API.
func (c *SlackClient) SetTitle(channelID, threadTS, title string) error {
	payload := map[string]any{
		"channel_id": channelID,
		"thread_ts":  threadTS,
		"title":      title,
	}
	return c.rawAPICall("assistant.threads.setTitle", payload)
}

// AddReaction adds an emoji reaction to a message.
// Uses the reactions.add Slack API method via rawAPICall.
// channelID is the channel containing the message, timestamp identifies the
// specific message to react to (msg.TS, not thread_ts), and emoji is the
// reaction name without colons (e.g., "eyes" not ":eyes:").
func (c *SlackClient) AddReaction(channelID, timestamp, emoji string) error {
	payload := map[string]any{
		"channel":   channelID,
		"timestamp": timestamp,
		"name":      emoji,
	}
	return c.rawAPICall("reactions.add", payload)
}

// HealthCheck verifies connectivity to the Slack API by calling auth.test.
func (c *SlackClient) HealthCheck(ctx context.Context) error {
	_, err := c.api.AuthTest()
	if err != nil {
		return fmt.Errorf("slack auth.test failed: %w", err)
	}
	return nil
}

// rawAPICall makes a raw HTTP POST to a Slack API method with a JSON payload.
// Used for assistant thread APIs that are not yet supported by slack-go.
func (c *SlackClient) rawAPICall(method string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s payload: %w", method, err)
	}

	baseURL := c.rawAPIBaseURL
	if baseURL == "" {
		baseURL = "https://slack.com/api/"
	}
	url := baseURL + method
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.botToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", method, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s response: %w", method, err)
	}

	// Slack returns 200 even on errors; check the "ok" field.
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse %s response: %w", method, err)
	}
	if !result.OK {
		return fmt.Errorf("slack %s: %s", method, result.Error)
	}
	return nil
}
