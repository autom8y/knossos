# Playbook: Refactoring & Code Quality

> Systematic code improvement without changing behavior

## When to Use

- Improving code structure without adding features
- Consolidating duplicate code (DRY violations)
- Simplifying complex functions or modules
- Updating patterns to modern conventions
- Preparing codebase for new feature work

## Prerequisites

- Working tests (or commitment to add them first)
- Clear scope boundaries
- No concurrent feature work in affected areas

## Command Sequence

### Phase 1: Initialize

```bash
/hygiene
```
**Expected output**: Team switched to hygiene-pack, roster displayed
**Decision point**: If just linting/formatting, consider direct fix.

### Phase 2: Start Session

```bash
/start "Refactoring scope" --complexity=MODULE
```
**Expected output**: Session created, context established
**Decision point**: Adjust complexity level:
- SPOT: Single function or class
- MODULE: Component or service
- CODEBASE: Cross-cutting patterns

### Phase 3: Code Audit

Code Auditor assesses current state.

**Expected output**: smell-report identifying issues with severity
**Decision point**: Prioritize by impact/effort ratio.

### Phase 4: Refactoring Design

```bash
/handoff refactoring-specialist
```
**Expected output**: refactoring-plan with sequenced changes
**Decision point**: Confirm no behavioral changes planned.

### Phase 5: Implementation

Refactoring Specialist executes changes.

**Expected output**: Code changes with preserved test suite
**Decision point**: Run tests after each logical change.

### Phase 6: Validation

```bash
/handoff quality-guardian
```
**Expected output**: Before/after metrics, test confirmation
**Decision point**: Verify no regressions introduced.

### Phase 7: Finalize

```bash
/wrap
```
**Expected output**: Session summary, quality metrics improvement

### Phase 8: Ship

```bash
/pr
```
**Expected output**: Pull request with refactoring rationale

## Variations

- **SPOT complexity**: Skip formal audit, direct fix
- **Breaking changes**: Document migration path for consumers
- **Performance refactoring**: Add benchmarks before/after

## Success Criteria

- [ ] All tests pass (no behavioral changes)
- [ ] Code smells addressed
- [ ] Metrics improved (complexity, duplication, etc.)
- [ ] PR approved by team

## Rollback

If refactoring breaks something:
```bash
git stash                      # Preserve attempted changes
git checkout .                 # Reset to clean state
/continue                      # Reassess scope
```
