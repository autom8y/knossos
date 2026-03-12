# Sprint Decomposition

Per-sprint definitions that tell Potnia what each sprint accomplishes without dictating how agents do the work. Sprint internals are NON-PRESCRIPTIVE -- define the mission and constraints, not the agent task list.

## Schema

```yaml
sprints:
  - id: sprint-1
    rite: <rite-slug>
    mission: "<what this sprint accomplishes -- one sentence>"
    agents: [<agent-1>, <agent-2>, ...]
    entry_criteria:
      - "<precondition that must be true before sprint starts>"
    exit_criteria:
      - "<condition that must be true for sprint to be complete>"
    exit_artifacts:
      - path: "<relative file path>"
        description: "<what this artifact contains and who consumes it>"
    context:
      - "<file path to Read() at sprint start>"
    notes: "<optional operational context for Potnia>"
```

## Guidance

- **Mission is a one-sentence charter.** Potnia reads this to understand what success looks like for this sprint. Agents read this to understand their collective goal.
- **Agent roster comes from `orchestrator.yaml`.** List agents by their slug as defined in the rite's orchestrator. Do not invent agents.
- **Entry criteria gate sprint start.** Potnia checks these before dispatching agents. If entry criteria reference artifacts from a prior sprint, name the specific file path.
- **Exit criteria gate sprint completion.** These feed directly into PT-NN checkpoints. Write them as boolean-evaluable conditions.
- **Exit artifacts name what the sprint produces.** Include the file path and a description. Downstream sprints reference these in their `context` field.
- **Context lists what to read, not what to do.** Prior spike results, architecture docs, configuration files. Potnia loads these at sprint start.
- **Non-prescriptive internals.** "Evaluate three caching strategies and recommend one" is a mission. "First read the Redis docs, then benchmark Memcached, then write a comparison table" is a task list. Missions empower; task lists constrain.

## Anti-Patterns

- **Prescriptive agent task lists.** Writing "agent-1 does X, agent-2 does Y" defeats the purpose of orchestration. Potnia coordinates agent work dynamically based on the mission and constraints.
- **Missing exit artifacts.** If a sprint produces nothing downstream sprints can reference, it is either the final sprint or the decomposition is incomplete.
