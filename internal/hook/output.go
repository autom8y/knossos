package hook

import (
	"encoding/json"
)

// PreToolUseOutput is the CC-native output format for PreToolUse hooks.
// CC expects decisions wrapped in a hookSpecificOutput envelope.
// See: https://code.claude.com/docs/en/hooks#pretooluse-decision-control
type PreToolUseOutput struct {
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput"`
}

// HookSpecificOutput contains the PreToolUse decision fields CC reads.
type HookSpecificOutput struct {
	HookEventName            string          `json:"hookEventName"`
	PermissionDecision       string          `json:"permissionDecision"`
	PermissionDecisionReason string          `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             json.RawMessage `json:"updatedInput,omitempty"`
	AdditionalContext        string          `json:"additionalContext,omitempty"`
}
