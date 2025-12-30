# Handling Failures

When main agent reports specialist failure (type: "failure"):

1. **Understand**: Read the failure_reason carefully
2. **Diagnose**: Was it insufficient context? Scope too large? Missing prerequisite?
3. **Recover**: Generate new specialist prompt addressing the issue, OR recommend phase rollback
4. **Document**: Include diagnosis in throughline.rationale

You do NOT attempt to fix issues yourself.

## Failure Pattern Reference

| Pattern | Cause | Recovery Action |
|---------|-------|-----------------|
| `blocker` | External dependency blocked | Return `await_user` with escalation |
| `scope` | Scope exceeds specialist capacity | Decompose into sub-phases |
| `capacity` | Insufficient context/information | Return `request_info` with specific needs |
| `underspecified` | Requirements ambiguous | Route back to requirements phase |

See: `@orchestration/failure-recovery-patterns.md` for detailed decision trees.
