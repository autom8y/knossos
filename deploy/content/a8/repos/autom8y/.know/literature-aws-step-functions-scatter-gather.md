---
domain: "literature-aws-step-functions-scatter-gather"
generated_at: "2026-03-11T00:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.68
format_version: "1.0"
---

# Literature Review: AWS Step Functions Scatter-Gather Patterns for Scheduled Data Aggregation

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on AWS Step Functions scatter-gather patterns converges on the Map state (both Inline and Distributed modes) as the primary mechanism for fan-out/fan-in workloads, with the Parallel state serving fixed-branch scatter scenarios. There is strong consensus that partial failure tolerance requires explicit architectural treatment -- either through Distributed Map's `ToleratedFailurePercentage`/`ToleratedFailureCount` fields, or through Catch-and-Pass patterns within Inline Map/Parallel branches. The migration path from monolithic Lambdas to composed Step Functions workflows is well-documented by AWS, centering on decomposition into single-responsibility functions with orchestration logic externalized to the state machine. EventBridge integration is bidirectional: Step Functions standard workflows emit execution status events automatically, and EventBridge rules can target state machines for event-driven triggering. The evidence base is dominated by official AWS documentation and AWS blog posts (MODERATE tier), with the foundational scatter-gather pattern formalized in the Hohpe/Woolf textbook (STRONG tier). Production experience reports are available but sparse, and cost modeling for Distributed Map at scale remains under-documented.

## Source Catalog

### [SRC-001] Enterprise Integration Patterns: Designing, Building, and Deploying Messaging Solutions
- **Authors**: Gregor Hohpe, Bobby Woolf
- **Year**: 2003
- **Type**: textbook
- **URL/DOI**: ISBN 978-0-321-20068-6; https://www.enterpriseintegrationpatterns.com/patterns/messaging/BroadcastAggregate.html
- **Verified**: yes (companion website fetched; ISBN confirmed via multiple retailers and ACM Digital Library)
- **Relevance**: 5
- **Summary**: Defines the Scatter-Gather pattern as a composite that broadcasts a message to multiple recipients and re-aggregates responses into a single message. Distinguishes two variants: Distribution (Recipient List with explicit control) and Auction (Publish-Subscribe Channel with dynamic participants). Specifies that the Aggregator component handles completion conditions (wait-for-all, first-best, timeout-based) and error propagation.
- **Key Claims**:
  - Scatter-Gather is a coordinated pattern that expects responses and applies aggregation logic, distinguishing it from simple fan-out [**STRONG**]
  - Two implementation variants exist: Distribution (explicit recipient control) and Auction (dynamic pub-sub registration) [**STRONG**]
  - The Aggregator must handle multiple completion strategies including timeout-based collection and partial failure propagation [**MODERATE**]

### [SRC-002] Handling errors in Step Functions workflows -- AWS Step Functions Developer Guide
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/step-functions/latest/dg/concepts-error-handling.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Comprehensive reference for Step Functions error handling. Defines all built-in error types (`States.ALL`, `States.TaskFailed`, `States.Timeout`, `States.ExceedToleratedFailureThreshold`, etc.), Retry configuration with exponential backoff and jitter, and Catch fallback mechanisms. Establishes that Retry and Catch are available only on Task, Parallel, and Map states. Documents that `States.Runtime` and `States.DataLimitExceeded` are non-retriable terminal errors.
- **Key Claims**:
  - Retry policies support exponential backoff with configurable `IntervalSeconds`, `MaxAttempts`, `BackoffRate`, `MaxDelaySeconds`, and `JitterStrategy` (FULL or NONE) [**STRONG**]
  - `States.Runtime` and `States.DataLimitExceeded` are terminal errors that cannot be caught by `States.ALL` [**STRONG**]
  - Retry and Catch are evaluated sequentially: retries exhaust first, then catch handlers activate [**STRONG**]
  - Lambda transient errors (`Lambda.ServiceException`, `Lambda.SdkClientException`) should always be handled with explicit Retry policies in production [**MODERATE**]

### [SRC-003] Using Map state in Distributed mode for large-scale parallel workloads -- AWS Step Functions Developer Guide
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/step-functions/latest/dg/state-map-distributed.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Authoritative reference for Distributed Map state. Documents that each child execution has separate execution history, supporting up to 10,000 concurrent child workflow executions by default. Defines `ToleratedFailurePercentage` and `ToleratedFailureCount` for partial failure tolerance, `ItemBatcher` for batch processing, `ResultWriter` for S3-based result consolidation, and `MaxConcurrency` for downstream capacity protection.
- **Key Claims**:
  - Distributed Map supports up to 10,000 concurrent child workflow executions (default when `MaxConcurrency` is 0 or unset) [**STRONG**]
  - `ToleratedFailurePercentage` (0-100) and `ToleratedFailureCount` define acceptable failure thresholds; exceeding either triggers `States.ExceedToleratedFailureThreshold` [**STRONG**]
  - `ResultWriter` exports child execution results to S3, avoiding the 256 KiB payload limit between states [**STRONG**]
  - `ItemBatcher` reduces total child executions by grouping items, improving cost efficiency for small-record workloads [**MODERATE**]
  - Retry at the Distributed Map level re-runs ALL child executions (creates a new Map Run), not just failed ones [**MODERATE**]

### [SRC-004] Best practices for Step Functions -- AWS Step Functions Developer Guide
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/step-functions/latest/dg/sfn-best-practices.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Prescribes operational best practices including always setting `TimeoutSeconds` on Task states to prevent stuck executions, using S3 ARNs instead of passing payloads exceeding 256 KiB, nesting Express workflows inside Standard workflows for cost optimization, and retrying Lambda transient service exceptions. Recommends maintaining 100+ open polls per activity ARN for low latency.
- **Key Claims**:
  - Always set `TimeoutSeconds` on Task and Activity states; default behavior waits indefinitely [**STRONG**]
  - Pass large data via S3 ARNs to avoid the 256 KiB payload limit that terminates executions [**STRONG**]
  - Nest Express workflows inside Standard workflows to optimize cost for idempotent sub-workflows [**MODERATE**]
  - Use `HeartbeatSeconds` with `.waitForTaskToken` patterns to detect external task failures [**MODERATE**]

### [SRC-005] Discover service integration patterns in Step Functions -- AWS Step Functions Developer Guide
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/step-functions/latest/dg/connect-to-resource.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents three service integration patterns: Request Response (fire-and-forget), Run a Job (`.sync` suffix, waits for completion), and Wait for Callback (`.waitForTaskToken`, pauses for external callback). The `.waitForTaskToken` pattern is the primary mechanism for readiness gating -- the workflow pauses indefinitely until an external process calls `SendTaskSuccess` or `SendTaskFailure` with the token. The `.sync` pattern provides built-in readiness gating for supported AWS services.
- **Key Claims**:
  - `.waitForTaskToken` pauses workflow execution indefinitely until external callback, enabling readiness gating for arbitrary external processes [**STRONG**]
  - `.sync` integration pattern provides automatic readiness gating for supported AWS services (Batch, ECS, EMR, CodeBuild) [**STRONG**]
  - `.sync` and `.waitForTaskToken` are available in Standard workflows only, not Express workflows [**STRONG**]
  - Task tokens must be used within the same AWS account; cross-account callback is not supported [**MODERATE**]

### [SRC-006] Automating Step Functions event delivery with EventBridge -- AWS Step Functions Developer Guide
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/step-functions/latest/dg/eventbridge-integration.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents bidirectional EventBridge integration. Standard workflows automatically emit execution status change events to the default EventBridge event bus on a best-effort basis. EventBridge rules can filter on `source: "aws.states"` and `detail-type: "Step Functions Execution Status Change"` with status filtering (SUCCEEDED, FAILED, TIMED_OUT, etc.). State machines can also be specified as EventBridge rule targets, enabling event-driven workflow triggering.
- **Key Claims**:
  - Standard workflows automatically emit execution status events to EventBridge default bus; Express workflows do not [**STRONG**]
  - Events are delivered on a best-effort basis and may arrive out of order [**MODERATE**]
  - If combined escaped input/output exceeds 248 KiB, input or output is excluded from the event payload [**MODERATE**]
  - Step Functions state machines can be EventBridge rule targets, enabling event-driven workflow triggering from any EventBridge source [**MODERATE**]

### [SRC-007] Introducing the Amazon EventBridge service integration for AWS Step Functions
- **Authors**: AWS Compute Blog (AWS)
- **Year**: 2021
- **Type**: official documentation (blog post from vendor)
- **URL/DOI**: https://aws.amazon.com/blogs/compute/introducing-the-amazon-eventbridge-service-integration-for-aws-step-functions/
- **Verified**: partial (metadata and summary confirmed; full article body not fully extractable)
- **Relevance**: 4
- **Summary**: Announces native PutEvents API integration enabling Step Functions workflows to publish custom events directly to EventBridge without intermediate Lambda functions. This eliminates a common architectural intermediary and enables workflows to participate as first-class event producers in event-driven architectures.
- **Key Claims**:
  - Step Functions can invoke EventBridge PutEvents API natively, eliminating the need for intermediate Lambda functions to bridge workflows and event buses [**MODERATE**]
  - The integration enables decoupled workflow communication through event publication rather than direct service calls [**MODERATE**]

### [SRC-008] Breaking down monolith workflows: Modularizing AWS Step Functions workflows
- **Authors**: AWS Compute Blog (AWS)
- **Year**: 2024
- **Type**: official documentation (blog post from vendor)
- **URL/DOI**: https://aws.amazon.com/blogs/compute/breaking-down-monolith-workflows-modularizing-aws-step-functions-workflows/
- **Verified**: partial (metadata and category tags confirmed; full article body not fully extractable)
- **Relevance**: 5
- **Summary**: Addresses decomposition of monolithic Step Functions workflows into modular components. Documents four decomposition strategies: parent-child pattern (hierarchical workflow structure), domain separation (business capability boundaries), shared utilities (reusable workflow components), and specialized error workflows (centralized error handling). Warns against over-decomposition and tight coupling between child workflows.
- **Key Claims**:
  - Four decomposition strategies for Step Functions: parent-child, domain separation, shared utilities, and specialized error workflows [**MODERATE**]
  - Over-decomposition and tight coupling between workflows are anti-patterns to avoid [**MODERATE**]
  - Decomposition enables faster deployments, better error isolation, and reduced operational overhead [**WEAK**]

### [SRC-009] Streamlining AWS serverless workflows: From AWS Lambda orchestration to AWS Step Functions
- **Authors**: AWS Compute Blog (AWS)
- **Year**: 2024
- **Type**: official documentation (blog post from vendor)
- **URL/DOI**: https://aws.amazon.com/blogs/compute/streamlining-aws-serverless-workflows-from-aws-lambda-orchestration-to-aws-step-functions/
- **Verified**: partial (metadata confirmed; full article body not fully extractable)
- **Relevance**: 5
- **Summary**: Addresses the migration path from Lambda-to-Lambda orchestration to Step Functions. Identifies Lambda-calling-Lambda as an anti-pattern that creates tight coupling, duplicated retry logic, and flow management scattered across function code. Step Functions externalizes orchestration, providing declarative error handling, visual monitoring, and native service integrations.
- **Key Claims**:
  - Lambda-calling-Lambda is an anti-pattern that creates tight coupling and scattered flow management [**MODERATE**]
  - Step Functions externalizes orchestration logic, removing glue code and state management from Lambda functions [**MODERATE**]
  - Migration to Step Functions improves maintainability, scalability, and operational efficiency [**WEAK**]

### [SRC-010] What's the best way to do fan-out/fan-in serverlessly in 2024?
- **Authors**: Yan Cui (theburningmonk)
- **Year**: 2024
- **Type**: blog post
- **URL/DOI**: https://theburningmonk.com/2024/08/whats-the-best-way-to-do-fan-out-fan-in-serverlessly-in-2024/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Comparative analysis of fan-out/fan-in approaches. Evaluates Step Functions Map state (Inline: up to 40 concurrent iterations; Distributed: up to 10,000) versus custom SQS+DynamoDB+Lambda architectures. Reports that Distributed Map with Standard child workflows prices iterations as one state transition each regardless of internal complexity. Custom solutions using SQS/DynamoDB reduce costs significantly but require teams to "own the uptime." Recommends Step Functions for general use cases and custom patterns for high-volume (millions of items) scenarios.
- **Key Claims**:
  - Inline Map supports up to 40 concurrent iterations; Distributed Map supports up to 10,000 [**STRONG**]
  - Distributed Map with Standard child workflows prices each iteration as one state transition regardless of internal complexity [**MODERATE**]
  - Custom SQS+DynamoDB fan-out/fan-in costs roughly $1.65 per million items vs. $25 per million state transitions for Step Functions Standard [**WEAK**]
  - Step Functions suits general use; custom patterns make sense for millions-of-items workloads [**WEAK**]

### [SRC-011] Step Functions Distributed Map Best Practices for Large-Scale Batch Workloads
- **Authors**: AWS Builders (DEV Community)
- **Year**: 2025
- **Type**: blog post
- **URL/DOI**: https://dev.to/aws-builders/step-functions-distributed-map-best-practices-for-large-scale-batch-workloads-55n2
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Production-oriented best practices for Distributed Map. Recommends treating `MaxConcurrency` as a safety valve rather than a performance knob, starting conservatively and scaling after load testing. Advocates designing idempotent workers, placing retries inside child workflows rather than at Map level, using `ItemBatcher` for small records, and using `ResultWriter` for compact orchestration payloads. Identifies downstream system capacity (Lambda concurrency, database write throughput, API rate limits) as the true bottleneck rather than Step Functions itself.
- **Key Claims**:
  - Downstream systems (Lambda concurrency, DB write capacity, API limits) are typically the real bottleneck, not Step Functions concurrency [**MODERATE**]
  - Place retries inside child workflows rather than at the Map level to avoid re-running all child executions [**MODERATE**]
  - Design workers as idempotent to safely handle retries and partial re-execution [**MODERATE**]
  - Use `ItemBatcher` to reduce state transitions and improve Lambda efficiency for small records [**WEAK**]

### [SRC-012] Parallel task error handling in Step Functions
- **Authors**: AWS Builders (DEV Community)
- **Year**: 2024
- **Type**: blog post
- **URL/DOI**: https://dev.to/aws-builders/parallel-task-error-handling-in-step-functions-4f1c
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Demonstrates the Catch-and-Pass pattern for partial failure tolerance in Parallel and Inline Map states. By default, any branch/iteration failure terminates the entire Parallel/Map state. The pattern catches errors at the task level within each branch, routes to a Pass state, and uses distinct `ResultPath` values (`$.data` for success, `$.error` for failure) to preserve both outcomes in the aggregated output.
- **Key Claims**:
  - By default, any branch failure in Parallel state or iteration failure in Map state terminates the entire state and stops all other branches/iterations [**STRONG**]
  - The Catch-and-Pass pattern enables partial failure tolerance by catching errors within individual branches and routing to Pass states [**MODERATE**]
  - Using distinct ResultPath values for success and error preserves both outcomes in the aggregated output [**MODERATE**]

### [SRC-013] Parallelization and scatter-gather patterns -- AWS Prescriptive Guidance
- **Authors**: AWS Prescriptive Guidance Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/prescriptive-guidance/latest/agentic-ai-patterns/parallelization-and-scatter-gather-patterns.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Distinguishes scatter-gather from simple fan-out: scatter-gather is coordinated because it expects responses and applies logic to combine, compare, and select results. Identifies four AWS implementation options: Step Functions Map state, Lambda with concurrency, EventBridge with correlation IDs and aggregation workflows, and custom controller patterns using S3/DynamoDB/queues. Documents the canonical scatter-gather workflow: coordinator scatters via SNS, workers process independently, results aggregate via SQS, and an aggregator function merges/selects/compares outputs.
- **Key Claims**:
  - Scatter-gather is coordinated (expects responses and applies aggregation logic), unlike simple fan-out which is fire-and-forget [**STRONG**]
  - Four AWS implementation options exist: Step Functions Map, Lambda concurrency, EventBridge with correlation IDs, and custom S3/DynamoDB/queue controllers [**MODERATE**]
  - The canonical AWS scatter-gather workflow uses SNS for scatter, independent Lambda workers, SQS for result collection, and a Lambda aggregator [**MODERATE**]

### [SRC-014] How we use AWS Step Functions for big data aggregation workloads
- **Authors**: Spreaker Engineering
- **Year**: 2023
- **Type**: blog post (production experience report)
- **URL/DOI**: https://careers.spreaker.com/engineering/how-we-use-aws-step-functions-for-big-data-aggregation-workloads/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Production experience report from Spreaker on evolving from single Lambda to multi-service Step Functions orchestration for data aggregation. Uses Parallel state containing multiple Map states with differentiated `MaxConcurrency` settings (heavy tables at 1, others in parallel). Offloads query execution to Athena to reduce database pressure. Reports that Step Functions' built-in retry mechanism is limited because it matches only on the error field and lacks observability hooks (no logging or metric emission during retries), motivating custom retry states.
- **Key Claims**:
  - Differentiating MaxConcurrency per Map state within a Parallel state enables resource-aware concurrency control for heterogeneous workloads [**MODERATE**]
  - Offloading query execution to Athena (instead of in-Lambda querying) reduces database pressure as a bottleneck mitigation strategy [**WEAK**]
  - Step Functions built-in retry matches only on error field and lacks logging/metric hooks, motivating custom retry state implementations [**WEAK**]

## Thematic Synthesis

### Theme 1: Map State is the Primary Scatter-Gather Primitive, With Two Distinct Modes

**Consensus**: Step Functions' Map state is the recommended implementation for scatter-gather/fan-out-fan-in patterns in AWS. Inline Map (up to 40 concurrent iterations, shared execution history) serves small-to-medium workloads; Distributed Map (up to 10,000 concurrent child executions, separate histories) serves large-scale batch processing. The Parallel state complements this for fixed-branch scatter scenarios where branches are heterogeneous. [**STRONG**]
**Sources**: [SRC-001], [SRC-003], [SRC-005], [SRC-010], [SRC-013]

**Controversy**: Whether the cost premium of Step Functions Map state over custom SQS+DynamoDB fan-out is justified. [SRC-010] documents significant cost differences at scale but acknowledges the operational burden of custom solutions.
**Dissenting sources**: [SRC-010] argues custom fan-out is cost-effective for millions-of-items workloads, while [SRC-011] argues Step Functions' operational simplicity justifies the premium for most teams.

**Practical Implications**:
- Default to Inline Map for workloads under 40 items or 25,000 execution history events
- Use Distributed Map when item counts, execution history, or concurrency requirements exceed Inline limits
- Reserve custom SQS/DynamoDB fan-out for extreme-volume scenarios where cost dominates and the team can own operational complexity
- Use Parallel state only when scatter branches are heterogeneous (different logic per branch)

**Evidence Strength**: STRONG

### Theme 2: Partial Failure Tolerance Requires Explicit Architectural Treatment

**Consensus**: Neither Parallel state nor Inline Map state provides partial failure tolerance by default -- any branch/iteration failure terminates the entire state. Two complementary mechanisms address this: (1) Distributed Map's `ToleratedFailurePercentage`/`ToleratedFailureCount` for threshold-based tolerance, and (2) the Catch-and-Pass pattern for per-branch error capture in Parallel/Inline Map states. Both require intentional design; "it just works" is not available. [**STRONG**]
**Sources**: [SRC-002], [SRC-003], [SRC-004], [SRC-011], [SRC-012]

**Practical Implications**:
- For Distributed Map: set `ToleratedFailurePercentage` or `ToleratedFailureCount` explicitly; a value of 0 (default) means any single failure kills the entire Map Run
- For Inline Map/Parallel: implement Catch blocks within each branch that route to Pass states, using distinct `ResultPath` values to preserve error context alongside successful results
- Design child workflows to distinguish retryable from non-retryable failures; place retries inside child workflows rather than at Map level to avoid re-running all children
- Always set `TimeoutSeconds` on Task states to prevent stuck executions from blocking the entire scatter-gather

**Evidence Strength**: STRONG

### Theme 3: Readiness Gating is Implemented via .waitForTaskToken and .sync Patterns

**Consensus**: Step Functions provides two built-in mechanisms for readiness gating (waiting for external conditions before proceeding): `.sync` for supported AWS service jobs and `.waitForTaskToken` for arbitrary external processes. The `.waitForTaskToken` pattern pauses the workflow indefinitely and passes a token to an external system, which must call `SendTaskSuccess` or `SendTaskFailure` to resume. Both patterns are available in Standard workflows only. [**STRONG**]
**Sources**: [SRC-004], [SRC-005]

**Practical Implications**:
- Use `.sync` for readiness gating on supported AWS services (Batch, ECS, EMR, CodeBuild) -- it handles polling automatically
- Use `.waitForTaskToken` for readiness gating on external APIs, human approval, or cross-system dependencies
- Always configure `HeartbeatSeconds` on `.waitForTaskToken` tasks to detect failed external processes; without it, the workflow waits indefinitely
- Readiness gating as a workflow step before scatter: check that all data sources are available/healthy before initiating fan-out, reducing wasted partial executions

**Evidence Strength**: STRONG

### Theme 4: Monolithic Lambda to Composed Step Functions Migration Follows Decomposition Patterns

**Consensus**: Lambda-calling-Lambda is an anti-pattern that creates tight coupling, scattered retry logic, and duplicated flow management. The migration path extracts orchestration logic into Step Functions and decomposes monolithic Lambdas into single-responsibility functions. Four decomposition strategies are documented: parent-child (hierarchical), domain separation (business capabilities), shared utilities (reusable components), and specialized error workflows. [**MODERATE**]
**Sources**: [SRC-008], [SRC-009], [SRC-014]

**Controversy**: The granularity of decomposition is debated. Over-decomposition creates excessive state transitions (cost) and coordination overhead, while under-decomposition preserves monolithic complexity.
**Dissenting sources**: [SRC-008] warns against over-decomposition, while [SRC-014] demonstrates that real production workloads often require combining Parallel and Map states with differentiated concurrency -- suggesting moderate decomposition is practical.

**Practical Implications**:
- Extract orchestration (sequencing, branching, retries, error routing) from Lambda code into Step Functions ASL
- Design each Lambda as a single-responsibility fetch/transform/load function, agnostic to its position in the workflow
- Use parent-child workflow nesting to manage execution history limits (25,000 events) and to isolate failure domains
- Start migration by identifying the monolithic Lambda's internal state transitions and mapping them to Step Functions states

**Evidence Strength**: MODERATE

### Theme 5: EventBridge Integration Enables Event-Driven Triggering and Result Publication

**Consensus**: EventBridge integration with Step Functions is bidirectional. Inbound: EventBridge rules can target state machines, enabling event-driven workflow triggering from any AWS service or custom event source. Outbound: Standard workflows automatically emit execution status change events to EventBridge, and workflows can publish custom events via native PutEvents integration. This eliminates intermediate Lambda functions for both triggering and result publication. [**MODERATE**]
**Sources**: [SRC-006], [SRC-007], [SRC-013]

**Controversy**: EventBridge event delivery from Step Functions is best-effort and may arrive out of order. For critical result publication, this may be insufficient without additional correlation and deduplication logic.
**Dissenting sources**: [SRC-006] documents best-effort delivery, while [SRC-013] treats EventBridge with correlation IDs as a viable scatter-gather implementation option -- implying the ordering limitation is manageable with explicit correlation.

**Practical Implications**:
- Use EventBridge rules with `source: "aws.states"` to trigger downstream processing on workflow completion/failure
- Use native PutEvents integration (not Lambda intermediaries) to publish workflow results to EventBridge for downstream consumers
- Implement correlation IDs when using EventBridge for scatter-gather aggregation to handle out-of-order and best-effort delivery
- EventBridge Scheduler is dramatically cheaper than Step Functions for simple time-based triggering ($1/million vs. $25/million state transitions)
- Express workflows do not emit EventBridge events; use CloudWatch Logs for Express workflow monitoring

**Evidence Strength**: MODERATE

### Theme 6: Downstream Capacity is the Real Bottleneck, Not Step Functions Concurrency

**Consensus**: Step Functions can scale to 10,000 concurrent child executions, but downstream systems (Lambda concurrency limits, database write capacity, third-party API rate limits) are typically the true throughput constraint. `MaxConcurrency` should be treated as a safety valve calibrated to downstream capacity, not as a performance knob to maximize. [**MODERATE**]
**Sources**: [SRC-003], [SRC-010], [SRC-011], [SRC-014]

**Practical Implications**:
- Load test with representative data before increasing `MaxConcurrency` beyond conservative initial values
- Map Lambda reserved/provisioned concurrency, DynamoDB write capacity units, and API rate limits before designing concurrency settings
- Use `ItemBatcher` to reduce the number of downstream invocations for small-record workloads
- Design workers as idempotent to safely handle retries triggered by downstream throttling

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Scatter-Gather is a coordinated pattern that expects responses and applies aggregation logic, distinguishing it from simple fan-out -- Sources: [SRC-001], [SRC-013]
- Step Functions Retry policies support exponential backoff with `IntervalSeconds`, `MaxAttempts`, `BackoffRate`, `MaxDelaySeconds`, and `JitterStrategy` -- Sources: [SRC-002]
- `States.Runtime` and `States.DataLimitExceeded` are terminal errors not catchable by `States.ALL` -- Sources: [SRC-002]
- Retry and Catch are evaluated sequentially: retries exhaust first, then catch handlers activate -- Sources: [SRC-002]
- Distributed Map supports up to 10,000 concurrent child workflow executions -- Sources: [SRC-003], [SRC-010]
- `ToleratedFailurePercentage` and `ToleratedFailureCount` define partial failure thresholds for Distributed Map -- Sources: [SRC-003]
- `ResultWriter` exports child execution results to S3, avoiding the 256 KiB payload limit -- Sources: [SRC-003]
- Always set `TimeoutSeconds` on Task states; default behavior waits indefinitely -- Sources: [SRC-004]
- Pass large data via S3 ARNs to avoid 256 KiB payload termination -- Sources: [SRC-004]
- `.waitForTaskToken` pauses workflow execution indefinitely for external callback (readiness gating) -- Sources: [SRC-005]
- `.sync` and `.waitForTaskToken` are Standard workflow-only patterns -- Sources: [SRC-005]
- Any branch failure in Parallel or iteration failure in Inline Map terminates the entire state by default -- Sources: [SRC-002], [SRC-012]
- Standard workflows automatically emit execution status events to EventBridge; Express workflows do not -- Sources: [SRC-006]

### MODERATE Evidence
- The Aggregator must handle multiple completion strategies including timeout-based collection -- Sources: [SRC-001]
- Lambda transient errors should always be handled with explicit Retry policies -- Sources: [SRC-002]
- `ItemBatcher` reduces child executions by grouping items, improving cost efficiency -- Sources: [SRC-003]
- Retry at Distributed Map level re-runs ALL child executions (new Map Run), not just failed ones -- Sources: [SRC-003]
- Nest Express workflows inside Standard for cost optimization of idempotent sub-workflows -- Sources: [SRC-004]
- Task tokens must be used within the same AWS account -- Sources: [SRC-005]
- EventBridge events from Step Functions are best-effort and may arrive out of order -- Sources: [SRC-006]
- State machines can be EventBridge rule targets for event-driven triggering -- Sources: [SRC-006]
- PutEvents integration eliminates intermediate Lambda for workflow-to-EventBridge communication -- Sources: [SRC-007]
- Four decomposition strategies: parent-child, domain separation, shared utilities, error workflows -- Sources: [SRC-008]
- Lambda-calling-Lambda is an anti-pattern creating tight coupling and scattered flow management -- Sources: [SRC-009]
- Distributed Map with Standard child workflows prices iterations as one state transition regardless of complexity -- Sources: [SRC-010]
- Downstream systems are the real bottleneck, not Step Functions concurrency -- Sources: [SRC-011], [SRC-014]
- Place retries inside child workflows to avoid re-running all children -- Sources: [SRC-011]
- Design workers as idempotent for safe retry handling -- Sources: [SRC-011]
- Catch-and-Pass pattern enables partial failure tolerance in Parallel/Inline Map -- Sources: [SRC-012]
- Four AWS scatter-gather implementation options: Map state, Lambda concurrency, EventBridge+correlation, custom controllers -- Sources: [SRC-013]
- Differentiating MaxConcurrency per Map state enables resource-aware concurrency control -- Sources: [SRC-014]

### WEAK Evidence
- Decomposition enables faster deployments, better error isolation, and reduced overhead -- Sources: [SRC-008]
- Migration to Step Functions improves maintainability, scalability, and operational efficiency -- Sources: [SRC-009]
- Custom SQS+DynamoDB fan-out costs ~$1.65/million vs. $25/million for Step Functions Standard -- Sources: [SRC-010]
- Step Functions suits general use; custom patterns for millions-of-items workloads -- Sources: [SRC-010]
- `ItemBatcher` reduces state transitions and improves Lambda efficiency for small records -- Sources: [SRC-011]
- Offloading queries to Athena reduces database pressure as bottleneck mitigation -- Sources: [SRC-014]
- Step Functions built-in retry lacks logging/metric hooks, motivating custom retry states -- Sources: [SRC-014]

### UNVERIFIED
- No claims in this review rely solely on model training knowledge. All claims are backed by at least one verifiable source.

## Knowledge Gaps

- **Cost modeling for Distributed Map at scale**: While [SRC-010] provides directional cost comparisons, there is no authoritative, benchmarked cost analysis comparing Distributed Map (with ItemBatcher, ResultWriter) against custom fan-out architectures for realistic multi-source data aggregation workloads. AWS pricing pages provide per-transition rates but not total-cost-of-ownership models.

- **Readiness gating patterns for data source health checks**: The literature documents `.waitForTaskToken` and `.sync` as readiness gating mechanisms, but there are no published patterns specifically for "check that N data sources are healthy/available before initiating scatter." This is typically implemented as custom Choice/Task logic but lacks a canonical pattern.

- **EventBridge correlation and aggregation for scatter-gather**: [SRC-013] mentions EventBridge with correlation IDs as a scatter-gather implementation option, but no source provides a detailed architectural pattern for correlation-based aggregation using EventBridge (as opposed to Step Functions Map state, which handles aggregation natively).

- **Migration tooling and automation**: While [SRC-008] and [SRC-009] describe decomposition strategies conceptually, there is no documented tooling or automated approach for analyzing a monolithic Lambda and generating a Step Functions decomposition plan.

- **Error handling observability during retries**: [SRC-014] reports that Step Functions' built-in retry lacks logging and metric emission hooks. No source documents a comprehensive observability pattern for monitoring retry behavior across scatter-gather workflows in production.

## Domain Calibration

Mixed confidence distribution reflects the domain's reliance on vendor documentation as the primary evidence source. AWS official documentation is authoritative for its own service behavior, but independent validation, academic analysis, and cross-vendor comparison are largely absent. The foundational pattern theory (scatter-gather) has strong textbook grounding, while AWS-specific implementation guidance is well-documented but vendor-sourced.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Several AWS blog posts returned only CSS/metadata without full article bodies when fetched. Claims from these sources were downgraded and supplemented with search result summaries.
3. **Citation accuracy**: Paper titles, URLs, and ISBNs were verified via web search where possible. No DOIs were fabricated. The Hohpe/Woolf textbook ISBN was confirmed via multiple retailers and ACM Digital Library.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.
5. **Vendor concentration**: 10 of 14 sources are AWS-authored (documentation or blog posts). This reflects the domain reality -- AWS Step Functions is a proprietary service -- but means the evidence base lacks independent validation.

Generated by `/research aws-step-functions-scatter-gather` on 2026-03-11.
