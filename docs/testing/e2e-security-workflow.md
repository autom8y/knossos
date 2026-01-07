# E2E Test: Security-Sensitive Change Workflow

> End-to-end test scenario for SYSTEM complexity with security consultation gates.
> Version: 1.0.0

## Overview

This document defines a complete test scenario for security-sensitive feature development, validating that threat-modeler gates trigger correctly and security assessment HANDOFFs are produced.

**Complexity Level**: SYSTEM (highest complexity in 10x-dev-pack)
**Workflow Path**: Requirements -> Security Consultation -> Design -> Code -> Security Assessment -> QA
**Primary Team**: 10x-dev-pack
**Cross-Team Handoffs**:
- 10x-dev-pack -> security-pack (threat modeling gate)
- 10x-dev-pack -> security-pack (security assessment)

---

## Test Scenario: Payment Processing Integration

### Scenario Description

Implement Stripe Connect integration for processing user payments, including checkout flow, refund handling, and payment webhook processing.

**Why This Scenario**: This represents a SYSTEM-complexity feature that:
- Handles sensitive financial data (PCI-DSS implications)
- Requires threat modeling before design proceeds
- Requires security assessment before production deployment
- Demonstrates blocking cross-rite handoff patterns
- Tests priority=critical and blocking=true handoff flags

---

## Phase 1: Requirements

### Entry Criteria
- [ ] User request or product initiative received
- [ ] Session initialized with `/start`

### Agent
**requirements-analyst**

### Input
User request: "We need to integrate Stripe Connect for processing payments. Users should be able to checkout, and merchants should receive payouts. We also need to handle refunds."

### Expected Artifact: PRD

```markdown
# PRD: Payment Processing Integration

## Problem Statement
Platform lacks payment processing capability, preventing monetization.

## Success Criteria
- Users can complete checkout with card payments
- Merchants receive automatic payouts via Stripe Connect
- Refunds can be processed within 30-day window
- All transactions are logged for audit

## Security Considerations
- PCI-DSS compliance required (SAQ-A minimum)
- No raw card data stored on our systems
- Payment tokens handled securely
- Audit trail for all financial transactions

## User Stories
- US-001: As a buyer, I can pay with credit card
- US-002: As a merchant, I receive payouts to my bank
- US-003: As a buyer, I can request refunds
- US-004: As an admin, I can view payment audit logs
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| PRD exists at `docs/requirements/PRD-payment-processing.md` | YES | [ ] |
| PRD contains Security Considerations section | YES | [ ] |
| Complexity assessed as SYSTEM | YES | [ ] |
| Security gate triggered (auth/crypto/PII detected) | YES | [ ] |

### Phase Transition: SECURITY GATE TRIGGERED

- **From**: requirements
- **To**: security-consultation (BLOCKING)
- **Handoff Type**: Cross-team (10x-dev-pack -> security-pack)
- **Trigger**: PRD contains security considerations AND complexity=SYSTEM

---

## Phase 2: Security Consultation (Cross-Team)

### Entry Criteria
- [ ] PRD complete with security considerations
- [ ] Complexity = SYSTEM
- [ ] Security gate triggered

### HANDOFF Artifact (10x -> security)

```yaml
---
source_team: 10x-dev-pack
target_team: security-pack
handoff_type: assessment
created: 2026-01-02
initiative: Payment Processing Integration
priority: critical
blocking: true
---

## Context

Payment processing feature requires threat modeling before design proceeds.
PRD identifies PCI-DSS compliance and financial transaction security as concerns.

## Source Artifacts
- `docs/requirements/PRD-payment-processing.md`

## Items

### SEC-001: Payment flow threat model
- **Priority**: Critical
- **Summary**: STRIDE analysis for checkout, refund, and payout flows
- **Assessment Questions**:
  - What are the STRIDE threats for payment flows?
  - Are there token exposure risks?
  - Is session binding sufficient for payment sessions?
  - What are the replay attack vectors?

### SEC-002: PCI-DSS compliance review
- **Priority**: Critical
- **Summary**: Validate design approach meets PCI-DSS requirements
- **Assessment Questions**:
  - Does Stripe Connect approach qualify for SAQ-A?
  - Are logging requirements compatible with PCI?
  - Is key rotation needed for API keys?

### SEC-003: Data flow analysis
- **Priority**: High
- **Summary**: Identify sensitive data movement through system
- **Assessment Questions**:
  - What PII flows through the system?
  - Where are payment tokens handled?
  - What audit data is retained?

## Notes for Target Team

This blocks design phase. Architect is waiting on threat model output.
Timeline: 48-hour turnaround requested.
Architect available for design consultation: @architect-lead
```

### Security Team Response: Threat Model

```markdown
# Threat Model: Payment Processing

## STRIDE Analysis

### Spoofing
- Risk: Attacker impersonates user during checkout
- Mitigation: Session binding, CSRF tokens, re-authentication for high-value

### Tampering
- Risk: Price manipulation in checkout
- Mitigation: Server-side price validation, Stripe Checkout Sessions

### Repudiation
- Risk: Disputed transactions without proof
- Mitigation: Comprehensive audit logging, Stripe receipts

### Information Disclosure
- Risk: Card data exposure
- Mitigation: Never store card data, use Stripe Elements

### Denial of Service
- Risk: Checkout flooding
- Mitigation: Rate limiting, payment intent limits

### Elevation of Privilege
- Risk: User processes payment as different user
- Mitigation: User ID binding in payment intent metadata

## PCI-DSS Assessment
- SAQ-A qualification: CONFIRMED (no card data handling)
- Logging: OK (no sensitive data in logs)
- Key rotation: RECOMMENDED (90-day API key rotation)

## Recommendations
1. Use Stripe Checkout (hosted) for checkout flow
2. Implement payment session timeout (15 min)
3. Add fraud detection signals to Stripe Radar
4. Log all payment events without card details

## Verdict: APPROVED TO PROCEED
Design may continue with above mitigations incorporated.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| HANDOFF artifact follows schema | YES | [ ] |
| `blocking: true` set | YES | [ ] |
| `priority: critical` set | YES | [ ] |
| Assessment questions specific | YES | [ ] |
| Threat model returned | YES | [ ] |
| Verdict allows design to proceed | YES | [ ] |

### Phase Transition
- **From**: security-consultation
- **To**: design
- **Handoff Type**: Return to 10x-dev-pack
- **Trigger**: Security team verdict = APPROVED

---

## Phase 3: Design (with Security Input)

### Entry Criteria
- [ ] Threat model approved
- [ ] Security recommendations documented
- [ ] PRD available

### Agent
**architect**

### Input
- PRD from Phase 1
- Threat Model from Phase 2
- Security recommendations

### Expected Artifact: TDD

```markdown
# TDD: Payment Processing Integration

## Design Overview
Stripe Connect integration using Checkout Sessions (hosted) for PCI-DSS SAQ-A compliance.

## Architecture

### Payment Flow
1. User initiates checkout -> Create Stripe Checkout Session
2. Redirect to Stripe hosted checkout
3. Stripe processes payment -> Webhook notification
4. Update order status, trigger payout schedule

### Security Controls (per Threat Model)
- Session binding: Payment intent tied to user session
- Price validation: Server-side price from product catalog
- CSRF protection: Token on checkout initiation
- Rate limiting: 10 checkout attempts per user per hour
- Re-authentication: Required for orders > $500

## Interface Definitions

### API Endpoints
POST /api/v1/checkout/sessions
  Body: { cartId }
  Response: { checkoutUrl, sessionId }
  Security: Authenticated, CSRF token

POST /api/v1/webhooks/stripe (Stripe signature verified)
  Events: checkout.session.completed, charge.refunded

### Webhook Security
- Stripe signature validation (HMAC)
- Idempotency key tracking
- Event replay protection

## Error Handling
- Payment failed: Show user-friendly message, log event
- Webhook replay: Idempotent processing, no double-charge

## ADR Reference
- ADR-0042: Stripe Connect over direct Stripe API
- ADR-0043: Hosted Checkout over embedded Elements
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| TDD exists | YES | [ ] |
| TDD incorporates threat model mitigations | YES | [ ] |
| Security controls section present | YES | [ ] |
| Webhook security documented | YES | [ ] |
| ADRs created for architectural decisions | YES | [ ] |

### Phase Transition
- **From**: design
- **To**: implementation
- **Handoff Type**: Internal (within 10x-dev-pack)
- **Trigger**: TDD complete

---

## Phase 4: Implementation

### Entry Criteria
- [ ] TDD complete with security controls
- [ ] Threat model mitigations understood

### Agent
**principal-engineer**

### Expected Artifacts

1. **Checkout Service**
   - `src/services/checkout/CheckoutService.ts`
   - `src/services/checkout/StripeClient.ts`

2. **Webhook Handler**
   - `src/api/webhooks/stripe.ts`
   - `src/services/checkout/WebhookProcessor.ts`

3. **Security Controls**
   - `src/middleware/checkoutRateLimit.ts`
   - `src/middleware/stripeSignature.ts`

4. **Tests**
   - `tests/services/checkout.test.ts`
   - `tests/webhooks/stripe.test.ts`
   - `tests/security/checkout-security.test.ts`

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| All security controls implemented | YES | [ ] |
| Rate limiting middleware present | YES | [ ] |
| Webhook signature validation present | YES | [ ] |
| Security-focused tests exist | YES | [ ] |
| No card data in logs (grep verification) | YES | [ ] |

### Phase Transition: SECURITY ASSESSMENT GATE

- **From**: implementation
- **To**: security-assessment (BLOCKING)
- **Handoff Type**: Cross-team (10x-dev-pack -> security-pack)
- **Trigger**: Implementation complete for SYSTEM complexity with security considerations

---

## Phase 5: Security Assessment (Cross-Team)

### Entry Criteria
- [ ] Implementation complete
- [ ] Unit tests passing
- [ ] Security controls implemented per TDD

### HANDOFF Artifact (10x -> security)

```yaml
---
source_team: 10x-dev-pack
target_team: security-pack
handoff_type: validation
created: 2026-01-02
initiative: Payment Processing Integration
priority: critical
blocking: true
---

## Context

Payment processing implementation complete. Threat model mitigations implemented.
Needs security validation before QA and production deployment.

## Source Artifacts
- `docs/requirements/PRD-payment-processing.md`
- `docs/design/TDD-payment-processing.md`
- `docs/security/THREAT-MODEL-payment-processing.md`
- `src/services/checkout/` (implementation)

## Items

### VAL-001: Threat mitigation verification
- **Priority**: Critical
- **Summary**: Verify all threat model mitigations are implemented
- **Validation Scope**:
  - Session binding implementation
  - CSRF protection on checkout
  - Rate limiting effectiveness
  - Re-authentication trigger for high-value

### VAL-002: PCI-DSS control verification
- **Priority**: Critical
- **Summary**: Confirm SAQ-A eligibility is maintained
- **Validation Scope**:
  - No card data in application logs
  - No card data in database
  - Stripe hosted checkout used exclusively
  - API keys properly secured

### VAL-003: Webhook security validation
- **Priority**: High
- **Summary**: Validate webhook endpoint security
- **Validation Scope**:
  - Signature validation working
  - Replay protection effective
  - Idempotency keys enforced

## Notes for Target Team

Implementation followed threat model recommendations.
Staging environment available for penetration testing.
Engineer available for walkthroughs: @principal-engineer
```

### Security Team Response: Validation Report

```markdown
# Security Validation: Payment Processing

## Threat Mitigation Verification
- [x] Session binding: PASS - Payment intent metadata includes user_id
- [x] CSRF protection: PASS - Token validated on POST /checkout/sessions
- [x] Rate limiting: PASS - 10/hour limit verified
- [x] Re-auth for high-value: PASS - $500+ triggers re-authentication

## PCI-DSS Control Verification
- [x] No card data in logs: PASS - Grep analysis confirms
- [x] No card data in DB: PASS - Schema review confirms
- [x] Hosted checkout only: PASS - No embedded card fields
- [x] API key security: PASS - Environment variables, not in code

## Webhook Security Validation
- [x] Signature validation: PASS - Tested with invalid signatures
- [x] Replay protection: PASS - Idempotency key rejects replays
- [x] Idempotency: PASS - Double-submit does not double-charge

## Penetration Test Results
- No critical or high vulnerabilities found
- 1 medium: Verbose error messages (FIXED)
- 2 low: Missing security headers (REMEDIATED)

## Verdict: APPROVED FOR QA
All security controls validated. May proceed to QA phase.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| HANDOFF artifact follows schema | YES | [ ] |
| Validation scope clear | YES | [ ] |
| Penetration test conducted | YES | [ ] |
| All critical/high issues resolved | YES | [ ] |
| Verdict allows QA to proceed | YES | [ ] |

### Phase Transition
- **From**: security-assessment
- **To**: validation (QA)
- **Handoff Type**: Return to 10x-dev-pack
- **Trigger**: Security validation verdict = APPROVED

---

## Phase 6: Validation (QA)

### Entry Criteria
- [ ] Security validation passed
- [ ] All security issues resolved
- [ ] Implementation complete

### Agent
**qa-adversary**

### Expected Artifact: Test Plan/Report

```markdown
# Test Plan: Payment Processing

## Functional Tests
- [x] Checkout completes successfully
- [x] Payment appears in Stripe dashboard
- [x] Webhook updates order status
- [x] Refund processes correctly
- [x] Payout scheduled for merchant

## Security Tests (Adversarial)
- [x] Cannot manipulate price client-side
- [x] Cannot checkout as different user
- [x] Rate limiting triggers at threshold
- [x] Invalid webhook signatures rejected
- [x] Replay attacks blocked

## Edge Cases
- [x] Card declined handling
- [x] Webhook timeout and retry
- [x] Partial refund
- [x] Currency conversion

## Result: PASS
All 23 test cases pass. No blocking defects.
Security posture validated.
```

### Handoff Verification
| Check | Expected | Verified |
|-------|----------|----------|
| Test plan includes security tests | YES | [ ] |
| Adversarial testing conducted | YES | [ ] |
| All security controls re-verified in QA | YES | [ ] |
| Overall result documented | YES | [ ] |

---

## Complete Test Checklist

### Phase Completeness
- [ ] Phase 1 (Requirements): PRD with security considerations
- [ ] Phase 2 (Security Consultation): Threat model produced
- [ ] Phase 3 (Design): TDD with security controls
- [ ] Phase 4 (Implementation): Code with security controls
- [ ] Phase 5 (Security Assessment): Validation passed
- [ ] Phase 6 (QA): Full test suite passed

### Security Gates Verification
- [ ] Security gate triggered after PRD (detected SYSTEM + auth/crypto/PII)
- [ ] Blocking handoff to security-pack produced
- [ ] Design waited for threat model approval
- [ ] Security assessment gate triggered after implementation
- [ ] Blocking handoff to security-pack produced
- [ ] QA waited for security validation approval

### Handoff Artifacts
- [ ] HANDOFF-10x-to-security (Phase 2) follows schema
- [ ] HANDOFF-10x-to-security (Phase 5) follows schema
- [ ] Both handoffs set `blocking: true`
- [ ] Both handoffs set `priority: critical`

---

## Running This Test

### Manual Execution

1. Initialize session:
   ```
   /start initiative="Payment Processing" complexity=SYSTEM team=10x-dev-pack
   ```

2. Execute Phase 1:
   ```
   Task(requirements-analyst, "Create PRD for Stripe payment integration...")
   ```

3. Security gate triggers - produce handoff:
   ```
   # Session should detect security concerns and prompt for handoff
   ```

4. Switch to security-pack (or await response):
   ```
   /team security-pack
   Task(threat-modeler, "Analyze payment flow threats per HANDOFF...")
   ```

5. Return to 10x-dev-pack with threat model:
   ```
   /team 10x-dev-pack
   Task(architect, "Design payment system incorporating threat model...")
   ```

6. Continue through implementation, security assessment, and QA.

### Validation Points

| Gate | Trigger Condition | Expected Behavior |
|------|-------------------|-------------------|
| Post-PRD Security | SYSTEM + security considerations | Block until threat model |
| Post-Implementation Security | SYSTEM + security controls | Block until validation |

---

## Related Documents

- [Cross-Team Coordination Playbook](../playbooks/cross-rite-coordination.md)
- [Edge Cases: Cross-Team Workflows](../edge-cases/cross-rite-workflows.md)
- [Security Pack Workflow](../../rites/security-pack/workflow.md)
- [Cross-Team Handoff Schema](../../.claude/skills/shared/cross-rite-handoff/schema.md)
