# QA Test Summary: Ariadne Team Domain

**Date**: 2026-01-04
**Tester**: QA Adversary Agent
**Build**: Static build with CGO_ENABLED=0, Go 1.22.3
**Recommendation**: **NO-GO** - Critical defects block release

---

## Executive Summary

The Ariadne Team Domain implementation has **3 Critical** and **2 High** severity defects that must be resolved before proceeding to the manifest domain. The switch command is completely unusable due to a flag conflict panic, and exit codes are not propagated correctly making the tool unreliable for scripted automation.

---

## Test Execution Summary

| Category | Passed | Failed | Skipped |
|----------|--------|--------|---------|
| Unit Tests (go test) | 22 | 0 | 0 |
| Race Detection Tests | 22 | 0 | 0 |
| CLI Integration Tests | 6 | 8 | 2 |
| Edge Case Tests | 4 | 0 | 0 |

---

## Defect Report

### D-001: [CRITICAL] Switch Command Panics on Any Invocation

**Severity**: Critical
**Priority**: P0 - Blocker
**Component**: `internal/cmd/team/switch.go`, `internal/cmd/root/root.go`

**Description**: The `ari team switch` command panics with "unable to redefine 'p' shorthand" on any invocation. The global flag `--project-dir` uses shorthand `-p` and the switch command's `--promote-all` also uses `-p`, causing a cobra flag conflict.

**Reproduction Steps**:
```bash
/tmp/ari_static team switch rnd-pack --output json
```

**Expected**: Switch to rnd-pack team or appropriate error
**Actual**:
```
panic: unable to redefine 'p' shorthand in "switch" flagset: it's already used for "promote-all" flag
```

**Impact**: The entire switch functionality is unusable. Users cannot switch teams.

**Fix**: Change `--promote-all` shorthand from `-p` to a different letter (suggestion: no shorthand, or `-P` uppercase).

---

### D-002: [CRITICAL] Exit Codes Not Propagated for Errors

**Severity**: Critical
**Priority**: P0 - Blocker
**Component**: `internal/cmd/team/*.go`, `internal/output/output.go`

**Description**: When errors occur, the CLI returns exit code 0 instead of the appropriate error code per TDD Section 4.1. The `PrintError()` function returns `nil` and commands return this value instead of the original error.

**Reproduction Steps**:
```bash
/tmp/ari_static team status --team=nonexistent-team --output json
echo $?  # Shows 0, should show 6
```

**Expected Exit Codes per TDD**:
| Error | Expected | Actual |
|-------|----------|--------|
| TEAM_NOT_FOUND | 6 | 0 |
| FILE_NOT_FOUND | 6 | 0 |
| ORPHAN_CONFLICT | 5 | N/A (switch broken) |
| PROJECT_NOT_FOUND | 9 | 9 (correct) |

**Impact**: Automation scripts cannot detect errors. CI/CD pipelines will falsely succeed.

**Fix**: Commands must return the error after PrintError, not `nil`. Alternatively, PrintError should return the error passed to it.

---

### D-003: [CRITICAL] Manifest Schema Mismatch - Orphan Detection Broken

**Severity**: Critical
**Priority**: P0 - Blocker
**Component**: `internal/team/manifest.go`

**Description**: The Go code expects JSON field `"team"` for agent origin, but the existing production manifest uses `"origin"` field. This causes:
1. All agents appear to have empty team field
2. DetectOrphans() incorrectly flags all agents as orphans
3. Team status shows active team's agents as orphans

**Evidence**:
```json
// Actual manifest on disk
"architect.md": {"source": "team", "origin": "10x-dev-pack", ...}

// Go struct expects
Team string `json:"team,omitempty"`

// Result: team="" (empty), all agents flagged as orphans
```

**Reproduction**:
```bash
/tmp/ari_static team status --output json | jq '.orphans'
# Returns all 5 agents as orphans even though they're from active team
```

**Impact**:
- False orphan detection blocks legitimate team switches
- Manifest sync validation always fails
- Data integrity compromised

**Fix**: Either:
1. Change JSON tag from `"team"` to `"origin"` to match existing manifests
2. OR add migration logic to handle both field names

---

### D-004: [HIGH] Validate Agent Count Message Corrupted

**Severity**: High
**Priority**: P1

**Component**: `internal/team/validate.go:195`

**Description**: The validation message "All N agent files present" displays corrupted output because of incorrect integer-to-string conversion using `string(rune(len(expectedAgents)))`.

**Code**:
```go
Message: "All " + string(rune(len(expectedAgents))) + " agent files present"
// When len=4, rune(4) = U+0004 (control character), not "4"
```

**Reproduction**:
```bash
/tmp/ari_static team validate --output json
# Shows: "All \u0004 agent files present"

/tmp/ari_static team validate --output text
# Shows: "All  agent files present" (blank where number should be)
```

**Fix**: Use `fmt.Sprintf` or `strconv.Itoa`:
```go
Message: fmt.Sprintf("All %d agent files present", len(expectedAgents))
```

---

### D-005: [HIGH] Manifest Sync Validation Always Warns

**Severity**: High
**Priority**: P1

**Component**: `internal/team/validate.go`

**Description**: Due to D-003 (schema mismatch), `MANIFEST_SYNC` check always produces a warning "Installed agent count differs from manifest" even when manifests are correctly synced.

**Reproduction**:
```bash
/tmp/ari_static team validate --output json | jq '.checks[] | select(.check=="MANIFEST_SYNC")'
# Returns: {"check":"MANIFEST_SYNC","status":"warn","message":"Installed agent count differs from manifest"}
```

**Root Cause**: Cascading from D-003. Fix D-003 first.

---

## Tests Passed

| Test | Result | Notes |
|------|--------|-------|
| `team list --output json` | PASS | Returns correct team data |
| `team list --format name-only` | PASS | Returns team names |
| `team status` (active team) | PASS* | Works but shows false orphans |
| `team validate` (structure checks) | PASS | TEAM_EXISTS, AGENTS_DIR, WORKFLOW_YAML pass |
| Project not found handling | PASS | Correct exit code 9 |
| Validation failed exit code | PASS | Correct exit code 12 |
| Injection in team name | PASS | Safely rejected |
| Path traversal attempt | PASS | Safely rejected |
| Long team name | PASS | Graceful handling |
| Race detector tests | PASS | No races detected |

---

## Tests Failed

| Test | Expected | Actual | Defect |
|------|----------|--------|--------|
| `team switch` any args | Switch or error | Panic | D-001 |
| Exit code for TEAM_NOT_FOUND | 6 | 0 | D-002 |
| Exit code for FILE_NOT_FOUND | 6 | 0 | D-002 |
| Orphan detection accuracy | Empty for active team | 5 false positives | D-003 |
| Agent count in validate msg | "All 4 agent files" | "All \u0004 agent files" | D-004 |
| MANIFEST_SYNC validation | PASS | WARN | D-005 |
| Switch with orphan flags | Execute switch | Panic before flag parse | D-001 |
| Dry-run switch | Preview changes | Panic | D-001 |

---

## Tests Not Executed

| Test | Reason |
|------|--------|
| Switch --remove-all behavior | Blocked by D-001 |
| Switch --keep-all behavior | Blocked by D-001 |
| Switch --promote-all behavior | Blocked by D-001 |
| Transaction rollback on failure | Blocked by D-001 |
| Concurrent switch operations | Blocked by D-001 |
| CLAUDE.md satellite updates | Blocked by D-001 |
| Backup/restore mechanism | Blocked by D-001 |

---

## Security Assessment

| Check | Status | Notes |
|-------|--------|-------|
| Path traversal | PASS | Team names with `..` are rejected |
| Command injection | PASS | Shell metacharacters in team names are safe |
| Input length limits | PASS | Long inputs handled gracefully |
| Sensitive data exposure | PASS | No credentials in output |

---

## Integration Gap Assessment

1. **Unit tests use synthetic data**: Tests pass because `manifest_test.go` creates manifests with correct `"team"` field. Real manifests use `"origin"`.

2. **No CLI integration tests**: The `internal/cmd/team/` package has no test files. Cobra command construction is untested.

3. **Missing testdata for broken scenarios**: `testdata/teams/broken-team/` has no agents directory, making it not discoverable by the real teams/ path.

4. **Schema backward compatibility**: No tests verify compatibility with existing bash-generated manifests.

---

## Recommendation: NO-GO

**Do NOT proceed to manifest domain** until the following are resolved:

### Required for Remediation (must fix)
- [ ] D-001: Fix flag shorthand conflict (1-2 hours)
- [ ] D-002: Fix exit code propagation (1-2 hours)
- [ ] D-003: Fix manifest schema mismatch (2-4 hours)
- [ ] D-004: Fix agent count string conversion (5 minutes)

### Recommended Improvements
- [ ] Add CLI integration tests for cmd/team package
- [ ] Add backward compatibility test with real manifest
- [ ] Add testdata fixture with production-like manifest

---

## Documentation Impact

- [ ] No documentation changes needed
- [x] Existing docs remain accurate
- [ ] Doc updates needed: None identified
- [ ] doc-team-pack notification: NO

---

## Security Handoff

- [x] Not applicable (TRIVIAL/ALERT complexity)

---

## SRE Handoff

- [x] Not applicable (TRIVIAL/ALERT/FEATURE complexity)

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| Test Summary | `/Users/tomtenuta/Code/roster/docs/analysis/QA-ariadne-team-domain.md` | Write |
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-ariadne-rite.md` | Read |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-ariadne.md` | Read |
| Switch command | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/team/switch.go` | Read |
| Root command | `/Users/tomtenuta/Code/roster/ariadne/internal/cmd/root/root.go` | Read |
| Manifest | `/Users/tomtenuta/Code/roster/ariadne/internal/team/manifest.go` | Read |
| Validate | `/Users/tomtenuta/Code/roster/ariadne/internal/team/validate.go` | Read |
| Discovery | `/Users/tomtenuta/Code/roster/ariadne/internal/team/discovery.go` | Read |
| Output | `/Users/tomtenuta/Code/roster/ariadne/internal/output/output.go` | Read |
| Errors | `/Users/tomtenuta/Code/roster/ariadne/internal/errors/errors.go` | Read |

---

*Report generated by QA Adversary Agent*
