---
domain: "literature-aws-fargate-spot-capacity-providers"
generated_at: "2026-03-25T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.64
format_version: "1.0"
---

# Literature Review: AWS Fargate Spot Capacity Providers

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

AWS Fargate Spot offers up to 70% cost savings over standard Fargate by running tasks on spare AWS capacity, but carries significant trade-offs around interruption reliability and capacity availability that limit its suitability for production-facing workloads. The literature broadly agrees that Fargate Spot is well-suited for batch processing, queue workers, CI/CD pipelines, and development environments, while consensus is weaker on its viability for production web services. Key controversies center on whether mixed Fargate/Fargate-Spot fleet configurations can adequately mitigate interruption risk for user-facing services, and on the absence of any automatic fallback mechanism from Spot to on-demand capacity. Evidence quality is moderate overall: official AWS documentation provides authoritative configuration guidance, but real-world interruption frequency data is almost entirely anecdotal, and AWS provides no Spot Advisor equivalent for Fargate (unlike EC2 Spot).

## Source Catalog

### [SRC-001] Amazon ECS clusters for Fargate -- Fargate Capacity Providers
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/fargate-capacity-providers.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Authoritative reference for Fargate and Fargate Spot capacity provider configuration. Documents weight/base strategy parameters, task distribution mechanics, SIGTERM handling with 2-minute warning, EventBridge integration for SpotInterruption events, and the constraint that capacity provider strategies cannot mix Fargate and Auto Scaling group providers. Specifies stopTimeout max of 120 seconds for Spot tasks and provides EventBridge event payload examples.
- **Key Claims**:
  - Fargate Spot tasks receive a 2-minute warning via both EventBridge task state change event and SIGTERM signal before termination [**STRONG**]
  - Only one capacity provider in a strategy can have a base value defined; weight determines proportional distribution of remaining tasks [**STRONG**]
  - Capacity provider strategies cannot contain both Fargate and Auto Scaling group providers [**MODERATE**]
  - Services with a single Spot task will remain interrupted until capacity returns; ECS does not fallback to on-demand [**STRONG**]

### [SRC-002] Deep Dive into Fargate Spot to Run Your ECS Tasks for Up to 70% Less
- **Authors**: Pritam Pal (Sr. EC2 Spot Specialist SA, AWS)
- **Year**: 2020
- **Type**: official documentation (AWS Compute Blog)
- **URL/DOI**: https://aws.amazon.com/blogs/compute/deep-dive-into-fargate-spot-to-run-your-ecs-tasks-for-up-to-70-less/
- **Verified**: partial (title and authorship confirmed via WebSearch; full article content not fully extractable)
- **Relevance**: 5
- **Summary**: Official AWS deep dive introducing Fargate Spot capacity providers. Explains that Spot price ranges between 50-70% off on-demand pricing (not a fixed 70% discount). Introduces capacity provider strategy concepts (base, weight) and details the interruption lifecycle: EventBridge event, SIGTERM, graceful shutdown window. Recommends catching SIGTERM and configuring Container Stop Timeout.
- **Key Claims**:
  - Fargate Spot discount ranges from 50% to 70% off on-demand pricing, not a fixed discount [**MODERATE**]
  - Capacity providers and capacity provider strategies are the mechanisms for managing Fargate/Fargate Spot task placement [**STRONG**]
  - Applications must implement SIGTERM handlers and set stopTimeout to 120 seconds for graceful shutdown [**STRONG**]

### [SRC-003] AWS Fargate Pricing Page
- **Authors**: AWS
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://aws.amazon.com/fargate/pricing/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Official pricing reference. Fargate Spot offers "up to 70% discount" on spare capacity. ARM/Graviton pricing is approximately 20% cheaper than x86 for both on-demand and Spot. Savings Plans offer up to 50% savings with 1- or 3-year commitments. Billing is per-second with 1-minute minimum (Linux) or 5-minute minimum (Windows). 20 GB ephemeral storage included at no charge; additional storage billed separately.
- **Key Claims**:
  - Fargate Spot offers up to 70% discount off standard Fargate rates [**STRONG**]
  - ARM/Graviton processors are approximately 20% cheaper per vCPU-hour and per GB-hour than x86 [**STRONG**]
  - Savings Plans offer up to 50% savings with commitment, representing an alternative to Spot for cost optimization [**MODERATE**]

### [SRC-004] The Trouble with Fargate Spot
- **Authors**: The Scale Factory (attributed to blog team)
- **Year**: 2020
- **Type**: blog post
- **URL/DOI**: https://scalefactory.com/blog/2020/07/27/the-trouble-with-fargate-spot/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Critical analysis of Fargate Spot for web-serving workloads. Identifies a fundamental timing issue: load balancers continue sending traffic to tasks that have received SIGTERM but have not yet been deregistered from target groups, causing 500 errors. Reports that interruptions can occur "several times per hour or more" depending on region and capacity. Concludes that Fargate Spot is unsuitable for customer-facing web applications and recommends EC2 Spot fleets or standard Fargate with Savings Plans instead. Notes there is no price history API for Fargate Spot (unlike EC2 Spot).
- **Key Claims**:
  - Load balancers continue routing traffic to Fargate Spot tasks after SIGTERM but before target group deregistration, causing 500 errors [**MODERATE**]
  - Fargate Spot interruption frequency can reach several times per hour in some regions/periods [**WEAK**]
  - There is no Fargate Spot price history API or equivalent of EC2 Spot Instance Advisor [**MODERATE**]
  - For customer-facing web applications, EC2 Spot fleets or standard Fargate with Savings Plans are preferable to Fargate Spot [**WEAK**]

### [SRC-005] Running a Web Application with 100% AWS Fargate Spot Containers
- **Authors**: Mauro Morales (DEV Community / AWS Builders)
- **Year**: 2023
- **Type**: blog post
- **URL/DOI**: https://dev.to/aws-builders/running-a-web-application-with-100-aws-fargate-spot-containers-5876
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Reports one year of production operation running 4 microservices on 100% Fargate Spot, serving approximately 2,000 daily users. Claims zero Fargate Spot capacity outages during the year. Tasks restarted "within a few seconds" of interruption with "virtually no effect on response times." Describes a dual-service fallback architecture: primary service on Spot with a dormant on-demand fallback service activated via Lambda when SERVICE_TASK_PLACEMENT_FAILURE events fire. Recommends minimum 2 desired tasks per service.
- **Key Claims**:
  - 100% Fargate Spot can sustain a production web application for 1+ year without capacity outages (in the author's specific region/config) [**WEAK**]
  - Dual-service architecture (Spot primary + dormant on-demand fallback triggered by Lambda) provides effective interruption mitigation [**WEAK**]
  - Minimum 2 desired tasks per service is necessary for Spot fault tolerance [**MODERATE**]
  - Task restart after Spot interruption occurs within seconds when capacity is available [**WEAK**]

### [SRC-006] AWS Fargate Spot vs. Fargate Price Comparison
- **Authors**: Tom Gregory
- **Year**: 2024 (updated)
- **Type**: blog post
- **URL/DOI**: https://tomgregory.com/aws-fargate-spot-vs-fargate-price-comparison
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Provides concrete pricing comparison in eu-west-1 region. Fargate on-demand: $0.04048/vCPU-hour, $0.004445/GB-hour. Fargate Spot: $0.01334053/vCPU-hour, $0.00146489/GB-hour. Real-world test with 100 containers per service running 24 hours showed $25.92 (Fargate) vs $8.93 (Fargate Spot), yielding 66% actual savings. Notes that the discount applies equally to both CPU and memory components.
- **Key Claims**:
  - Fargate Spot provides approximately 67% savings on both CPU and memory in eu-west-1 [**MODERATE**]
  - Real-world test of 100 containers confirmed 66% actual cost reduction ($25.92 vs $8.93 over 24 hours) [**WEAK**]

### [SRC-007] AWS Fargate Spot: Cost Optimization with Managed Container Workloads
- **Authors**: ElasticScale
- **Year**: 2025
- **Type**: blog post
- **URL/DOI**: https://elasticscale.com/blog/aws-fargate-spot-cost-optimization-with-managed-container-workloads/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Reports eu-west-1 pricing of $32.34/month (Fargate) vs $9.96/month (Fargate Spot) for 1 vCPU/1GB, yielding ~69% savings. Documents that capacity constraints typically affect specific AZs rather than entire regions. Observes container longevity variance: some tasks run 15+ days continuously while others face more frequent reclamation. Provides Fargate vs EC2 premium analysis: ~16% premium over c5.xlarge, ~21% over m5.xlarge, ~40% over t3.xlarge.
- **Key Claims**:
  - Fargate Spot provides ~69% savings in eu-west-1 for a 1 vCPU/1GB configuration [**MODERATE**]
  - Capacity constraints are typically AZ-scoped rather than region-wide [**WEAK**]
  - Fargate Spot task longevity varies widely: some run 15+ days, others face frequent reclamation [**WEAK**]
  - Fargate carries a 16-40% pricing premium over equivalent EC2 instance types [**WEAK**]

### [SRC-008] How to Use Spot Instances with ECS Fargate
- **Authors**: OneUptime
- **Year**: 2026
- **Type**: blog post
- **URL/DOI**: https://oneuptime.com/blog/post/2026-02-12-use-spot-instances-with-ecs-fargate/view
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Provides concrete capacity provider strategy configuration examples. Documents a standard mixed pattern: FARGATE base=2, weight=1 with FARGATE_SPOT weight=3, resulting in 2 guaranteed on-demand tasks plus a 1:3 ratio for additional scaling. Explains the SQS worker pattern where unacknowledged messages return to queue on interruption. Reports Spot costs approximately 30% of on-demand pricing. Recommends starting with queue workers and batch jobs before expanding to load-balanced services.
- **Key Claims**:
  - A base=2 FARGATE + weight=3 FARGATE_SPOT strategy ensures minimum on-demand availability while maximizing Spot usage [**MODERATE**]
  - SQS worker pattern naturally tolerates Spot interruptions because unacknowledged messages return to queue [**MODERATE**]
  - Spot pricing is approximately 30% of on-demand Fargate pricing [**MODERATE**]
  - Implementation priority should be: queue workers first, then batch jobs, then load-balanced services (ascending risk order) [**WEAK**]

### [SRC-009] Fargate Spot Interruptions Do Not Deregister Tasks from Target Groups (containers-roadmap #2673)
- **Authors**: AWS community reporters (GitHub issue)
- **Year**: 2025
- **Type**: blog post (community issue tracker)
- **URL/DOI**: https://github.com/aws/containers-roadmap/issues/2673
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Documents a production incident beginning September 3, 2025 where Fargate Spot interruptions caused 502 errors because tasks were not deregistered from ALB target groups before receiving SIGTERM. Multiple users confirmed the issue across different ECS services. AWS acknowledged the issue, referencing a February 2023 improvement to ECS load balancing accuracy. Issue was closed as COMPLETED on December 3, 2025, implying a fix was deployed.
- **Key Claims**:
  - Fargate Spot tasks can receive SIGTERM before or simultaneously with target group deregistration, causing 502 errors under load [**MODERATE**]
  - This deregistration timing issue affected multiple production ECS services in September-December 2025 [**MODERATE**]
  - AWS closed the issue as COMPLETED in December 2025, indicating a fix was applied [**WEAK**]

### [SRC-010] Request: Spot Advisor for Fargate-Spot (containers-roadmap #1358)
- **Authors**: AWS community (GitHub feature request)
- **Year**: 2021 (filed), still open as of 2026
- **Type**: blog post (community issue tracker)
- **URL/DOI**: https://github.com/aws/containers-roadmap/issues/1358
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Community request (49 thumbs-up) for a Fargate Spot equivalent of the EC2 Spot Instance Advisor, which provides interruption probability rates by instance type and region. Highlights a critical data gap: Fargate Spot users have no visibility into interruption probability, capacity availability by container size, or regional reliability data. No AWS response or timeline has been provided. The absence of this tool forces users to rely on anecdotal experience rather than data-driven capacity planning.
- **Key Claims**:
  - AWS provides no interruption probability data for Fargate Spot (unlike EC2 Spot Instance Advisor) [**STRONG**]
  - Users cannot make data-driven decisions about Fargate Spot capacity planning without interruption frequency data [**MODERATE**]

### [SRC-011] Maximizing Cost Efficiency on ECS Fargate: ARM Architecture and Fargate Spot Strategies
- **Authors**: suzuki0430 (DEV Community)
- **Year**: 2024
- **Type**: blog post
- **URL/DOI**: https://dev.to/suzuki0430/maximizing-cost-efficiency-on-ecs-fargate-arm-architecture-and-fargate-spot-strategies-3dff
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Documents a practical development environment configuration combining ARM and Fargate Spot. Provides a Spot-primary strategy: FARGATE_SPOT base=1, weight=2 with FARGATE base=0, weight=1. Confirms that Fargate Spot has supported ARM since September 2024. Notes that ARM migration requires multi-arch container builds (--platform linux/arm64) and task definition updates to specify ARM64 cpuArchitecture.
- **Key Claims**:
  - Fargate Spot ARM support was added in September 2024, enabling combined ARM + Spot savings [**MODERATE**]
  - Combined ARM (20% savings) + Spot (up to 70% savings) can reduce costs by up to 76% vs x86 on-demand [**WEAK**]
  - ARM migration requires multi-arch container builds and task definition cpuArchitecture specification [**MODERATE**]

### [SRC-012] Troubleshooting Amazon ECS SpotInterruption Errors
- **Authors**: AWS Documentation Team
- **Year**: 2025 (continuously updated)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/spot-interruption-errors.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Official troubleshooting guide for SpotInterruption errors. Confirms that the error occurs both when Fargate Spot capacity is initially unavailable and when previously allocated capacity is reclaimed. Primary mitigation is multi-AZ deployment to increase available capacity pool. References stopped task error inspection via the AWS Console.
- **Key Claims**:
  - SpotInterruption errors occur both on initial placement failure and on capacity reclamation [**STRONG**]
  - Multi-AZ deployment is the primary AWS-recommended mitigation for Spot capacity unavailability [**STRONG**]

### [SRC-013] Container Spot Capacity (SST)
- **Authors**: SST (framework documentation)
- **Year**: 2025
- **Type**: official documentation (framework)
- **URL/DOI**: https://sst.dev/blog/container-spot-capacity/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: SST framework documentation for Fargate Spot integration. Reports approximately 50% cost reduction (lower than AWS's claimed 70%) in their pricing examples. Notes that if Spot capacity is unavailable, service deployment fails with no automatic fallback. Explicitly recommends Spot for dev/staging environments and conditional environment-based configuration for production.
- **Key Claims**:
  - In SST's pricing examples, Spot saves approximately 50% (lower bound of the 50-70% range) [**WEAK**]
  - Fargate Spot capacity unavailability causes deployment failure with no automatic fallback to on-demand [**MODERATE**]
  - Spot is explicitly recommended for dev/PR environments, not production, by SST framework authors [**WEAK**]

## Thematic Synthesis

### Theme 1: Fargate Spot Delivers 50-70% Cost Savings but the Discount Is Variable, Not Fixed

**Consensus**: Fargate Spot provides significant cost savings, with real-world observations ranging from 50% to 70% off on-demand pricing depending on region and configuration. The "up to 70%" marketing figure represents the upper bound, not a guaranteed rate. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-006], [SRC-007], [SRC-008], [SRC-013]

**Controversy**: The exact discount magnitude varies across sources. AWS documentation says "up to 70%"; the original AWS deep dive [SRC-002] specifies "50% to 70%"; independent benchmarks in eu-west-1 [SRC-006] measured 66%; SST [SRC-013] reports approximately 50% in their examples. No source provides a mechanism to predict or query the current discount rate.
**Dissenting sources**: [SRC-013] reports ~50% savings, while [SRC-003] and [SRC-007] report 69-70%, suggesting the discount varies by region and time.

**Practical Implications**:
- Budget for 50-60% savings as a conservative planning estimate, not 70%
- The absence of a Fargate Spot price history API means savings cannot be predicted or tracked programmatically
- Combining ARM/Graviton (20% savings) with Spot (50-70% savings) yields the highest aggregate discount (up to ~76% off x86 on-demand)
- Compare against Compute Savings Plans (up to 50% with commitment) as an alternative that provides predictable pricing without interruption risk

**Evidence Strength**: STRONG (consensus on significant savings) / MODERATE (exact discount magnitude)

### Theme 2: No Automatic Fallback from Spot to On-Demand -- A Critical Architectural Gap

**Consensus**: When Fargate Spot capacity is unavailable, ECS does not automatically fall back to on-demand Fargate capacity. Tasks simply fail to launch (SERVICE_TASK_PLACEMENT_FAILURE) and the scheduler retries until Spot capacity returns. This is a fundamental architectural limitation that requires explicit mitigation. [**STRONG**]
**Sources**: [SRC-001], [SRC-004], [SRC-005], [SRC-008], [SRC-012], [SRC-013]

**Controversy**: Whether this gap is adequately addressable through architectural patterns. [SRC-005] demonstrates a working dual-service workaround (Spot primary + dormant on-demand fallback activated by Lambda), while [SRC-004] argues this complexity makes Fargate Spot unsuitable for web workloads entirely.
**Dissenting sources**: [SRC-005] argues the dual-service Lambda fallback pattern makes 100% Spot viable for production web apps, while [SRC-004] argues the complexity and failure modes make EC2 Spot fleets or standard Fargate preferable.

**Practical Implications**:
- Never configure production services with FARGATE_SPOT as the sole capacity provider without an explicit fallback mechanism
- The capacity provider strategy's base parameter should guarantee minimum on-demand tasks (e.g., base=2 on FARGATE)
- For critical services, implement the dual-service pattern from [SRC-005] or maintain a sufficient FARGATE base count
- Monitor SERVICE_TASK_PLACEMENT_FAILURE events via EventBridge to detect capacity unavailability early

**Evidence Strength**: STRONG (the gap exists) / MIXED (adequacy of workarounds)

### Theme 3: Interruption Frequency Is Unknown and Unknowable for Planning Purposes

**Consensus**: AWS provides no official data on Fargate Spot interruption rates, and unlike EC2 Spot, there is no Spot Advisor tool for Fargate. Anecdotal reports vary wildly: from "several times per hour" [SRC-004] to "zero outages in one year" [SRC-005]. Interruption frequency depends on region, AZ, time, and competing demand -- variables invisible to users. [**MODERATE**]
**Sources**: [SRC-004], [SRC-005], [SRC-007], [SRC-010], [SRC-012]

**Controversy**: The lack of data makes it impossible to settle the debate. [SRC-005] reports excellent reliability over a year, while [SRC-004] reports frequent interruptions. Both could be accurate for their respective regions and time periods.
**Dissenting sources**: [SRC-005] reports zero capacity outages in one year of 100% Spot production, while [SRC-004] reports interruptions "several times per hour or more."

**Practical Implications**:
- Do not rely on any single anecdotal report for capacity planning; interruption rates are region-specific and time-varying
- The open feature request for a Fargate Spot Advisor [SRC-010] (49 upvotes, no AWS response) confirms this is a recognized community need
- Implement interruption monitoring from day one: track SERVICE_TASK_PLACEMENT_FAILURE and SpotInterruption events to build your own baseline
- Capacity constraints are typically AZ-scoped [SRC-007]; multi-AZ deployment is the primary mitigation [SRC-012]

**Evidence Strength**: MIXED (consensus that data is absent; no consensus on actual frequency)

### Theme 4: Target Group Deregistration Timing Creates a 502 Error Window for Load-Balanced Services

**Consensus**: When Fargate Spot tasks are interrupted, there is a timing gap where the task receives SIGTERM while still registered in the ALB target group, causing the load balancer to route requests to a terminating container and generating 502 errors. This was documented as a production issue in September-December 2025. [**MODERATE**]
**Sources**: [SRC-004], [SRC-009]

**Controversy**: Whether AWS's December 2025 fix fully resolves the timing issue. [SRC-009] was closed as COMPLETED, but the fix has not been independently verified by the community in subsequent reports.
**Dissenting sources**: [SRC-009] (closed as COMPLETED) implies the issue is resolved, while [SRC-004] (2020, predating the fix) documents the same architectural concern as fundamental to how Fargate Spot interacts with ALBs.

**Practical Implications**:
- Configure aggressive target group deregistration delays (shorter deregistration_delay.timeout_seconds) to minimize the 502 window
- Implement health check failure on SIGTERM receipt so the load balancer stops routing traffic before the task terminates
- Use connection draining configuration to allow in-flight requests to complete
- Monitor 502 error rates correlated with Spot interruption events to detect if the timing issue recurs

**Evidence Strength**: MODERATE

### Theme 5: Capacity Provider Strategy Patterns Follow a Small Set of Established Configurations

**Consensus**: The community has converged on a small number of capacity provider strategy patterns for mixed fleets, all using the base/weight mechanism. [**MODERATE**]
**Sources**: [SRC-001], [SRC-002], [SRC-005], [SRC-008], [SRC-011]

**Practical Implications**:
- **Production web services**: FARGATE base=2, weight=1; FARGATE_SPOT weight=3. Guarantees 2 on-demand tasks, scales 75% on Spot.
- **Development/staging**: FARGATE_SPOT base=1, weight=2; FARGATE weight=1. Cost-optimized with minimal on-demand fallback.
- **Batch/queue workers**: FARGATE_SPOT base=0, weight=1. 100% Spot acceptable because work items are idempotent and retriable.
- **Critical path with Spot burst**: FARGATE base=N (matching steady-state load), weight=0; FARGATE_SPOT weight=1. All burst capacity on Spot, baseline fully on-demand.
- Only one capacity provider can have a non-zero base value per strategy

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Fargate Spot tasks receive a 2-minute warning via EventBridge task state change event and SIGTERM signal before termination -- Sources: [SRC-001], [SRC-002], [SRC-008], [SRC-012]
- Fargate Spot offers up to 70% discount off standard Fargate rates, with real-world savings consistently in the 50-70% range -- Sources: [SRC-001], [SRC-002], [SRC-003], [SRC-006], [SRC-007]
- ARM/Graviton processors are approximately 20% cheaper per vCPU-hour and per GB-hour than x86 on Fargate -- Sources: [SRC-003], [SRC-011]
- Capacity provider strategies use base (minimum guaranteed tasks) and weight (proportional distribution) parameters; only one provider can have a non-zero base -- Sources: [SRC-001], [SRC-002]
- AWS provides no interruption probability data for Fargate Spot, unlike EC2 Spot Instance Advisor -- Sources: [SRC-004], [SRC-010]
- ECS does not automatically fall back from Fargate Spot to on-demand Fargate when Spot capacity is unavailable -- Sources: [SRC-001], [SRC-005], [SRC-008], [SRC-012], [SRC-013]
- SpotInterruption errors occur both on initial placement failure and on capacity reclamation -- Sources: [SRC-001], [SRC-012]
- Multi-AZ deployment is the primary AWS-recommended mitigation for Spot capacity unavailability -- Sources: [SRC-001], [SRC-012]

### MODERATE Evidence
- Fargate Spot discount ranges from 50% to 70% off on-demand pricing, not a fixed discount -- Sources: [SRC-002]
- Fargate Spot tasks can receive SIGTERM before or simultaneously with target group deregistration, causing 502 errors -- Sources: [SRC-004], [SRC-009]
- The target group deregistration timing issue affected multiple production services September-December 2025 and was marked as resolved by AWS -- Sources: [SRC-009]
- Minimum 2 desired tasks per service is necessary for Spot fault tolerance -- Sources: [SRC-005], [SRC-008]
- SQS worker pattern naturally tolerates Spot interruptions because unacknowledged messages return to queue -- Sources: [SRC-008]
- Fargate Spot ARM support was added in September 2024 -- Sources: [SRC-011]
- A base=2 FARGATE + weight=3 FARGATE_SPOT strategy is the most commonly documented mixed fleet pattern -- Sources: [SRC-008], [SRC-011]
- Savings Plans (up to 50% with commitment) represent a no-interruption alternative to Spot for cost optimization -- Sources: [SRC-003]
- Capacity provider strategies cannot mix Fargate and Auto Scaling group providers -- Sources: [SRC-001]
- Fargate Spot capacity unavailability causes deployment/scaling failure with no automatic fallback -- Sources: [SRC-001], [SRC-013]

### WEAK Evidence
- Fargate Spot interruption frequency can reach several times per hour in some regions/periods -- Sources: [SRC-004]
- 100% Fargate Spot can sustain a production web application for 1+ year without capacity outages in some configurations -- Sources: [SRC-005]
- Task restart after Spot interruption occurs within seconds when capacity is available -- Sources: [SRC-005]
- Capacity constraints are typically AZ-scoped rather than region-wide -- Sources: [SRC-007]
- Fargate Spot task longevity varies widely: some tasks run 15+ days, others face frequent reclamation -- Sources: [SRC-007]
- Combined ARM + Spot can reduce costs by up to ~76% vs x86 on-demand Fargate -- Sources: [SRC-011]
- Fargate carries a 16-40% pricing premium over equivalent EC2 instance types -- Sources: [SRC-007]
- Implementation priority should be queue workers first, then batch jobs, then load-balanced services (ascending risk) -- Sources: [SRC-008]
- AWS closed the target group deregistration timing issue as COMPLETED in December 2025 -- Sources: [SRC-009]

### UNVERIFIED
- Exact current Fargate Spot discount rates by region and container size -- Basis: AWS publishes current Spot pricing but provides no historical data or forecasting; the rate could change at any time
- Long-term (multi-year) Fargate Spot interruption frequency trends -- Basis: No public dataset or AWS-published statistics exist; all available data is anecdotal from individual operators
- Whether the December 2025 fix for target group deregistration timing [SRC-009] fully resolves 502 errors during Spot interruptions -- Basis: Issue closed by AWS but no independent community verification published
- Comparative interruption rates between Fargate Spot and EC2 Spot for equivalent workloads -- Basis: No published comparison exists; the underlying capacity pools may differ
- Impact of AWS Graviton4 (announced 2025) on Fargate Spot pricing relative to Graviton2/3 -- Basis: Graviton4 pricing published for some services but Fargate Spot + Graviton4 specific data not yet confirmed

## Knowledge Gaps

- **Interruption frequency data by region/AZ/container-size**: The most critical gap. No public data exists, no Spot Advisor equivalent, and AWS has not responded to the 2021 feature request [SRC-010]. Without this data, capacity planning for Fargate Spot is fundamentally guess-based.

- **Fargate Spot vs. Compute Savings Plans break-even analysis**: While individual pricing for each is documented, a rigorous analysis of when Spot's variable discount with interruption risk outperforms a Savings Plan's guaranteed discount with commitment is absent from the literature. The Medium article by Watanabe (2025) appeared to address this but was paywalled.

- **Post-December-2025 target group deregistration behavior**: The fix referenced in [SRC-009] needs independent verification. No post-fix production reports have been published confirming that 502 errors during Spot interruptions are resolved.

- **Multi-region Fargate Spot capacity correlation**: Whether Spot capacity constraints in one region correlate with constraints in other regions (suggesting systemic AWS capacity pressure vs. localized events) has not been studied.

- **Fargate Spot behavior under sustained high-demand events**: How Fargate Spot performs during AWS-wide capacity crunches (e.g., re:Invent week, holiday shopping season) is undocumented. Operators need worst-case scenario data for capacity planning.

## Domain Calibration

Low confidence distribution reflects a domain with sparse or paywalled primary literature. Many claims could not be independently corroborated. Treat findings as starting points for manual research, not as settled knowledge.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.
5. **No peer-reviewed papers**: This domain has no peer-reviewed academic literature. All sources are vendor documentation, blog posts, and community issue trackers. Evidence tiers are calibrated accordingly -- STRONG in this review means corroboration across multiple authoritative sources, not academic peer review.

Generated by `/research AWS Fargate Spot capacity providers: production reliability data, interruption rates, capacity provider strategy patterns, mixed Fargate/Fargate-Spot fleet configurations, and real-world cost savings benchmarks as of 2025-2026` on 2026-03-25.
