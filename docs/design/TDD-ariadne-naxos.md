# TDD: Ariadne Naxos Domain

> Technical Design Document for the naxos domain of the Ariadne Go CLI

**Status**: Approved
**Author**: Architect Agent
**Date**: 2026-01-07
**PRD**: docs/requirements/PRD-ariadne.md

---

## 1. Overview

This Technical Design Document specifies the implementation of the **naxos domain** for Ariadne (`ari`), the Go binary that provides session lifecycle management for the Knossos platform. The naxos domain provides cleanup tooling for identifying abandoned or orphaned sessions that may need user attention.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-ariadne.md` |
| Session TDD | `docs/design/TDD-ariadne-session.md` |
| Implementation | `internal/cmd/naxos/`, `internal/naxos/` |

### 1.2 Naming Context

**Naxos** is named after the island in Greek mythology where Theseus abandoned Ariadne after she helped him escape the Labyrinth. In the Knossos platform, Naxos represents the cleanup mechanism that identifies sessions that have been "abandoned" -- left inactive, parked with stale status indicators, or started but never completed.

### 1.3 Scope

**In Scope**:
- `ari naxos scan` command for detecting orphaned sessions
- Detection criteria: inactive sessions, stale gray sails, incomplete wraps
- Report generation with suggested actions
- Configurable thresholds for detection sensitivity

**Out of Scope**:
- Automatic cleanup or deletion (Naxos is report-only)
- Session repair or modification
- Archive management beyond scanning

### 1.4 Design Goals

1. **Report-Only**: Naxos identifies problems but never automatically deletes or modifies sessions
2. **Configurable Sensitivity**: Thresholds are adjustable to match workflow patterns
3. **Actionable Output**: Suggestions guide users toward appropriate remediation
4. **Non-Invasive**: Scanning does not acquire locks or modify any state

---

## 2. Architecture

### 2.1 Package Structure

```
ariadne/
├── internal/
│   ├── cmd/
│   │   └── naxos/
│   │       ├── naxos.go          # Parent command registration
│   │       └── scan.go           # ari naxos scan implementation
│   └── naxos/
│       ├── types.go              # Domain types (OrphanReason, SuggestedAction, etc.)
│       ├── scanner.go            # Session scanning logic
│       ├── scanner_test.go       # Scanner unit tests
│       ├── report.go             # Output formatting (ScanOutput)
│       └── report_test.go        # Report formatting tests
```

### 2.2 Dependency Graph

```
                    ┌─────────────────────────────┐
                    │  internal/cmd/naxos/        │
                    │  (CLI layer)                │
                    └─────────────┬───────────────┘
                                  │
                                  v
                    ┌─────────────────────────────┐
                    │  internal/naxos/            │
                    │  (domain logic)             │
                    └─────────────┬───────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
              v                   v                   v
        ┌───────────┐      ┌───────────┐      ┌───────────┐
        │ paths/    │      │ session/  │      │ output/   │
        │           │      │           │      │           │
        └───────────┘      └───────────┘      └───────────┘
```

---

## 3. Interface Contracts

### 3.1 Command Summary

| Command | Description | Requires Lock | Modifies State |
|---------|-------------|---------------|----------------|
| `naxos scan` | Scan for orphaned sessions | No | No |

### 3.2 Command: `ari naxos scan`

Scans the sessions directory for orphaned sessions that may need cleanup attention.

**Signature**:
```
ari naxos scan [--inactive-threshold=DURATION] [--stale-threshold=DURATION] [--include-archived]
```

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--inactive-threshold` | duration | `24h` | How long a session can be inactive before flagging |
| `--stale-threshold` | duration | `7d` (168h) | How long gray sails can persist before flagging |
| `--include-archived` | bool | `false` | Include archived sessions in scan |

**Global Flags** (inherited from root):

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | `text` | Output format: text, json |
| `--verbose` | `-v` | bool | `false` | Enable verbose output |
| `--project` | `-p` | string | (auto-detect) | Project root directory |

**Output (JSON)**:
```json
{
  "orphaned_sessions": [
    {
      "session_id": "session-20260102-120000-abcd1234",
      "session_dir": ".sos/sessions/session-20260102-120000-abcd1234",
      "status": "ACTIVE",
      "initiative": "Feature implementation",
      "reason": "INACTIVE",
      "suggested_action": "RESUME",
      "age": 432000000000000,
      "inactive_for": 172800000000000,
      "created_at": "2026-01-02T12:00:00Z",
      "last_activity": "2026-01-04T12:00:00Z",
      "additional_info": "2 days since last activity"
    }
  ],
  "total_scanned": 5,
  "total_orphaned": 1,
  "scanned_at": "2026-01-07T10:00:00Z",
  "by_reason": {
    "inactive": 1,
    "stale_sails": 0,
    "incomplete_wrap": 0
  },
  "config": {
    "inactive_threshold": "1 day",
    "stale_sails_threshold": "7 days",
    "include_archived": false
  }
}
```

**Output (text)**:
```
Naxos Session Scan Report
==================================================

Scanned: 5 sessions
Orphaned: 1 sessions

By Reason:
  [!] Inactive (>1 day): 1

Orphaned Sessions:
--------------------------------------------------

[!] session-20260102-120000-abcd1234
  Status: ACTIVE
  Initiative: Feature implementation
  Reason: Inactive for too long
  Inactive: 2 days
  Info: 2 days since last activity
  Suggested: ari session resume

--------------------------------------------------
Actions:
  To wrap:   ari session wrap --session <id>
  To resume: ari session resume <id>
  To delete: rm -rf .sos/sessions/<id>
```

**Exit Codes**:

| Code | Condition |
|------|-----------|
| 0 | Scan completed successfully (regardless of orphan count) |
| 1 | General error during scan |
| 9 | No .claude/ directory found (PROJECT_NOT_FOUND) |

---

## 4. Detection Criteria

### 4.1 Orphan Reasons

The scanner identifies three categories of orphaned sessions:

| Reason | Code | Description | Detection Logic |
|--------|------|-------------|-----------------|
| Inactive | `INACTIVE` | Session active but no recent activity | Status=ACTIVE AND (now - last_activity) > inactive_threshold |
| Stale Sails | `STALE_SAILS` | Parked session with gray sails past threshold | Status=PARKED AND sails_color=GRAY AND (now - parked_at) > stale_threshold |
| Incomplete Wrap | `INCOMPLETE_WRAP` | Wrap initiated but never completed | Status=ACTIVE AND current_phase="wrap" |

### 4.2 Last Activity Determination

The scanner determines last activity by examining timestamps in order of precedence:

1. `resumed_at` - Most recent resume timestamp
2. `parked_at` - Most recent park timestamp
3. `archived_at` - Archive timestamp
4. `created_at` - Session creation (fallback)

The most recent timestamp is used as the last activity time.

### 4.3 Sails Color Detection

For parked sessions, the scanner checks for a `sails.yaml` file in the session directory:

- If file doesn't exist: Treated as GRAY (unknown)
- If file exists: Parses for `color:` field
- Values: WHITE, BLACK, GRAY
- GRAY or missing sails trigger stale sails detection

### 4.4 Suggested Actions

Based on the orphan reason and session state, the scanner suggests one of three actions:

| Action | Code | When Suggested |
|--------|------|----------------|
| Wrap | `WRAP` | Incomplete wrap sessions; parked sessions without explicit reason |
| Resume | `RESUME` | Recently inactive sessions (<30 days); parked sessions with explicit reason |
| Delete | `DELETE` | Very old inactive sessions (>30 days) with no apparent progress |

**Action Descriptions** (shown in text output):

| Action | Command |
|--------|---------|
| WRAP | `ari session wrap` |
| RESUME | `ari session resume` |
| DELETE | `rm -rf <session-dir>` |

---

## 5. Data Model

### 5.1 ScanConfig

Configuration for the session scanner:

```go
type ScanConfig struct {
    // InactiveThreshold is how long a session can be inactive before flagging.
    // Default: 24 hours
    InactiveThreshold time.Duration

    // StaleSailsThreshold is how long gray sails can persist before flagging.
    // Default: 7 days
    StaleSailsThreshold time.Duration

    // IncludeArchived controls whether to scan archived sessions.
    // Default: false
    IncludeArchived bool
}
```

### 5.2 OrphanedSession

Represents a session flagged for cleanup review:

```go
type OrphanedSession struct {
    SessionID       string          `json:"session_id"`
    SessionDir      string          `json:"session_dir"`
    Status          string          `json:"status"`
    Initiative      string          `json:"initiative"`
    Reason          OrphanReason    `json:"reason"`
    SuggestedAction SuggestedAction `json:"suggested_action"`
    Age             time.Duration   `json:"age"`
    InactiveFor     time.Duration   `json:"inactive_for"`
    CreatedAt       time.Time       `json:"created_at"`
    LastActivity    time.Time       `json:"last_activity"`
    SailsColor      string          `json:"sails_color,omitempty"`
    AdditionalInfo  string          `json:"additional_info,omitempty"`
}
```

### 5.3 ScanResult

Results of a session scan:

```go
type ScanResult struct {
    OrphanedSessions []OrphanedSession     `json:"orphaned_sessions"`
    TotalScanned     int                   `json:"total_scanned"`
    TotalOrphaned    int                   `json:"total_orphaned"`
    ScannedAt        time.Time             `json:"scanned_at"`
    Config           ScanConfig            `json:"config"`
    ByReason         map[OrphanReason]int  `json:"by_reason"`
}
```

### 5.4 Enumerations

**OrphanReason**:
```go
const (
    ReasonInactive       OrphanReason = "INACTIVE"
    ReasonStaleSails     OrphanReason = "STALE_SAILS"
    ReasonIncompleteWrap OrphanReason = "INCOMPLETE_WRAP"
)
```

**SuggestedAction**:
```go
const (
    ActionWrap   SuggestedAction = "WRAP"
    ActionResume SuggestedAction = "RESUME"
    ActionDelete SuggestedAction = "DELETE"
)
```

---

## 6. Scanner Implementation

### 6.1 Scanner Structure

```go
type Scanner struct {
    config   ScanConfig
    resolver *paths.Resolver
    now      func() time.Time  // Injectable for testing
}
```

### 6.2 Scan Algorithm

```
1. Initialize empty ScanResult with current config
2. Get sessions directory path from resolver
3. For each entry in sessions directory:
   a. Skip if not a directory
   b. Skip if not a session directory (doesn't match session-* pattern)
   c. Increment TotalScanned
   d. Load SESSION_CONTEXT.md
   e. Skip if context cannot be parsed
   f. Check for orphan conditions (see 4.1)
   g. If orphan detected, add to result
4. If --include-archived, repeat for archive directory
5. Return ScanResult
```

### 6.3 Directory Filtering

The scanner ignores:
- Non-directory entries (files)
- Hidden directories (`.locks`, `.audit`)
- Directories not matching `session-*` pattern
- Sessions with unparseable SESSION_CONTEXT.md

---

## 7. Output Formatting

### 7.1 Text Output Structure

The text output provides a human-readable report with:

1. **Header**: Report title and separator
2. **Summary**: Total scanned and orphaned counts
3. **Breakdown by Reason**: Count per reason category with symbol indicators
4. **Detailed List**: Each orphaned session with full details
5. **Actions Footer**: Quick reference for remediation commands

### 7.2 Reason Symbols

Visual indicators used in text output:

| Reason | Symbol | Meaning |
|--------|--------|---------|
| INACTIVE | `[!]` | Attention needed - session inactive |
| STALE_SAILS | `[~]` | Warning - status unclear |
| INCOMPLETE_WRAP | `[x]` | Error state - wrap incomplete |

### 7.3 Table Output

For tabular display (implements `output.Tabular`):

| Column | Content |
|--------|---------|
| SESSION ID | Session ID (truncated to 35 chars) |
| STATUS | Session status (ACTIVE, PARKED, etc.) |
| REASON | Symbol + reason code |
| INACTIVE | Human-readable duration |
| SUGGESTED ACTION | Action code |

---

## 8. Integration

### 8.1 Session Lifecycle Integration

Naxos integrates with the session lifecycle at the read-only level:

```
┌─────────────────────────────────────────────────────────┐
│                   Session Lifecycle                      │
│                                                         │
│  create ──▶ ACTIVE ──▶ park ──▶ PARKED ──▶ wrap        │
│              │                    │         │           │
│              │                    │         ▼           │
│              │                    │      ARCHIVED       │
│              │                    │                     │
│              ▼                    ▼                     │
│         ┌─────────────────────────────────┐            │
│         │          Naxos Scan             │            │
│         │    (read-only observation)      │            │
│         └─────────────────────────────────┘            │
└─────────────────────────────────────────────────────────┘
```

### 8.2 White Sails Integration

The sails system indicates session health:

| Sails Color | Meaning | Naxos Treatment |
|-------------|---------|-----------------|
| WHITE | Session completed successfully | Not flagged (healthy) |
| BLACK | Session failed or has issues | Not flagged (requires manual review) |
| GRAY | Status unknown/not determined | Flagged if past stale threshold |

### 8.3 Workflow Recommendations

**Periodic Cleanup**:
```bash
# Run weekly scan
ari naxos scan --json > scan-report.json

# Or with custom thresholds for more aggressive cleanup
ari naxos scan --inactive-threshold 12h --stale-threshold 3d
```

**Pre-Sprint Hygiene**:
```bash
# Before starting new work, check for abandoned sessions
ari naxos scan
```

---

## 9. Test Strategy

### 9.1 Unit Tests

Location: `internal/naxos/*_test.go`

| Test Category | Coverage Target |
|---------------|-----------------|
| Scanner logic | 100% |
| Reason detection | 100% |
| Suggested action logic | 100% |
| Duration formatting | 100% |
| Output formatting | 90% |

### 9.2 Test Scenarios

| Test ID | Scenario | Expected Result |
|---------|----------|-----------------|
| `scan_001` | No sessions exist | TotalScanned=0, TotalOrphaned=0 |
| `scan_002` | Healthy active session | Not flagged |
| `scan_003` | Inactive active session (>threshold) | Flagged as INACTIVE |
| `scan_004` | Parked with gray sails (>threshold) | Flagged as STALE_SAILS |
| `scan_005` | Parked with white sails | Not flagged |
| `scan_006` | Active with current_phase=wrap | Flagged as INCOMPLETE_WRAP |
| `scan_007` | Archived session (not included) | Not scanned |
| `scan_008` | Archived session (included) | Scanned, may be flagged |
| `scan_009` | Custom inactive threshold | Respects custom threshold |
| `scan_010` | Custom stale threshold | Respects custom threshold |
| `scan_011` | Non-session directories | Ignored |
| `scan_012` | Very old inactive (>30d) | Suggested action = DELETE |
| `scan_013` | Recent inactive (<30d) | Suggested action = RESUME |
| `scan_014` | Stale sails with park reason | Suggested action = RESUME |
| `scan_015` | Stale sails without reason | Suggested action = WRAP |

### 9.3 Test Fixtures

Tests use a temporary directory structure:

```
temp-dir/
└── .claude/
    └── sessions/
        ├── session-YYYYMMDD-HHMMSS-xxxxxxxx/
        │   ├── SESSION_CONTEXT.md
        │   └── sails.yaml (optional)
        └── ...
```

---

## 10. Error Handling

### 10.1 Error Conditions

| Condition | Behavior |
|-----------|----------|
| No .claude/ directory | Exit code 9 (PROJECT_NOT_FOUND) |
| Sessions directory doesn't exist | Return empty result (not an error) |
| Cannot read session directory | Skip session, continue scanning |
| Cannot parse SESSION_CONTEXT.md | Skip session, continue scanning |
| Cannot read sails.yaml | Treat as GRAY sails |

### 10.2 Graceful Degradation

The scanner is designed to be resilient:
- Missing directories are not errors
- Unparseable sessions are skipped silently
- The scan always completes with whatever data is available

---

## 11. Use Cases

### 11.1 Daily Hygiene Check

```bash
# Quick check for abandoned work
ari naxos scan

# If orphans found:
#   - Resume: ari session resume <id>
#   - Wrap:   ari session wrap --session <id>
#   - Delete: rm -rf .sos/sessions/<id>
```

### 11.2 Automated CI Check

```bash
#!/bin/bash
# Fail build if orphaned sessions exceed threshold
result=$(ari naxos scan --json)
orphan_count=$(echo "$result" | jq '.total_orphaned')

if [ "$orphan_count" -gt 5 ]; then
    echo "Warning: $orphan_count orphaned sessions found"
    echo "$result" | jq '.orphaned_sessions[].session_id'
fi
```

### 11.3 Team Workspace Audit

```bash
# Generate comprehensive report including archived sessions
ari naxos scan --include-archived --json > audit-report.json
```

---

## 12. ADRs

| ADR | Status | Topic |
|-----|--------|-------|
| N/A | - | No architectural decisions requiring ADR for this domain |

The naxos domain follows patterns established in the session domain TDD and requires no new architectural decisions.

---

## 13. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| False positives (healthy sessions flagged) | Medium | Low | Configurable thresholds; user makes final decision |
| Large session count impacts scan time | Low | Low | Linear scan is acceptable for typical session counts |
| sails.yaml format changes | Low | Low | Graceful fallback to GRAY on parse errors |

---

## 14. Handoff Criteria

Ready for Implementation when:

- [x] Scan command interface fully specified
- [x] Detection criteria documented
- [x] Output formats defined (JSON and text)
- [x] Test scenarios enumerated
- [x] Integration with session lifecycle documented
- [x] Error handling strategy defined

---

## 15. Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-naxos.md` | Written |
| Command Implementation | `/Users/tomtenuta/Code/roster/internal/cmd/naxos/naxos.go` | Read |
| Scan Command | `/Users/tomtenuta/Code/roster/internal/cmd/naxos/scan.go` | Read |
| Domain Types | `/Users/tomtenuta/Code/roster/internal/naxos/types.go` | Read |
| Scanner Implementation | `/Users/tomtenuta/Code/roster/internal/naxos/scanner.go` | Read |
| Report Formatting | `/Users/tomtenuta/Code/roster/internal/naxos/report.go` | Read |
| Scanner Tests | `/Users/tomtenuta/Code/roster/internal/naxos/scanner_test.go` | Read |
| Session TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-session.md` | Read |
