---
description: "Staleness Detection Mode companion for doc-reviews skill."
---

# Staleness Detection Mode

> Focused analysis of documentation freshness by cross-referencing document modification dates with related code changes.

## Invocation

```
/doc-audit --staleness [path]
```

**Options:**
- `path` (optional): Scope detection to specific directory (default: entire repository)
- `--threshold=<days>`: Flag docs unchanged for N days after related code changed (default: 30)
- `--output=<format>`: Report format: `summary`, `detailed`, or `json` (default: `detailed`)

## Detection Logic

Staleness detection operates through three analysis passes:

**1. Temporal Correlation**
- Extract last-modified timestamp for each doc file (via git log)
- Identify related code files through:
  - Explicit references (imports, file paths mentioned in doc)
  - Directory proximity (docs in same module as code)
  - Naming conventions (e.g., `auth.md` relates to `auth.py`, `auth/`)
- Compare doc modification date to most recent related code change
- Flag when: `code_last_modified - doc_last_modified > threshold`

**2. Reference Validation**
- Parse docs for code element references (function names, class names, file paths)
- Verify each reference exists in current codebase
- Flag docs referencing:
  - Deleted files or directories
  - Renamed functions/classes (detected via git history)
  - Deprecated APIs (marked with deprecation decorators/comments)
- Severity: Critical when doc describes non-existent behavior

**3. Semantic Drift Detection**
- For docs with inline code examples, extract code snippets
- Compare against current implementation signatures
- Flag mismatches in:
  - Function signatures (changed parameters, return types)
  - Configuration keys (renamed or removed settings)
  - CLI flags (changed command-line interfaces)

## Staleness Categories

| Category | Criteria | Severity |
|----------|----------|----------|
| **Orphaned** | References deleted code elements | Critical |
| **Stale** | Related code changed significantly after doc update | High |
| **Drifted** | Code examples no longer match implementation | Medium |
| **Aging** | No updates in 6+ months, code stable | Low |

## Example Usage

```bash
# Full repository staleness audit
/doc-audit --staleness

# Scope to API documentation
/doc-audit --staleness docs/api/

# Strict threshold with JSON output
/doc-audit --staleness --threshold=14 --output=json
```

## Example Output

```
STALENESS AUDIT REPORT
======================
Scope: docs/
Threshold: 30 days
Analyzed: 47 documents

CRITICAL (Orphaned) - 2 files
─────────────────────────────
docs/api/oauth1-guide.md
  References: auth/oauth1.py (DELETED in commit abc123, 2024-03-15)
  Evidence: Line 23 documents `oauth1_token_exchange()` which no longer exists

docs/setup/legacy-config.md
  References: config/legacy.yaml (RENAMED to config/v1-compat.yaml)
  Evidence: File path on line 8 points to non-existent location

HIGH (Stale) - 5 files
──────────────────────
docs/architecture/auth-flow.md
  Last updated: 2024-01-10
  Related code changed: 2024-06-22 (163 days drift)
  Changed files: auth/flow.py (+127/-89), auth/tokens.py (+45/-12)

[...]

SUMMARY
───────
Critical: 2 (4%)
High: 5 (11%)
Medium: 8 (17%)
Low: 12 (26%)
Current: 20 (43%)
```
