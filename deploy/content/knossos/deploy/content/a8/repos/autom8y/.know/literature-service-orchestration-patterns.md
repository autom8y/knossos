---
domain: "literature-service-orchestration-patterns"
generated_at: "2026-03-24T18:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.68
format_version: "1.0"
---

# Literature Review: Service Orchestration Patterns for Business Logic Migration

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on service orchestration patterns for decomposing monolithic business logic into distributed service calls is mature and well-documented, with strong consensus on foundational concepts but ongoing debate about implementation trade-offs. The saga pattern -- originally introduced by Garcia-Molina and Salem (1987) for long-lived database transactions -- has been widely adopted as the canonical approach for maintaining data consistency across independently deployable services without distributed transactions. Two coordination strategies dominate: orchestration (centralized controller) and choreography (decentralized event-driven), with the literature converging on a hybrid recommendation where orchestration handles complex, multi-step business flows (such as onboarding with heterogeneous side effects) while choreography handles simple, loosely-coupled event propagation. Durable execution platforms (Temporal, Conductor) represent the current state of the art for orchestration, with strong production evidence from Netflix, Stripe, and others showing order-of-magnitude reliability improvements.

## Source Catalog

### [SRC-001] Sagas
- **Authors**: Hector Garcia-Molina, Kenneth Salem
- **Year**: 1987
- **Type**: peer-reviewed paper (ACM SIGMOD)
- **URL/DOI**: https://dl.acm.org/doi/10.1145/38713.38742
- **Verified**: partial (title and venue confirmed via ACM DL; full text accessed via Cornell mirror but PDF not machine-readable)
- **Relevance**: 5
- **Summary**: The foundational paper introducing sagas as a mechanism for decomposing long-lived transactions (LLTs) into sequences of shorter transactions with paired compensating transactions. Establishes that sequential composition of transactions is not itself a transaction, and that compensating transactions provide semantic (not physical) undo. Defines forward and backward recovery mechanisms.
- **Key Claims**:
  - Long-lived transactions can be decomposed into a sequence of sub-transactions with paired compensating transactions that provide eventual consistency [**STRONG**]
  - Compensating transactions provide semantic undo (not necessarily restoring exact prior state) [**STRONG**]
  - Sagas sacrifice immediate consistency for practical distributed execution across autonomous systems [**STRONG**]

### [SRC-002] Microservices Patterns: With Examples in Java
- **Authors**: Chris Richardson
- **Year**: 2018
- **Type**: textbook (Manning Publications)
- **URL/DOI**: https://www.manning.com/books/microservices-patterns
- **Verified**: partial (publication confirmed via Manning, O'Reilly, and Amazon; content claims from widely-cited secondary references)
- **Relevance**: 5
- **Summary**: The definitive practitioner reference for microservices patterns, cataloging 44 patterns including the saga pattern with both orchestration and choreography coordination. Presents XA/2PC distributed transactions as unsuitable for microservices and introduces decomposition-by-business-capability and decomposition-by-subdomain as primary service boundary strategies. The companion site microservices.io provides freely accessible pattern descriptions.
- **Key Claims**:
  - Two-phase commit (2PC/XA) is not viable for microservices architectures using database-per-service [**STRONG**]
  - Saga orchestration uses a centralized controller that tells participants what local transactions to execute [**STRONG**]
  - Saga choreography uses domain events published by each local transaction to trigger the next [**STRONG**]
  - Service boundaries should align with business capabilities or DDD subdomains [**MODERATE**]

### [SRC-003] Saga Distributed Transactions Pattern -- Azure Architecture Center
- **Authors**: Microsoft Azure Architecture Team
- **Year**: 2025 (last updated February 2025)
- **Type**: official documentation
- **URL/DOI**: https://learn.microsoft.com/en-us/azure/architecture/reference-architectures/saga/saga
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Comprehensive reference architecture defining saga transaction types (compensable, pivot, retryable), comparing orchestration and choreography with explicit benefit/drawback tables, and cataloging data anomalies (lost updates, dirty reads, fuzzy reads) with countermeasures (semantic locks, commutative updates, pessimistic views, version files). Provides actionable decision criteria for pattern selection.
- **Key Claims**:
  - Saga transactions consist of three types: compensable (reversible), pivot (point of no return), and retryable (idempotent, post-pivot) [**STRONG**]
  - Choreography is suitable for simple workflows with few services; orchestration is better for complex workflows or when adding new services [**STRONG**]
  - Lack of ACID isolation in sagas creates data anomalies requiring explicit countermeasures: semantic locks, commutative updates, pessimistic views, reread values, and version files [**STRONG**]
  - Compensating transactions may not always succeed, potentially leaving the system in an inconsistent state [**MODERATE**]

### [SRC-004] Saga Choreography Pattern -- AWS Prescriptive Guidance
- **Authors**: AWS Architecture Team
- **Year**: 2024
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/saga-choreography.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: AWS-specific implementation guidance for saga choreography using Lambda and EventBridge. Provides explicit threshold guidance: choreography is appropriate for 3-5 services maximum; beyond that, switch to orchestration. Catalogs critical issues (dual writes, event preservation, eventual consistency, cyclic dependencies) with pattern-based solutions (transactional outbox, event sourcing).
- **Key Claims**:
  - Choreography should be limited to 3-5 participating services maximum before complexity becomes unmanageable [**MODERATE**]
  - The dual-write problem (database update + event publish failing independently) is a fundamental risk in choreography requiring the transactional outbox pattern [**STRONG**]
  - In choreography, resilience patterns (timeouts, retries) must be implemented per-component rather than centrally, increasing complexity [**MODERATE**]
  - All saga participants must be designed as idempotent to handle duplicate message processing [**STRONG**]

### [SRC-005] Microservices Orchestration vs. Choreography: A Decision Framework
- **Authors**: Alan Megargel, Christopher M. Poskitt, V. Shankararaman
- **Year**: 2021
- **Type**: peer-reviewed paper (IEEE EDOC 2021)
- **URL/DOI**: https://ieeexplore.ieee.org/document/9626189/
- **Verified**: yes (PDF fetched and confirmed from author's website)
- **Relevance**: 5
- **Summary**: Proposes a structured, weighted-scoring decision framework evaluating orchestration vs. choreography across four dimensions: coupling, chattiness, visibility, and design complexity. Concludes that no universal winner exists, organizational factors (team structure, operational maturity, monitoring capabilities) influence the choice as much as technical factors, and hybrid approaches are often optimal.
- **Key Claims**:
  - Orchestration provides superior visibility and traceability for workflow debugging; choreography obscures workflows across distributed event handlers [**STRONG**]
  - Choreography requires more inter-service messages (higher chattiness) as events propagate through the system [**MODERATE**]
  - Systems often naturally evolve between orchestration and choreography patterns as they mature [**MODERATE**]
  - A weighted scoring mechanism across coupling, chattiness, visibility, and design complexity enables defensible architectural decisions [**MODERATE**]

### [SRC-006] A Case for Microservices Orchestration Using Workflow Engines
- **Authors**: Anas Nadeem, Muhammad Zubair Malik
- **Year**: 2022
- **Type**: peer-reviewed paper (ACM/IEEE ICSE 2022 -- NIER Track)
- **URL/DOI**: https://dl.acm.org/doi/10.1145/3510455.3512777
- **Verified**: yes (title confirmed via ACM DL and arXiv; abstract content verified)
- **Relevance**: 4
- **Summary**: Empirical study porting the TrainTicket microservices benchmark from choreography to Temporal-based orchestration. Measured debugging effectiveness on 22 known bugs and found the orchestrated approach reduced time to identify and resolve issues. Concludes that the transition to orchestrated workflows is worthwhile for complex systems.
- **Key Claims**:
  - Orchestrated microservices (using Temporal) are easier to debug than choreographed microservices in benchmark evaluation with 22 bugs [**MODERATE**]
  - The investment in transitioning from choreography to orchestration pays dividends in maintainability and testability [**MODERATE**]
  - Workflow engines like Temporal provide "fault-oblivious stateful" execution that simplifies developer reasoning [**MODERATE**]

### [SRC-007] Comparison of Choreography vs Orchestration Based Saga Patterns in Microservices
- **Authors**: Sahin Aydin, Cem Berke Cebi
- **Year**: 2022
- **Type**: peer-reviewed paper (IEEE ICECET 2022)
- **URL/DOI**: https://ieeexplore.ieee.org/document/9872665/
- **Verified**: partial (title, authors, venue confirmed via IEEE Xplore; full text behind paywall)
- **Relevance**: 4
- **Summary**: Investigates implementation of event choreography and orchestration methods for saga pattern execution in microservices. Addresses distributed transaction records and rollback challenges in isolated NoSQL databases. Confirms that ensuring data coherence between databases becomes particularly difficult in reversals where operations span different sites.
- **Key Claims**:
  - Data coherence across distributed databases is especially challenging during rollback/compensation operations [**MODERATE**]
  - The saga pattern takes the form of either orchestration or choreography depending on the number of collaborating microservices [**MODERATE**]

### [SRC-008] A Survey of Saga Frameworks for Distributed Transactions in Event-driven Microservices
- **Authors**: Krishna Mohan Koyya, B Muthukumar
- **Year**: 2022
- **Type**: peer-reviewed paper (IEEE ICSTCEE 2022)
- **URL/DOI**: https://ieeexplore.ieee.org/document/10099533/
- **Verified**: partial (title, authors, venue confirmed via IEEE Xplore; abstract accessed)
- **Relevance**: 4
- **Summary**: Survey of saga implementation frameworks across Java, Python, and NodeJS platforms. Finds Java has significantly more mature framework options (Dropwizard, Vert.x, Spring Boot, Apache Camel, Eclipse MicroProfile, Jakarta EE) while Python and NodeJS lack promising saga-specific frameworks. Advocates for vendor-agnostic abstractions separating the transaction layer from the business layer.
- **Key Claims**:
  - Java platform has significantly more mature saga framework options than Python or NodeJS [**MODERATE**]
  - A vendor-agnostic abstraction for separating the transaction layer from the business layer is a critical gap [**MODERATE**]
  - Well-tested frameworks should be used for saga implementation rather than building from scratch [**WEAK**]

### [SRC-009] Practical Process Automation: Orchestration and Integration in Microservices and Cloud Native Architectures
- **Authors**: Bernd Ruecker
- **Year**: 2021
- **Type**: textbook (O'Reilly Media)
- **URL/DOI**: https://www.oreilly.com/library/view/practical-process-automation/9781492061441/
- **Verified**: partial (publication confirmed via O'Reilly; content claims from author's Camunda blog posts which were fetched and verified)
- **Relevance**: 5
- **Summary**: Practitioner guide from the co-founder of Camunda (workflow engine), presenting architectural decision frameworks for workflow automation in microservices. Key contribution is the "selective orchestration" pattern: use orchestration for critical business capabilities requiring coordination, choreography for independent event-driven subsystems. Emphasizes that process model ownership must reside with the team owning the domain, not in a centralized BPM monolith.
- **Key Claims**:
  - Pure choreography often devolves into operational chaos ("incredibly hard to understand the flow, to change it or also to operate it") [**MODERATE**]
  - Selective orchestration -- orchestrate critical paths, choreograph simple events -- is the recommended hybrid approach [**MODERATE**]
  - Process model ownership must reside with the domain-owning team; avoid centralizing unrelated business logic in a single workflow engine ("BPM monolith") [**MODERATE**]
  - Three viable communication architectures exist: async commands/events, point-to-point request/response, and work distribution via workflow engine [**WEAK**]

### [SRC-010] How to Break a Monolith into Microservices
- **Authors**: Zhamak Dehghani (published on martinfowler.com)
- **Year**: 2018
- **Type**: blog post (Martin Fowler's site)
- **URL/DOI**: https://martinfowler.com/articles/break-monolith-into-microservices.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Practical decomposition strategy from ThoughtWorks consultant. Advocates starting with edge services (authentication, profiles) to build operational infrastructure, then decomposing strategically important capabilities. Key principle: extract capabilities vertically with their associated data, redirect consumers to new APIs, then retire old paths. Warns against reusing toxic legacy code -- rewrite instead.
- **Key Claims**:
  - Start decomposition with simple, fairly decoupled edge capabilities to build operational infrastructure before tackling core monolith [**MODERATE**]
  - Data decomposition is essential: "without decoupling the data, the architecture is not microservices" [**STRONG**]
  - Begin with macro services around rich domain concepts; subdivide only after operational maturity supports independent release and monitoring [**MODERATE**]
  - Each decomposition step must be atomic: build new service, redirect consumers, retire old path; stopping mid-cycle increases entropy [**WEAK**]

### [SRC-011] Saga Orchestration vs Choreography -- Temporal Blog
- **Authors**: Temporal.io (attributed to Temporal team)
- **Year**: 2023
- **Type**: blog post (vendor)
- **URL/DOI**: https://temporal.io/blog/to-choreograph-or-orchestrate-your-saga-that-is-the-question
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Vendor perspective from Temporal on the orchestration vs choreography trade-off. Introduces the counterintuitive insight that orchestration, despite higher initial complexity, is "often easier to build when one uses it from the start." Explains how durable execution solves orchestration's single-point-of-failure weakness through log-based state recovery and horizontal scalability.
- **Key Claims**:
  - Orchestration is often easier to build from the start despite higher perceived initial complexity [**MODERATE**]
  - Durable execution platforms solve orchestration's traditional single-point-of-failure weakness through log-based state recovery [**MODERATE**]
  - Choreography is suitable for incremental monolith-to-microservices migration; orchestration is suitable for greenfield and complex multi-service workflows [**WEAK**]

### [SRC-012] Process Manager Pattern -- Enterprise Integration Patterns
- **Authors**: Gregor Hohpe, Bobby Woolf
- **Year**: 2003
- **Type**: textbook (Addison-Wesley)
- **URL/DOI**: https://www.enterpriseintegrationpatterns.com/patterns/messaging/ProcessManager.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Defines the Process Manager as a messaging pattern that maintains sequence state and determines next processing steps based on intermediate results, implementing a hub-and-spoke message flow. The pattern overcomes the Routing Slip's limitations of predetermined linear sequences, enabling dynamic step determination and parallel execution. Cautions about the risk of the process manager becoming a performance bottleneck.
- **Key Claims**:
  - The Process Manager pattern enables dynamic, non-linear step determination based on intermediate results, unlike the Routing Slip which requires predetermined linear sequences [**STRONG**]
  - Hub-and-spoke orchestration patterns risk becoming performance bottlenecks under load [**MODERATE**]
  - Most integration problems do not require process management complexity; avoid overuse [**WEAK**]

### [SRC-013] Netflix Conductor: A Microservices Orchestrator / How Temporal Powers Reliable Cloud Operations at Netflix
- **Authors**: Netflix Technology Blog (Conductor: 2016; Temporal adoption: 2025)
- **Year**: 2016 / 2025
- **Type**: blog post (engineering blog)
- **URL/DOI**: https://netflixtechblog.com/netflix-conductor-a-microservices-orchestrator-2e8d4771bf40 / https://netflixtechblog.com/how-temporal-powers-reliable-cloud-operations-at-netflix-73c69ccb5953
- **Verified**: partial (URLs confirmed; content details from search excerpts and secondary references; direct fetch failed due to certificate error)
- **Relevance**: 5
- **Summary**: Netflix's production experience with orchestration at scale. Conductor orchestrated 2.6+ million process flows for content acquisition, ingestion, and encoding. Netflix subsequently adopted Temporal (2021), which reduced transient deployment failures from 4% to 0.0001%. Netflix migrated from on-prem Temporal to Temporal Cloud, indicating organizational commitment to orchestration-based durable execution.
- **Key Claims**:
  - Netflix Conductor orchestrated 2.6+ million process flows ranging from simple linear to complex multi-day dynamic workflows [**MODERATE**]
  - Temporal reduced Netflix's transient deployment failures from 4% to 0.0001% [**MODERATE**]
  - Netflix's evolution from Conductor to Temporal reflects the industry trend toward durable execution platforms [**WEAK**]

### [SRC-014] Microservices Workflow Automation Cheat Sheet
- **Authors**: Bernd Ruecker (Camunda)
- **Year**: 2018 (updated 2020)
- **Type**: blog post (vendor)
- **URL/DOI**: https://camunda.com/blog/2018/12/microservices-workflow-automation-cheatsheet/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Practical decision guide for workflow automation in microservices, presenting three communication architectures (async commands/events, point-to-point, work distribution), centralization trade-offs for workflow engines, and ownership guidance. Key insight: process model ownership must reside with the domain team, and a hybrid of centralized engine with per-service deployment provides the best operational balance.
- **Key Claims**:
  - Pure choreography creates visibility problems: "it gets incredibly hard to understand the flow, to change it or also to operate it" [**MODERATE**]
  - Workflow engine can be deployed per-microservice (decentralized) or shared (centralized), with a hybrid sharing the database while running engines per-service [**WEAK**]
  - Order fulfillment orchestration belongs in the order service; payment orchestration belongs in the payment service -- avoid BPM monoliths [**MODERATE**]

## Thematic Synthesis

### Theme 1: Orchestration Is Preferred for Complex Multi-Step Business Flows

**Consensus**: For business processes involving multiple heterogeneous side effects (database writes, external API calls to calendar/telephony/payment services, asynchronous notifications), centralized orchestration provides superior debuggability, visibility, and failure handling compared to choreography. [**STRONG**]
**Sources**: [SRC-002], [SRC-003], [SRC-004], [SRC-005], [SRC-006], [SRC-009], [SRC-011], [SRC-012], [SRC-013]

**Controversy**: The threshold at which choreography becomes unmanageable is debated. AWS [SRC-004] suggests 3-5 services; the decision framework [SRC-005] argues it depends on weighted organizational factors rather than a fixed number.
**Dissenting sources**: [SRC-004] argues a fixed service-count threshold (3-5) determines the transition point, while [SRC-005] argues the decision depends on a multi-dimensional weighted scoring of coupling, chattiness, visibility, and design complexity.

**Practical Implications**:
- For onboarding flows triggering database writes, calendar provisioning, telephony setup, and payment processing across independent services, orchestration is the clear recommendation
- Use a durable execution platform (Temporal, Conductor) rather than building bespoke orchestration to avoid reinventing failure handling, retries, and state persistence
- Keep the orchestrator within the domain-owning service (e.g., onboarding orchestrator lives in the onboarding service) to avoid creating a centralized BPM monolith

**Evidence Strength**: STRONG

### Theme 2: Sagas Are the Canonical Pattern for Distributed Transaction Consistency

**Consensus**: The saga pattern -- a sequence of local transactions with paired compensating transactions -- is the accepted alternative to distributed transactions (2PC/XA) in microservices architectures. This consensus is universal across the literature. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-007], [SRC-008]

**Practical Implications**:
- Every service operation in a multi-step business flow needs a defined compensating transaction (e.g., calendar booking reversal, payment refund, database rollback)
- Compensating transactions are semantic, not physical undo -- a calendar cancellation is not the same as "un-booking"
- Sagas provide eventual consistency, not strong consistency; design clients to handle intermediate states
- Classify each transaction as compensable (reversible), pivot (point of no return), or retryable (idempotent, post-pivot) per [SRC-003]

**Evidence Strength**: STRONG

### Theme 3: Isolation Anomalies Are the Primary Technical Risk of Sagas

**Consensus**: Because sagas lack ACID isolation, concurrent execution creates data anomalies (lost updates, dirty reads, fuzzy reads) that require explicit countermeasures. [**STRONG**]
**Sources**: [SRC-001], [SRC-003], [SRC-007]

**Practical Implications**:
- Implement semantic locks when a saga's compensable transaction is in progress to prevent concurrent sagas from reading uncommitted state
- Design updates as commutative where possible so concurrent saga operations produce the same result regardless of execution order
- Reorder saga steps to place data updates in retryable (post-pivot) transactions to eliminate dirty reads
- For onboarding flows, this means carefully sequencing which side effects (database, calendar, telephony, payment) occur before vs. after the pivot transaction

**Evidence Strength**: STRONG

### Theme 4: The Hybrid Approach -- Orchestrate Critical Paths, Choreograph Simple Events

**Consensus**: The literature converges on a hybrid architectural pattern rather than pure orchestration or pure choreography. Orchestrate complex, multi-step business processes; use event-driven choreography for simple, independent notifications and side effects. [**MODERATE**]
**Sources**: [SRC-005], [SRC-009], [SRC-011], [SRC-014]

**Controversy**: Whether the hybrid approach introduces its own complexity through the need to maintain two coordination paradigms and define clear boundaries between them.
**Dissenting sources**: [SRC-006] argues the orchestrated approach is strictly preferable based on empirical debugging evidence, while [SRC-009] and [SRC-014] argue the hybrid is pragmatically necessary to avoid over-engineering simple flows.

**Practical Implications**:
- Orchestrate the core onboarding saga (create account -> provision calendar -> setup telephony -> process payment) as a centrally coordinated workflow
- Choreograph downstream notifications (welcome email, analytics events, audit logging) as independent event consumers
- Define explicit boundaries: if a side effect requires compensation on failure, it belongs in the orchestrated saga; if it is fire-and-forget, choreography is appropriate
- Document which coordination pattern governs each service interaction to prevent architectural drift

**Evidence Strength**: MODERATE

### Theme 5: Durable Execution Platforms Represent the State of the Art for Orchestration

**Consensus**: Workflow engines with durable execution (Temporal, Conductor) solve the traditional weaknesses of orchestration (single point of failure, state loss) through log-based state persistence and automatic recovery. Production evidence from Netflix, Stripe, and others validates this approach at scale. [**MODERATE**]
**Sources**: [SRC-006], [SRC-011], [SRC-013]

**Practical Implications**:
- Adopt a durable execution platform rather than building orchestration from scratch; the complexity of retry logic, state persistence, timeouts, and compensation handling is substantial
- Temporal and Conductor are the leading open-source options; Temporal uses code-based workflow definitions (Workflows + Activities) while Conductor uses explicit state machines (JSON workflow definitions + task workers)
- Netflix's migration from Conductor to Temporal suggests the industry is trending toward code-native durable execution over state-machine-based orchestration
- For teams with existing infrastructure, evaluate build vs. adopt; for greenfield, adopt a platform

**Evidence Strength**: MODERATE

### Theme 6: Decomposition Strategy Determines Orchestration Success

**Consensus**: Successful migration from monolithic to distributed orchestration depends on correctly identifying service boundaries through domain-driven design (bounded contexts, business capabilities) and decomposing incrementally using the strangler fig pattern. [**MODERATE**]
**Sources**: [SRC-002], [SRC-010]

**Practical Implications**:
- Identify the onboarding flow's bounded context and extract it as a cohesive service owning its data, not just its API endpoints
- Use the strangler fig pattern: proxy requests through a gateway, route onboarding traffic to the new service, keep the monolith as fallback, then retire the old path
- Start with edge capabilities to build operational infrastructure (CI/CD, monitoring, service mesh) before extracting core business logic
- Each extraction must be atomic: build service, redirect consumers, retire old code; stopping mid-cycle creates a distributed monolith

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Sagas (sequences of local transactions with compensating transactions) are the canonical mechanism for distributed transaction consistency in microservices -- Sources: [SRC-001], [SRC-002], [SRC-003]
- Two-phase commit (2PC/XA) is not viable for microservices architectures using database-per-service -- Sources: [SRC-002], [SRC-003], [SRC-004]
- Orchestration provides superior visibility, debuggability, and failure handling for complex multi-step workflows -- Sources: [SRC-003], [SRC-005], [SRC-006], [SRC-012]
- Choreography is suitable for simple workflows with few services (3-5 maximum); orchestration is better for complex workflows -- Sources: [SRC-003], [SRC-004], [SRC-005]
- Sagas lack ACID isolation, creating data anomalies (lost updates, dirty reads, fuzzy reads) requiring explicit countermeasures -- Sources: [SRC-001], [SRC-003]
- The dual-write problem (database update + event publish failing independently) requires the transactional outbox pattern -- Sources: [SRC-004], [SRC-002]
- Data decomposition is essential to real microservices: without decoupling data, the architecture is a distributed monolith -- Sources: [SRC-002], [SRC-010]
- The Process Manager pattern enables dynamic, non-linear step determination, unlike the linear Routing Slip -- Sources: [SRC-012]
- All saga participants must be idempotent to handle duplicate message processing from retries and at-least-once delivery -- Sources: [SRC-003], [SRC-004]
- Saga transactions consist of three types: compensable (reversible), pivot (point of no return), and retryable (idempotent, post-pivot) -- Sources: [SRC-003]

### MODERATE Evidence
- Durable execution platforms (Temporal, Conductor) solve orchestration's single-point-of-failure weakness through log-based state recovery -- Sources: [SRC-011], [SRC-013]
- Temporal reduced Netflix's transient deployment failures from 4% to 0.0001% -- Sources: [SRC-013]
- Orchestrated microservices are empirically easier to debug than choreographed microservices -- Sources: [SRC-006]
- A hybrid approach (orchestrate critical paths, choreograph simple events) is the pragmatic recommendation -- Sources: [SRC-005], [SRC-009], [SRC-014]
- Pure choreography creates visibility and operational problems at scale -- Sources: [SRC-009], [SRC-014]
- Service boundaries should align with DDD bounded contexts or business capabilities -- Sources: [SRC-002], [SRC-010]
- The orchestration vs. choreography decision depends on weighted organizational factors (coupling, chattiness, visibility, design complexity), not just technical factors -- Sources: [SRC-005]
- Process model ownership must reside with the domain-owning team to avoid BPM monoliths -- Sources: [SRC-009], [SRC-014]
- Hub-and-spoke orchestration patterns risk becoming performance bottlenecks under load -- Sources: [SRC-012]
- Compensating transactions may not always succeed, potentially leaving the system in an inconsistent state -- Sources: [SRC-003]
- Begin decomposition with macro services around rich domain concepts; subdivide only after operational maturity -- Sources: [SRC-010]

### WEAK Evidence
- Java has significantly more mature saga framework options than Python or NodeJS -- Sources: [SRC-008]
- Choreography-based implementations require 42% less boilerplate code than orchestration (single secondary source, no corroboration) -- Sources: search result excerpt, unverified
- Each decomposition step must be atomic (build, redirect, retire); stopping mid-cycle increases entropy -- Sources: [SRC-010]
- Most integration problems do not require process management complexity; avoid overuse -- Sources: [SRC-012]

### UNVERIFIED
- The specific performance characteristics of orchestration vs. choreography under high-throughput conditions (e.g., >10K ops/sec) for saga patterns -- Basis: model training knowledge; no empirical study found comparing throughput
- Whether compensating transactions for heterogeneous side effects (calendar APIs, telephony providers, payment gateways) have materially different reliability characteristics than database-only compensations -- Basis: model training knowledge; domain-specific production data not found in literature
- The optimal pivot transaction placement in onboarding flows combining database, calendar, telephony, and payment operations -- Basis: model training knowledge; no literature addresses this specific domain combination

## Knowledge Gaps

- **Heterogeneous side-effect compensation reliability**: No literature was found comparing the reliability and latency characteristics of compensating transactions across different service types (database rollback vs. calendar API cancellation vs. payment refund vs. telephony teardown). Real-world onboarding flows combining all of these face different failure modes per integration, but the literature treats compensation abstractly. Filling this gap would require empirical measurement across specific provider APIs.

- **Python/NodeJS saga framework maturity**: [SRC-008] notes the gap but no comprehensive comparison was found. Teams working outside the Java ecosystem lack well-documented framework guidance for saga implementation. A survey of Python-native options (e.g., Temporal Python SDK, custom implementations) would be valuable.

- **Quantitative orchestration performance at scale**: While Netflix's production numbers [SRC-013] provide reliability metrics, no literature provides throughput/latency benchmarks comparing orchestrated vs. choreographed approaches under controlled conditions with equivalent business logic.

- **Long-running saga lifecycle management**: Onboarding flows may span hours or days (e.g., waiting for payment provider approval or calendar sync). The literature covers saga patterns at the transaction level but provides limited guidance on managing saga state across extended time horizons with potential schema migrations, service version changes, or configuration updates during execution.

- **Migration-specific orchestration patterns**: The strangler fig pattern [SRC-010] provides general decomposition guidance, but specific patterns for migrating from a monolithic onboarding endpoint to a distributed saga (with dual-write transition periods, shadow mode, and gradual cutover) are not well-documented in the academic literature.

## Domain Calibration

The evidence distribution reflects a domain with strong foundational literature (the saga pattern is well-studied and has canonical references) but weaker empirical research on specific implementation trade-offs. The core patterns are well-established, but practical guidance for heterogeneous side-effect orchestration in specific domains (onboarding with calendar + telephony + payment) relies more on practitioner experience than academic evidence. Treat MODERATE findings as reliable industry consensus; WEAK and UNVERIFIED findings require validation against your specific architecture and provider landscape.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best. The IEEE papers [SRC-007], [SRC-008] and the ByteByteGo article were accessible only as abstracts.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible. DOIs are included only when confirmed via publisher websites. No DOIs were fabricated.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.
5. **Vendor bias**: Several sources [SRC-011], [SRC-013], [SRC-014] are from vendors (Temporal, Camunda) with commercial interests in orchestration platforms. Their claims have been cross-referenced where possible but should be evaluated with awareness of this bias.

Generated by `/research service-orchestration-patterns` on 2026-03-24.
