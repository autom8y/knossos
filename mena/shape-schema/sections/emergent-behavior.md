# Emergent Behavior Constraints

Three categories that define where agents have freedom and where they do not. Agents need explicit boundaries to optimize locally without violating global constraints.

## Schema

```yaml
behavior_constraints:
  prescribed:
    - "<non-negotiable rule -- agents MUST follow>"
    - "<non-negotiable rule>"
  emergent:
    - "<area of agent discretion -- local optimization allowed>"
    - "<area of agent discretion>"
  out_of_scope:
    - "<must NOT touch -- explicitly excluded from this initiative>"
    - "<must NOT touch>"
```

## Guidance

- **Prescribed = non-negotiable.** These are hard constraints: "All new endpoints must have integration tests," "No changes to the authentication middleware." Agents cannot override prescribed rules regardless of local optimization pressure.
- **Emergent = agent discretion.** These are areas where agents choose the approach: "File organization within the new module," "Error message wording," "Test fixture design." Potnia does not micromanage these -- agents optimize based on local context.
- **Out of Scope = do not touch.** Explicitly name files, modules, or concerns that this initiative must not modify. This prevents scope creep from ambitious agents who see adjacent improvements.
- **Err toward emergent.** When unsure whether something should be prescribed or emergent, default to emergent. Over-prescribing produces brittle plans that break when agents encounter unexpected conditions.
- **Out of Scope prevents scope creep.** Agents regularly discover adjacent improvements. Without explicit out-of-scope boundaries, a "payment integration" initiative drifts into "refactor the entire billing module."

## Anti-Patterns

- **Empty emergent category.** If everything is prescribed or out of scope, agents have no room to optimize. This produces mechanical execution that misses opportunities visible only at implementation time.
- **Vague out-of-scope.** "Don't change unrelated things" is not actionable. "Do not modify `internal/auth/`, `internal/billing/legacy.go`, or CI pipeline configuration" is.
