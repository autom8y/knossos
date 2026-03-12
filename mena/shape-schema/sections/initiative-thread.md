# Initiative Thread

The throughline that Potnia carries across all sprints and rite transitions. Every checkpoint evaluates against this thread. If the thread is vague, checkpoints devolve into "is it done?" rubber stamps.

## Schema

```yaml
initiative_thread:
  throughline: "<one-sentence invariant -- what must remain true from first sprint to last>"
  success_criteria:
    - "<measurable outcome with concrete evidence>"
    - "<measurable outcome with concrete evidence>"
  failure_signals:
    - "<observable early warning -- something you can detect mid-initiative>"
    - "<observable early warning>"
```

## Guidance

- **Throughline is an invariant, not a goal.** "Integrate Stripe payments" is a goal. "Every payment path settles within 30 seconds with idempotent retry" is a throughline. Goals describe what you build; throughlines describe what must hold true.
- **Success criteria must be evaluable.** Potnia needs to answer yes/no at checkpoints. "Improved performance" fails. "P95 latency under 200ms on the read path" passes.
- **Failure signals are early warnings, not post-mortems.** They should be detectable mid-initiative: "Sprint 2 prototype cannot connect to upstream API" not "Project shipped late."
- **One throughline per shape.** If you have two throughlines, you have two initiatives. Split them.
- **Thread carries across rite transitions.** When work moves from rnd to dev, the thread is the continuity. Agents in the new rite should understand the throughline without reading prior sprint artifacts.

## Anti-Patterns

- **Restating the brief as throughline.** The user's brief is the request; the throughline is the extracted invariant. Copy-pasting the brief means no analysis happened.
- **Success criteria without measurement.** "System is secure" is not evaluable. "All endpoints require authentication and rate limiting is configured per the security rite's standards" is.
