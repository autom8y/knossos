# Critical Path

Visual dependency graph of sprints identifying which sprints block others and where parallel execution is possible. Use for initiatives with non-linear sprint dependencies.

## Schema

```yaml
critical_path:
  diagram: |
    sprint-1 (rnd) ──→ sprint-2 (rnd) ──→ sprint-4 (dev)
                                      ╲
    sprint-3 (security) ──────────────→ sprint-5 (dev)
                                              │
                                              ▼
                                        sprint-6 (ops)
  parallel_opportunities:
    - group: [sprint-1, sprint-3]
      rationale: "<why these can run concurrently>"
  bottlenecks:
    - sprint: sprint-2
      reason: "<why this sprint is on the critical path>"
```

## Guidance

- **ASCII diagrams for portability.** Use `──→` for dependencies, `╲` for branching. Potnia and humans both need to parse this without rendering tools.
- **Name rites in the diagram.** Each sprint node should include its rite in parentheses so the critical path doubles as a rite transition map.
- **Parallel opportunities must be safe.** Two sprints can run in parallel only if they have no shared exit artifacts and no entry criteria that reference each other's outputs.
- **Bottlenecks identify risk.** A bottleneck is a sprint where delay cascades to all downstream work. Name it explicitly so risk mitigation can focus there.

## Anti-Patterns

- **Linear chains presented as critical paths.** If every sprint depends on the previous one, the critical path is just the sprint list. Omit this section -- it adds no information.
- **Parallel opportunities without safety analysis.** Saying "these can run in parallel" without confirming artifact independence creates race conditions.
