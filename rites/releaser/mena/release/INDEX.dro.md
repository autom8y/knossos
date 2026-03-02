---
name: release
description: "Orchestrate a multi-repo release from reconnaissance through deployment verification. Use when: releasing SDKs, bumping consumers, pushing and verifying across repos, running PATCH/RELEASE/PLATFORM workflows. Triggers: /release, ship it, release all, push and verify, platform release."
argument-hint: "<glob-or-repos> [--complexity=PATCH|RELEASE|PLATFORM] [--dry-run]"
allowed-tools: Bash, Read, Write, Glob, Grep, Task, Skill, AskUserQuestion, TodoWrite
model: opus
---

# /release - Orchestrated Release Workflow

One command from code to verified deployment. `/release` captures your intent, consults Pythia for complexity gating and phase routing, dispatches specialists in order, and loops until the verification report renders a chain-aware verdict.

## Arguments

| Argument | Description | Example |
|----------|-------------|---------|
| positional | Glob pattern or comma-separated repo names | `~/code/autom8y*`, `autom8y-data,autom8y-ads` |
| `--complexity` | Override auto-detection (PATCH, RELEASE, PLATFORM) | `--complexity=RELEASE` |
| `--dry-run` | Run reconnaissance only, produce state map, stop before execution | |

If no positional argument is given, ask the user what they want to release.

## Behavior

### Phase 0 -- Pre-Flight

1. **Verify releaser rite is active**: Check `.claude/CLAUDE.md` for `releaser` in the agent configurations section. If not active, run `ari sync --rite releaser` first and inform the user.
2. **Verify gh auth**: Run `gh auth status` -- if unauthenticated, ERROR and stop.
3. **Clean slate**: Check `.sos/wip/release/` for artifacts from a prior run. If found, ask the user whether to clear them or resume from where the prior run left off.
4. **Parse arguments**: Extract glob pattern or repo list from `$ARGUMENTS`. Extract `--complexity` and `--dry-run` flags.

### Phase 1 -- Initial Pythia Consultation

Construct a startup CONSULTATION_REQUEST and invoke Pythia:

```
Task(subagent_type="pythia", prompt="
## CONSULTATION_REQUEST

consultation:
  type: startup

  initiative:
    title: 'Release: {user_intent}'
    goal: '{parsed goal from arguments}'

  state:
    current_phase: none
    completed_phases: []
    blocked_on: none

  results:
    last_specialist: none
    last_outcome: none
    artifacts_ready: []

  context_summary: |
    User invoked /release with scope: {glob_or_repos}
    Complexity override: {--complexity value or 'auto-detect'}
    Dry run: {true|false}
    Artifact directory: .sos/wip/release/
    {any additional context from pre-flight}
")
```

Store Pythia's agent ID for throughline resumption.

### Phase 2 -- Orchestration Loop

Loop until Pythia returns `directive.action: complete`:

#### On `invoke_specialist`:

1. Read `specialist.agent` and `specialist.prompt` from Pythia's response
2. Invoke the specialist via Task tool:
   ```
   Task(subagent_type="{specialist.agent}", prompt="{specialist.prompt}")
   ```
3. When the specialist completes, verify its artifacts exist:
   - cartographer: `platform-state-map.yaml` at `.sos/wip/release/`
   - dependency-resolver: `dependency-graph.yaml`
   - release-planner: `release-plan.yaml`
   - release-executor: `execution-ledger.yaml`
   - pipeline-monitor: `verification-report.yaml`
4. Construct a continuation CONSULTATION_REQUEST and re-consult Pythia (with `resume` if agent ID available):
   ```
   Task(subagent_type="pythia", resume="{pythia_agent_id}", prompt="
   ## CONSULTATION_REQUEST

   consultation:
     type: continuation
     initiative:
       title: 'Release: {user_intent}'
       goal: '{goal}'
     state:
       current_phase: '{completed phase name}'
       completed_phases: [{list of completed phases}]
       blocked_on: none
     results:
       last_specialist: '{agent name}'
       last_outcome: success
       artifacts_ready:
         - '{artifact}: .sos/wip/release/{filename}'
     context_summary: |
       {Brief summary of specialist results}
       {Any notable findings -- dirty repos, dependency issues, failures}
   ")
   ```

#### On `await_user`:

1. Read `user_question.prompt` and `user_question.context`
2. Present to the user via AskUserQuestion
3. Feed the user's answer back to Pythia in the next consultation

#### On specialist failure:

1. Construct a failure CONSULTATION_REQUEST:
   ```
   consultation:
     type: failure
     results:
       last_specialist: '{agent}'
       last_outcome: failure
       artifacts_ready: []
     context_summary: |
       {failure reason and any partial output}
   ```
2. Pythia will either retry with adjusted prompt, recommend phase rollback, or escalate to user

### Phase 3 -- Dry-Run Exit (if `--dry-run`)

After the reconnaissance phase (cartographer produces `platform-state-map.yaml`):
1. Read and display the state map summary to the user
2. If complexity was auto-detected, report it
3. If `has_dependents` was discovered, note the auto-escalation that WOULD occur
4. Stop. Do not proceed to execution.

### Phase 4 -- Completion

When Pythia returns `directive.action: complete`:

1. Read `verification-report.yaml` from `.sos/wip/release/`
2. Display the verdict to the user:

   **On PASS**:
   ```
   Release verified. All CI green, all pipeline chains resolved, all deployments healthy.
   {N} repos released, {M} PRs created.
   ```

   **On FAIL**:
   ```
   Release verification FAILED.
   {failure summary from verification report}
   Recommendation: {pipeline-monitor's recommendation}
   ```

   **On PARTIAL**:
   ```
   Release partially verified. CI green but {chain timeouts / dispatch issues / deployment concerns}.
   Review verification-report.md for details.
   ```

3. If a session is active, note the release outcome for session context.

## Specialist Agent Reference

| Phase | Agent | Produces | Typical Duration |
|-------|-------|----------|-----------------|
| Reconnaissance | cartographer | platform-state-map.yaml | 2-5 min |
| Dependency Analysis | dependency-resolver | dependency-graph.yaml | 1-3 min |
| Release Planning | release-planner | release-plan.yaml | 1-3 min |
| Execution | release-executor | execution-ledger.yaml | 5-15 min |
| Verification | pipeline-monitor | verification-report.yaml | 5-30 min (chain-dependent) |

## Throughline Resumption

Store Pythia's agent ID after the first consultation. On subsequent consultations, pass `resume: {agent_id}` to maintain Pythia's decision history across the full workflow. If resume fails (ID expired, session changed), fall back to a fresh consultation with full context in `context_summary`.

## Error Recovery

| Scenario | Action |
|----------|--------|
| Specialist times out (maxTurns exceeded) | Report to Pythia as failure; Pythia may retry with narrower scope |
| gh auth expires mid-workflow | Stop, inform user, they re-auth, then `/release --resume` |
| Artifact missing after specialist claims completion | Report to Pythia as failure with details |
| User Ctrl+C during long CI monitoring | Artifacts remain in `.sos/wip/release/`; next `/release` detects and offers resume |

## Anti-Patterns

- **Skipping Pythia**: Never invoke specialists directly without consulting Pythia first. Pythia gates complexity, manages auto-escalation, and verifies handoff criteria.
- **Ignoring specialist failures**: Always report failures back to Pythia. Never silently retry or skip.
- **Hardcoding phase order**: Let Pythia determine routing. PATCH skips dependency-analysis and release-planning. Auto-escalation may add phases mid-flight.
- **Treating dry-run as no-op**: Dry-run still runs full reconnaissance. The state map is valuable even without execution.
