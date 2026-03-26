---
description: "Four-Phase Rollout Lifecycle (F6) companion for ui-ref skill."
---

# Four-Phase Rollout Lifecycle (F6)

Migration playbook for design system evolution. Derived from Saarinen's design system lifecycle [MODERATE evidence -- industry four-phase rollout is well-established]. The four phases ensure the system remains coherent at every intermediate state: no intermediate state is broken.

## The Invariant: Never Skip Phase 2

**NEVER skip Phase 2 (block new usage).** This is the most violated and most critical constraint in design system migration.

Why Phase 2 is non-negotiable: Without blocking new adoption of the deprecated pattern, the migration target is a moving target. While you are migrating existing usage (Phase 3), new usage appears (because the deprecated pattern is still usable). By the time you finish migrating, you are back to the beginning. Phase 2 freezes the problem before you solve it.

## The Four Phases

### Phase 1: Warn

**What happens**: Mark the deprecated pattern as deprecated. Add deprecation warnings in tooling (TypeScript types, ESLint rules, Storybook annotations). Document the intended migration path.

**What does NOT happen**: Nothing is broken. Deprecated patterns continue to work. No forced migrations.

**Gate criteria to advance to Phase 2**:
- [ ] All deprecated usages identified and catalogued
- [ ] Deprecation warnings active in all relevant tooling
- [ ] Migration path documented and accessible
- [ ] Team notified of deprecation and migration timeline

**Duration**: Sufficient for teams to be notified and plan. Typically 1-2 sprint cycles.

### Phase 2: Block New (NEVER SKIP)

**What happens**: Prevent NEW usage of the deprecated pattern. Lint rules error on new adoption. Type errors on new usage. CI gates fail on new deprecated usage in new or modified files.

**What does NOT happen**: Existing deprecated usage is not yet required to migrate. Old code still works. Only NEW code is blocked from adopting the deprecated pattern.

**Gate criteria to advance to Phase 3**:
- [ ] Lint rule or type error blocking new usage in CI
- [ ] Existing usage count documented (baseline for Phase 3)
- [ ] Migration timeline communicated
- [ ] Support resources available (migration guide, office hours, codemod)

**Why this is the critical phase**: Phase 2 stops the bleeding. Without it, the deprecated pattern continues growing while you are migrating it. The migration never ends.

### Phase 3: Budget Down

**What happens**: Actively migrate existing deprecated usage. Set and enforce decreasing usage budgets. Provide codemods where automatable. Track remaining usage count; the number must decrease each sprint.

**Budget mechanism**: Set a maximum deprecated usage count. Reduce the budget each sprint. CI fails if deprecated usage exceeds budget. Treat each sprint's budget reduction as a delivery requirement.

**Gate criteria to advance to Phase 4**:
- [ ] Deprecated usage at zero (or explicitly deferred edge cases documented with owner and timeline)
- [ ] Codemods applied and verified
- [ ] Adapter layers removed (or retained for external API stability)
- [ ] All tests passing without deprecated pattern

**Codemod specification template**:
```
Codemod: {name}
Transforms: {old pattern} -> {new pattern}
Coverage: {what it handles} / {what requires manual migration}
Test cases: {link to test fixtures}
Known limitations: {edge cases the codemod cannot handle}
```

### Phase 4: Remove

**What happens**: Remove the deprecated pattern from the design system entirely. Delete the deprecated token, component, or API. Remove the deprecation warnings (no longer needed). Remove the blocking lint rules.

**What does NOT happen**: This phase should be anticlimactic. If Phase 2 and Phase 3 were executed correctly, there is no deprecated usage to break. Removal is a cleanup operation, not a breaking change.

**Gate criteria (completion)**:
- [ ] Zero deprecated usage in codebase (verified by tooling)
- [ ] Deprecated pattern removed from design system
- [ ] Documentation updated
- [ ] Deprecation tooling (lint rules, type errors) removed
- [ ] Visual regression baseline updated to reflect removal

## Adapter Layer Pattern

When the migration requires backward compatibility during Phase 3 (external APIs, published packages, multi-team coordination):

**Adapter layer structure**:
```typescript
// Adapter: maps deprecated API to new API
// Named after the deprecated pattern, documents the target
// TODO(phase-4): remove this adapter when Phase 3 complete
export const DeprecatedComponent = (props: DeprecatedProps) => {
  return <NewComponent {...mapProps(props)} />;
};
```

**Adapter layer invariants**:
- Adapters are temporary -- they exist only for Phase 3 duration
- Adapters must be tracked: every adapter has a Phase 4 removal ticket
- Adapters must not add behavior -- pure mapping only
- Adapters must have their own tests (they can break too)

## Bottom-Up Extraction (Product -> System)

The four-phase rollout also supports bottom-up extraction: promoting a product-level pattern to the design system.

**Phase 1 (Warn)**: Identify the pattern in product code. Annotate it as "design system candidate." Inform teams that this pattern will be promoted.

**Phase 2 (Block New)**: Freeze the pattern in product code (no variants, no modifications). The design system version is being built. New usage must wait for the system version.

**Phase 3 (Budget Down)**: Replace product-specific instances with the design system component as it ships.

**Phase 4 (Remove)**: Delete the product-level implementation. Design system version is canonical.
