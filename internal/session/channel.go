package session

import "slices"

// ChannelLifecycleMap maps channel-specific lifecycle event names to Knossos session FSM transitions.
// Each AI assistant channel (Claude, Gemini, etc.) uses different event names for
// session lifecycle events. This map normalizes them into Knossos Status values.
type ChannelLifecycleMap struct {
	StartEvents   []string
	EndEvents     []string
	SuspendEvents []string
	ResumeEvents  []string
}

// GeminiLifecycleMap returns the lifecycle event mapping for Google Gemini sessions.
func GeminiLifecycleMap() ChannelLifecycleMap {
	return ChannelLifecycleMap{
		StartEvents:   []string{"session_start", "conversation_start"},
		EndEvents:     []string{"session_end", "conversation_end"},
		SuspendEvents: []string{"session_suspend"},
		ResumeEvents:  []string{"session_resume"},
	}
}

// ClaudeLifecycleMap returns the lifecycle event mapping for Claude Code sessions.
func ClaudeLifecycleMap() ChannelLifecycleMap {
	return ChannelLifecycleMap{
		StartEvents:   []string{"SessionStart"},
		EndEvents:     []string{"SessionEnd", "Stop"},
		SuspendEvents: []string{},
		ResumeEvents:  []string{},
	}
}

// MapToFSMTransition maps a channel-specific event name to a Knossos session Status.
// Returns the target status and true if the event maps to a known transition,
// or empty string and false if the event is not recognized.
func (m ChannelLifecycleMap) MapToFSMTransition(eventName string) (Status, bool) {
	if slices.Contains(m.StartEvents, eventName) {
		return StatusActive, true
	}
	if slices.Contains(m.EndEvents, eventName) {
		return StatusArchived, true
	}
	if slices.Contains(m.SuspendEvents, eventName) {
		return StatusParked, true
	}
	if slices.Contains(m.ResumeEvents, eventName) {
		return StatusActive, true
	}
	return "", false
}
