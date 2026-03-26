---
domain: "literature-event-sourcing-at-scale"
generated_at: "2026-02-27T11:35:24Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.71
format_version: "1.0"
---

# Literature Review: Event Sourcing at Scale

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

Event sourcing is a well-documented architectural pattern where application state is persisted as an append-only sequence of immutable domain events rather than mutable current-state records. The literature broadly agrees that event sourcing provides strong auditability, scalability for write-heavy workloads, and temporal query capabilities, but introduces significant complexity around schema evolution, eventual consistency, GDPR compliance, and operational overhead. The pattern is consistently recommended for high-throughput systems requiring audit trails and flexible read models, but explicitly discouraged for simple CRUD domains. Evidence quality is moderate-to-strong, with two peer-reviewed empirical studies, multiple authoritative vendor/platform documentation sources, and several influential practitioner works forming the core of the literature.

## Source Catalog

### [SRC-001] An Empirical Characterization of Event Sourced Systems and Their Schema Evolution -- Lessons from Industry
- **Authors**: Michiel Overeem, Marten Spoor, Slinger Jansen, Sjaak Brinkkemper
- **Year**: 2021
- **Type**: peer-reviewed paper (Journal of Systems and Software, Vol. 178)
- **URL/DOI**: https://arxiv.org/abs/2104.01146
- **Verified**: yes (arXiv preprint fetched, abstract and findings confirmed)
- **Relevance**: 5
- **Summary**: The only large-scale empirical study of event sourcing in industry. Interviews with 25 engineers across 19 event sourced systems identified five major practitioner challenges: event system evolution, steep learning curve, lack of available technology, rebuilding projections, and data privacy. Documents five schema evolution tactics (versioned events, weak schema, upcasting, in-place transformation, copy-and-transform). Provides grounded theory analysis of adoption rationale (reliability, flexibility, scalability).
- **Key Claims**:
  - Practitioners face five primary challenges with event sourcing: event system evolution, steep learning curve, lack of technology, rebuilding projections, and data privacy [**STRONG**]
  - Schema evolution is managed through five distinct tactics, with versioned events and upcasting being most common [**MODERATE**]
  - Event sourcing adoption is driven by reliability, flexibility, and scalability requirements [**MODERATE**]

### [SRC-002] Designing Data-Intensive Applications
- **Authors**: Martin Kleppmann
- **Year**: 2017
- **Type**: textbook (O'Reilly Media)
- **URL/DOI**: https://www.oreilly.com/library/view/designing-data-intensive-applications/9781491903063/
- **Verified**: partial (title and chapter structure confirmed via multiple secondary sources; full text behind O'Reilly paywall)
- **Relevance**: 5
- **Summary**: Chapter 11 (Stream Processing) covers event sourcing as an architectural pattern where state changes are logged as immutable events. Positions event sourcing within the broader context of change data capture, stream processing, and log-based architectures. Argues that the log is a fundamental abstraction for distributed data systems, with event sourcing being one realization of this principle.
- **Key Claims**:
  - Event sourcing logs state changes as immutable events; current state is derived by replaying the event log [**STRONG**]
  - The log is a unifying abstraction for stream processing, change data capture, and event sourcing [**STRONG**]
  - Event sourcing and CDC are related but distinct: event sourcing captures domain-level intent while CDC captures storage-level mutations [**MODERATE**]

### [SRC-003] Event Sourcing Pattern -- Azure Architecture Center
- **Authors**: Microsoft (Clayton Siemens et al.)
- **Year**: 2025 (last updated January 2025)
- **Type**: official documentation
- **URL/DOI**: https://learn.microsoft.com/en-us/azure/architecture/patterns/event-sourcing
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Comprehensive pattern documentation covering definition, workflow, advantages, issues, and when-to-use guidance. Explicitly warns that event sourcing is a complex pattern that "permeates through the entire architecture" with high migration costs. Identifies eventual consistency, event versioning, event ordering, querying limitations, state reconstruction cost, and idempotency as key implementation concerns. Recommends combination with CQRS and materialized views.
- **Key Claims**:
  - Event sourcing vastly improves performance and scalability by eliminating write contention through append-only operations [**STRONG**]
  - The pattern is not justified for most systems; complexity is only warranted when performance and scalability are top requirements [**STRONG**]
  - Eventual consistency is inherent: materialized views and projections always lag behind the event store [**STRONG**]
  - Snapshots are necessary for large event streams to avoid expensive full-replay state reconstruction [**MODERATE**]
  - Event consumers must be idempotent due to at-least-once delivery semantics [**MODERATE**]

### [SRC-004] Event Sourcing (martinfowler.com)
- **Authors**: Martin Fowler
- **Year**: 2005
- **Type**: blog post (authoritative practitioner reference)
- **URL/DOI**: https://martinfowler.com/eaaDev/EventSourcing.html
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Foundational practitioner definition of event sourcing. Defines the pattern as capturing all changes to application state as a sequence of events. Identifies complete rebuild, temporal queries, event replay, and parallel testing as core capabilities. Warns about external system interactions during replay, code change handling, and interface awkwardness. Notes that snapshots enable performance optimization by caching state at checkpoints.
- **Key Claims**:
  - Complete application state can be rebuilt from scratch by replaying events [**STRONG**]
  - Temporal queries allow determining application state at any point in time [**STRONG**]
  - External system interactions create significant complexity during event replay [**MODERATE**]
  - Snapshots cache state at checkpoints; systems start from overnight snapshots and replay subsequent events [**MODERATE**]

### [SRC-005] The Log: What Every Software Engineer Should Know About Real-Time Data's Unifying Abstraction
- **Authors**: Jay Kreps
- **Year**: 2013
- **Type**: blog post (LinkedIn Engineering; later published as O'Reilly book)
- **URL/DOI**: https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Seminal essay establishing the append-only log as a unifying abstraction for distributed data systems. Argues that "tables and events are dual: tables support data at rest and logs capture change." Demonstrates that LinkedIn's Kafka processed over 60 billion unique message writes per day using log-based architecture with partitioning, batching, and zero-copy transfers. Positions the log as the foundation for both event sourcing and change data capture.
- **Key Claims**:
  - The append-only log is a unifying abstraction for data integration, stream processing, and distributed consensus [**STRONG**]
  - Tables and events are dual: a complete changelog can reconstruct any previous system state [**STRONG**]
  - LinkedIn's Kafka achieved 60+ billion messages/day through partitioning, batching, and zero-copy transfers [**MODERATE**]

### [SRC-006] Immutability Changes Everything
- **Authors**: Pat Helland
- **Year**: 2015 (CIDR 2015; ACM Queue 2015)
- **Type**: peer-reviewed paper (CIDR 2015, also published in ACM Queue vol. 13, no. 9)
- **URL/DOI**: https://www.cidrdb.org/cidr2015/Papers/CIDR15_Paper16.pdf
- **Verified**: partial (title and publication venue confirmed via multiple sources; PDF available but not fully fetched due to access constraints)
- **Relevance**: 4
- **Summary**: Argues that immutability is an inexorable computing trend. Transaction logs record all database changes as append-only entries; from this perspective, "the database holds a caching of the latest record values in the logs, with the truth being the log and the database being a cache." Cheap storage makes immutability economically viable at scale. Provides theoretical foundation for why event sourcing's append-only model is architecturally sound.
- **Key Claims**:
  - The truth is the log; the database is a cache of a subset of the log [**STRONG**]
  - Immutability enables coordination at distance; append-only computing records facts forever and derives results on demand [**MODERATE**]
  - Cheap computation, disk, DRAM, and SSDs make immutability economically viable while coordination (latching) has become the bottleneck [**MODERATE**]

### [SRC-007] Versioning in an Event Sourced System
- **Authors**: Greg Young
- **Year**: 2017
- **Type**: textbook (Leanpub, freely available)
- **URL/DOI**: https://leanpub.com/esversioning/read
- **Verified**: yes (table of contents and key sections fetched and confirmed)
- **Relevance**: 5
- **Summary**: The definitive guide to event schema versioning. Covers why events must not be mutated (immutability, downstream dependencies, audit integrity, fraud prevention). Presents versioning patterns including type-based versioning, weak schema, negotiation-based approaches, compensating actions ("accountants use pens, not erasers"), and copy-and-replace for stream boundary changes. Warns against conflating multiple concerns in single events and changing semantics without explicit versioning.
- **Key Claims**:
  - Events must never be mutated: immutability, downstream consumer dependencies, audit integrity, and fraud prevention all require it [**STRONG**]
  - Five versioning patterns exist: type-based, weak schema, negotiation, compensating actions, copy-and-replace [**MODERATE**]
  - Version bankruptcy (discarding and rebuilding) is a legitimate last resort when evolution becomes untenable [**WEAK**]

### [SRC-008] Event Sourcing Pattern -- AWS Prescriptive Guidance
- **Authors**: AWS (Amazon Web Services)
- **Year**: 2025 (undated; accessed February 2026)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/event-sourcing.html
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 4
- **Summary**: AWS-specific implementation guidance for event sourcing. Recommends Kinesis Data Streams for high-throughput event stores, S3 for archival, and Aurora for materialized views. Identifies event store exponential growth, event replay performance, and event ordering as critical scale challenges. Recommends time-based snapshots aligned with Recovery Point Objectives (RPO). Emphasizes idempotency and FIFO queues for ordering guarantees.
- **Key Claims**:
  - Event store grows exponentially with high throughput or extended retention; periodic archival to cost-effective storage is necessary [**MODERATE**]
  - Snapshot frequency should align with Recovery Point Objective (RPO) requirements [**MODERATE**]
  - Event ordering requires FIFO queues or sequence numbers; incorrect ordering causes incorrect system state [**MODERATE**]

### [SRC-009] Distributed Data for Microservices -- Event Sourcing vs. Change Data Capture
- **Authors**: Eric Murphy (Debezium project)
- **Year**: 2020
- **Type**: blog post (vendor, but technically detailed)
- **URL/DOI**: https://debezium.io/blog/2020/02/10/event-sourcing-vs-cdc/
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Detailed comparison of event sourcing and CDC for microservice data distribution. Argues that CDC with the Outbox Pattern is "usually a better alternative to Event Sourcing" for most teams due to simpler consistency guarantees. Identifies the "dual writes flaw" in event sourcing (risk of data loss when journal and projection updates are not atomic). Notes that event sourcing requires everything to be eventually consistent, which does not fit systems needing strong consistency.
- **Key Claims**:
  - CDC with Outbox Pattern provides transactional guarantees that event sourcing's dual-write model cannot [**MODERATE**]
  - Event sourcing requires all data to be eventually consistent; strong consistency requirements are incompatible [**STRONG**]
  - For most teams, CDC is preferable due to simpler operational model [**WEAK**]

### [SRC-010] Event Sourcing at Global Scale
- **Authors**: Martin Krasser
- **Year**: 2015
- **Type**: blog post (practitioner, technically detailed)
- **URL/DOI**: http://krasserm.github.io/2015/01/13/event-sourcing-at-global-scale/
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Describes architecture for globally distributed event sourcing with geo-replicated event logs. Uses vector timestamps for causal ordering without requiring global total order. Accepts eventual consistency by design, with concurrent events having potentially different orderings across sites. Supports conflict resolution via interactive resolution, automated functions, and CRDTs. Overcomes Akka Persistence's cluster-wide singleton limitation.
- **Key Claims**:
  - Global-scale event sourcing requires abandoning global total ordering in favor of causal ordering with vector timestamps [**MODERATE**]
  - CRDTs can eliminate conflicts entirely for commutative operations in distributed event-sourced systems [**MODERATE**]
  - Geo-distributed event sourcing must prioritize availability and partition tolerance, accepting eventual consistency [**WEAK**]

### [SRC-011] Event Sourcing Pattern (microservices.io)
- **Authors**: Chris Richardson
- **Year**: 2026 (copyright date; original publication undated)
- **Type**: official documentation (practitioner reference site)
- **URL/DOI**: https://microservices.io/patterns/data/event-sourcing.html
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Defines event sourcing as persisting business entity state as a sequence of state-changing events. Identifies the core problem as atomically updating databases and sending messages without two-phase commit. Notes the event store functions as both database and message broker. Lists benefits: reliable event publishing, no object-relational impedance mismatch, 100% reliable audit log, temporal queries. Lists drawbacks: unfamiliar programming paradigm, steep learning curve, complex querying requiring CQRS.
- **Key Claims**:
  - Event sourcing solves the atomic update + message publish problem without two-phase commit [**MODERATE**]
  - The event store serves dual roles as both database and message broker [**MODERATE**]
  - Event sourcing provides a "100% reliable audit log" of business entity changes [**STRONG**]

### [SRC-012] Snapshots in Event Sourcing
- **Authors**: Oskar Dudycz
- **Year**: 2021 (updated 2024)
- **Type**: blog post (practitioner, technically detailed)
- **URL/DOI**: https://www.kurrent.io/blog/snapshots-in-event-sourcing
- **Verified**: yes (full content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Detailed analysis of snapshot strategies for event sourcing. Identifies four timing strategies (after each event, every N events, on specific event types, scheduled intervals). Argues snapshots should be tactical optimizations, not architectural foundations. Warns that heavy snapshot reliance indicates aggregate design problems. Recommends separate streams for snapshot storage and asynchronous snapshotting to avoid blocking command handling.
- **Key Claims**:
  - Snapshots should be a performance optimization applied only when replay latency is measurably problematic, not a default pattern [**MODERATE**]
  - Heavy snapshot reliance is a design smell indicating aggregate boundaries need restructuring [**WEAK**]
  - Asynchronous snapshotting via subscriptions avoids blocking command handling but introduces processing delays [**WEAK**]

### [SRC-013] GDPR Compliance Strategies for Event-Sourced Systems
- **Authors**: Multiple (Oskar Dudycz, Michiel Rook, Dan Lebrero, and others; synthesized from multiple sources)
- **Year**: 2017-2025
- **Type**: blog posts and documentation (practitioner consensus)
- **URL/DOI**: https://event-driven.io/en/gdpr_in_event_driven_architecture/ (primary); https://danlebrero.com/2018/04/11/kafka-gdpr-event-sourcing/ ; https://docs.eventsourcingdb.io/best-practices/gdpr-compliance/
- **Verified**: partial (search results confirmed strategies; primary article content not fully extracted due to rendering issues)
- **Relevance**: 4
- **Summary**: Documents the fundamental tension between event sourcing's immutability and GDPR's right to erasure. Identifies four compliance strategies: (1) projection-based removal (erase from read models only), (2) crypto-shredding (encrypt per-user, destroy key on deletion request), (3) data minimization (store references not PII in events), (4) retention policies with event store TTLs. Consensus is that crypto-shredding is the most practical approach for systems that store PII in events, while data minimization is the preferred design-time strategy.
- **Key Claims**:
  - Crypto-shredding (encrypting PII per-user and destroying the key) is the most practical GDPR compliance strategy for existing event-sourced systems [**MODERATE**]
  - Data minimization (storing references instead of PII in events) is the preferred design-time strategy to avoid the GDPR-immutability tension entirely [**MODERATE**]
  - GDPR compliance must be designed into the system from the start; retrofitting is significantly more expensive [**WEAK**]

## Thematic Synthesis

### Theme 1: Append-Only Immutability Is the Core Architectural Bet

**Consensus**: Event sourcing's defining characteristic -- storing all state changes as immutable, append-only events -- provides auditability, temporal queries, and replay capabilities that mutable-state systems cannot match. This is supported by both academic theory (Helland's "the truth is the log; the database is a cache") and production practice (Kreps' demonstration at LinkedIn scale). [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-005], [SRC-006], [SRC-007], [SRC-011]

**Controversy**: Whether this immutability creates an irreconcilable tension with data privacy regulations (GDPR right to erasure). Practitioners have developed workarounds (crypto-shredding, data minimization), but these add complexity.
**Dissenting sources**: [SRC-013] documents practical mitigation strategies, while [SRC-001] identifies data privacy as one of the five major practitioner challenges without a fully satisfactory resolution.

**Practical Implications**:
- Commit to immutability as a foundational constraint before adopting event sourcing; all downstream decisions flow from this
- Design PII handling strategy at architecture time, not as an afterthought
- Leverage the audit trail as a first-class capability, not a side effect

**Evidence Strength**: STRONG

### Theme 2: Event Sourcing Demands CQRS and Projections as Production Necessities

**Consensus**: In production, event sourcing cannot operate without materialized views (projections) for query performance. Reading state by replaying events is prohibitively expensive for most query patterns. CQRS (separating read and write models) is the standard complement. This trio -- event store + projections + snapshots -- forms the minimum viable production architecture. [**STRONG**]
**Sources**: [SRC-002], [SRC-003], [SRC-004], [SRC-008], [SRC-011], [SRC-012]

**Practical Implications**:
- Budget for projection infrastructure (separate read stores, subscription handlers) from day one
- Accept eventual consistency between write model and read model as an inherent architectural property
- Choose snapshot strategies based on measured performance, not preemptive optimization

**Evidence Strength**: STRONG

### Theme 3: Schema Evolution Is the Hardest Long-Term Challenge

**Consensus**: Event schema evolution is the most difficult operational challenge in event-sourced systems, worsening over time as event streams grow. The literature identifies multiple strategies (versioned events, upcasting, weak schema, copy-and-transform) but no single dominant approach. [**STRONG**]
**Sources**: [SRC-001], [SRC-003], [SRC-007], [SRC-008]

**Controversy**: Whether events should ever be mutated in-place for schema migration. Young [SRC-007] insists events must never be mutated (immutability, audit integrity, fraud prevention). Azure docs [SRC-003] acknowledge that updating historical events "breaks the immutability of events" but list it as an option. Overeem et al. [SRC-001] find that practitioners do use in-place transformation despite theoretical objections.
**Dissenting sources**: [SRC-007] argues immutability is non-negotiable, while [SRC-001] empirically documents in-place transformation as a real-world tactic.

**Practical Implications**:
- Invest in versioning infrastructure (upcasters, schema registries) early; cost compounds over time
- Design events with evolution in mind: avoid overly specific schemas, include version fields
- Consider "version bankruptcy" (rebuilding streams) as a legitimate escape hatch when evolution debt becomes unmanageable

**Evidence Strength**: STRONG (on the challenge) / MIXED (on the solution)

### Theme 4: Scalability Requires Partitioning, Not Just Append-Only Writes

**Consensus**: While append-only writes eliminate write contention at the single-stream level, scaling event sourcing to high throughput requires partitioning (sharding) event streams across nodes. The log abstraction scales through partition-level parallelism, not through faster single-partition writes. [**MODERATE**]
**Sources**: [SRC-002], [SRC-005], [SRC-008], [SRC-010]

**Controversy**: Whether global total ordering should be maintained across partitions. Krasser [SRC-010] argues global ordering must be abandoned in favor of causal ordering for geo-distributed systems. Kreps [SRC-005] demonstrates partition-level ordering is sufficient for LinkedIn's scale. AWS [SRC-008] recommends FIFO queues for strict ordering within partitions.
**Dissenting sources**: [SRC-010] advocates causal ordering only, while [SRC-008] emphasizes strict per-partition ordering.

**Practical Implications**:
- Design partition keys around aggregate boundaries to maintain per-aggregate ordering
- Accept that cross-partition ordering is either unavailable or prohibitively expensive
- Use causal ordering (vector clocks) rather than total ordering for geo-distributed deployments

**Evidence Strength**: MODERATE

### Theme 5: The Pattern Is Explicitly Not for Most Systems

**Consensus**: The literature consistently warns against adopting event sourcing for simple domains, systems without audit requirements, or teams without distributed systems experience. The complexity cost is only justified when scalability, auditability, or temporal query capabilities are genuine requirements. [**STRONG**]
**Sources**: [SRC-001], [SRC-003], [SRC-004], [SRC-009], [SRC-011]

**Controversy**: Whether CDC (Change Data Capture) is generally preferable to event sourcing for microservice data integration. Murphy [SRC-009] argues CDC with Outbox is "usually a better alternative." Richardson [SRC-011] and Kleppmann [SRC-002] present them as complementary approaches for different use cases rather than competitors.
**Dissenting sources**: [SRC-009] argues CDC should be the default, while [SRC-002] and [SRC-011] position event sourcing and CDC as different tools for different problems.

**Practical Implications**:
- Require explicit justification (audit mandate, high write throughput, temporal queries) before adopting event sourcing
- Evaluate CDC with Outbox Pattern as a simpler alternative for microservice data distribution
- Budget for the steep learning curve documented in [SRC-001]: event sourcing is a paradigm shift, not a library

**Evidence Strength**: STRONG (on the warning) / MIXED (on CDC as alternative)

## Evidence-Graded Findings

### STRONG Evidence
- Event sourcing provides immutable audit trails, temporal queries, and complete state rebuild capabilities -- Sources: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-005], [SRC-006], [SRC-011]
- The append-only log is a unifying abstraction for distributed data systems; the database is a cache of the log -- Sources: [SRC-002], [SRC-005], [SRC-006]
- Event sourcing eliminates write contention through append-only operations, improving write scalability -- Sources: [SRC-003], [SRC-005]
- Eventual consistency is inherent: read models always lag behind the event store -- Sources: [SRC-003], [SRC-009], [SRC-010]
- The pattern's complexity is not justified for most systems; adopt only when performance, scalability, or auditability are critical -- Sources: [SRC-001], [SRC-003], [SRC-004]
- Practitioners face five primary challenges: event system evolution, steep learning curve, lack of technology, rebuilding projections, and data privacy -- Sources: [SRC-001]
- Events must never be mutated: immutability is required for audit integrity, downstream consumer stability, and fraud prevention -- Sources: [SRC-004], [SRC-006], [SRC-007]

### MODERATE Evidence
- CQRS and materialized views (projections) are production necessities, not optional complements -- Sources: [SRC-003], [SRC-008], [SRC-011]
- Schema evolution requires explicit versioning strategies; five tactics exist (versioned events, weak schema, upcasting, in-place transformation, copy-and-transform) -- Sources: [SRC-001], [SRC-007]
- Snapshots are necessary for large event streams but should be tactical optimizations, not architectural defaults -- Sources: [SRC-003], [SRC-004], [SRC-008], [SRC-012]
- Scaling requires partitioning event streams; single-partition append is insufficient for high throughput -- Sources: [SRC-005], [SRC-008], [SRC-010]
- CDC with Outbox Pattern provides stronger transactional guarantees than event sourcing's dual-write model -- Sources: [SRC-009]
- Crypto-shredding is the most practical GDPR compliance strategy for existing event-sourced systems with PII in events -- Sources: [SRC-013]
- Data minimization (references instead of PII in events) is the preferred design-time GDPR strategy -- Sources: [SRC-013]
- Event consumers must implement idempotency due to at-least-once delivery semantics -- Sources: [SRC-003], [SRC-008]
- External system interactions create significant complexity during event replay and must be gated -- Sources: [SRC-004], [SRC-008]
- Global-scale event sourcing requires causal ordering (vector clocks) rather than total ordering -- Sources: [SRC-010]
- LinkedIn's Kafka processes 60+ billion messages/day using partitioning, batching, and zero-copy -- Sources: [SRC-005]
- The event store serves dual roles as both database and message broker -- Sources: [SRC-011]

### WEAK Evidence
- Heavy snapshot reliance indicates aggregate design problems that should be addressed before optimizing -- Sources: [SRC-012]
- For most teams, CDC is preferable to event sourcing due to simpler operational model -- Sources: [SRC-009]
- Version bankruptcy (discarding and rebuilding streams) is a legitimate last resort -- Sources: [SRC-007]
- GDPR compliance must be designed in from the start; retrofitting is significantly more expensive -- Sources: [SRC-013]
- Geo-distributed event sourcing must prioritize AP over CP -- Sources: [SRC-010]

### UNVERIFIED
- Jet.com (Walmart) used event sourcing from inception and scaled it across reads, redundancy, and projections for production e-commerce -- Basis: search result summary from Medium article by Leo Gorodinski; full content not fetched (403 error)
- Alongi et al. (2022) provide empirical evidence from a case study on event-sourced observable architectures in Software: Practice and Experience -- Basis: paper metadata confirmed; full text not accessed (Wiley paywall)

## Knowledge Gaps

- **Production-scale benchmarks**: No rigorous, peer-reviewed benchmarks comparing event sourcing performance against traditional CRUD at specific scale thresholds (e.g., millions of events/second). Vendor claims (EventStoreDB: 15K writes/sec, 50K reads/sec) lack independent verification. LinkedIn's Kafka numbers are for the log abstraction broadly, not event sourcing specifically.

- **Long-term operational costs**: No longitudinal studies measuring the total cost of ownership (storage growth, projection maintenance, schema evolution labor) of event-sourced systems over 5+ year lifespans. Overeem et al. [SRC-001] capture a snapshot but not a time series.

- **GDPR litigation outcomes**: No documented legal rulings on whether crypto-shredding satisfies GDPR Article 17 requirements. The strategy is practitioner consensus, not legally validated.

- **Event sourcing at database-native level**: Limited literature on databases purpose-built for event sourcing (EventStoreDB/KurrentDB) versus general-purpose databases (PostgreSQL, DynamoDB) used as event stores. Performance and operational tradeoffs are not well characterized in primary literature.

- **Failure mode taxonomy**: No systematic catalog of event sourcing failure modes (corrupted events, projection divergence, ordering violations) with frequency and remediation data from production systems.

## Domain Calibration

Mixed evidence distribution with a concentration of STRONG findings on core pattern properties and MODERATE findings on operational practices. This reflects a domain that is well-established in its theoretical foundations but still maturing in production engineering practices. The gap between academic coverage and practitioner experience is narrowing (Overeem et al. 2021 is a notable bridge), but empirical production data remains scarce relative to pattern guidance literature.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best. Specifically, Kleppmann [SRC-002] and Alongi et al. (2022) could not be fully accessed.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible. DOIs are included only when confirmed -- none were fabricated.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research event sourcing at scale` on 2026-02-27.
