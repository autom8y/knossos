# Edge Cases: Cross-Team Workflows

> Documentation of edge cases, failure modes, and recovery procedures for cross-rite handoffs.
> Version: 1.0.0

## Overview

This document covers edge cases that may occur during cross-rite coordination, including failed handoffs, circular dependencies, multi-team coordination needs, and urgent change bypass procedures.

---

## 1. Failed Handoffs

### 1.1 Target Team Rejects Handoff

**Scenario**: Target team cannot accept handoff due to missing information, incorrect type, or scope issues.

**Detection**:
- Target team returns HANDOFF with `status: rejected`
- Rejection reason documented in response

**Example Rejection Response**:
```yaml
---
source_team: security
target_team: 10x-dev
handoff_type: assessment
status: rejected
created: 2026-01-02
initiative: Payment Processing
---

## Rejection

### Reason
Missing threat model context. Cannot perform security assessment without:
- Data flow diagrams
- Trust boundaries
- Asset inventory

### Required for Resubmission
1. Data flow diagram showing payment token movement
2. Trust boundary identification (client/server/third-party)
3. Asset classification (what data is processed)

### Next Steps
Update PRD with security context section, resubmit HANDOFF.
```

**Recovery**:
1. Source team reviews rejection reason
2. Address missing requirements
3. Resubmit HANDOFF with updated artifacts
4. Reference original HANDOFF in resubmission

**Prevention**:
- Use pre-submission checklist per handoff type
- Validate against schema before sending
- Include all referenced source artifacts

---

### 1.2 Handoff Times Out (No Response)

**Scenario**: Blocking handoff sent but no response within expected timeframe.

**Detection**:
- Session tracks handoff with `pending_response` status
- Blocking flag prevents phase progression
- SLA exceeded (24h for critical, 72h for normal)

**Escalation Process**:

```markdown
## Handoff Timeout Escalation

**Original Handoff**: HANDOFF-10x-to-security-2026-01-02.md
**Status**: No response after 48 hours
**Impact**: Design phase blocked

### Escalation Request
Requesting user intervention to:
1. Confirm handoff was received by security
2. Expedite response or provide ETA
3. If team unavailable, identify alternative path

### Options
A. Wait for security response (preferred)
B. Proceed without security review (RISK: compliance gap)
C. Engage external security consultant (COST)
```

**Recovery**:
1. User investigates routing/receipt
2. Target team prioritizes or provides ETA
3. If unresolvable, escalate to alternative path

**Prevention**:
- Set realistic SLAs based on handoff type
- Include `blocking: true` to signal urgency
- Use `priority: critical` for time-sensitive work

---

### 1.3 Handoff Lost or Corrupted

**Scenario**: HANDOFF artifact cannot be found or is malformed.

**Detection**:
- Target team reports no handoff received
- Artifact file missing from expected location
- YAML parse errors in frontmatter

**Recovery**:
1. Check all possible storage locations:
   - `.claude/sessions/{session-id}/`
   - `docs/handoffs/`
   - Source team's working directory
2. Regenerate from session context if available
3. Recreate from source artifacts if necessary

**Prevention**:
- Use consistent file naming: `HANDOFF-[source]-to-[target]-[date].md`
- Store in versioned location (git tracked)
- Validate YAML frontmatter before saving

---

## 2. Circular Dependencies

### 2.1 A -> B -> A Loop

**Scenario**: Team A hands off to Team B, which needs input from Team A to proceed.

**Example**:
```
10x-dev -> security (threat model request)
security -> 10x-dev (needs architecture details)
10x-dev waiting on security before architecture
```

**Detection**:
- Session shows multiple pending handoffs between same teams
- Neither team can proceed
- Deadlock in workflow

**Resolution Pattern**:

```markdown
## Circular Dependency Resolution

### Deadlock Identified
- 10x-dev waiting on: HANDOFF-security-response
- security waiting on: Architecture details from 10x

### Breaking the Loop
1. Identify minimum viable input for one team
2. Produce partial artifact to unblock
3. Iterate with additional detail

### Resolution
10x-dev produces: Preliminary architecture (enough for threat modeling)
security proceeds: Threat model with architecture assumptions
10x-dev finalizes: Architecture incorporating threat mitigations
```

**Prevention**:
- Include sufficient context in initial handoff
- Use assessment questions to clarify needs upfront
- Design teams produce preliminary artifacts before handoff

---

### 2.2 Multi-Hop Loop (A -> B -> C -> A)

**Scenario**: Work cycles through multiple teams before returning to origin.

**Example**:
```
rnd -> strategy (evaluation)
strategy -> 10x-dev (implementation guidance needed)
10x-dev -> rnd (prototype questions)
```

**Detection**:
- Work has traversed 3+ teams
- Original team receives handoff referencing their work
- Initiative stalls with no clear owner

**Resolution Pattern**:

1. **Convene**: All involved teams review current state
2. **Consolidate**: Merge requirements into single work package
3. **Assign**: Designate primary owner with consulting support
4. **Proceed**: Primary owner drives with async consultation

**Prevention**:
- Use orchestrator for complex multi-team work
- Establish clear ownership at initiative start
- Limit handoff chains to 2 hops before review

---

## 3. Multi-Team Coordination

### 3.1 Parallel Team Dependencies

**Scenario**: Feature requires simultaneous work from multiple teams.

**Example**: New authentication system requires:
- security: Threat model and compliance review
- sre: Infrastructure provisioning
- 10x-dev: Implementation

**Coordination Pattern**:

```markdown
## Parallel Coordination: Authentication System

### Teams Involved
1. security - Threat model (blocking for 10x)
2. sre - Infrastructure (parallel with 10x)
3. 10x-dev - Implementation (after security, parallel with sre)

### Handoff Structure
```
                    ┌─────────────┐
                    │ Orchestrator│
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ security │    │   sre    │    │   10x    │
    └────┬─────┘    └────┬─────┘    └────┬─────┘
         │               │               │
         │ Threat Model  │ Infra Ready   │ Implementation
         └───────────────┴───────────────┘
                         │
                    Integration
```

### Synchronization Points
1. Security approval before 10x implementation
2. SRE + 10x integration for deployment
3. All teams signoff for production release
```

**Coordination Mechanisms**:
- Central initiative document with team assignments
- Shared milestone tracking
- Regular sync checkpoints (handoff artifacts)

---

### 3.2 Sequential Team Chain

**Scenario**: Work flows through multiple teams in sequence.

**Example**: Research to production pipeline:
```
intelligence -> strategy -> 10x-dev -> sre
```

**Chain Management**:

```markdown
## Sequential Chain: Customer Insights Feature

### Chain
1. intelligence: User research synthesis
   Output: HANDOFF-intelligence-to-strategy (strategic_input)

2. strategy: Strategic evaluation and go decision
   Output: HANDOFF-strategy-to-10x (implementation)

3. 10x-dev: Production implementation
   Output: HANDOFF-10x-to-sre (validation)

4. sre: Production readiness validation
   Output: Deployment approval

### Chain Integrity
- Each handoff references previous outputs
- Traceability from research to deployment
- Any team can trace back to original research
```

**Prevention of Chain Breaks**:
- Each handoff includes source artifact chain
- Context section summarizes prior work
- Initiative ID consistent across all handoffs

---

### 3.3 Hub-and-Spoke Coordination

**Scenario**: Central team coordinates multiple specialist teams.

**Example**: Major feature requiring multiple specializations:
```
             ┌─────────────┐
             │ 10x (hub)   │
             └──────┬──────┘
    ┌───────────────┼───────────────┐
    ▼               ▼               ▼
┌────────┐    ┌──────────┐    ┌──────────┐
│security│    │doc-team  │    │   sre    │
└────────┘    └──────────┘    └──────────┘
```

**Hub Responsibilities**:
- Produce handoffs to all spoke teams
- Track all pending responses
- Integrate outputs from spoke teams
- Coordinate timing and dependencies

**Spoke Responsibilities**:
- Respond to handoff within SLA
- Flag blocking issues early
- Reference hub initiative in output

---

## 4. Urgent Change Bypass

### 4.1 Emergency Hotfix Path

**Scenario**: Production incident requires immediate fix, cannot wait for full workflow.

**Bypass Authorization**:

```markdown
## Emergency Bypass: Production Auth Failure

### Incident
- Severity: P0 (production outage)
- Impact: All users unable to login
- Duration: Started 2026-01-02 14:30 UTC

### Bypass Justification
Normal security review would take 24+ hours.
User impact exceeds security review benefit for this fix.

### Bypass Scope
- Skip: Pre-implementation security assessment
- Keep: Post-deployment security review (mandatory)

### Authorization
- Authorized by: @engineering-lead
- Timestamp: 2026-01-02 14:45 UTC
- Tracking: INC-2026-001

### Post-Bypass Requirements
1. Hotfix deployed: 2026-01-02 15:00 UTC
2. Security review scheduled: 2026-01-02 (within 48h)
3. Postmortem: Include bypass justification
```

**Bypass Conditions**:
- P0/P1 incident in progress
- Explicit authorization from designated authority
- Documented justification
- Mandatory follow-up review

**Recovery After Bypass**:
1. Complete skipped handoff as soon as incident resolved
2. Document in postmortem
3. Review if bypass was appropriate

---

### 4.2 Time-Critical Feature Launch

**Scenario**: Business deadline requires accelerated workflow.

**Acceleration Pattern**:

```markdown
## Accelerated Workflow: Product Launch Feature

### Deadline
Launch event: 2026-01-15
Standard workflow would complete: 2026-01-20

### Acceleration Request
- Reduce security review from 48h to 24h
- Parallelize SRE validation with final QA
- Pre-allocate doc-team capacity

### Risk Assessment
- Security: Medium risk (payment adjacent, not direct)
- Operational: Low risk (similar to existing features)
- Documentation: Low risk (can complete post-launch)

### Approval
- Product: Approved
- Engineering: Approved with conditions
- Security: Approved for 24h review

### Conditions
1. Security team has advance notice (done)
2. Feature flag for gradual rollout
3. Enhanced monitoring for 7 days post-launch
```

**Acceleration Limits**:
- Cannot skip mandatory security gates for SYSTEM complexity
- Cannot skip SRE validation for production deployment
- Must document acceleration and conditions

---

### 4.3 Single-Team Override

**Scenario**: Work normally requiring cross-rite handoff done by single team.

**Override Justification**:

```markdown
## Single-Team Override: Security Hotfix

### Normal Path
10x-dev -> security (assessment) -> 10x-dev

### Override Request
Security team member embedded in 10x-dev session
Performing inline security review during implementation

### Justification
- Security specialist available in session
- Urgent fix, cannot wait for handoff round-trip
- Scope is narrow and well-understood

### Documentation
- Security review performed by: @security-lead (embedded)
- Review scope: CVE-2026-1234 patch
- Findings: None, patch approved

### Tracking
This override is logged for audit purposes.
Full security pack not invoked due to embedded specialist.
```

**Override Conditions**:
- Specialist from target team participates
- Scope is narrow and time-critical
- Full documentation of inline review

---

## 5. Recovery Procedures

### 5.1 State Recovery After Failure

**Scenario**: Session fails mid-handoff, state unclear.

**Recovery Steps**:

1. **Assess Current State**
   ```bash
   # Check session context
   cat .claude/sessions/{session-id}/SESSION_CONTEXT.md

   # Find any handoff artifacts
   find .claude/sessions/{session-id} -name "HANDOFF-*.md"

   # Check git status for uncommitted work
   git status
   ```

2. **Determine Last Known Good State**
   - Last completed phase
   - Last committed artifact
   - Pending handoffs

3. **Resume or Restart**
   - If handoff was sent: Wait for response
   - If handoff was in progress: Complete and send
   - If unclear: Regenerate from source artifacts

---

### 5.2 Rollback After Failed Execution

**Scenario**: Hygiene pack execution fails, need to revert.

**Rollback Process**:

```markdown
## Rollback: PKG-001 Email Validator Consolidation

### Failure
Tests failing after janitor commits.
Behavior change unintentionally broke downstream consumers.

### Rollback Steps
1. Revert commits: abc123, def456, ghi789
2. Restore original validator files
3. Confirm tests pass on reverted state

### Post-Rollback
1. Update debt-triage package with new constraints
2. Re-plan with behavior preservation stricter
3. Re-execute with additional testing

### Tracking
- Original package: PKG-001
- Rollback commit: xyz789
- Re-execution: PKG-001-v2
```

---

### 5.3 Handoff Replay

**Scenario**: Need to re-execute handoff due to changed conditions.

**Replay Pattern**:

```markdown
## Handoff Replay: HANDOFF-10x-to-security-2026-01-02

### Original Handoff
Filed 2026-01-02, threat model completed 2026-01-03

### Why Replay Needed
Significant architecture change after threat model.
Original threat model no longer applies.

### Replay Request
- New HANDOFF: HANDOFF-10x-to-security-2026-01-05-v2
- References: Original handoff, architecture change docs
- Scope: Delta analysis (what changed)

### Delta Scope
Only reassess:
- New data flows introduced
- Changed trust boundaries
- New third-party integration

Retain from original:
- Core payment flow analysis
- PCI-DSS assessment (unchanged)
```

---

## 6. Prevention Patterns

### 6.1 Pre-Flight Checklist

Before producing any handoff:

- [ ] Handoff type matches flow pattern
- [ ] All required frontmatter fields present
- [ ] Source artifacts exist and are accessible
- [ ] Each item has required type-specific fields
- [ ] Blocking flag set appropriately
- [ ] Priority reflects urgency
- [ ] Notes provide actionable guidance
- [ ] File named correctly and in right location

### 6.2 Handoff Review Gate

For critical handoffs, add review step:

```yaml
# In session context
handoff_review_required: true
handoff_reviewer: @senior-engineer
```

Reviewer checks:
- Schema compliance
- Completeness
- Clarity
- Appropriate target team

### 6.3 SLA Configuration

Configure expected response times:

| Handoff Type | Priority | Expected Response |
|--------------|----------|-------------------|
| assessment | critical | 24 hours |
| assessment | high | 48 hours |
| assessment | medium | 72 hours |
| execution | high | 24 hours |
| execution | medium | 72 hours |
| validation | critical | 12 hours |
| validation | high | 24 hours |

---

## Related Documents

- [Cross-Team Coordination Playbook](../playbooks/cross-rite-coordination.md)
- [Handoff Smoke Tests](../testing/handoff-smoke-tests.md)
- [Cross-Team Handoff Schema](../../.claude/skills/shared/cross-rite-handoff/schema.md)
