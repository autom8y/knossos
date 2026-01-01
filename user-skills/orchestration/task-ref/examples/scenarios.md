# Task Lifecycle Examples

Complete demonstrations of `/task` command across complexity levels.

## Example 1: Simple Script Task

```bash
/task "Add retry logic to API client" --complexity=SCRIPT
```

Output:
```
Task: Add retry logic to API client
Complexity: SCRIPT

[Phase 1] Requirements
✓ PRD created (lightweight): /docs/requirements/PRD-api-retry.md
  Acceptance criteria:
  - Retry on network failures (3 attempts)
  - Exponential backoff
  - Configurable retry count

[Phase 2] Design
⊘ Skipped (SCRIPT complexity - no TDD needed)

[Phase 3] Implementation
✓ Principal Engineer implementing...
✓ Code: /src/api/client.ts (added retryWithBackoff function)
✓ Tests: /src/api/client.test.ts (5 test cases)
✓ All tests passing

[Phase 4] Validation
✓ QA Adversary validating...
✓ All acceptance criteria met
✓ Edge cases tested: timeout, max retries, success after retry
✓ No defects found

Task complete: Ready for commit
```

## Example 2: Module Task

```bash
/task "Implement user authentication module"
```

Interactive prompts:
```
Task: Implement user authentication module
Complexity? [SCRIPT/MODULE/SERVICE/PLATFORM]: MODULE

[Phase 1] Requirements
✓ Requirements Analyst creating PRD...
✓ PRD: /docs/requirements/PRD-user-auth.md

[Phase 2] Design
✓ Architect creating TDD...
✓ TDD: /docs/design/TDD-user-auth.md
✓ ADR created: /docs/decisions/ADR-0050-jwt-vs-sessions.md

[Phase 3] Implementation
✓ Principal Engineer implementing...
  Files created:
  - /src/auth/authenticator.ts
  - /src/auth/jwt-manager.ts
  - /src/auth/middleware.ts
  - /tests/auth/authenticator.test.ts
✓ All tests passing (coverage: 94%)

[Phase 4] Validation
✓ QA Adversary validating...
✓ Test Plan: /docs/testing/TEST-user-auth.md
✓ Security review: PASSED
✓ Performance: Login < 100ms (requirement: < 200ms)
✓ Production ready

Task complete.
```

## Example 3: Quick Task (Skip PRD)

```bash
/task "Fix typo in error message" --skip-prd --complexity=SCRIPT
```

Output:
```
Task: Fix typo in error message
Complexity: SCRIPT

[Phase 1] Requirements
⊘ Skipped (--skip-prd specified)

[Phase 2] Design
⊘ Skipped (SCRIPT complexity)

[Phase 3] Implementation
✓ Principal Engineer implementing...
✓ Fixed: /src/errors/messages.ts
✓ Tests updated

[Phase 4] Validation
✓ QA Adversary: Quick validation passed

Task complete: Trivial change, ready for commit
```
