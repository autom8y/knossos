---
type: handoff
from_rite: arch
to_rite: 10x-dev
initiative: project-panopticon
phase: 2
sprint: 5
date: 2026-03-27
status: ready-for-execution
---

# BC-03 Boundary Extraction Design: `internal/citation`

## Problem

`internal/reason/response/generator.go` imports `internal/slack/streaming` solely to call
`streaming.ExtractCitations(text)` at line 377. This creates a dependency from the reason
layer into the slack layer — opposite to the intended dependency direction.

The boundary intent is explicitly documented at `internal/reason/response/stream.go:172`:
```
// streaming.ExtractCitations() -- not done here to avoid importing slack/.
```

The violation is pre-existing (not introduced by the Panopticon delivery) and is the
sole remaining structural issue preventing an A grade in Correctness.

## Current State

**Violation site**: `internal/reason/response/generator.go:12,377`
```go
import "github.com/autom8y/knossos/internal/slack/streaming"  // line 12
qualifiedNames := streaming.ExtractCitations(text)              // line 377
```

**Function location**: `internal/slack/streaming/citations.go`
- `ExtractCitations(text string) []string` — pure function, stdlib-only (`regexp`)
- `citationPattern` — compiled regex matching `[org::repo::domain]` markers
- 30 lines total, zero internal imports

**Callers** (exhaustive):
1. `internal/reason/response/generator.go:377` — the violation
2. `internal/slack/streaming/sender_test.go:279` — test within streaming package

**Test coverage**: `TestExtractCitations` at `sender_test.go:239` — 8 sub-cases covering
empty input, no citations, single, multiple, duplicates, mixed text, and edge cases.

## Design

### New Package: `internal/citation/`

Create a new leaf package owning citation marker parsing.

**`internal/citation/citation.go`**:
```go
// Package citation provides platform-wide citation marker parsing.
// This is a LEAF package — it imports only stdlib.
package citation

import "regexp"

// citationPattern matches inline citations in free-form text.
// Format: [org::repo::domain] e.g., [autom8y::knossos::architecture]
var citationPattern = regexp.MustCompile(`\[([a-zA-Z0-9_-]+::[a-zA-Z0-9_-]+::[a-zA-Z0-9_-]+)\]`)

// ExtractCitations parses inline citation markers from free-form text.
// Returns deduplicated qualified names in order of first appearance.
func ExtractCitations(text string) []string {
    matches := citationPattern.FindAllStringSubmatch(text, -1)
    seen := make(map[string]bool)
    var citations []string
    for _, m := range matches {
        if len(m) >= 2 && !seen[m[1]] {
            seen[m[1]] = true
            citations = append(citations, m[1])
        }
    }
    return citations
}
```

**`internal/citation/citation_test.go`**: Move `TestExtractCitations` from
`sender_test.go` (re-package as `package citation`). All 8 sub-cases preserved.

### Properties

| Property | Value |
|----------|-------|
| Internal imports | 0 (leaf package) |
| Stdlib imports | `regexp` |
| Exported symbols | `ExtractCitations` |
| Unexported symbols | `citationPattern` |
| Test functions | 1 (`TestExtractCitations` with 8 sub-cases) |

### Design Decision: Delete vs Delegate

**Option A (Recommended): Delete `streaming/citations.go`**
- Remove the file entirely from `internal/slack/streaming/`
- Move `TestExtractCitations` to `internal/citation/citation_test.go`
- Remove the test from `sender_test.go`
- Cleanest result: no dead code, no indirection, single source of truth

**Option B: Delegate**
- Keep `streaming.ExtractCitations` as a thin wrapper calling `citation.ExtractCitations`
- Preserves the streaming package's public API
- Adds unnecessary indirection

**Decision: Option A.** There are exactly two callers. Neither is outside the project.
The streaming package has no external consumers. Dead delegation is an anti-pattern
per project conventions ("Avoid backwards-compatibility hacks").

## Migration Steps (for Sprint 6 execution)

### Step 1: Create `internal/citation/`
1. Create `internal/citation/citation.go` with the function and regex (as above)
2. Create `internal/citation/citation_test.go` — copy test cases from `sender_test.go:239-288`
3. Verify: `CGO_ENABLED=0 go test ./internal/citation/...`

### Step 2: Update `internal/reason/response/generator.go`
1. Replace import `"github.com/autom8y/knossos/internal/slack/streaming"` with `"github.com/autom8y/knossos/internal/citation"`
2. Change line 377: `streaming.ExtractCitations(text)` → `citation.ExtractCitations(text)`
3. Verify no other `streaming.` references remain in the file

### Step 3: Remove `internal/slack/streaming/citations.go`
1. Delete `internal/slack/streaming/citations.go`
2. Remove `TestExtractCitations` from `sender_test.go` (it now lives in `citation_test.go`)
3. Verify: `CGO_ENABLED=0 go test ./internal/slack/streaming/...`

### Step 4: Verify boundary
1. Run: `CGO_ENABLED=0 go build ./...`
2. Run: `CGO_ENABLED=0 go test ./internal/citation/... ./internal/reason/response/... ./internal/slack/streaming/...`
3. Confirm: `go list -f '{{.Imports}}' ./internal/reason/response/` does NOT contain `slack/streaming`

## Import Graph Impact

### Before
```
internal/reason/response → internal/slack/streaming  (VIOLATION)
```

### After
```
internal/reason/response → internal/citation  (CLEAN — leaf package)
internal/slack/streaming  (no longer exports ExtractCitations)
```

No new boundary violations introduced. `internal/citation` is a leaf package with
zero internal imports — it cannot create transitive dependency issues.

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Hidden callers of `streaming.ExtractCitations` | Very Low | Exhaustive grep confirms exactly 2 callers |
| Test regression from moving test file | Very Low | Copy test cases verbatim; both packages use same test patterns |
| Future consumers wanting streaming-specific citation behavior | Low | If needed, streaming can import citation (correct direction) |

## Scope Boundary

This design covers ONLY WS-III-E (BC-03 boundary violation). Other Sprint 6 workstreams
(WS-III-A through WS-III-D) are independent and specified in the shape file.
