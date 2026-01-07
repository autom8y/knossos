# /resume Examples

> Usage scenarios for resuming parked work sessions.

## Example 1: Simple Resume (Same Day)

```bash
/resume
```

Output:
```
Resuming Session: Add dark mode toggle

Session Details:
- Started: 2025-12-24 10:00:00
- Parked: 2025-12-24 12:30:00 (3 hours ago)
- Park reason: Lunch break
- Complexity: MODULE
- Team: 10x-dev-pack
- Current phase: design
- Last agent: architect

Artifacts produced:
✓ PRD: /docs/requirements/PRD-dark-mode.md
✓ TDD: /docs/design/TDD-dark-mode.md

✓ No blockers
✓ Git status: Clean
✓ Team consistent

Continue with architect? [Y/n]: Y

Session resumed: Add dark mode toggle
Duration parked: 3 hours
Continuing with: architect
Current phase: design

Architect is reviewing TDD and ready to proceed.
```

---

## Example 2: Resume After Team Change

```bash
/resume
```

Output:
```
Resuming Session: Multi-tenant authentication

Session Details:
- Started: 2025-12-23 14:00:00
- Parked: 2025-12-23 17:00:00 (1 day ago)
- Park reason: Waiting for design review
- Complexity: SERVICE
- Team: 10x-dev-pack
- Current phase: design
- Last agent: architect

⚠ Team Mismatch Detected

Session started with: 10x-dev-pack
Current active rite: doc-rite-pack

Options:
1. Switch back to 10x-dev-pack (recommended)
2. Continue with doc-rite-pack (may cause issues)
3. Cancel resume

Choice [1/2/3]: 1

Switching to team: 10x-dev-pack
✓ Swapped to 10x-dev-pack (5 agents loaded)

Continue with architect? [Y/n]: Y

Session resumed: Multi-tenant authentication
Duration parked: 1 day
Continuing with: architect
```

---

## Example 3: Resume with Different Agent

```bash
/resume --agent=principal-engineer
```

Output:
```
Resuming Session: API retry logic

Session Details:
- Started: 2025-12-24 09:00:00
- Parked: 2025-12-24 12:00:00 (4 hours ago)
- Park reason: Design approved, ready for implementation
- Complexity: MODULE
- Team: 10x-dev-pack
- Current phase: design
- Last agent: architect

Artifacts produced:
✓ PRD: /docs/requirements/PRD-api-retry.md
✓ TDD: /docs/design/TDD-api-retry.md
✓ ADR-0043: Exponential backoff strategy

✓ No blockers
✓ Git status: Clean

Resuming with principal-engineer (overriding last_agent: architect)

Session resumed: API retry logic
Duration parked: 4 hours
Continuing with: principal-engineer
Current phase: implementation

Principal Engineer ready to implement based on approved TDD.
```

---

## Example 4: Resume with Git Changes

```bash
/resume
```

Output:
```
Resuming Session: Database migration

Session Details:
- Started: 2025-12-20 10:00:00
- Parked: 2025-12-20 16:00:00 (4 days ago)
- Park reason: End of sprint
- Complexity: SERVICE
- Current phase: implementation
- Last agent: principal-engineer

⚠ Git Changes Detected

Git status at park time: clean
Current git status: dirty

New/modified files since park:
- src/db/migrations/0012_add_indexes.sql (new)
- src/db/schema.py (modified)
- tests/test_migration.py (modified)

This may indicate work done outside this session.

Review changes before continuing? [y/n]: y

git diff --stat:
 src/db/migrations/0012_add_indexes.sql | 15 +++++++++++++++
 src/db/schema.py                       |  5 +++--
 tests/test_migration.py                |  8 ++++++++
 3 files changed, 26 insertions(+), 2 deletions(-)

Continue with principal-engineer? [Y/n]: Y

Session resumed: Database migration
Duration parked: 4 days
⚠ Note: Review external changes before proceeding
```
