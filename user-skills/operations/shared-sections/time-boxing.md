# Time-Boxing Pattern

> Enforce time limits with progress checkpoints.

## When to Apply

- `/spike`: Research time limits (30m-8h)
- `/hotfix`: Fix time limits (30-90 min by severity)

## Implementation

### Checkpoint Schedule

| Checkpoint | Action |
|------------|--------|
| 25% | Report initial findings |
| 50% | Preliminary conclusions |
| 75% | Start wrapping up |
| 100% | STOP and document |

### Severity-Based Limits (hotfix only)

| Severity | Target | Max |
|----------|--------|-----|
| CRITICAL | 30 min | 60 min |
| HIGH | 45 min | 90 min |
| MEDIUM | 30 min | 60 min |

### Time Exceeded Handling

If exceeding limit:
1. Stop current work
2. Document partial findings
3. For spikes: Recommend follow-up spike
4. For hotfixes: Escalate to full `/task`

## Customization Points

| Parameter | Description | Commands |
|-----------|-------------|----------|
| timebox | Total duration | spike, hotfix |
| checkpoints | Percentage markers | spike |
| severity | Hotfix priority | hotfix |

## Error Messages

| Condition | Message Template |
|-----------|------------------|
| Timebox too short | "Cannot complete in {timebox}. Recommend {min_duration}." |
| Timebox too long | "Timebox > 8h. Break into phases or use /task." |
| Time exceeded | "Time budget reached. Documenting {partial/incomplete} findings." |
