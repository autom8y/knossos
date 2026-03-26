// Package streaming implements progressive response rendering via Slack's streaming API.
//
// BC-03: StreamSender is wired via callback onChunk func(chunk string), NOT
// interface import. reason/ does NOT import slack/.
//
// Three-tier degradation:
//  1. Native streaming (chat.startStream -> appendStream -> stopStream)
//  2. Edit-based fallback (chat.postMessage -> chat.update on cadence)
//  3. Single message fallback (standard chat.postMessage)
package streaming

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Sender provides progressive response streaming via Slack's streaming API.
// Falls back to edit-based or single-message rendering on failure.
type Sender struct {
	botToken   string
	teamID     string // Cached from auth.test at startup.
	apiBaseURL string // Override for testing; defaults to "https://slack.com/api/".
	mu         sync.Mutex

	// Active streams tracked for cleanup.
	activeStreams map[string]*streamState
}

type streamState struct {
	channelID string
	threadTS  string
	mode      streamMode
	messageTS string // For edit-based fallback: the posted message timestamp.
	buffer    strings.Builder
}

type streamMode int

const (
	modeNative  streamMode = iota // chat.startStream/appendStream/stopStream
	modeEdit                      // chat.postMessage + chat.update
	modeSingle                    // Single chat.postMessage (final fallback)
)

// StreamChunkDelay is the minimum delay between streaming updates (Slack requirement).
const StreamChunkDelay = 300 * time.Millisecond

// NewSender creates a Sender with the given Slack bot token.
// teamID is obtained from auth.test at server startup for DM streaming.
func NewSender(botToken string, teamID string) *Sender {
	return &Sender{
		botToken:     botToken,
		teamID:       teamID,
		apiBaseURL:   "https://slack.com/api/",
		activeStreams: make(map[string]*streamState),
	}
}

// StartStream initiates a streaming response in a Slack channel/thread.
// Returns a stream_id for subsequent Append/Stop calls.
// On failure, automatically falls back to edit-based mode.
func (s *Sender) StartStream(ctx context.Context, channelID string, threadTS string) (string, error) {
	// Attempt native streaming first.
	streamID, err := s.startNativeStream(ctx, channelID, threadTS)
	if err == nil {
		s.mu.Lock()
		s.activeStreams[streamID] = &streamState{
			channelID: channelID,
			threadTS:  threadTS,
			mode:      modeNative,
		}
		s.mu.Unlock()
		return streamID, nil
	}

	slog.Info("native streaming unavailable, falling back to edit-based",
		"channel", channelID,
		"error", err,
	)

	// Fallback: post an initial message and track its timestamp for edits.
	messageTS, editErr := s.postInitialMessage(ctx, channelID, threadTS)
	if editErr != nil {
		slog.Warn("edit-based streaming unavailable, falling back to single message",
			"channel", channelID,
			"error", editErr,
		)
		// Final fallback: generate a pseudo stream ID for single-message mode.
		fallbackID := fmt.Sprintf("single-%s-%d", channelID, time.Now().UnixNano())
		s.mu.Lock()
		s.activeStreams[fallbackID] = &streamState{
			channelID: channelID,
			threadTS:  threadTS,
			mode:      modeSingle,
		}
		s.mu.Unlock()
		return fallbackID, nil
	}

	editID := fmt.Sprintf("edit-%s-%d", channelID, time.Now().UnixNano())
	s.mu.Lock()
	s.activeStreams[editID] = &streamState{
		channelID: channelID,
		threadTS:  threadTS,
		mode:      modeEdit,
		messageTS: messageTS,
	}
	s.mu.Unlock()

	return editID, nil
}

// AppendStream sends a text chunk to an active stream.
// For native mode: calls chat.appendStream.
// For edit mode: accumulates text and updates the message.
// For single mode: accumulates text (will be sent on StopStream).
func (s *Sender) AppendStream(ctx context.Context, streamID string, chunk string) error {
	s.mu.Lock()
	state, ok := s.activeStreams[streamID]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("unknown stream: %s", streamID)
	}

	switch state.mode {
	case modeNative:
		return s.appendNativeStream(ctx, streamID, chunk)
	case modeEdit:
		// Hold the mutex while writing to the buffer and copying the text.
		// Release before the HTTP call to avoid holding the lock during I/O.
		s.mu.Lock()
		state.buffer.WriteString(chunk)
		text := state.buffer.String()
		channelID := state.channelID
		messageTS := state.messageTS
		s.mu.Unlock()
		return s.updateMessage(ctx, channelID, messageTS, text)
	case modeSingle:
		s.mu.Lock()
		state.buffer.WriteString(chunk)
		s.mu.Unlock()
		return nil // Accumulate only; sent on stop.
	default:
		return fmt.Errorf("unknown stream mode: %d", state.mode)
	}
}

// StopStream finalizes a streaming response.
// BC-09: MUST be deferred in handler. On partial failure: append error indicator.
func (s *Sender) StopStream(ctx context.Context, streamID string) error {
	s.mu.Lock()
	state, ok := s.activeStreams[streamID]
	delete(s.activeStreams, streamID)
	s.mu.Unlock()

	if !ok {
		return nil // Already stopped or never started.
	}

	switch state.mode {
	case modeNative:
		return s.stopNativeStream(ctx, streamID, "")
	case modeEdit:
		// Final update with accumulated content.
		if state.buffer.Len() > 0 {
			return s.updateMessage(ctx, state.channelID, state.messageTS, state.buffer.String())
		}
		return nil
	case modeSingle:
		// Post the accumulated message.
		if state.buffer.Len() > 0 {
			return s.postMessage(ctx, state.channelID, state.threadTS, state.buffer.String())
		}
		return nil
	default:
		return nil
	}
}

// StopStreamWithError finalizes a stream with an error indicator appended.
// Used for mid-stream failure recovery (BC-09).
func (s *Sender) StopStreamWithError(ctx context.Context, streamID string, errorText string) error {
	s.mu.Lock()
	state, ok := s.activeStreams[streamID]
	delete(s.activeStreams, streamID)
	s.mu.Unlock()

	if !ok {
		return nil
	}

	switch state.mode {
	case modeNative:
		return s.stopNativeStream(ctx, streamID, errorText)
	case modeEdit:
		finalText := state.buffer.String() + errorText
		return s.updateMessage(ctx, state.channelID, state.messageTS, finalText)
	case modeSingle:
		finalText := state.buffer.String() + errorText
		return s.postMessage(ctx, state.channelID, state.threadTS, finalText)
	default:
		return nil
	}
}

// --- Native Slack streaming API (raw HTTP) ---

func (s *Sender) startNativeStream(ctx context.Context, channelID, threadTS string) (string, error) {
	payload := map[string]any{
		"channel":  channelID,
		"thread_ts": threadTS,
	}

	respBody, err := s.slackAPICall(ctx, "chat.startStream", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		OK       bool   `json:"ok"`
		StreamID string `json:"stream_id"`
		Error    string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse startStream response: %w", err)
	}
	if !result.OK {
		return "", fmt.Errorf("slack startStream: %s", result.Error)
	}
	return result.StreamID, nil
}

func (s *Sender) appendNativeStream(ctx context.Context, streamID, chunk string) error {
	payload := map[string]any{
		"stream_id":     streamID,
		"markdown_text": chunk,
	}

	respBody, err := s.slackAPICall(ctx, "chat.appendStream", payload)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse appendStream response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack appendStream: %s", result.Error)
	}
	return nil
}

func (s *Sender) stopNativeStream(ctx context.Context, streamID, finalText string) error {
	payload := map[string]any{
		"stream_id": streamID,
	}
	if finalText != "" {
		payload["markdown_text"] = finalText
	}

	respBody, err := s.slackAPICall(ctx, "chat.stopStream", payload)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse stopStream response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack stopStream: %s", result.Error)
	}
	return nil
}

// --- Edit-based fallback ---

func (s *Sender) postInitialMessage(ctx context.Context, channelID, threadTS string) (string, error) {
	payload := map[string]any{
		"channel":   channelID,
		"thread_ts": threadTS,
		"text":      "_Generating response..._",
	}

	respBody, err := s.slackAPICall(ctx, "chat.postMessage", payload)
	if err != nil {
		return "", err
	}

	var result struct {
		OK        bool   `json:"ok"`
		TS        string `json:"ts"`
		Error     string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse postMessage response: %w", err)
	}
	if !result.OK {
		return "", fmt.Errorf("slack postMessage: %s", result.Error)
	}
	return result.TS, nil
}

func (s *Sender) updateMessage(ctx context.Context, channelID, messageTS, text string) error {
	payload := map[string]any{
		"channel": channelID,
		"ts":      messageTS,
		"text":    text,
	}

	respBody, err := s.slackAPICall(ctx, "chat.update", payload)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse update response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack update: %s", result.Error)
	}
	return nil
}

func (s *Sender) postMessage(ctx context.Context, channelID, threadTS, text string) error {
	payload := map[string]any{
		"channel":   channelID,
		"thread_ts": threadTS,
		"text":      text,
	}

	respBody, err := s.slackAPICall(ctx, "chat.postMessage", payload)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse postMessage response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack postMessage: %s", result.Error)
	}
	return nil
}

// --- Common HTTP layer ---

func (s *Sender) slackAPICall(ctx context.Context, method string, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal %s payload: %w", method, err)
	}

	url := fmt.Sprintf("%s%s", s.apiBaseURL, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.botToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", method, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", method, err)
	}

	return respBody, nil
}
