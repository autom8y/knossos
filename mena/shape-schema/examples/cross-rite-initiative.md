---
type: shape
initiative: external-payment-gateway
frame: .sos/wip/frames/external-payment-gateway.md
created: 2026-03-10
rite: cross-rite
complexity: INITIATIVE
scope:
  rites: [rnd, security, dev, ops]
  sprints: 7
cross_rite_consultations:
  - from: rnd
    to: security
    artifact: .ledge/spikes/payment-gateway-rnd-handoff.md
  - from: security
    to: dev
    artifact: .ledge/spikes/payment-gateway-security-handoff.md
  - from: dev
    to: ops
    artifact: .ledge/spikes/payment-gateway-dev-handoff.md
---

# External Payment Gateway Integration

## Initiative Thread

```yaml
initiative_thread:
  throughline: "Every payment path from checkout to settlement is idempotent, PCI-compliant, and settles within 30 seconds under normal load"
  success_criteria:
    - "All payment endpoints enforce idempotency keys and return consistent results on retry"
    - "PCI DSS SAQ-A compliance verified by security rite with documented evidence"
    - "P95 settlement latency under 30 seconds measured against staging traffic profile"
    - "Rollback path tested: failed settlements reverse within 60 seconds without manual intervention"
  failure_signals:
    - "Prototype cannot achieve sub-30s settlement in isolated testing (sprint-2)"
    - "Security review identifies PCI scope expansion beyond SAQ-A (sprint-3)"
    - "Integration tests show non-idempotent behavior on retry paths (sprint-5)"
```

## Sprint Decomposition

```yaml
sprints:
  - id: sprint-1
    rite: rnd
    mission: "Evaluate gateway provider APIs and validate integration feasibility"
    agents: [technology-scout, integration-researcher]
    entry_criteria:
      - "Initiative frame exists at .sos/wip/frames/external-payment-gateway.md"
    exit_criteria:
      - "Provider comparison matrix with latency, idempotency support, and PCI burden documented"
      - "Recommended provider selected with rationale"
    exit_artifacts:
      - path: ".ledge/spikes/payment-gateway-provider-evaluation.md"
        description: "Provider comparison matrix and recommendation"
    context:
      - ".sos/wip/frames/external-payment-gateway.md"
      - ".know/architecture.md"

  - id: sprint-2
    rite: rnd
    mission: "Build proof-of-concept demonstrating checkout-to-settlement flow with selected provider"
    agents: [prototype-engineer, integration-researcher]
    entry_criteria:
      - "Provider selected in sprint-1 exit artifact"
    exit_criteria:
      - "Working prototype completes checkout-to-settlement in under 30 seconds"
      - "Idempotency key handling demonstrated with duplicate request test"
      - "Error paths documented: timeout, provider rejection, network failure"
    exit_artifacts:
      - path: ".ledge/spikes/payment-gateway-prototype-results.md"
        description: "Prototype findings, latency measurements, error path documentation"
    context:
      - ".ledge/spikes/payment-gateway-provider-evaluation.md"

  - id: sprint-3
    rite: security
    mission: "Validate PCI compliance posture and identify security requirements for production implementation"
    agents: [security-reviewer, compliance-architect]
    entry_criteria:
      - "Prototype results available from sprint-2 handoff artifact"
      - "Provider API documentation accessible"
    exit_criteria:
      - "PCI scope assessment complete with SAQ-A boundary confirmed or escalated"
      - "Security requirements document produced for dev rite consumption"
      - "Token handling and key management approach approved"
    exit_artifacts:
      - path: ".ledge/spikes/payment-gateway-security-requirements.md"
        description: "PCI scope assessment, security requirements, token handling design"
    context:
      - ".ledge/spikes/payment-gateway-rnd-handoff.md"

  - id: sprint-4
    rite: dev
    mission: "Implement payment service with idempotency, retry logic, and provider integration"
    agents: [architect, senior-engineer, engineer]
    entry_criteria:
      - "Security requirements available from sprint-3 handoff"
      - "Prototype code available for reference"
    exit_criteria:
      - "Payment service handles checkout, settlement, and refund flows"
      - "Idempotency enforced at API boundary with persistent key storage"
      - "Unit and integration tests cover happy path and all documented error paths"
    exit_artifacts:
      - path: ".ledge/decisions/ADR-payment-service-architecture.md"
        description: "Architecture decision record for payment service design"
    context:
      - ".ledge/spikes/payment-gateway-security-handoff.md"
      - ".ledge/spikes/payment-gateway-prototype-results.md"
      - ".know/architecture.md"
      - ".know/conventions.md"

  - id: sprint-5
    rite: dev
    mission: "Integrate payment service with existing checkout surfaces and validate end-to-end"
    agents: [senior-engineer, engineer]
    entry_criteria:
      - "Payment service passes unit and integration tests from sprint-4"
    exit_criteria:
      - "Checkout UI calls payment service for all payment methods"
      - "End-to-end test covers: successful payment, failed payment, retry, refund"
      - "Idempotency verified under concurrent request simulation"
    exit_artifacts:
      - path: ".ledge/spikes/payment-gateway-integration-report.md"
        description: "End-to-end test results, performance measurements, integration notes"
    context:
      - ".ledge/decisions/ADR-payment-service-architecture.md"

  - id: sprint-6
    rite: dev
    mission: "Performance validation and hardening under realistic load"
    agents: [senior-engineer]
    entry_criteria:
      - "End-to-end integration complete from sprint-5"
    exit_criteria:
      - "P95 settlement latency under 30 seconds at 2x expected peak load"
      - "Rollback path verified: failed settlements reverse within 60 seconds"
      - "Circuit breaker configured for provider outages"
    exit_artifacts:
      - path: ".ledge/spikes/payment-gateway-perf-report.md"
        description: "Load test results, latency measurements, circuit breaker configuration"
    context:
      - ".ledge/spikes/payment-gateway-integration-report.md"

  - id: sprint-7
    rite: ops
    mission: "Production deployment configuration, monitoring, and alerting"
    agents: [sre-engineer, platform-engineer]
    entry_criteria:
      - "Performance validation complete from sprint-6 handoff"
      - "All security requirements from sprint-3 implemented and verified"
    exit_criteria:
      - "Deployment pipeline configured with canary release strategy"
      - "Monitoring dashboards cover: settlement latency, error rates, provider health"
      - "Alerting configured for: P95 > 30s, error rate > 1%, provider circuit open"
      - "Runbook documents rollback procedure"
    exit_artifacts:
      - path: ".ledge/spikes/payment-gateway-ops-runbook.md"
        description: "Deployment runbook, monitoring configuration, alert definitions"
    context:
      - ".ledge/spikes/payment-gateway-dev-handoff.md"
      - ".ledge/spikes/payment-gateway-perf-report.md"
```

## Potnia Consultation Points

```yaml
checkpoints:
  - id: PT-01
    after: sprint-2
    evaluates: "Feasibility of sub-30s idempotent settlement with selected provider"
    questions:
      - "Does the prototype achieve settlement in under 30 seconds consistently (>95% of attempts)?"
      - "Does the provider API support native idempotency keys, or does the prototype implement application-level idempotency?"
      - "Are all three error paths (timeout, rejection, network failure) documented with observed behavior?"
    gate: hard
    on_fail: "Re-evaluate provider selection. If no provider meets latency requirements, escalate to user for scope adjustment."

  - id: PT-02
    after: sprint-3
    evaluates: "PCI compliance boundary and security feasibility"
    questions:
      - "Is PCI scope confirmed as SAQ-A, or has the assessment identified scope expansion?"
      - "Does the token handling approach avoid storing raw card data in any system under our control?"
      - "Are security requirements specific enough for dev rite agents to implement without further security consultation?"
    gate: hard
    on_fail: "If PCI scope exceeds SAQ-A, escalate to user -- this changes initiative cost and timeline significantly."

  - id: PT-03
    after: sprint-5
    evaluates: "End-to-end idempotency and integration completeness"
    questions:
      - "Do concurrent duplicate requests with the same idempotency key produce identical responses without double-charging?"
      - "Does the end-to-end test cover all payment methods exposed in the checkout UI?"
      - "Are retry paths tested with simulated provider failures, not just happy-path retries?"
    gate: hard
    on_fail: "Rework sprint-5 with tightened exit criteria on idempotency testing. Do not proceed to performance validation with unverified retry behavior."

  - id: PT-04
    after: sprint-6
    evaluates: "Throughline latency and resilience under load"
    questions:
      - "Does P95 settlement latency remain under 30 seconds at 2x expected peak load?"
      - "Does the circuit breaker correctly isolate provider outages without affecting non-payment flows?"
      - "Does the rollback path reverse failed settlements within 60 seconds in load test conditions?"
    gate: soft
    on_fail: "Log performance gap. If P95 is between 30-45s, proceed with ops sprint and add optimization to backlog. If P95 > 45s, insert remediation sprint before ops."

  - id: PT-05
    after: sprint-7
    evaluates: "Production readiness against full throughline"
    questions:
      - "Does the monitoring configuration cover all three throughline metrics: latency, idempotency, and settlement success rate?"
      - "Does the runbook include rollback procedures that match the tested rollback path from sprint-6?"
      - "Is the canary release configured to gate on settlement error rate, not just HTTP status codes?"
    gate: hard
    on_fail: "Rework sprint-7 monitoring and alerting gaps before production deployment."
```

## Cross-Rite Handoff Protocol

```yaml
cross_rite_handoffs:
  - from_rite: rnd
    to_rite: security
    artifact:
      path: ".ledge/spikes/payment-gateway-rnd-handoff.md"
      produces: "Provider selection rationale, prototype architecture, API integration patterns, latency measurements, error path documentation"
      consumes: "Provider API documentation references, data flow diagram showing where card data is handled, prototype code location"
    verification:
      - "Handoff artifact contains provider API endpoint documentation"
      - "Data flow diagram clearly marks PCI scope boundaries"

  - from_rite: security
    to_rite: dev
    artifact:
      path: ".ledge/spikes/payment-gateway-security-handoff.md"
      produces: "PCI scope assessment, security requirements with acceptance criteria, approved token handling design, key management requirements"
      consumes: "Security requirements as implementable acceptance criteria (not abstract policies)"
    verification:
      - "Each security requirement has a testable acceptance criterion"
      - "Token handling design specifies exact storage and transit encryption requirements"

  - from_rite: dev
    to_rite: ops
    artifact:
      path: ".ledge/spikes/payment-gateway-dev-handoff.md"
      produces: "Service architecture overview, deployment requirements, environment variables and secrets needed, health check endpoints, performance baseline from load tests"
      consumes: "Deployment configuration requirements, monitoring metric names, alert thresholds"
    verification:
      - "Health check endpoints documented with expected response format"
      - "Performance baseline includes P50, P95, P99 latency measurements"
```

## Emergent Behavior Constraints

```yaml
behavior_constraints:
  prescribed:
    - "All payment amounts stored as integer cents -- no floating point currency"
    - "Idempotency keys must be persistent (database-backed), not in-memory"
    - "No raw card data stored in any system -- tokenized at provider boundary"
    - "All provider API calls logged with correlation IDs for audit trail"
  emergent:
    - "Internal service API design (REST, gRPC, or hybrid)"
    - "Database schema for payment records and idempotency keys"
    - "Test fixture design and mock provider implementation"
    - "Error message wording and HTTP status code mapping"
    - "Code organization within the payment service module"
  out_of_scope:
    - "Existing billing module refactoring -- payment service is additive"
    - "User account management or authentication changes"
    - "Checkout UI redesign beyond adding payment method selection"
    - "Provider contract negotiation or pricing"
```

## Execution Sequence

```yaml
execution_sequence:
  - phase: setup
    commands:
      - "ari sync --rite=rnd"
      - "/sos start --initiative=external-payment-gateway"

  - phase: sprint-1
    commands:
      - "/sprint 1"
      - "Read('.sos/wip/frames/external-payment-gateway.md')"
      - "Read('.know/architecture.md')"
    notes: "Technology scout evaluates providers; integration researcher maps API compatibility"

  - phase: sprint-2
    commands:
      - "/sprint 2"
      - "Read('.ledge/spikes/payment-gateway-provider-evaluation.md')"
    notes: "Prototype engineer builds PoC with selected provider"

  - phase: checkpoint-PT-01
    commands:
      - "# Potnia evaluates prototype against sub-30s settlement and idempotency requirements"
    notes: "Hard gate -- blocks rite transition to security if prototype fails feasibility check"

  - phase: rite-transition-to-security
    commands:
      - "/sos wrap"
      - "ari sync --rite=security"
      - "/sos start --initiative=external-payment-gateway"
    notes: "Security rite Potnia loads rnd handoff artifact for PCI assessment context"

  - phase: sprint-3
    commands:
      - "/sprint 3"
      - "Read('.ledge/spikes/payment-gateway-rnd-handoff.md')"
    notes: "Security reviewers assess PCI scope; compliance architect validates token handling"

  - phase: checkpoint-PT-02
    commands:
      - "# Potnia evaluates PCI scope and security requirements completeness"
    notes: "Hard gate -- PCI scope expansion requires user escalation"

  - phase: rite-transition-to-dev
    commands:
      - "/sos wrap"
      - "ari sync --rite=dev"
      - "/sos start --initiative=external-payment-gateway"
    notes: "Dev rite Potnia loads security handoff for implementation requirements"

  - phase: sprint-4
    commands:
      - "/sprint 4"
      - "Read('.ledge/spikes/payment-gateway-security-handoff.md')"
      - "Read('.ledge/spikes/payment-gateway-prototype-results.md')"
      - "Read('.know/architecture.md')"
      - "Read('.know/conventions.md')"
    notes: "Core payment service implementation"

  - phase: sprint-5
    commands:
      - "/sprint 5"
      - "Read('.ledge/decisions/ADR-payment-service-architecture.md')"
    notes: "Integration with checkout surfaces and end-to-end validation"

  - phase: checkpoint-PT-03
    commands:
      - "# Potnia evaluates idempotency under concurrent requests and integration completeness"
    notes: "Hard gate -- idempotency failures block performance validation"

  - phase: sprint-6
    commands:
      - "/sprint 6"
      - "Read('.ledge/spikes/payment-gateway-integration-report.md')"
    notes: "Load testing and hardening"

  - phase: checkpoint-PT-04
    commands:
      - "# Potnia evaluates latency and resilience under load"
    notes: "Soft gate -- marginal performance gaps logged, severe gaps insert remediation sprint"

  - phase: rite-transition-to-ops
    commands:
      - "/sos wrap"
      - "ari sync --rite=ops"
      - "/sos start --initiative=external-payment-gateway"
    notes: "Ops rite Potnia loads dev handoff for deployment context"

  - phase: sprint-7
    commands:
      - "/sprint 7"
      - "Read('.ledge/spikes/payment-gateway-dev-handoff.md')"
      - "Read('.ledge/spikes/payment-gateway-perf-report.md')"
    notes: "Deployment pipeline, monitoring, alerting, runbook"

  - phase: checkpoint-PT-05
    commands:
      - "# Potnia evaluates production readiness against full throughline"
    notes: "Hard gate -- final validation before production deployment"

  - phase: wrap
    commands:
      - "/sos wrap"
    notes: "Initiative complete. Shape file serves as execution record."
```
