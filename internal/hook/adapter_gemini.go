package hook

import (
	"encoding/json"
	"io"
	"os"
)

// GeminiAdapter handles Gemini CLI hook payloads.
// Gemini sends the same snake_case JSON fields as Claude Code (session_id,
// hook_event_name, tool_name, tool_input, tool_response, cwd), so StdinPayload
// can parse Gemini payloads directly. Gemini-only fields (timestamp, mcp_context)
// are silently ignored by json.Unmarshal.
//
// The only Gemini-specific logic is event name translation: Gemini wire names
// (BeforeTool, AfterTool, etc.) are translated to CC canonical names (PreToolUse,
// PostToolUse, etc.) before validation, so all downstream hook commands operate
// on a uniform Env regardless of which CLI fired the event.
type GeminiAdapter struct{}

func (a *GeminiAdapter) ParsePayload(reader io.Reader) (*Env, error) {
	projectDir := os.Getenv(EnvProjectDir)

	data, err := io.ReadAll(reader)
	if err != nil || len(data) == 0 {
		return &Env{ProjectDir: projectDir}, nil
	}

	var payload StdinPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return &Env{ProjectDir: projectDir}, nil
	}

	// Translate Gemini wire event name to CC canonical before validation.
	// e.g. "BeforeTool" -> "PreToolUse"; unknown Gemini-only events (BeforeModel)
	// pass through TranslateInboundEvent unchanged, then fail isValidHookEvent.
	translatedEvent := TranslateInboundEvent(payload.HookEventName)
	event := HookEvent(translatedEvent)
	if event != "" && !isValidHookEvent(event) {
		// Gemini-only events (e.g. BeforeModel, AfterModel) -- silently ignore
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
		Signature:      payload.XKnossosSignature,
		RawPayload:     data,
	}, nil
}

// FormatResponse serializes a hook decision to JSON for stdout.
// Gemini reads the same {"decision": "...", "reason": "..."} format as Claude Code.
// No os.Exit: the caller prints the response bytes and the process exits naturally.
func (a *GeminiAdapter) FormatResponse(decision string, reason string) ([]byte, error) {
	resp := map[string]interface{}{
		"decision": decision,
	}
	if reason != "" {
		resp["reason"] = reason
	}
	return json.Marshal(resp)
}

func (a *GeminiAdapter) ChannelName() string {
	return "gemini"
}
