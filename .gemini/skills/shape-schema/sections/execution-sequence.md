# Execution Sequence

The operational playbook: what commands to run, in what order, to execute the shape. Covers rite transitions, session lifecycle, and per-sprint context loading.

## Schema

```yaml
execution_sequence:
  - phase: setup
    commands:
      - "ari sync --rite=<rite-slug>"
      - "/sos start --initiative=<initiative-slug>"
    notes: "<operational context>"
  - phase: sprint-<N>
    commands:
      - "/sprint <N>"
      - "Read('<context-file-1>')"
      - "Read('<context-file-2>')"
    notes: "<what Potnia should know before dispatching agents>"
  - phase: checkpoint-PT-<NN>
    commands:
      - "# Potnia evaluates PT-<NN> questions against sprint artifacts"
    notes: "<checkpoint evaluation context>"
  - phase: rite-transition
    commands:
      - "/sos wrap"
      - "ari sync --rite=<next-rite>"
      - "/sos start --initiative=<slug>"
    notes: "<what changes in the new rite context>"
  - phase: wrap
    commands:
      - "/sos wrap"
    notes: "<final state>"
```

## Guidance

- **Rite transitions follow a three-step pattern.** Wrap the current session, sync to the new rite, start a new session. This ensures agent rosters, skills, and hooks align with the new rite's orchestrator.
- **Context loading is explicit.** List every `Read()` call that should happen at sprint start. Do not assume Potnia will discover context files -- name them.
- **Checkpoint phases are evaluation-only.** No commands to run -- Potnia reads sprint exit artifacts and evaluates PT-NN questions. Include the phase to make checkpoint timing visible in the sequence.
- **Notes carry operational context.** Use notes for information Potnia needs that does not fit in the command list: "sprint-2 agents need access to the prototype from sprint-1" or "this rite's Potnia has not seen prior sprint work -- load the handoff artifact."
- **Single-rite shapes omit rite transitions.** The sequence simplifies to: setup, sprint phases, checkpoint phases, wrap.

## Anti-Patterns

- **Omitting context loading.** "Just run /sprint 2" without naming the files to read means Potnia starts blind. Always list context files explicitly.
- **Mixing execution with evaluation.** Sprint phases produce work; checkpoint phases evaluate it. Keep them separate so Potnia knows when she is coordinating vs. assessing.
