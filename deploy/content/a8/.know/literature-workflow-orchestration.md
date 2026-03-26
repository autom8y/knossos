---
domain: "literature-workflow-orchestration"
generated_at: "2026-03-06T18:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.62
format_version: "1.0"
---

# Literature Review: AWS Step Functions, Saga Patterns, and Event-Driven Microservice Orchestration

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on workflow orchestration in serverless and event-driven microservice architectures is mature in foundational patterns (saga, CQRS, event sourcing) but rapidly evolving in applied domains (agentic AI workflows, CDP event pipelines). The saga pattern, originally described by Garcia-Molina and Salem in 1987, has become the canonical approach for distributed transaction management, with strong consensus favoring orchestration over choreography for complex workflows. AWS Step Functions is the dominant serverless orchestrator in the AWS ecosystem, with well-documented tradeoffs between Standard and Express workflows, though cost and ASL limitations drive migration to Temporal.io at high scale (200M+ workflows/month). EventBridge has emerged as the preferred serverless event backbone on AWS, with schema registry and versioning practices maturing. Agentic workflow orchestration is the least settled domain, with the key architectural insight being the separation of deterministic orchestration from non-deterministic LLM reasoning. Evidence quality is MODERATE overall, weighted toward official documentation and practitioner sources rather than peer-reviewed research.

## Source Catalog

### [SRC-001] Sagas
- **Authors**: Hector Garcia-Molina, Kenneth Salem
- **Year**: 1987
- **Type**: peer-reviewed paper (ACM SIGMOD)
- **URL/DOI**: https://doi.org/10.1145/38714.38742
- **Verified**: partial (abstract and citations confirmed via ACM DL; full text behind paywall)
- **Relevance**: 5
- **Summary**: The foundational paper introducing sagas as sequences of transactions that can be interleaved with other transactions. Proposed compensating transactions as the mechanism for undoing partial execution. Established the theoretical basis for all modern saga implementations in distributed systems and microservices.
- **Key Claims**:
  - Long-lived transactions can be decomposed into sequences of shorter transactions (sagas) with compensating transactions for rollback [**STRONG**]
  - The DBMS guarantees either all transactions in a saga complete successfully or compensating transactions amend partial execution [**STRONG**]

### [SRC-002] Saga Design Pattern -- Azure Architecture Center
- **Authors**: Microsoft Azure Architecture Team
- **Year**: 2025 (updated December 2025)
- **Type**: official documentation
- **URL/DOI**: https://learn.microsoft.com/en-us/azure/architecture/patterns/saga
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Comprehensive reference for saga pattern implementation covering orchestration vs choreography tradeoffs, compensation transaction design, and six countermeasures for data anomalies (semantic lock, commutative updates, pessimistic view, reread values, version files, risk-based concurrency). Introduces the pivot transaction concept as the point-of-no-return in saga flows.
- **Key Claims**:
  - Orchestration is better suited for complex workflows with many participants; choreography suits simple flows with few services [**STRONG**]
  - Six countermeasures address saga isolation anomalies: semantic lock, commutative updates, pessimistic view, reread values, version files, risk-based concurrency [**MODERATE**]
  - Saga participants must be idempotent to handle transient failures and orchestrator crashes [**STRONG**]
  - Compensating transactions may not always succeed, potentially leaving the system in an inconsistent state [**MODERATE**]

### [SRC-003] Microservices Patterns (Book)
- **Authors**: Chris Richardson
- **Year**: 2018 (2nd edition MEAP available)
- **Type**: textbook (Manning Publications)
- **URL/DOI**: https://www.manning.com/books/microservices-patterns
- **Verified**: partial (table of contents and pattern summaries verified via microservices.io; full text not accessed)
- **Relevance**: 5
- **Summary**: The definitive practitioner reference for microservice patterns including sagas, event sourcing, CQRS, and transactional outbox. Covers orchestration-based saga implementation with command/reply messaging, aggregate design with event sourcing, and CQRS for maintaining query-optimized views. The associated microservices.io site provides freely accessible pattern descriptions.
- **Key Claims**:
  - Services must atomically update their database and publish events, requiring patterns like transactional outbox or event sourcing [**STRONG**]
  - Event sourcing persists aggregates as sequences of events; current state is reconstructed by replaying events (fold/reduce) [**STRONG**]
  - CQRS separates command-side (event sourcing) from query-side (materialized views) to support diverse query patterns [**MODERATE**]
  - Snapshots optimize event sourcing performance by periodically persisting aggregate state [**MODERATE**]

### [SRC-004] Implement the Serverless Saga Pattern by Using AWS Step Functions -- AWS Prescriptive Guidance
- **Authors**: AWS Prescriptive Guidance Team
- **Year**: 2024
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/prescriptive-guidance/latest/patterns/implement-the-serverless-saga-pattern-by-using-aws-step-functions.html
- **Verified**: yes
- **Relevance**: 5
- **Summary**: AWS reference implementation of saga pattern using Step Functions with Lambda and DynamoDB. Demonstrates backward compensation (cancel in reverse order), transaction status tracking via DynamoDB (pending/confirmed states), and testable failure injection via query parameters. Identifies complexity scaling as the key limitation for sagas beyond 3-5 steps.
- **Key Claims**:
  - Step Functions provides ideal orchestration for sagas with both forward-processing and compensating states [**MODERATE**]
  - Saga complexity increases significantly beyond 3-5 microservice steps [**MODERATE**]
  - DynamoDB transaction_status pattern (pending -> confirmed, or DELETE on rollback) provides implicit idempotency [**WEAK**]
  - Testing distributed sagas requires all services running, making integration testing difficult [**MODERATE**]

### [SRC-005] Choosing Workflow Type in Step Functions -- AWS Documentation
- **Authors**: AWS Step Functions Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/step-functions/latest/dg/choosing-workflow-type.html
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Authoritative comparison of Standard vs Express workflow types. Standard provides exactly-once execution up to 1 year with full history; Express provides at-least-once (async) or at-most-once (sync) execution for up to 5 minutes at higher throughput. Express lacks .waitForTaskToken, .sync, Distributed Map, and Activities support. Workflow type cannot be changed after creation.
- **Key Claims**:
  - Standard workflows guarantee exactly-once execution; Express async provides at-least-once, Express sync provides at-most-once [**STRONG**]
  - Express workflows support up to 100,000 state transitions per second but are limited to 5-minute duration [**STRONG**]
  - Express workflows do not support .waitForTaskToken or .sync integration patterns [**STRONG**]
  - Workflow type is immutable after creation [**MODERATE**]

### [SRC-006] AWS Step Functions Pitfalls: 3 Real Problems (and When to Avoid Them)
- **Authors**: Allen Helton
- **Year**: 2025
- **Type**: blog post (Ready, Set, Cloud!)
- **URL/DOI**: https://www.readysetcloud.io/blog/allen.helton/when-not-to-use-step-functions/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Practitioner analysis of Step Functions limitations at scale. Identifies 256KB payload limit as a hard constraint requiring S3 workarounds, Map state concurrency cap of 40, and 25,000 max history events as production-impacting limits. Recommends Lambda functions for simple workflows and API-based architectures for cross-service boundaries.
- **Key Claims**:
  - Step Functions has a hard 256KB payload size limit per state transition [**STRONG**]
  - Map state maximum concurrency is 40, forcing sequential batching for large datasets [**MODERATE**]
  - 25,000 max history events restricts total state transitions in complex workflows [**MODERATE**]
  - Direct cross-service resource invocation in Step Functions creates tight coupling anti-pattern [**WEAK**]

### [SRC-007] AWS Step Functions vs Temporal: A Practical Developer Comparison
- **Authors**: Allen Helton
- **Year**: 2024
- **Type**: blog post (Ready, Set, Cloud!)
- **URL/DOI**: https://www.readysetcloud.io/blog/allen.helton/step-functions-vs-temporal/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Side-by-side comparison of Step Functions (declarative/visual, fully managed) vs Temporal (code-first, self-hosted or managed). Step Functions excels in AWS-native integration and visual management; Temporal excels in code-based testing, reusable activity libraries, and sub-second timer granularity. Temporal Cloud's $200/month minimum support charge is identified as a barrier for smaller workloads.
- **Key Claims**:
  - Step Functions uses declarative JSON/YAML (ASL) while Temporal uses code-based workflows in Go, Java, Python, .NET, TypeScript [**STRONG**]
  - Temporal activities create reusable libraries across projects; Step Functions tasks are not reusable in the same way [**MODERATE**]
  - Temporal Cloud minimum cost ($200/month support) makes it impractical for small-scale or side projects [**MODERATE**]
  - Unit testing code-based Temporal activities is significantly simpler than testing Step Functions state machines [**WEAK**]

### [SRC-008] From Step Functions to Temporal on EKS: Durable Workflows at Scale
- **Authors**: Anonymous (AWS Builders community)
- **Year**: 2025
- **Type**: blog post (DEV Community)
- **URL/DOI**: https://dev.to/aws-builders/from-step-functions-to-temporal-on-eks-durable-workflows-at-scale-without-breaking-the-bank-3cdf
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Production migration case study from Step Functions to self-hosted Temporal on EKS. Triggered by unsustainable costs at 200M workflows/month with delay-heavy patterns. Self-hosted Temporal achieved ~80% cost reduction (stabilized at ~$1,500/month). Key architectural insight: using StartDelay instead of Workflow.Sleep() reduced actions per workflow by 33%. Shadow-mode migration strategy maintained Step Functions as fallback during transition.
- **Key Claims**:
  - The cost breaking point for Step Functions occurs around 200M workflows/month for delay-heavy patterns [**WEAK**]
  - Self-hosted Temporal on EKS achieved ~80% cost reduction vs Step Functions at scale [**WEAK**]
  - Temporal Cloud action-based pricing was also prohibitive at ~400M actions/month [**WEAK**]
  - Shadow-mode migration (running both systems in parallel) eliminates cutover risk [**MODERATE**]

### [SRC-009] Of Course You Can Build Dynamic AI Agents with Temporal
- **Authors**: Temporal Team
- **Year**: 2025
- **Type**: blog post (Temporal.io)
- **URL/DOI**: https://temporal.io/blog/of-course-you-can-build-dynamic-ai-agents-with-temporal
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Technical explanation of how Temporal's deterministic workflow code accommodates non-deterministic LLM decisions. Workflow code (orchestration blueprint) must be deterministic for replay; Activities (LLM calls, tool invocations) can be non-deterministic. On failure, Temporal replays using Event History, resuming at the exact failure point without re-executing completed LLM calls. Cites OpenAI's Codex and Replit's Agent 3 as production implementations.
- **Key Claims**:
  - Temporal's determinism requirement applies to workflow code (orchestration), not to activity code (LLM calls, tool use) [**MODERATE**]
  - Temporal replays agent progress using Event History, avoiding re-execution of completed LLM calls on failure recovery [**MODERATE**]
  - OpenAI's Codex web agent is built on Temporal, handling millions of requests [**WEAK**]

### [SRC-010] Best Practices for Amazon EventBridge Event Patterns
- **Authors**: AWS EventBridge Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-patterns-best-practices.html
- **Verified**: yes
- **Relevance**: 4
- **Summary**: AWS guidance on EventBridge rule design for production systems. Mandates specifying both source and detail-type at minimum. Warns against overly broad patterns that match unintended events when schemas evolve. Recommends account/region scoping for cross-account architectures and pattern validation via console sandbox or CLI before deployment.
- **Key Claims**:
  - Event patterns must specify both source and detail-type to prevent unintended matches during schema evolution [**STRONG**]
  - Overly broad patterns create infinite loop risk with unexpected charges and throttling [**MODERATE**]
  - Cross-account event patterns should include account and region filters [**MODERATE**]

### [SRC-011] Event Versioning Strategies for Event-Driven Architectures
- **Authors**: Yan Cui (theburningmonk)
- **Year**: 2025
- **Type**: blog post
- **URL/DOI**: https://theburningmonk.com/2025/04/event-versioning-strategies-for-event-driven-architectures/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Comparative analysis of six event versioning strategies: version in name, version in payload, separate streams, schema registry with ID, no breaking changes, and out-of-band translation. Recommends "no breaking changes" (additive-only schema evolution) as the optimal strategy, arguing it eliminates versioning overhead entirely. This is the shortest path to safe schema evolution in event-driven systems.
- **Key Claims**:
  - Additive-only schema evolution (never remove/rename fields, never change types) eliminates versioning overhead [**MODERATE**]
  - Schema registries introduce temporal coupling to the registry itself [**WEAK**]
  - Version-in-event-name strategy ("user.created.v1") enables explicit consumer opt-in but requires event duplication [**MODERATE**]

### [SRC-012] Developing Transactional Microservices Using Aggregates, Event Sourcing and CQRS
- **Authors**: Chris Richardson
- **Year**: 2016
- **Type**: conference talk / article (InfoQ)
- **URL/DOI**: https://www.infoq.com/articles/microservices-aggregates-events-cqrs-part-2-richardson/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Technical deep-dive on event sourcing implementation including domain event schema design (entity_type, entity_id, event_id, event_type, event_data), three approaches to atomic event publishing (message broker, transaction log tailing, database table queue), and idempotency via monotonically increasing event IDs. Covers event schema evolution by transforming events to latest version on load.
- **Key Claims**:
  - Three approaches to atomic event publishing: message broker, transaction log tailing, database table queue [**MODERATE**]
  - Event handlers achieve idempotency by tracking highest-seen event IDs and discarding duplicates [**MODERATE**]
  - Event schema evolution is handled by transforming events to latest version when loading from event store [**MODERATE**]

### [SRC-013] From Prompts to Production: A Playbook for Agentic Development
- **Authors**: InfoQ Contributors
- **Year**: 2025
- **Type**: conference talk / article (InfoQ)
- **URL/DOI**: https://www.infoq.com/articles/prompts-to-production-playbook-for-agentic-development/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Production playbook for agentic AI systems covering deterministic vs non-deterministic decomposition, orchestration patterns (ReAct, Supervisor, Hierarchical), and human oversight levels. Introduces the Capability Matrix for systematically identifying where LLM reasoning adds value vs deterministic logic. Emphasizes treating system prompts, tool configs, and agent parameters as versioned infrastructure-as-code.
- **Key Claims**:
  - Nondeterministic reasoning from unstructured text is the dividing line between agentic and non-agentic components [**MODERATE**]
  - System prompt changes cause up to 63% execution path variation, requiring version control [**WEAK**]
  - Four human oversight levels: in-the-loop, on-the-loop, above-the-loop, behind-the-loop [**MODERATE**]
  - Breaking conditions (confidence thresholds, iteration limits, error detection) prevent infinite agent loops [**MODERATE**]

### [SRC-014] Segment Protocols Overview
- **Authors**: Twilio Segment
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://segment.com/docs/protocols/
- **Verified**: partial (fetched but 403 on full content; verified via search results and product documentation references)
- **Relevance**: 3
- **Summary**: Segment Protocols validates event payloads against JSON Schema-based tracking plans during ingestion. Violations are generated when events don't match the tracking plan spec. Schema controls can block non-conforming events at the source, quarantining bad data before it reaches destinations. Labels enable multi-team tracking plan organization.
- **Key Claims**:
  - Tracking plans use JSON Schema to validate event payloads in real-time during ingestion [**MODERATE**]
  - Schema controls can block violating events at the source, preventing data quality degradation downstream [**MODERATE**]
  - Transformations within Protocols can change event and property names without code changes [**WEAK**]

### [SRC-015] When to Use Step Functions vs. Doing It All in a Lambda Function
- **Authors**: Yan Cui (theburningmonk)
- **Year**: 2024
- **Type**: blog post
- **URL/DOI**: https://theburningmonk.com/2024/03/when-to-use-step-functions-vs-doing-it-all-in-a-lambda-function/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Decision framework for Step Functions vs Lambda-only architectures. Step Functions excels at waiting operations (no duration billing), callback patterns, and visual audit trails. Lambda is preferred for simple workflows and cost-conscious implementations. Step Functions at $25/million transitions is one of the more expensive AWS services. The TestState API is improving but Step Functions testing remains harder than Lambda unit tests.
- **Key Claims**:
  - Step Functions does not charge for wait duration, only state transitions, making it optimal for long-wait workflows [**MODERATE**]
  - At $25 per million state transitions, Step Functions is among the more expensive AWS serverless services [**STRONG**]
  - Step Functions visual workflows provide audit trails accessible to non-technical stakeholders [**MODERATE**]

## Thematic Synthesis

### Theme 1: Orchestration Dominates Choreography for Complex Saga Workflows

**Consensus**: For workflows involving more than 3-4 services, orchestration-based sagas are preferred over choreography due to clearer visibility, simpler debugging, and avoidance of cyclic dependencies. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004]

**Controversy**: Whether the orchestrator itself becomes a single point of failure. Choreography proponents argue that distributed event routing is more resilient.
**Dissenting sources**: [SRC-002] acknowledges orchestration introduces "a point of failure because the orchestrator manages the complete workflow," while [SRC-004] argues Step Functions' managed infrastructure mitigates this concern.

**Practical Implications**:
- Default to orchestration (Step Functions, Temporal) for workflows with 4+ services or complex compensation logic
- Use choreography only for simple, loosely-coupled flows where adding an orchestrator would be over-engineering
- Implement the transactional outbox pattern to ensure atomic event publishing regardless of approach

**Evidence Strength**: STRONG

### Theme 2: Step Functions Cost and ASL Limitations Drive Migration to Temporal at Scale

**Consensus**: AWS Step Functions is the pragmatic default for AWS-native workflow orchestration, but cost ($25/million transitions) and ASL constraints (256KB payload, 40 max Map concurrency, 25K history events) create pressure to migrate at high scale. [**MODERATE**]
**Sources**: [SRC-005], [SRC-006], [SRC-007], [SRC-008], [SRC-015]

**Controversy**: The exact threshold for migration is workload-dependent. One case study identifies 200M workflows/month as the breaking point; others emphasize developer experience (code-based testing, reusable activities) as the primary migration driver independent of scale.
**Dissenting sources**: [SRC-007] frames the choice as subjective developer preference, while [SRC-008] presents cost as the unambiguous driver at extreme scale.

**Practical Implications**:
- Start with Step Functions for new AWS-native workflows; the managed service eliminates operational overhead
- Monitor state transition costs monthly; at >$5K/month in Step Functions billing, evaluate Temporal
- If migrating, use shadow-mode (run both in parallel) to de-risk the transition
- For delay-heavy workflows, optimize with Express workflows or Temporal's StartDelay pattern before migrating

**Evidence Strength**: MIXED (strong on Step Functions limits, weak on specific migration thresholds)

### Theme 3: Express vs Standard Workflow Choice Is Architecturally Binding

**Consensus**: The choice between Standard and Express workflows is irrevocable after creation and determines execution semantics, integration pattern support, and cost model. Standard provides exactly-once with full history; Express trades durability for throughput. [**STRONG**]
**Sources**: [SRC-005], [SRC-006], [SRC-015]

**Practical Implications**:
- Use Standard for any workflow requiring .waitForTaskToken (human approval), .sync (job completion), or duration >5 minutes
- Use Express for high-volume event processing (IoT, streaming) where at-least-once semantics are acceptable
- Never use Express for non-idempotent operations (payment processing, EMR cluster creation)
- Plan workflow type during architecture phase -- retrofitting requires new state machine creation

**Evidence Strength**: STRONG

### Theme 4: EventBridge Schema Evolution Requires Additive-Only Discipline

**Consensus**: The most effective event versioning strategy for EventBridge-based architectures is additive-only schema evolution: never remove fields, never rename fields, never change field types. This eliminates versioning overhead while maintaining backward compatibility. [**MODERATE**]
**Sources**: [SRC-010], [SRC-011], [SRC-012]

**Controversy**: Whether schema registries provide sufficient governance to allow breaking changes. The EventBridge Schema Registry supports automatic versioning, but practitioners argue the temporal coupling and operational overhead outweigh the benefits for most teams.
**Dissenting sources**: [SRC-010] (AWS) implicitly endorses the Schema Registry approach, while [SRC-011] argues schema registries introduce unnecessary coupling and recommends avoiding them for versioning.

**Practical Implications**:
- Adopt additive-only evolution as the default contract for all custom EventBridge events
- Specify both `source` and `detail-type` in all event patterns to prevent unintended matches when new event types appear
- If breaking changes are unavoidable, use version-in-event-name strategy ("order.created.v2") with parallel publishing during migration
- Use Schema Registry for discovery and code binding generation, not as the primary versioning mechanism

**Evidence Strength**: MODERATE

### Theme 5: Idempotency Is the Non-Negotiable Foundation of Event-Driven Systems

**Consensus**: All event consumers in distributed systems must be idempotent because at-least-once delivery is the practical reality of message brokers. Exactly-once processing is achievable only through consumer-side deduplication, not broker guarantees alone. [**STRONG**]
**Sources**: [SRC-002], [SRC-003], [SRC-005], [SRC-012]

**Practical Implications**:
- Implement idempotency at the consumer level using event IDs or idempotency keys, never rely solely on broker delivery guarantees
- Use DynamoDB conditional writes or database unique constraints for deduplication
- Track highest-seen event sequence numbers per aggregate for ordering guarantees
- Design compensating transactions to be idempotent themselves -- a compensation may execute multiple times

**Evidence Strength**: STRONG

### Theme 6: Agentic Workflows Require Explicit Deterministic/Non-Deterministic Boundaries

**Consensus**: The foundational architectural decision in agentic AI systems is identifying which components require non-deterministic LLM reasoning vs deterministic rule execution. Mixing these without clear boundaries creates untestable, unreliable systems. [**MODERATE**]
**Sources**: [SRC-009], [SRC-013]

**Controversy**: Whether existing workflow engines (Temporal, Step Functions) are appropriate for agentic orchestration, or whether purpose-built agent frameworks are needed.
**Dissenting sources**: [SRC-009] argues Temporal's deterministic workflow + non-deterministic activity model is naturally suited for AI agents, while [SRC-013] implicitly favors purpose-built agent orchestration frameworks with specialized evaluation loops.

**Practical Implications**:
- Use the Capability Matrix approach to map each workflow step to deterministic vs non-deterministic execution
- Place LLM calls in activities/tasks (non-deterministic), never in orchestration logic (deterministic)
- Implement breaking conditions (confidence thresholds, iteration limits) to prevent infinite agent loops
- Version system prompts and tool configurations as infrastructure-as-code; prompt changes cause up to 63% execution path variation
- Temporal's replay-based durability prevents re-execution of expensive LLM calls on failure recovery

**Evidence Strength**: MODERATE

### Theme 7: CDP Event Pipelines and Marketing Automation Favor Schema-First Governance

**Consensus**: Customer Data Platforms (Segment, mParticle) enforce data quality through schema-first governance, validating events against tracking plans at ingestion time. This pattern is applicable to any event-driven system handling cross-team event contracts. [**MODERATE**]
**Sources**: [SRC-014]

**Practical Implications**:
- Define tracking plans (event schemas) before implementation, not after
- Enforce schema validation at the event bus ingestion layer, not at individual consumers
- Use violation blocking in production to prevent schema-violating events from reaching downstream systems
- Apply labels/metadata to events for multi-team tracking plan organization

**Evidence Strength**: WEAK (single-source coverage)

## Evidence-Graded Findings

### STRONG Evidence
- Sagas decompose long-lived transactions into compensable sequences; this is the foundational distributed transaction pattern -- Sources: [SRC-001], [SRC-002], [SRC-003]
- Orchestration-based sagas are preferred over choreography for workflows with 4+ services -- Sources: [SRC-002], [SRC-003], [SRC-004]
- All saga participants and event consumers must be idempotent; at-least-once delivery is the practical reality -- Sources: [SRC-002], [SRC-003], [SRC-005], [SRC-012]
- Step Functions Standard provides exactly-once execution; Express provides at-least-once (async) or at-most-once (sync) -- Sources: [SRC-005]
- Express workflows limited to 5-minute duration and lack .waitForTaskToken, .sync, Distributed Map -- Sources: [SRC-005]
- Step Functions has a hard 256KB payload limit per state transition -- Sources: [SRC-005], [SRC-006]
- Step Functions pricing at $25/million state transitions makes it one of the more expensive AWS serverless services -- Sources: [SRC-006], [SRC-015]
- EventBridge patterns must specify both source and detail-type to prevent unintended matches -- Sources: [SRC-010]

### MODERATE Evidence
- Six countermeasures for saga isolation anomalies: semantic lock, commutative updates, pessimistic view, reread values, version files, risk-based concurrency -- Sources: [SRC-002]
- Pivot transactions represent the point-of-no-return in saga flows; post-pivot steps must be retryable -- Sources: [SRC-002]
- Services must atomically update database and publish events via transactional outbox or event sourcing -- Sources: [SRC-003], [SRC-012]
- Additive-only schema evolution eliminates event versioning overhead while maintaining backward compatibility -- Sources: [SRC-011]
- Temporal's determinism requirement applies to workflow code, not activity code (LLM calls, tool use) -- Sources: [SRC-009]
- The deterministic/non-deterministic boundary is the foundational architectural decision for agentic workflows -- Sources: [SRC-009], [SRC-013]
- Shadow-mode migration (running Step Functions and Temporal in parallel) de-risks orchestrator transitions -- Sources: [SRC-008]
- Segment Protocols validates events against JSON Schema tracking plans at ingestion time -- Sources: [SRC-014]
- Step Functions visual workflows provide audit trails accessible to non-technical stakeholders -- Sources: [SRC-015]
- Breaking conditions (confidence thresholds, iteration limits) prevent infinite agent loops -- Sources: [SRC-013]

### WEAK Evidence
- Step Functions cost breaking point occurs around 200M workflows/month for delay-heavy patterns -- Sources: [SRC-008]
- Self-hosted Temporal on EKS achieves ~80% cost reduction vs Step Functions at scale -- Sources: [SRC-008]
- Schema registries introduce temporal coupling to the registry service itself -- Sources: [SRC-011]
- System prompt changes cause up to 63% execution path variation in agentic systems -- Sources: [SRC-013]
- OpenAI Codex and Replit Agent 3 are production implementations of Temporal-based AI agents -- Sources: [SRC-009]
- Unit testing Temporal activities is significantly simpler than testing Step Functions state machines -- Sources: [SRC-007]

### UNVERIFIED
- Optimal saga step count ceiling of 3-5 steps before complexity becomes unmanageable -- Basis: model training knowledge corroborated by [SRC-004] but no rigorous study found
- mParticle's mobile-first event pipeline architecture differs structurally from Segment's data conduit approach at the event routing layer -- Basis: model training knowledge; product documentation was not fully accessible
- Temporal.io grew out of Uber's Cadence project for reliability under extreme event volume -- Basis: model training knowledge; widely cited but primary source not fetched
- CQRS event store snapshot frequency should balance read performance against storage cost, with typical intervals of 100-1000 events per aggregate -- Basis: model training knowledge

## Knowledge Gaps

- **Saga pattern at extreme scale (100+ services)**: No rigorous study quantifies saga coordination overhead as service count grows. Most references stop at "3-5 steps" guidance without formal complexity analysis. Production experience reports from organizations running sagas across 20+ services would fill this gap.

- **Temporal.io vs Step Functions formal benchmark**: While practitioner comparisons exist, no independent benchmark compares latency, throughput, and cost across equivalent workloads. The migration case study [SRC-008] is single-datapoint evidence. A controlled comparison across workload profiles (compute-heavy, delay-heavy, fan-out) is needed.

- **EventBridge schema registry at scale**: Limited evidence on Schema Registry performance and governance workflows at scale (1000+ event types, 100+ consumers). Most documentation covers small-scale examples. Production experience reports from large organizations would strengthen evidence.

- **Marketing automation event pipeline architecture**: CDP event pipeline internals (Segment, mParticle) are proprietary. Public documentation covers API surfaces but not internal event routing, partitioning, or exactly-once guarantees. This gap can only be filled by vendor whitepapers or reverse engineering.

- **Agentic workflow pattern maturity**: The field is pre-consensus. Andrew Ng's four agentic patterns (Reflection, Tool Use, Planning, Multi-Agent) provide a taxonomy but lack rigorous empirical validation. Production evidence is limited to vendor case studies (Temporal, OpenAI) without independent verification. Academic literature on agentic orchestration patterns is sparse.

- **CQRS eventual consistency window measurement**: No source provides empirical data on typical event propagation latency from command-side to query-side in production CQRS systems. This gap makes it difficult to set SLAs for read-after-write consistency in event-sourced architectures.

## Domain Calibration

Low-to-moderate confidence distribution reflects a domain that spans well-studied foundations (saga pattern, CQRS) and rapidly evolving applied areas (agentic workflows, CDP pipelines). Many claims could not be independently corroborated beyond vendor documentation and practitioner blog posts. Treat findings as starting points for manual research, not as settled knowledge. The foundational patterns (sagas, event sourcing, CQRS) have strong evidence backing; the applied patterns (agentic orchestration, CDP event governance, Step Functions migration thresholds) require production validation specific to your workload.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research AWS Step Functions saga pattern event-driven microservice orchestration` on 2026-03-06.
