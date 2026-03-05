package clewcontract

// typed_constructors.go -- Constructor functions for all v3 TypedEvent types.
//
// Each constructor corresponds to one event type from the SESSION-1 spec type catalog.
// Naming convention: NewTyped{EventName}Event()
//
// These constructors produce TypedEvent values (v3 format).
// The existing New*Event() constructors in event.go continue to produce v2 flat Event values.
// Both formats write to the same events.jsonl via EventWriter.Write() / EventWriter.WriteTyped().

// --- Curated types (projected to timeline) ---

// NewTypedSessionCreatedEvent creates a v3 "session.created" TypedEvent.
// Source: cli (emitted by `ari session create`).
func NewTypedSessionCreatedEvent(sessionID, initiative, complexity, rite string) TypedEvent {
	return newTypedEvent(EventTypeSessionCreated, SourceCLI, SessionCreatedData{
		SessionID:  sessionID,
		Initiative: initiative,
		Complexity: complexity,
		Rite:       rite,
	})
}

// NewTypedSessionParkedEvent creates a v3 "session.parked" TypedEvent.
// Source: cli (emitted by `ari session park`).
func NewTypedSessionParkedEvent(sessionID, reason string) TypedEvent {
	return newTypedEvent(EventTypeSessionParked, SourceCLI, SessionParkedData{
		SessionID: sessionID,
		Reason:    reason,
	})
}

// NewTypedSessionResumedEvent creates a v3 "session.resumed" TypedEvent.
// Source: cli (emitted by `ari session resume`).
func NewTypedSessionResumedEvent(sessionID string) TypedEvent {
	return newTypedEvent(EventTypeSessionResumed, SourceCLI, SessionResumedData{
		SessionID: sessionID,
	})
}

// NewTypedSessionWrappedEvent creates a v3 "session.wrapped" TypedEvent.
// Source: cli (emitted by `ari session wrap`).
// This is the curated "session intentionally concluded" signal.
// sailsColor is optional (WHITE, GRAY, BLACK); pass empty string if not yet generated.
func NewTypedSessionWrappedEvent(sessionID, sailsColor string, durationMs int64) TypedEvent {
	return newTypedEvent(EventTypeSessionWrapped, SourceCLI, SessionWrappedData{
		SessionID:  sessionID,
		SailsColor: sailsColor,
		DurationMs: durationMs,
	})
}

// NewTypedSessionFrayedEvent creates a v3 "session.frayed" TypedEvent.
// Source: cli (emitted by `ari session fray`).
func NewTypedSessionFrayedEvent(parentID, childID, frayPoint string) TypedEvent {
	return newTypedEvent(EventTypeSessionFrayed, SourceCLI, SessionFrayedData{
		ParentID:  parentID,
		ChildID:   childID,
		FrayPoint: frayPoint,
	})
}

// NewTypedPhaseTransitionedEvent creates a v3 "phase.transitioned" TypedEvent.
// Source: cli (emitted by `ari session transition`).
func NewTypedPhaseTransitionedEvent(sessionID, fromPhase, toPhase string) TypedEvent {
	return newTypedEvent(EventTypePhaseTransitioned, SourceCLI, PhaseTransitionedData{
		SessionID: sessionID,
		From:      fromPhase,
		To:        toPhase,
	})
}

// NewTypedAgentDelegatedEvent creates a v3 "agent.delegated" TypedEvent.
// Source: hook (emitted by SubagentStart hook) or cli (emitted by `ari handoff execute`).
// agentType and taskID are optional; agentID is the CC-assigned subagent identifier.
func NewTypedAgentDelegatedEvent(source EventSource, agentName, agentType, taskID, agentID string) TypedEvent {
	return newTypedEvent(EventTypeAgentDelegated, source, AgentDelegatedData{
		AgentName: agentName,
		AgentType: agentType,
		TaskID:    taskID,
		AgentID:   agentID,
	})
}

// NewTypedAgentCompletedEvent creates a v3 "agent.completed" TypedEvent.
// Source: hook (emitted by SubagentStop hook) or cli (emitted by `ari handoff prepare`).
// agentType, taskID, agentID, outcome, and durationMs are optional.
// artifacts is optional; pass nil or empty slice when no artifacts produced.
func NewTypedAgentCompletedEvent(source EventSource, agentName, agentType, taskID, agentID, outcome string, durationMs int64, artifacts []string) TypedEvent {
	return newTypedEvent(EventTypeAgentCompleted, source, AgentCompletedData{
		AgentName:  agentName,
		AgentType:  agentType,
		TaskID:     taskID,
		AgentID:    agentID,
		Outcome:    outcome,
		DurationMs: durationMs,
		Artifacts:  artifacts,
	})
}

// NewTypedCommitCreatedEvent creates a v3 "commit.created" TypedEvent.
// Source: hook (emitted by PostToolUse hook when git commit detected in Bash output).
// sha may be short (7-char) or full (40-char) commit hash.
// message should be the first line of the commit message only.
func NewTypedCommitCreatedEvent(sha, message string) TypedEvent {
	return newTypedEvent(EventTypeCommitCreated, SourceHook, CommitCreatedData{
		SHA:     sha,
		Message: message,
	})
}

// NewTypedDecisionRecordedEvent creates a v3 "decision.recorded" TypedEvent.
// Source: agent (emitted via `ari session log --type=decision`).
// rejected is optional; pass nil when no alternatives were considered.
func NewTypedDecisionRecordedEvent(decision, rationale string, rejected []string) TypedEvent {
	return newTypedEvent(EventTypeDecisionRecorded, SourceAgent, DecisionRecordedData{
		Decision:  decision,
		Rationale: rationale,
		Rejected:  rejected,
	})
}

// NewTypedCommandInvokedEvent creates a v3 "command.invoked" TypedEvent.
// Source: hook (emitted by PostToolUse hook when Skill tool call detected).
// commandType should be "skill" or "command".
func NewTypedCommandInvokedEvent(command, commandType string) TypedEvent {
	return newTypedEvent(EventTypeCommandInvoked, SourceHook, CommandInvokedData{
		Command: command,
		Type:    commandType,
	})
}

// --- Backplane-only types ---

// NewTypedToolInvokedEvent creates a v3 "tool.invoked" TypedEvent.
// Source: hook (emitted by PostToolUse hook for every tool call).
// meta is optional; use for tool-specific fields (exit_code, command, etc.).
func NewTypedToolInvokedEvent(tool, path, summary string, meta map[string]any) TypedEvent {
	return newTypedEvent(EventTypeToolInvoked, SourceHook, ToolInvokedData{
		Tool:    tool,
		Path:    path,
		Summary: summary,
		Meta:    meta,
	})
}

// NewTypedFileModifiedEvent creates a v3 "file.modified" TypedEvent.
// Source: hook (emitted by PostToolUse hook via emitSupplementalEvents).
func NewTypedFileModifiedEvent(path string, linesChanged int) TypedEvent {
	return newTypedEvent(EventTypeFileModified, SourceHook, FileModifiedData{
		Path:         path,
		LinesChanged: linesChanged,
	})
}

// NewTypedArtifactCreatedEvent creates a v3 "artifact.created" TypedEvent.
// Source: hook (emitted by PostToolUse hook via emitSupplementalEvents).
// validatesAgainst, wipType, and slug are optional.
func NewTypedArtifactCreatedEvent(artifactType, path, phase, validatesAgainst, wipType, slug string) TypedEvent {
	return newTypedEvent(EventTypeArtifactCreatedV3, SourceHook, ArtifactCreatedData{
		ArtifactType:     artifactType,
		Path:             path,
		Phase:            phase,
		ValidatesAgainst: validatesAgainst,
		WipType:          wipType,
		Slug:             slug,
	})
}

// NewTypedErrorOccurredEvent creates a v3 "error.occurred" TypedEvent.
// Source: hook or cli.
func NewTypedErrorOccurredEvent(source EventSource, errorCode, message, context string, recoverable bool, suggestedAction string) TypedEvent {
	return newTypedEvent(EventTypeErrorOccurred, source, ErrorOccurredData{
		ErrorCode:       errorCode,
		Message:         message,
		Context:         context,
		Recoverable:     recoverable,
		SuggestedAction: suggestedAction,
	})
}

// NewTypedSessionStartedEvent creates a v3 "session.started" TypedEvent.
// Source: hook (emitted by SessionStart CC lifecycle hook).
// Distinct from NewTypedSessionCreatedEvent: started = CC session began; created = knossos session created via CLI.
func NewTypedSessionStartedEvent(sessionID, initiative, complexity, rite string) TypedEvent {
	return newTypedEvent(EventTypeSessionStart, SourceHook, SessionStartedData{
		SessionID:  sessionID,
		Initiative: initiative,
		Complexity: complexity,
		Rite:       rite,
	})
}

// NewTypedSessionEndedEvent creates a v3 "session.ended" TypedEvent.
// Source: hook or cli.
// cognitiveBudget is optional; pass nil when not available.
func NewTypedSessionEndedEvent(source EventSource, sessionID, status string, durationMs int64, cognitiveBudget map[string]any) TypedEvent {
	return newTypedEvent(EventTypeSessionEnd, source, SessionEndedData{
		SessionID:       sessionID,
		Status:          status,
		DurationMs:      durationMs,
		CognitiveBudget: cognitiveBudget,
	})
}

// NewTypedSessionArchivedEvent creates a v3 "session.archived" TypedEvent.
// Source: cli (emitted by `ari session wrap`).
// fromStatus is the status the session transitions from: "ACTIVE" or "PARKED".
func NewTypedSessionArchivedEvent(sessionID, fromStatus string) TypedEvent {
	return newTypedEvent(EventTypeSessionArchived, SourceCLI, SessionArchivedData{
		SessionID:  sessionID,
		FromStatus: fromStatus,
	})
}

// NewTypedStrandResolvedEvent creates a v3 "session.strand_resolved" TypedEvent.
// Source: cli (emitted by `ari session wrap` when a frayed child wraps).
// resolution is one of: "wrapped", "abandoned".
func NewTypedStrandResolvedEvent(parentID, childID, resolution string) TypedEvent {
	return newTypedEvent(EventTypeStrandResolved, SourceCLI, StrandResolvedData{
		ParentID:   parentID,
		ChildID:    childID,
		Resolution: resolution,
	})
}

// NewTypedSchemaMigratedEvent creates a v3 "session.schema_migrated" TypedEvent.
// Source: cli (emitted by `ari session migrate`).
func NewTypedSchemaMigratedEvent(sessionID, fromVersion, toVersion string) TypedEvent {
	return newTypedEvent(EventTypeSchemaMigrated, SourceCLI, SchemaMigratedData{
		SessionID:   sessionID,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
	})
}

// NewTypedLockAcquiredEvent creates a v3 "lock.acquired" TypedEvent.
// Source: cli (emitted by session lock operations).
func NewTypedLockAcquiredEvent(sessionID, holder string) TypedEvent {
	return newTypedEvent(EventTypeLockAcquired, SourceCLI, LockAcquiredData{
		SessionID: sessionID,
		Holder:    holder,
	})
}

// NewTypedLockReleasedEvent creates a v3 "lock.released" TypedEvent.
// Source: cli (emitted by session lock operations; currently declared but not emitted).
func NewTypedLockReleasedEvent(sessionID, holder string) TypedEvent {
	return newTypedEvent(EventTypeLockReleased, SourceCLI, LockReleasedData{
		SessionID: sessionID,
		Holder:    holder,
	})
}

// NewTypedSailsGeneratedEvent creates a v3 "quality.sails_generated" TypedEvent.
// Source: cli (emitted by `ari session wrap`).
// evidence is optional; keys: tests, build, lint, adversarial, integration.
func NewTypedSailsGeneratedEvent(sessionID, color, computedBase string, reasons []string, filePath string, evidence map[string]string) TypedEvent {
	return newTypedEvent(EventTypeSailsGenerated, SourceCLI, SailsGeneratedTypedData{
		SessionID:    sessionID,
		Color:        color,
		ComputedBase: computedBase,
		Reasons:      reasons,
		FilePath:     filePath,
		Evidence:     evidence,
	})
}

// NewTypedContextSwitchEvent creates a v3 "context_switch" TypedEvent.
// Source: hook.
// Note: context_switch uses a legacy non-dotted type string retained for backward compat.
func NewTypedContextSwitchEvent(summary, path string) TypedEvent {
	return newTypedEvent(EventTypeContextSwitch, SourceHook, ContextSwitchData{
		Summary: summary,
		Path:    path,
	})
}

// NewTypedHandoffPreparedEvent creates a v3 "agent.handoff_prepared" TypedEvent.
// Source: cli (emitted by `ari handoff prepare`).
func NewTypedHandoffPreparedEvent(fromAgent, toAgent, sessionID string) TypedEvent {
	return newTypedEvent(EventTypeHandoffPrepared, SourceCLI, HandoffPreparedData{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		SessionID: sessionID,
	})
}

// NewTypedHandoffExecutedEvent creates a v3 "agent.handoff_executed" TypedEvent.
// Source: cli (emitted by `ari handoff execute`).
// artifacts may be an empty slice; it is always serialized (never omitted).
func NewTypedHandoffExecutedEvent(fromAgent, toAgent, sessionID string, artifacts []string) TypedEvent {
	if artifacts == nil {
		artifacts = []string{}
	}
	return newTypedEvent(EventTypeHandoffExecuted, SourceCLI, HandoffExecutedData{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		SessionID: sessionID,
		Artifacts: artifacts,
	})
}

// NewTypedArtifactPromotedEvent creates a v3 "artifact.promoted" TypedEvent.
// Source: cli (emitted by `ari session wrap --auto-promote`).
func NewTypedArtifactPromotedEvent(sessionID, sourcePath, shelfPath, category string) TypedEvent {
	return newTypedEvent(EventTypeArtifactPromoted, SourceCLI, ArtifactPromotedData{
		SessionID:  sessionID,
		SourcePath: sourcePath,
		ShelfPath:  shelfPath,
		Category:   category,
	})
}

// --- Future event constructors (no current producers) ---

// NewTypedFieldUpdatedEvent creates a v3 "field.updated" TypedEvent.
// Source: cli (future `ari session field-set` command).
// oldValue and newValue may be nil, string, number, bool, or any JSON-serializable type.
func NewTypedFieldUpdatedEvent(sessionID, key string, oldValue, newValue any) TypedEvent {
	return newTypedEvent(EventTypeFieldUpdated, SourceCLI, FieldUpdatedData{
		SessionID: sessionID,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  newValue,
	})
}

// NewTypedHookFiredEvent creates a v3 "hook.fired" TypedEvent for observability.
// Source: hook (future hook runner observability implementation).
// eventType is the CC hook event name (e.g., "PostToolUse", "SessionStart").
func NewTypedHookFiredEvent(hookName, eventType string) TypedEvent {
	return newTypedEvent(EventTypeHookFired, SourceHook, HookFiredData{
		HookName:  hookName,
		EventType: eventType,
	})
}
