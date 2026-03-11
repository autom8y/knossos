package hook

import (
	"encoding/json"
	"io"
	"os"
)

type GeminiPayload struct {
	SessionID      string          `json:"sessionId"`
	ConversationID string          `json:"conversationId"`
	CWD            string          `json:"cwd"`
	HookEventName  string          `json:"event"`
	ToolName       string          `json:"toolName"`
	ToolInput      json.RawMessage `json:"toolInput"`
	ToolResponse   json.RawMessage `json:"toolResponse"`
	Prompt         string          `json:"prompt"`
}

type GeminiAdapter struct{}

func (a *GeminiAdapter) ParsePayload(reader io.Reader) (*Env, error) {
	projectDir := os.Getenv(EnvProjectDir)

	data, err := io.ReadAll(reader)
	if err != nil || len(data) == 0 {
		return &Env{ProjectDir: projectDir}, nil
	}

	var payload GeminiPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return &Env{ProjectDir: projectDir}, nil
	}

	event := HookEvent(payload.HookEventName)
	if event != "" && !isValidHookEvent(event) {
		event = ""
	}

	var cwd string
	if payload.CWD != "" {
		cwd = payload.CWD
		if projectDir == "" {
			projectDir = payload.CWD
		}
	}

	var toolInput string
	if len(payload.ToolInput) > 0 && string(payload.ToolInput) != "null" {
		toolInput = unwrapJSONValue(payload.ToolInput)
	}

	var toolResult string
	if len(payload.ToolResponse) > 0 && string(payload.ToolResponse) != "null" {
		toolResult = unwrapJSONValue(payload.ToolResponse)
	}

	return &Env{
		Event:          event,
		ToolName:       payload.ToolName,
		ToolInput:      toolInput,
		ToolResult:     toolResult,
		SessionID:      payload.SessionID,
		ProjectDir:     projectDir,
		CWD:            cwd,
		ConversationID: payload.ConversationID,
		UserMessage:    payload.Prompt,
	}, nil
}

func (a *GeminiAdapter) FormatResponse(decision string, reason string) ([]byte, error) {
	// Gemini CLI uses exit codes: 0 = allow, 1 = block
	// We return the reason as stderr output if blocked
	if decision == "block" || decision == "deny" { // "deny" just in case
		os.Stderr.WriteString(reason + "\n")
		os.Exit(1)
	}
	
	// Exit 0 implicitly happens if we just return normally, but we can also os.Exit(0)
	// Return empty since Gemini reads stderr and exit codes, not a JSON response.
	return []byte{}, nil
}

func (a *GeminiAdapter) ChannelName() string {
	return "gemini"
}
