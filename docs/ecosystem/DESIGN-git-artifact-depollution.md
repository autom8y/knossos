# Context Design: Git Artifact Depollution

**Date**: 2026-02-10
**Author**: Context Architect (ecosystem rite)
**Status**: READY FOR IMPLEMENTATION
**Stakeholder Decisions**: 12 interview decisions, all resolved

## Problem Statement

The CE audit identified ~15 source files with ambient branch/worktree pressure that causes Claude to proactively create git branches and worktrees even when the user has not requested parallel work. The root cause is threefold:

1. **Hook Text() output** includes Git Branch and Base Branch as visible table rows, priming Claude to treat branch context as actionable state rather than background metadata.
2. **Dromena examples and output templates** embed branch names (`Branch: {branch}`, `Git: feature/auth-system`) that Claude mirrors in its output, normalizing branch creation as a default workflow step.
3. **Worktree promotion** in sprint, start, and hotfix commands positions worktrees as a primary workflow pattern rather than an opt-in advanced feature triggered only by explicit user keywords.

### Trigger Mechanism

When the SessionStart hook fires, the Text() output includes `| Git Branch | feature/xyz |` in the session context table. Every dromena command then sees this as conversational context. Combined with examples showing `Branch: {branch}` in output templates and entire sections dedicated to "Integration with Git" and "Parallel Sprint Pattern," Claude infers that branch management is an expected part of every workflow.

## Approach: Surgical Depollution (Selected)

**Alternative considered**: Remove all git context from the hook entirely (JSON fields too). Rejected because JSON fields are consumed programmatically by `ari session status` and worktree operations. The issue is Text() output entering Claude's conversational context, not the structured JSON payload.

**Selected approach**: Remove Git Branch and Base Branch from Text() output only. Keep JSON fields for programmatic consumers. Remove branch/worktree signals from all non-worktree dromena examples and output templates. Add explicit anti-pattern documentation.

### Backward Compatibility: COMPATIBLE

All changes are subtractive (removing ambient signals) or additive (adding anti-pattern docs). No schema changes. No pipeline changes. JSON API unchanged. The only behavioral shift: Claude stops proactively suggesting branches/worktrees unless the user explicitly asks.

## Edit Specifications

### Edit 1: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/context.go`

**Rationale**: Remove Git Branch and Base Branch from Text() method output. These fields prime Claude to treat branch context as actionable. JSON fields remain for programmatic consumers (`ari session status`, worktree operations). The struct fields `GitBranch` and `BaseBranch` stay because they are part of the JSON API contract. Only the human-readable Text() rendering is affected.

**Action**: Remove the Git Branch and Base Branch conditional blocks from the Text() method.

**Old text**:
```
	b.WriteString(fmt.Sprintf("| Mode | %s |\n", c.ExecutionMode))
	if c.GitBranch != "" {
		b.WriteString(fmt.Sprintf("| Git Branch | %s |\n", c.GitBranch))
	}
	if c.BaseBranch != "" {
		b.WriteString(fmt.Sprintf("| Base Branch | %s |\n", c.BaseBranch))
	}
	if len(c.AvailableRites) > 0 {
```

**New text**:
```
	b.WriteString(fmt.Sprintf("| Mode | %s |\n", c.ExecutionMode))
	if len(c.AvailableRites) > 0 {
```

---

### Edit 2: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/context_test.go`

**Rationale**: The test case "with git and ecosystem fields" asserts that Git Branch and Base Branch appear in Text() output. Since we removed those lines from Text(), the test must be updated. The JSON struct still carries these fields (verified by other tests), so we only need to remove the Text() assertions.

**Action**: Remove the Git Branch and Base Branch assertions from the "with git and ecosystem fields" test case.

**Old text**:
```
		{
			name: "with git and ecosystem fields",
			output: ContextOutput{
				SessionID:       "session-001",
				Status:          "ACTIVE",
				ExecutionMode:   "orchestrated",
				HasSession:      true,
				GitBranch:       "feat/backtick-migration",
				BaseBranch:      "main",
				AvailableRites:  []string{"10x-dev", "hygiene", "ecosystem"},
				AvailableAgents: []string{"orchestrator", "context-architect"},
			},
			contains: []string{
				"| Git Branch | feat/backtick-migration |",
				"| Base Branch | main |",
				"| Available Rites | 10x-dev, hygiene, ecosystem |",
				"| Available Agents | orchestrator, context-architect |",
			},
		},
```

**New text**:
```
		{
			name: "with git and ecosystem fields",
			output: ContextOutput{
				SessionID:       "session-001",
				Status:          "ACTIVE",
				ExecutionMode:   "orchestrated",
				HasSession:      true,
				GitBranch:       "feat/backtick-migration",
				BaseBranch:      "main",
				AvailableRites:  []string{"10x-dev", "hygiene", "ecosystem"},
				AvailableAgents: []string{"orchestrator", "context-architect"},
			},
			contains: []string{
				"| Available Rites | 10x-dev, hygiene, ecosystem |",
				"| Available Agents | orchestrator, context-architect |",
			},
		},
```

---

### Edit 3: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/context_test.go` (second location)

**Rationale**: The `TestContextOutput_Text_OmitsEmptyFields` test checks that empty Git Branch and Base Branch do not appear in output. Since we now NEVER render these fields in Text() regardless of value, these assertions are still correct (they will pass). However, we should also verify that non-empty values are now also omitted. Add a new test case for this.

**Action**: Replace the existing `TestContextOutput_Text_OmitsEmptyFields` function to also verify that populated git fields are excluded.

**Old text**:
```
func TestContextOutput_Text_OmitsEmptyFields(t *testing.T) {
	// Verify that empty git/ecosystem fields do not appear in output
	out := ContextOutput{
		SessionID:     "session-003",
		Status:        "ACTIVE",
		ExecutionMode: "cross-cutting",
		HasSession:    true,
		GitBranch:     "",
		BaseBranch:    "",
	}
	text := out.Text()
	if strings.Contains(text, "Git Branch") {
		t.Error("Text() should not contain Git Branch when empty")
	}
	if strings.Contains(text, "Base Branch") {
		t.Error("Text() should not contain Base Branch when empty")
	}
	if strings.Contains(text, "Available Rites") {
		t.Error("Text() should not contain Available Rites when nil")
	}
	if strings.Contains(text, "Available Agents") {
		t.Error("Text() should not contain Available Agents when nil")
	}
}
```

**New text**:
```
func TestContextOutput_Text_OmitsEmptyFields(t *testing.T) {
	// Verify that empty git/ecosystem fields do not appear in output
	out := ContextOutput{
		SessionID:     "session-003",
		Status:        "ACTIVE",
		ExecutionMode: "cross-cutting",
		HasSession:    true,
		GitBranch:     "",
		BaseBranch:    "",
	}
	text := out.Text()
	if strings.Contains(text, "Git Branch") {
		t.Error("Text() should not contain Git Branch when empty")
	}
	if strings.Contains(text, "Base Branch") {
		t.Error("Text() should not contain Base Branch when empty")
	}
	if strings.Contains(text, "Available Rites") {
		t.Error("Text() should not contain Available Rites when nil")
	}
	if strings.Contains(text, "Available Agents") {
		t.Error("Text() should not contain Available Agents when nil")
	}

	// Verify that populated git fields are also excluded from Text()
	// (git fields are only in JSON, never in Text() output)
	out2 := ContextOutput{
		SessionID:     "session-004",
		Status:        "ACTIVE",
		ExecutionMode: "orchestrated",
		HasSession:    true,
		GitBranch:     "feature/something",
		BaseBranch:    "main",
	}
	text2 := out2.Text()
	if strings.Contains(text2, "Git Branch") {
		t.Error("Text() should never contain Git Branch (removed from text output)")
	}
	if strings.Contains(text2, "Base Branch") {
		t.Error("Text() should never contain Base Branch (removed from text output)")
	}
}
```

---

### Edit 4: `/Users/tomtenuta/Code/knossos/mena/navigation/go.dro.md`

**Rationale**: The Scenario 1 (ALREADY_ACTIVE) output template includes `Branch: {branch}`, and the Scenario 4 (ORIENTATION) dashboard includes both `Worktree: {worktree_name or "main"}` and `Branch: {branch_name}`. These prime Claude to display branch state as a primary status signal and normalize worktree context. Remove `Branch:` from Scenario 1 and remove both `Worktree:` and `Branch:` lines from Scenario 4.

**Action**: Remove `Branch: {branch}` from Scenario 1 output template.

**Old text**:
```
Rite: {active_rite} | Phase: {current_phase} | Branch: {branch}
```

**New text**:
```
Rite: {active_rite} | Phase: {current_phase}
```

---

### Edit 5: `/Users/tomtenuta/Code/knossos/mena/navigation/go.dro.md` (second location)

**Rationale**: Remove git state collection command from Phase 1 parallel state collection. Claude should not gather branch/worktree info as part of the cold-start dispatcher since it primes branch-oriented thinking.

**Action**: Remove the git state collection step from Phase 1.

**Old text**:
```
# 2. Session list (recent, any status)
ari session list --output json 2>/dev/null

# 3. Git state
git rev-parse --abbrev-ref HEAD 2>/dev/null
git status --porcelain 2>/dev/null | head -20

# 4. Active rite
```

**New text**:
```
# 2. Session list (recent, any status)
ari session list --output json 2>/dev/null

# 3. Active rite
```

---

### Edit 6: `/Users/tomtenuta/Code/knossos/mena/navigation/go.dro.md` (third location)

**Rationale**: Remove Worktree and Branch lines from Scenario 4 (ORIENTATION) dashboard. These are the most potent ambient signals because ORIENTATION is the "nothing happening" state where Claude has maximum latitude to suggest actions, and displaying branch/worktree state implies those are relevant next steps.

**Old text**:
```
Worktree: {worktree_name or "main"}
Branch: {branch_name}
Rite: {active_rite or "none"}
Sessions: {count active} active, {count parked} parked
```

**New text**:
```
Rite: {active_rite or "none"}
Sessions: {count active} active, {count parked} parked
```

---

### Edit 7: `/Users/tomtenuta/Code/knossos/mena/navigation/consult/reference.md`

**Rationale**: The example output under "Mode 1: No Arguments (General Help)" includes `Git: feature/auth-system (clean)` which normalizes branch display as a standard ecosystem status element. Remove it. Also remove `- Git branch and status` from the state summary checklist.

**Action**: Remove the Git line from the example output.

**Old text**:
```
Active Rite: 10x-dev (5 agents)
Session: ACTIVE - "Add authentication" (MODULE complexity)
Git: feature/auth-system (clean)
```

**New text**:
```
Active Rite: 10x-dev (5 agents)
Session: ACTIVE - "Add authentication" (MODULE complexity)
```

---

### Edit 8: `/Users/tomtenuta/Code/knossos/mena/navigation/consult/reference.md` (second location)

**Rationale**: The state summary checklist under "Mode 1: No Arguments" includes `Git branch and status` as a first-class item to display. Remove it.

**Old text**:
```
1. **Summarize Current State**
   - Active rite (from `.claude/ACTIVE_RITE`)
   - Active session (from `.claude/sessions/`)
   - Git branch and status
   - Current complexity level (if in session)
```

**New text**:
```
1. **Summarize Current State**
   - Active rite (from `.claude/ACTIVE_RITE`)
   - Active session (from `.claude/sessions/`)
   - Current complexity level (if in session)
```

---

### Edit 9: `/Users/tomtenuta/Code/knossos/mena/navigation/consult/reference.md` (third location)

**Rationale**: The playbook example under "Mode 3" includes `Branch: main` in the current context block. Remove it.

**Old text**:
```
Current Context:
  Rite: none (will sync to 10x-dev)
  Session: none (will start new)
  Branch: main
```

**New text**:
```
Current Context:
  Rite: none (will sync to 10x-dev)
  Session: none (will start new)
```

---

### Edit 10: `/Users/tomtenuta/Code/knossos/mena/session/start/INDEX.dro.md`

**Rationale**: Replace worktree option 4 in the "session already exists" options block with a "See also" footer. The current design presents `/worktree create` as a first-class option alongside `/continue`, `/park`, and `/wrap`, which normalizes worktree creation as a routine step. Moving it to a footer makes worktrees discoverable but not proactively suggested. Also remove the promotional paragraph that follows the options block.

**Action**: Replace the options block and the promotional paragraph.

**Old text**:
```
**When session already exists, offer these options:**

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first
4. /worktree create "<name>" - Start in ISOLATED worktree (parallel work)

Tip: Use worktrees when you want to work on something different
without affecting the current session/rite.
```

The `/worktree` option is especially useful when:
- Different rite needed for the new work
- Want to keep current sprint context intact
- Need true parallel sessions on same project
```

**New text**:
```
**When session already exists, offer these options:**

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first

See also: /worktree for parallel work in isolated worktrees
```
```

---

### Edit 11: `/Users/tomtenuta/Code/knossos/mena/workflow/sprint.dro.md`

**Rationale**: The "Parallel Sprint Pattern" section (lines 75-107) is a 32-line block that positions worktree-based parallel sprints as a primary workflow pattern. Per stakeholder decision, move this content to a companion file and leave a 1-line reference. This reduces ambient worktree pressure in the sprint command while keeping the knowledge discoverable.

**Action**: Replace the entire "Parallel Sprint Pattern" section with a 1-line reference.

**Old text**:
```
## Parallel Sprint Pattern

For truly parallel sprints across multiple rites/focuses, use **worktrees**:

```bash
# Create isolated worktrees per sprint
/worktree create "sprint-backend" --rite=10x-dev
/worktree create "sprint-frontend" --rite=10x-dev
/worktree create "sprint-docs" --rite=docs

# In each terminal, navigate and start sprint independently:
# Terminal 1:
cd worktrees/wt-xxx && claude
/sprint "Backend Sprint" --tasks="API,Database,Auth"

# Terminal 2:
cd worktrees/wt-yyy && claude
/sprint "Frontend Sprint" --tasks="Components,State,Tests"

# Terminal 3:
cd worktrees/wt-zzz && claude
/sprint "Docs Sprint" --tasks="API Docs,User Guide,Examples"
```

**Why worktrees for parallel sprints?**
- Each sprint gets isolated SPRINT_CONTEXT (no collision)
- Different rites can work simultaneously
- Changes don't affect each other
- Use `/sessions --all` to monitor all sprints

**Single sprint, multiple tasks** → use this command directly
**Multiple parallel sprints** → use `/worktree` per sprint
```

**New text**:
```
## Parallel Sprints

For multiple parallel sprints, see `/worktree` command.
```

---

### Edit 12: `/Users/tomtenuta/Code/knossos/mena/workflow/hotfix/examples.md`

**Rationale**: The "Integration with Git" section (lines 236-296) is a 60-line block that teaches Claude to proactively create hotfix branches, push them, and create PRs as part of the hotfix workflow. This is the most potent branch-creation trigger in the codebase because it shows explicit `git checkout -b` commands. Per stakeholder decision, remove the entire section. The commit message example that follows is useful reference material and can stay.

**Action**: Remove the "Integration with Git" section entirely.

**Old text**:
```
## Integration with Git

**Recommended git workflow for hotfixes**:

```bash
# Create hotfix branch
git checkout -b hotfix/login-500-error

# Make fix
/hotfix "API returning 500 on login"

# Commit (done by Principal Engineer during /hotfix)
# Message should include: what, why, rollback

# Push and create PR
git push origin hotfix/login-500-error
gh pr create --title "HOTFIX: Fix login 500 error" --body "..."

# Fast-track review and merge
# Deploy immediately
```

**For CRITICAL fixes, consider**:
- Merge directly to main (skip PR if necessary)
- Deploy immediately
- Create PR retroactively for audit trail

---

## Example Commit Message
```

**New text**:
```
## Example Commit Message
```

---

### Edit 13: `/Users/tomtenuta/Code/knossos/mena/session/common/anti-patterns.md`

**Rationale**: Add a new anti-pattern entry documenting proactive branch/worktree creation as an anti-pattern. Per stakeholder decision, this goes in the session common anti-patterns file (not CLAUDE.md inscription). This is added to the "Cross-Cutting Anti-Patterns" section before the existing "Treating Sessions as Todo Lists" entry.

**Action**: Add new anti-pattern entry after the "Cross-Cutting Anti-Patterns" heading.

**Old text**:
```
## Cross-Cutting Anti-Patterns

### Treating Sessions as Todo Lists
```

**New text**:
```
## Cross-Cutting Anti-Patterns

### Proactive Branch/Worktree Creation

**Pattern**: Creating git branches or worktrees without the user explicitly requesting parallel work

**Symptoms**:
- Branch created that the user did not ask for
- Worktree suggested when user just wants to work in current directory
- `git checkout -b` commands appearing in workflow without user keyword trigger
- Output displays branch state as a primary status element

**Fix**:
- Only create branches or worktrees when the user uses explicit keywords: `branch`, `worktree`, `isolate`, `parallel`
- Work on the current branch unless the user requests otherwise
- Do not display branch name as a primary status field
- Do not suggest worktree creation as a default option

**Prevention**: Branch/worktree operations require explicit user intent

**Why Bad**: Creates unwanted git artifacts, pollutes branch namespace, adds complexity the user did not request, and derails the user's intended workflow

---

### Treating Sessions as Todo Lists
```

---

### Edit 14: `/Users/tomtenuta/Code/knossos/mena/operations/commit/INDEX.dro.md`

**Rationale**: The commit INDEX file contains the line `Your current branch is available in session context ('git_branch' field).` which explicitly calls Claude's attention to the branch field, priming branch-aware behavior. Since we removed the branch from Text() output, this pointer to the git_branch JSON field should also be removed.

**Action**: Remove the branch context pointer.

**Old text**:
```
Auto-injected by SessionStart hook (project, rite, session, git).

Your current branch is available in session context (`git_branch` field).
```

**New text**:
```
Auto-injected by SessionStart hook (project, rite, session).
```

---

### Edit 15: `/Users/tomtenuta/Code/knossos/mena/operations/pr/INDEX.dro.md`

**Rationale**: The PR INDEX file contains `Base branch is available in session context ('base_branch' field).` This is a legitimate programmatic pointer for the PR command (it needs the base branch to create the PR correctly). However, since we are removing branch fields from Text() output and this field is still available in JSON, we should keep this pointer but note that it comes from JSON context rather than the visible session table. Actually, the PR command is one of the few places where branch context IS operationally necessary (it needs to know what base to target). Keep this as-is but update the context line to be accurate.

**Action**: Update the context reference to be more accurate without promoting branch display.

**Old text**:
```
## Context
Auto-injected by SessionStart hook (project, rite, session, git).

Base branch is available in session context (`base_branch` field).
```

**New text**:
```
## Context
Auto-injected by SessionStart hook (project, rite, session).

Base branch for PR targeting: detect via `git symbolic-ref refs/remotes/origin/HEAD` or default to `main`.
```

---

## New Files

### `/Users/tomtenuta/Code/knossos/mena/workflow/sprint-parallel-worktrees.md`

**Rationale**: The parallel sprint pattern content removed from `sprint.dro.md` (Edit 11) needs a home. This companion file preserves the knowledge for users who explicitly seek it via `/worktree` or when the sprint command references it.

**Content**:

```markdown
# Parallel Sprint Pattern with Worktrees

> Reference for running multiple sprints in parallel using git worktrees.

For truly parallel sprints across multiple rites or focuses, use worktrees to get filesystem isolation:

```bash
# Create isolated worktrees per sprint
/worktree create "sprint-backend" --rite=10x-dev
/worktree create "sprint-frontend" --rite=10x-dev
/worktree create "sprint-docs" --rite=docs

# In each terminal, navigate and start sprint independently:
# Terminal 1:
cd worktrees/wt-xxx && claude
/sprint "Backend Sprint" --tasks="API,Database,Auth"

# Terminal 2:
cd worktrees/wt-yyy && claude
/sprint "Frontend Sprint" --tasks="Components,State,Tests"

# Terminal 3:
cd worktrees/wt-zzz && claude
/sprint "Docs Sprint" --tasks="API Docs,User Guide,Examples"
```

## Why Worktrees for Parallel Sprints

- Each sprint gets isolated SPRINT_CONTEXT (no collision)
- Different rites can work simultaneously
- Changes don't affect each other
- Use `/sessions --all` to monitor all sprints

## When to Use

- **Single sprint, multiple tasks**: Use `/sprint` directly
- **Multiple parallel sprints**: Use `/worktree` per sprint

## Related

- `/worktree` command for worktree management
- `/sprint` command for single-sprint orchestration
- `/sessions --all` to view sessions across worktrees
```

---

## No-Change Files

| File | Reason |
|------|--------|
| `/Users/tomtenuta/Code/knossos/mena/operations/pr/behavior.md` | References to `feature branch` and `main/master` are operationally correct for PR creation. PRs inherently require branch context. No ambient pressure -- these are functional requirements. |
| `/Users/tomtenuta/Code/knossos/mena/operations/pr/examples.md` | Branch names in examples (`feature/auth`, `hotfix/double-charge`) are part of PR output display where branch IS the subject. No depollution needed -- this is the correct context for branch references. |
| `/Users/tomtenuta/Code/knossos/mena/operations/commit/behavior.md` | References to branches are limited to error cases (`Detached HEAD`) and state validation. No proactive branch creation. |
| `/Users/tomtenuta/Code/knossos/mena/operations/commit/examples.md` | No branch references. Clean. |
| `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/INDEX.dro.md` | This IS the worktree command -- branch/worktree references are its purpose. No changes needed. |
| `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/behavior.md` | Worktree behavior spec -- all references are operationally appropriate. |
| `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/examples.md` | Worktree examples -- references to `/start` option 4 will be stale after Edit 10 but the examples file itself correctly describes worktree workflows. The stale `/start` cross-reference in this file should be updated. |
| `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/integration.md` | Contains `/start` option 4 cross-reference that will be stale. See Edit 16. |
| `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/troubleshooting.md` | All references are operationally appropriate for worktree troubleshooting. |
| `/Users/tomtenuta/Code/knossos/mena/operations/code-review/INDEX.dro.md` | References `branch` in argument hint and `git diff main...<branch>` which is the operational requirement for code review. No ambient pressure. |
| `/Users/tomtenuta/Code/knossos/mena/navigation/sessions.dro.md` | References to `worktrees` are limited to the `--all` flag which lists sessions across worktrees. Operationally correct. |
| `/Users/tomtenuta/Code/knossos/mena/session/park/behavior.md` | Git status capture during park is operationally correct -- it records state, not creates branches. |
| `/Users/tomtenuta/Code/knossos/mena/guidance/lexicon/anti-patterns.md` | This file covers mena syntax anti-patterns (stale syntax, invocation confusion). Git artifact depollution is a different concern and belongs in session common anti-patterns. |

---

## Stale Cross-Reference Updates

### Edit 16: `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/examples.md`

**Rationale**: After Edit 10, the `/start` options block no longer shows worktree as option 4. The examples file and integration file both cross-reference this exact options block. Update them to match the new footer-style reference.

**Action**: Update the `/start` cross-reference in the examples file.

**Old text**:
```
### With /start

When `/start` detects an existing session, it now offers worktree as an option:

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first
4. /worktree create "<name>" - Start in ISOLATED worktree (parallel work)
```
```

**New text**:
```
### With /start

When `/start` detects an existing session, it offers these options and references worktree:

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first

See also: /worktree for parallel work in isolated worktrees
```
```

---

### Edit 17: `/Users/tomtenuta/Code/knossos/mena/navigation/worktree/integration.md`

**Rationale**: Same stale cross-reference as Edit 16, in the integration file.

**Action**: Update the `/start` cross-reference.

**Old text**:
```
### /start Interaction

When `/start` detects an existing session, it offers worktree as an option:

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first
4. /worktree create "<name>" - Start in ISOLATED worktree (parallel work)
```
```

**New text**:
```
### /start Interaction

When `/start` detects an existing session, it offers these options and references worktree:

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first

See also: /worktree for parallel work in isolated worktrees
```
```

---

## Regression Check

### What Could Break

1. **Test failure in context_test.go**: The test `TestContextOutput_Text` case "with git and ecosystem fields" currently asserts `| Git Branch | feat/backtick-migration |` appears in Text() output. Edit 2 removes this assertion. The test `TestRunContext_WithActiveSession_IncludesRitesAndAgents` checks `result.BaseBranch != "main"` which validates the JSON field (not Text()) -- this test is UNAFFECTED.

2. **Worktree operations that read Text() output**: The `internal/cmd/worktree/` package reads ContextOutput via JSON, not Text(). Verified by grep: no callers parse the markdown table output. JSON fields `git_branch` and `base_branch` are preserved.

3. **Existing hook consumers**: The SessionStart hook outputs ContextOutput. Any consumer that parses the JSON payload sees no change. The Text() output is only displayed in Claude's conversation context -- it is not parsed programmatically.

4. **Sprint companion file projection**: The new file `sprint-parallel-worktrees.md` sits alongside `sprint.dro.md` in the `mena/workflow/` directory. Since `sprint.dro.md` is a standalone dromena file (not in an INDEX directory), this companion file will project independently. Given the companion file depollution design (DESIGN-companion-file-depollution.md), this file should receive `user-invocable: false` frontmatter injection if it ends up in a dromena directory. Since `mena/workflow/` contains standalone `.dro.md` files, not INDEX directories, this companion file will project as a standalone command. To prevent this, it should NOT have a `.dro.md` extension (it does not -- it is plain `.md`). However, the pipeline routes ALL `.md` files from the `workflow/` directory. We need to verify the pipeline behavior for non-`.dro.md` files in dromena grouping directories.

### Verification Steps

1. Run `CGO_ENABLED=0 go test ./internal/cmd/hook/...` -- all context tests must pass
2. Run `CGO_ENABLED=0 go test ./...` -- full test suite
3. Run `ari sync` on a satellite -- verify no regression in materialized output
4. Manual verification: start a Claude session, invoke `/go`, confirm no branch in output
5. Manual verification: invoke `/sprint`, confirm parallel pattern section replaced with 1-liner
6. Manual verification: invoke `/hotfix`, confirm no "Integration with Git" section in examples

## Integration Test Matrix

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| Knossos (self-host) | `ari sync` after edits | All commands materialize. Sprint companion file projects. No branch in `/go` output template. Anti-pattern entry present in session-common. |
| Minimal satellite | Fresh `ari sync` | No errors. Hook context table lacks Git Branch/Base Branch rows. Commands work normally. |
| Standard satellite | `ari sync` update | No errors. Hotfix examples lack git integration section. Sprint has 1-line parallel reference. |
| Complex satellite (custom hooks) | `ari sync` + manual test | Custom hooks still receive `git_branch` and `base_branch` in JSON payload. Only Text() output changed. |

## Implementation Sequence

All edits are independent and can be applied in any order. Stakeholder requested a single atomic commit.

1. Apply Edits 1-3 (Go source: context.go and context_test.go)
2. Apply Edits 4-6 (go.dro.md: branch removal from output templates)
3. Apply Edits 7-9 (consult reference: git state removal from examples)
4. Apply Edit 10 (/start: worktree option to footer)
5. Apply Edit 11 (sprint: parallel pattern to companion)
6. Create new file: sprint-parallel-worktrees.md
7. Apply Edit 12 (hotfix: remove git integration section)
8. Apply Edit 13 (anti-patterns: add new entry)
9. Apply Edits 14-15 (commit/pr INDEX: branch context pointers)
10. Apply Edits 16-17 (worktree cross-references)
11. Run tests: `CGO_ENABLED=0 go test ./internal/cmd/hook/...`
12. Run full test suite: `CGO_ENABLED=0 go test ./...`
13. Single atomic commit

## Design Decisions Log

| Decision | Rationale | Alternatives Rejected |
|----------|-----------|----------------------|
| Remove from Text() only, keep JSON fields | JSON fields are consumed by `ari session status`, worktree operations, and other programmatic consumers. Text() is the sole channel into Claude's conversational context. | Remove struct fields entirely: breaks JSON API contract and programmatic consumers. |
| Keep branch context in /pr and /code-review | These commands operationally REQUIRE branch context (PR targets a base branch, code review diffs against a branch). Removing branch context here would break functionality. | Depollute /pr too: would break the command's core purpose. |
| Anti-pattern in session-common, not CLAUDE.md | CLAUDE.md inscription is for universal rules that must be in every session. The anti-pattern is reference knowledge that Claude loads on demand when session lifecycle topics arise. | CLAUDE.md: too heavy for this concern; wastes inscription budget. |
| Single companion file for sprint parallel content | The removed content is cohesive (one topic: parallel sprints with worktrees). A single file is easier to discover and maintain than splitting across worktree skill files. | Inline in worktree INDEX: would make worktree INDEX too long and mix concerns. |
| Explicit keyword trigger list for branch creation | Provides clear, auditable criteria for when branch/worktree operations are appropriate. Keywords are: `branch`, `worktree`, `isolate`, `parallel`. | No keyword list: too vague, hard to audit. Longer keyword list: dilutes signal. |
| Remove git state collection from /go Phase 1 | The `/go` command collected `git rev-parse --abbrev-ref HEAD` and `git status --porcelain` as part of parallel state gathering. This primes branch-aware output even when no session exists. Session/rite state is sufficient for dispatch decisions. | Keep git state but hide from output: inconsistent -- if we collect it, Claude uses it. |
