# SPIKE: Sails Architectural Review

**Date**: 2026-01-08
**Timebox**: 30 minutes
**Status**: Complete

## Question

Is the White Sails architecture correctly designed? Specifically:
1. Should `sails` commands require project context?
2. Is the current WHITE_SAILS.yaml placement (session-level) correct?
3. Are there architectural smells in the implementation?

## Context

User observed that `sails` is marked as "needsProject = false" in the CLI, but sails functionality is inherently session-based. This raised the question of whether this is an architectural smell.

## Approach

1. Reviewed `internal/cmd/sails/check.go` for CLI behavior
2. Reviewed `internal/sails/gate.go` for CheckGate implementation
3. Reviewed `internal/sails/generator.go` for WHITE_SAILS.yaml generation
4. Analyzed WHITE_SAILS.yaml schema and existing files
5. Compared with session lifecycle in Knossos doctrine

## Findings

### 1. Current CLI Behavior

The `sails check` command has **two modes**:

```go
// From check.go lines 66-79
if len(args) > 0 {
    // Check specified path - can be any path
    result, err = sails.CheckGate(args[0])
} else {
    // Check current session - REQUIRES project
    if projectDir == "" {
        return errors.New(errors.CodeProjectNotFound,
            "no project directory specified and none discovered")
    }
    result, err = sails.CheckGateForCurrentSession(projectDir)
}
```

**Key insight**: The command already errors when no project is provided AND no explicit path is given.

### 2. Session Dependency Analysis

| Component | Session Required? | Reason |
|-----------|------------------|--------|
| WHITE_SAILS.yaml location | Yes | Lives in `.claude/sessions/{session-id}/` |
| Generator | Yes | Reads SESSION_CONTEXT.md, session ID is in schema |
| CheckGateForCurrentSession | Yes | Reads `.current-session` to find active session |
| CheckGate (with path) | No | Can validate any WHITE_SAILS.yaml file |
| schema.json session_id pattern | Yes | Requires `session-YYYYMMDD-HHMMSS-hexid` format |

### 3. Use Cases for Path Mode

The explicit path mode (`ari sails check /path/to/WHITE_SAILS.yaml`) enables:

1. **CI/CD integration** - Validate sails files without project discovery
2. **Multi-worktree checks** - Check sails from another worktree
3. **Archive validation** - Validate archived session sails
4. **QA workflows** - Check sails from a different session than current

### 4. Schema Confirms Session Binding

The WHITE_SAILS schema (lines 26-29) **requires** a session ID:

```json
"session_id": {
    "type": "string",
    "pattern": "^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$",
    "description": "Session identifier"
}
```

This confirms sails are fundamentally session-bound artifacts.

## Assessment

| Question | Answer | Rationale |
|----------|--------|-----------|
| Is `needsProject=false` a smell? | **Partially** | The common case (no path arg) already requires project and errors appropriately |
| Should we change it to `needsProject=true`? | **No** | Would break valid CI/CD use cases for explicit paths |
| Is the design sound? | **Yes** | The implementation correctly handles both cases |

### Not a Smell Because:

1. **Graceful degradation** - Command errors with clear message when project needed but not found
2. **Explicit override works** - Providing a path bypasses project requirement intentionally
3. **Schema enforces session binding** - Even standalone checks validate session_id format
4. **Use cases are valid** - CI/CD and cross-worktree scenarios need path mode

### Minor Improvement Opportunity

The help text could be clearer about when project context is needed:

```
Current: [session-path] - A session directory or direct path to WHITE_SAILS.yaml
Better:  [session-path] - If omitted, uses current project session (requires project context)
```

## Comparison Matrix

| Approach | Project Required | Path Override | CI/CD Friendly | Smell? |
|----------|-----------------|---------------|----------------|--------|
| Current (needsProject=false) | When no path | Yes | Yes | No |
| Alternative (needsProject=true) | Always | Breaks | No | Yes (over-constrains) |

## Recommendation

**Keep current design.** The implementation is correct:

1. `needsProject=false` is appropriate because the command can operate in standalone mode
2. The implementation correctly errors when project context is needed but missing
3. The explicit path mode serves valid use cases (CI/CD, cross-worktree, QA)

### Follow-up Actions

| Action | Priority | Description |
|--------|----------|-------------|
| None required | - | Current design is sound |
| Optional: Improve help text | LOW | Clarify when project context is required |

## Related

- `internal/cmd/sails/sails.go` - Annotation set to `needsProject=false`
- `internal/sails/gate.go` - Core gate checking logic
- `internal/validation/schemas/white-sails.schema.json` - Schema with session_id requirement
- `docs/philosophy/knossos-doctrine.md` - White Sails doctrine

---
*Generated by /spike on 2026-01-08*
