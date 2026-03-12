# Estimated Duration

Per-sprint time estimates and velocity assumptions. Use when timeline commitments exist or when coordinating across teams with scheduling constraints.

## Schema

```yaml
duration_estimates:
  velocity_assumptions:
    - "<assumption about agent throughput or environment>"
  sprints:
    - id: sprint-<N>
      estimate: "<time range, e.g., 1-2 sessions>"
      basis: "<what the estimate is based on>"
      buffer: "<contingency time>"
  total:
    optimistic: "<best case>"
    expected: "<likely case>"
    pessimistic: "<worst case with risk materialization>"
```

## Guidance

- **Estimates are ranges, not points.** "1-2 sessions" acknowledges uncertainty. "1 session" implies false precision.
- **State the basis.** "Based on similar prototype work in the rnd rite" or "based on sprint-1 actual velocity." Unbased estimates are guesses.
- **Buffer at the sprint level, not just the total.** Each sprint with medium or high risks in the Risk Map should have its own buffer. A single global buffer hides which sprints are driving timeline risk.
- **Three-point total estimate.** Optimistic (everything goes right), expected (normal friction), pessimistic (major risk materializes). This gives the human operator a realistic range.
- **Sessions, not hours.** Agent work happens in sessions. "2-3 sessions" is more meaningful than "4-6 hours" because session productivity varies with context complexity.

## Anti-Patterns

- **Single-point estimates.** "This will take 3 sessions" is always wrong. Ranges communicate uncertainty honestly.
- **Estimates without basis.** If you cannot say why a sprint will take 2 sessions, the estimate is a guess. State "no reliable basis -- estimate is speculative" rather than presenting guesses as plans.
