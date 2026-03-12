package hook

import (
	"encoding/json"
	"io"
	"os"
)

type ClaudeAdapter struct{}

func (a *ClaudeAdapter) ParsePayload(reader io.Reader) (*Env, error) {
	projectDir := os.Getenv(EnvProjectDir)

	data, err := io.ReadAll(reader)
	if err != nil || len(data) == 0 {
		return &Env{ProjectDir: projectDir}, nil
	}

	var stdin StdinPayload
	if err := json.Unmarshal(data, &stdin); err != nil {
		return &Env{ProjectDir: projectDir}, nil
	}

	// Translate CC wire event name to knossos canonical (ADR-0032).
	// e.g. "PreToolUse" -> "pre_tool"; unknown events pass through, then fail isValidHookEvent.
	event := HookEvent(WireToCanonical(stdin.HookEventName))
	if event != "" && !isValidHookEvent(event) {
		event = ""
	}

	var cwd string
	if stdin.CWD != "" {
		cwd = stdin.CWD
		if projectDir == "" {
			projectDir = stdin.CWD
		}
	}

	var toolInput string
	if len(stdin.ToolInput) > 0 && string(stdin.ToolInput) != "null" {
		toolInput = unwrapJSONValue(stdin.ToolInput)
	}

	var toolResult string
	if len(stdin.ToolResponse) > 0 && string(stdin.ToolResponse) != "null" {
		toolResult = unwrapJSONValue(stdin.ToolResponse)
	}

	return &Env{
		Event:             event,
		ToolName:          stdin.ToolName,
		ToolInput:         toolInput,
		ToolResult:        toolResult,
		SessionID:         stdin.SessionID,
		ProjectDir:        projectDir,
		CWD:               cwd,
		ConversationID:    stdin.ConversationID,
		UserMessage:       stdin.Prompt,
		Signature:         stdin.XKnossosSignature,
		RawPayload:        data,
	}, nil
}

func (a *ClaudeAdapter) FormatResponse(decision string, reason string) ([]byte, error) {
	// Claude Code hook response format
	resp := map[string]interface{}{
		"decision": decision,
	}
	if reason != "" {
		resp["reason"] = reason
	}
	return json.Marshal(resp)
}

func (a *ClaudeAdapter) ChannelName() string {
	return "claude"
}
