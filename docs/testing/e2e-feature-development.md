# E2E Test: Feature Development Workflow

> End-to-end test scenario for FEATURE complexity development through the 10x-dev-pack.
> Version: 1.0.0

## Overview

This document defines a complete test scenario for the standard feature development workflow, validating HANDOFF artifact production at each phase transition.

**Complexity Level**: FEATURE (MODULE complexity in 10x-dev-pack terms)
**Workflow Path**: Requirements -> Design -> Code -> QA -> Doc (optional handoff)
**Primary Team**: 10x-dev-pack
**Cross-Team Handoffs**: Optional handoff to doc-team-pack for documentation

---

## Test Scenario: User Profile Settings Feature

### Scenario Description

Implement a user profile settings page allowing users to update their display name, email preferences, and notification settings.

**Why This Scenario**: This represents a typical MODULE-complexity feature that:
- Requires all four 10x-dev-pack phases
- Produces artifacts at each phase transition
- May trigger doc-team-pack handoff for user documentation
- Has clear acceptance criteria for each phase

---

## Phase 1: Requirements

### Entry Criteria
- [ ] User request or product initiative received
- [ ] Session initialized with `/start`

### Agent
**requirements-analyst**

### Input
User request: "We need a profile settings page where users can update their display name, email preferences, and notification settings."

### Expected Artifact: PRD

```markdown
# PRD: User Profile Settings

## Problem Statement
Users cannot modify their profile settings after initial account creation.

## Success Criteria
- Users can update display name (validated, 2-50 chars)
- Users can toggle email preferences (marketing, product updates)
- Users can configure notification settings (push, email, SMS)
- Changes persist immediately with optimistic UI

## User Stories
- US-001: As a user, I want to change my display name
- US-002: As a user, I want to control which emails I receive
- US-003: As a user, I want to choose my notification channels
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| PRD exists at `docs/requirements/PRD-user-profile-settings.md` | YES | [ ] |
| PRD contains Problem Statement | YES | [ ] |
| PRD contains Success Criteria | YES | [ ] |
| PRD contains User Stories | YES | [ ] |
| Complexity assessed as MODULE or higher | YES | [ ] |

### Phase Transition
- **From**: requirements
- **To**: design
- **Handoff Type**: Internal (within 10x-dev-pack)
- **Trigger**: PRD marked complete in session context

---

## Phase 2: Design

### Entry Criteria
- [ ] PRD complete and approved
- [ ] Complexity >= MODULE (design phase required)

### Agent
**architect**

### Input
- PRD from Phase 1
- Existing codebase context

### Expected Artifact: TDD

```markdown
# TDD: User Profile Settings

## Design Overview
RESTful API for profile settings with React form component.

## Architecture
- API: `/api/v1/users/:id/settings` (GET, PATCH)
- Frontend: `ProfileSettingsForm` component
- State: React Query for server state

## Interface Definitions
### API Endpoints
PATCH /api/v1/users/:id/settings
  Body: { displayName?, emailPrefs?, notifications? }
  Response: 200 OK with updated settings

## Component Structure
ProfileSettingsPage
  -> ProfileSettingsForm
    -> DisplayNameField
    -> EmailPreferencesSection
    -> NotificationSettingsSection

## Error Handling
- 400: Invalid display name format
- 401: Not authenticated
- 403: Cannot edit other user's settings
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| TDD exists at `docs/design/TDD-user-profile-settings.md` | YES | [ ] |
| TDD references source PRD | YES | [ ] |
| TDD contains Interface Definitions | YES | [ ] |
| TDD contains Error Handling | YES | [ ] |
| ADR created if architectural decisions made | CONDITIONAL | [ ] |

### Phase Transition
- **From**: design
- **To**: implementation
- **Handoff Type**: Internal (within 10x-dev-pack)
- **Trigger**: TDD marked complete in session context

---

## Phase 3: Implementation

### Entry Criteria
- [ ] TDD complete and approved
- [ ] Implementation scope clear from TDD

### Agent
**principal-engineer**

### Input
- TDD from Phase 2
- PRD for acceptance criteria reference

### Expected Artifacts

1. **API Implementation**
   - `src/api/routes/users/settings.ts`
   - `src/api/schemas/userSettings.ts`

2. **Frontend Implementation**
   - `src/components/ProfileSettingsForm.tsx`
   - `src/pages/ProfileSettingsPage.tsx`

3. **Tests**
   - `tests/api/userSettings.test.ts`
   - `tests/components/ProfileSettingsForm.test.tsx`

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| All files from TDD component structure exist | YES | [ ] |
| Unit tests exist for API endpoints | YES | [ ] |
| Component tests exist for frontend | YES | [ ] |
| Tests pass (`npm test`) | YES | [ ] |
| Linting passes (`npm run lint`) | YES | [ ] |
| Implementation notes documented | YES | [ ] |

### Phase Transition
- **From**: implementation
- **To**: validation
- **Handoff Type**: Internal (within 10x-dev-pack)
- **Trigger**: Code complete, tests passing

---

## Phase 4: Validation (QA)

### Entry Criteria
- [ ] Code complete per TDD
- [ ] Unit tests passing
- [ ] Implementation notes available

### Agent
**qa-adversary**

### Input
- Implemented code
- PRD for success criteria
- TDD for interface contracts

### Expected Artifact: Test Plan/Report

```markdown
# Test Plan: User Profile Settings

## Functional Tests
- [x] Display name updates correctly
- [x] Display name validation enforced (2-50 chars)
- [x] Email preferences toggle correctly
- [x] Notification settings persist

## Edge Cases Tested
- [x] Empty display name rejected
- [x] Display name with special characters
- [x] Concurrent updates handled
- [x] Network failure shows error state

## Security Tests
- [x] Cannot update other user's settings
- [x] Unauthenticated requests rejected
- [x] Rate limiting on PATCH endpoint

## Result: PASS
All 15 test cases pass. No blocking defects.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Test plan exists at `docs/qa/TEST-PLAN-user-profile-settings.md` | YES | [ ] |
| All PRD success criteria covered | YES | [ ] |
| Edge cases from PRD tested | YES | [ ] |
| Security considerations tested | YES | [ ] |
| Overall result documented | YES | [ ] |

### Phase Transition
- **From**: validation
- **To**: complete (or handoff to doc-team-pack)
- **Handoff Type**: Cross-team (if documentation needed)
- **Trigger**: QA signoff

---

## Optional Cross-Team Handoff: Documentation

If user-facing documentation is required, produce a HANDOFF artifact for doc-team-pack:

### HANDOFF Artifact

```yaml
---
source_team: 10x-dev-pack
target_team: doc-team-pack
handoff_type: assessment
created: 2026-01-02
initiative: User Profile Settings Feature
priority: medium
---

## Context

User profile settings feature is complete and QA-validated.
Needs user documentation for help center and in-app guidance.

## Source Artifacts
- `docs/requirements/PRD-user-profile-settings.md`
- `docs/design/TDD-user-profile-settings.md`
- `docs/qa/TEST-PLAN-user-profile-settings.md`

## Items

### DOC-001: Help center article
- **Priority**: High
- **Summary**: User-facing documentation for profile settings
- **Assessment Questions**:
  - What level of detail is needed?
  - Should screenshots be included?
  - Is video walkthrough needed?

### DOC-002: In-app tooltips
- **Priority**: Medium
- **Summary**: Contextual help for settings fields
- **Assessment Questions**:
  - Which fields need explanation?
  - What tone/voice guidelines apply?

## Notes for Target Team

UI screenshots available in design system.
Product owner: @product-lead for content approval.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| HANDOFF follows schema | YES | [ ] |
| Source artifacts listed | YES | [ ] |
| Assessment questions clear | YES | [ ] |
| Target team actionable | YES | [ ] |

---

## Complete Test Checklist

### Phase Completeness
- [ ] Phase 1 (Requirements): PRD produced
- [ ] Phase 2 (Design): TDD produced
- [ ] Phase 3 (Implementation): Code + tests produced
- [ ] Phase 4 (Validation): Test plan/report produced

### Artifact Trail
- [ ] Each artifact references its inputs
- [ ] Session context updated at each transition
- [ ] Handoff artifacts follow schema (if cross-rite)

### Quality Gates
- [ ] No phase skipped for MODULE complexity
- [ ] All tests pass before QA phase
- [ ] QA validates against PRD success criteria

---

## Running This Test

### Manual Execution

1. Initialize session:
   ```
   /start initiative="User Profile Settings" complexity=MODULE team=10x-dev-pack
   ```

2. Execute Phase 1:
   ```
   Task(requirements-analyst, "Create PRD for user profile settings feature...")
   ```

3. Execute Phase 2:
   ```
   Task(architect, "Design profile settings system per PRD...")
   ```

4. Execute Phase 3:
   ```
   Task(principal-engineer, "Implement profile settings per TDD...")
   ```

5. Execute Phase 4:
   ```
   Task(qa-adversary, "Validate profile settings implementation...")
   ```

6. Verify all artifacts exist and pass schema validation.

### Automated Validation

Future: Hook into CI to validate artifact existence and schema compliance at each phase gate.

---

## Related Documents

- [Cross-Team Coordination Playbook](../playbooks/cross-rite-coordination.md)
- [Handoff Smoke Tests](handoff-smoke-tests.md)
- [10x-dev-pack Workflow](../../rites/10x-dev-pack/workflow.md)
- [Cross-Team Handoff Schema](../../.claude/skills/shared/cross-rite-handoff/schema.md)
