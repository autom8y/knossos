# Migration Runbook: Permission Alignment (PATCH)

> **Complexity**: PATCH (no user action required)
> **Date**: 2025-12-27
> **Affects**: hygiene, intelligence, rnd, security

## Problem Statement

Agent configurations specified artifact production requirements in their instructions but lacked the corresponding tool permissions to fulfill those requirements. This created a gap where agents were instructed to produce documents but could not actually write them.

**Root Cause**: Disconnect between agent instructions ("Produce X using template Y") and tool permissions (missing `Write` or `Edit` tools).

**Discovery Method**: Permission alignment audit comparing agent instructions against tool permissions.

## Changes Made

| Team | Agent | File | Change |
|------|-------|------|--------|
| hygiene | code-smeller | `rites/hygiene/agents/code-smeller.md` | Added `Write` to tools |
| intelligence | user-researcher | `rites/intelligence/agents/user-researcher.md` | Added `Edit` to tools |
| rnd | technology-scout | `rites/rnd/agents/technology-scout.md` | Added `Edit` to tools |
| security | penetration-tester | `rites/security/agents/penetration-tester.md` | Added `Edit` to tools |
| security | threat-modeler | `rites/security/agents/threat-modeler.md` | Added `Edit` to tools |

### Before/After Examples

**code-smeller.md (line 28)**
```yaml
# Before
tools: Bash, Glob, Grep, Read, TodoWrite

# After
tools: Bash, Glob, Grep, Read, Write, TodoWrite
```

**penetration-tester.md (line 18)**
```yaml
# Before
tools: Bash, Glob, Grep, Read, Write, TodoWrite

# After
tools: Bash, Edit, Glob, Grep, Read, Write, TodoWrite
```

## Validation Steps

### 1. Verify Tool Permissions

Run this command from the roster root to confirm all modified agents have the correct tools:

```bash
# Check code-smeller has Write
grep "^tools:" rites/hygiene/agents/code-smeller.md | grep -q "Write" && echo "PASS: code-smeller" || echo "FAIL: code-smeller"

# Check user-researcher has Edit
grep "^tools:" rites/intelligence/agents/user-researcher.md | grep -q "Edit" && echo "PASS: user-researcher" || echo "FAIL: user-researcher"

# Check technology-scout has Edit
grep "^tools:" rites/rnd/agents/technology-scout.md | grep -q "Edit" && echo "PASS: technology-scout" || echo "FAIL: technology-scout"

# Check penetration-tester has Edit
grep "^tools:" rites/security/agents/penetration-tester.md | grep -q "Edit" && echo "PASS: penetration-tester" || echo "FAIL: penetration-tester"

# Check threat-modeler has Edit
grep "^tools:" rites/security/agents/threat-modeler.md | grep -q "Edit" && echo "PASS: threat-modeler" || echo "FAIL: threat-modeler"
```

Expected output: All lines show `PASS`.

### 2. Verify Agent Invocation

After syncing to a satellite, verify agents can produce artifacts:

```bash
# In satellite directory, invoke agent and confirm it can write
# Example: code-smeller should be able to produce smell-report.md
```

## Prevention: CI Validation

To prevent this class of bug, add a validation script that checks instruction-permission alignment:

### Recommended CI Check

Add to `justfile` or CI pipeline:

```bash
#!/bin/bash
# validate-agent-permissions.sh
# Checks that agents with artifact production instructions have Write/Edit tools

ERRORS=0

for agent in rites/*/agents/*.md; do
  # Check if agent produces artifacts
  if grep -q "Produce.*using\|produces.*\.md\|Artifact Production" "$agent"; then
    # Must have Write or Edit
    if ! grep "^tools:" "$agent" | grep -qE "(Write|Edit)"; then
      echo "ERROR: $agent produces artifacts but lacks Write/Edit permission"
      ERRORS=$((ERRORS + 1))
    fi
  fi
done

if [ $ERRORS -gt 0 ]; then
  echo "Found $ERRORS permission alignment issues"
  exit 1
fi

echo "All agent permissions aligned"
```

### Template Validation

Update `@team-development/templates/agent-template.md` Tool Selection Guide to include:

| Use Case | Tools |
|----------|-------|
| Produces any artifacts | Must include `Write` or `Edit` |

## Rollback

If issues arise, revert to previous tool configurations:

```bash
# Rollback all changes
git checkout HEAD~1 -- \
  rites/hygiene/agents/code-smeller.md \
  rites/intelligence/agents/user-researcher.md \
  rites/rnd/agents/technology-scout.md \
  rites/security/agents/penetration-tester.md \
  rites/security/agents/threat-modeler.md
```

**Risk Assessment**: Low. Adding permissions is additive; removing them could break agent functionality.

## Compatibility Matrix

| Roster Version | CEM Version | Status |
|----------------|-------------|--------|
| Pre-patch | Any | Agents may fail when attempting artifact production |
| Post-patch | Any | Full functionality restored |

**Note**: This is a roster-only change. No CEM or skeleton updates required.

## Summary

This PATCH-level migration adds missing tool permissions to 5 agents across 4 rites. No user action is required for satellites that sync from roster. The validation script prevents future occurrences.
