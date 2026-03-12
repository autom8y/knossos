# Potnia Consultation Points

PT-NN numbered checkpoints where Potnia evaluates throughline satisfaction. These are the quality gates of the initiative -- not "is it done?" but "does the work so far satisfy the invariant?"

## Schema

```yaml
checkpoints:
  - id: PT-01
    after: sprint-1
    evaluates: "<which aspect of the throughline this checkpoint tests>"
    questions:
      - "<specific evaluative question with a yes/no or measurable answer>"
      - "<specific evaluative question>"
    gate: <hard|soft>
    on_fail: "<concrete action -- rework sprint, add remediation sprint, escalate to user>"
```

## Guidance

- **Every checkpoint evaluates the throughline.** The `evaluates` field names the specific throughline aspect being tested. If a checkpoint cannot trace back to the throughline, it does not belong.
- **Questions must be specific and answerable.** "Is the API working?" fails. "Does the API contract match the schema defined in sprint-1's exit artifact?" passes. Potnia needs to evaluate evidence, not make subjective judgments.
- **Hard gates block progression.** A hard gate means the next sprint cannot start until this checkpoint passes. Use for foundational work that downstream sprints depend on.
- **Soft gates flag but continue.** A soft gate logs a concern and proceeds. Use for quality checks where rework can happen in a later sprint without blocking progress.
- **`on_fail` must be actionable.** "Fix it" is not actionable. "Re-run sprint-2 with tightened exit criteria on error handling" is. Options: rework current sprint, insert remediation sprint, escalate to user, accept risk and proceed.
- **Place checkpoints at rite boundaries.** When work transitions between rites, a checkpoint validates that the outgoing rite's work is sufficient for the incoming rite. This is the most critical checkpoint placement.
- **Business-domain granularity.** For domain-specific initiatives, checkpoints should reflect domain concerns. A payment integration initiative needs checkpoints about idempotency and settlement, not just "API responds 200."

## Anti-Patterns

- **Generic "is everything OK?" questions.** These produce rubber-stamp passes. Every question must reference specific artifacts, criteria, or measurements.
- **Checkpoints without `on_fail`.** A gate without a failure path is not a gate -- it is a suggestion.
