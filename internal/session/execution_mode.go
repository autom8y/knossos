package session

// DeriveExecutionMode determines the execution mode from session status and rite.
//
// Returns:
//   - "native" — no session or archived session
//   - "cross-cutting" — session without active rite, or parked session
//   - "orchestrated" — active session with a rite
func DeriveExecutionMode(status Status, activeRite string) string {
	if status == StatusParked {
		return "cross-cutting"
	}
	if status == StatusArchived {
		return "native"
	}
	if activeRite == "" || activeRite == "none" {
		return "cross-cutting"
	}
	return "orchestrated"
}
