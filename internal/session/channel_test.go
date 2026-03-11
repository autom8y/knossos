package session

import "testing"

func TestGeminiLifecycleMap_StartEvents(t *testing.T) {
	t.Parallel()
	m := GeminiLifecycleMap()
	tests := []struct {
		event string
		want  Status
	}{
		{"session_start", StatusActive},
		{"conversation_start", StatusActive},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			t.Parallel()
			got, ok := m.MapToFSMTransition(tt.event)
			if !ok {
				t.Fatalf("MapToFSMTransition(%q) returned false, want true", tt.event)
			}
			if got != tt.want {
				t.Errorf("MapToFSMTransition(%q) = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}

func TestGeminiLifecycleMap_EndEvents(t *testing.T) {
	t.Parallel()
	m := GeminiLifecycleMap()
	tests := []struct {
		event string
		want  Status
	}{
		{"session_end", StatusArchived},
		{"conversation_end", StatusArchived},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			t.Parallel()
			got, ok := m.MapToFSMTransition(tt.event)
			if !ok {
				t.Fatalf("MapToFSMTransition(%q) returned false, want true", tt.event)
			}
			if got != tt.want {
				t.Errorf("MapToFSMTransition(%q) = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}

func TestGeminiLifecycleMap_SuspendEvents(t *testing.T) {
	t.Parallel()
	m := GeminiLifecycleMap()
	got, ok := m.MapToFSMTransition("session_suspend")
	if !ok {
		t.Fatal("MapToFSMTransition(session_suspend) returned false, want true")
	}
	if got != StatusParked {
		t.Errorf("MapToFSMTransition(session_suspend) = %q, want %q", got, StatusParked)
	}
}

func TestGeminiLifecycleMap_ResumeEvents(t *testing.T) {
	t.Parallel()
	m := GeminiLifecycleMap()
	got, ok := m.MapToFSMTransition("session_resume")
	if !ok {
		t.Fatal("MapToFSMTransition(session_resume) returned false, want true")
	}
	if got != StatusActive {
		t.Errorf("MapToFSMTransition(session_resume) = %q, want %q", got, StatusActive)
	}
}

func TestClaudeLifecycleMap_StartEvents(t *testing.T) {
	t.Parallel()
	m := ClaudeLifecycleMap()
	got, ok := m.MapToFSMTransition("SessionStart")
	if !ok {
		t.Fatal("MapToFSMTransition(SessionStart) returned false, want true")
	}
	if got != StatusActive {
		t.Errorf("MapToFSMTransition(SessionStart) = %q, want %q", got, StatusActive)
	}
}

func TestClaudeLifecycleMap_EndEvents(t *testing.T) {
	t.Parallel()
	m := ClaudeLifecycleMap()
	tests := []struct {
		event string
		want  Status
	}{
		{"SessionEnd", StatusArchived},
		{"Stop", StatusArchived},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			t.Parallel()
			got, ok := m.MapToFSMTransition(tt.event)
			if !ok {
				t.Fatalf("MapToFSMTransition(%q) returned false, want true", tt.event)
			}
			if got != tt.want {
				t.Errorf("MapToFSMTransition(%q) = %q, want %q", tt.event, got, tt.want)
			}
		})
	}
}

func TestUnknownEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    ChannelLifecycleMap
	}{
		{"gemini unknown", GeminiLifecycleMap()},
		{"claude unknown", ClaudeLifecycleMap()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := tt.m.MapToFSMTransition("completely_unknown_event")
			if ok {
				t.Error("MapToFSMTransition(unknown) should return false")
			}
		})
	}
}
