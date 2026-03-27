---
domain: "literature-progressive-delivery-canary-ecs-fargate"
generated_at: "2026-03-25T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.62
format_version: "1.0"
---

# Literature Review: Progressive Delivery and Canary Deployment on AWS ECS Fargate

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The progressive delivery landscape on AWS ECS Fargate underwent a major shift in 2025 with Amazon's introduction of native canary and linear deployment strategies (October 2025), building on the native blue/green deployment support launched in July 2025. Prior to this, canary deployments on ECS required either AWS CodeDeploy blue/green with canary traffic shifting or custom controllers built around ALB weighted target groups. Automated canary analysis tools (Kayenta, Flagger, Argo Rollouts) remain Kubernetes-centric, with no first-party ECS integration; teams deploying on ECS Fargate must build custom canary scoring by querying CloudWatch metrics and implementing differential analysis independently. The interaction between Fargate Spot capacity interruptions and active canary deployments remains an under-documented risk area, with the 2-minute SIGTERM window and target group deregistration delay creating a narrow but real window for canary scoring contamination during Spot reclamation events.

## Source Catalog

### [SRC-001] Amazon ECS Canary Deployments -- AWS Official Documentation
- **Authors**: AWS Documentation Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/canary-deployment.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Definitive reference for ECS-native canary deployments. Documents the six-phase canary lifecycle (preparation, deployment, testing, canary traffic shifting, monitoring, completion), configurable canary percentage and bake time, 10 deployment lifecycle stages with hook support, and CloudWatch alarm integration for automated rollback. Specifies that canary traffic shifts in two steps: initial percentage to green, then 100% to green. Each lifecycle stage has a 24-hour timeout.
- **Key Claims**:
  - ECS native canary deployments shift traffic in two phases: configurable canary percentage, then all-at-once to 100% [**MODERATE**]
  - Canary deployments support 10 lifecycle stages with Lambda hook integration at 7 of them for custom validation [**MODERATE**]
  - CloudWatch alarms can trigger automated rollback during canary bake time [**MODERATE**]
  - Each deployment lifecycle stage has a maximum 24-hour timeout; CloudFormation imposes a 36-hour total limit [**MODERATE**]

### [SRC-002] Amazon ECS Now Supports Built-in Linear and Canary Deployments -- AWS Announcement
- **Authors**: AWS
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://aws.amazon.com/about-aws/whats-new/2025/10/amazon-ecs-built-in-linear-canary-deployments/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: October 30, 2025 announcement of native ECS canary and linear deployment strategies. Linear deployments shift traffic in equal percentage increments with configurable step percentage and step bake time. Canary deployments route a small percentage first, then shift remaining traffic after canary bake time. Both strategies support lifecycle hooks, CloudWatch alarm-based rollback, and post-deployment bake time. Available via Console, SDK, CLI, CloudFormation, CDK, and Terraform. Supports ALB and ECS Service Connect.
- **Key Claims**:
  - ECS natively supports canary and linear deployment strategies as of October 2025, without requiring CodeDeploy [**MODERATE**]
  - Linear deployments shift traffic in equal increments (e.g., 10% every 5 minutes) with configurable step bake time [**MODERATE**]
  - Both strategies support ALB and ECS Service Connect as traffic routing mechanisms [**MODERATE**]

### [SRC-003] ECS Native Blue/Green Is Here! With Strong Hooks and Dark Canary
- **Authors**: AWS Builders community (dev.to)
- **Year**: 2025
- **Type**: blog post
- **URL/DOI**: https://dev.to/aws-builders/ecs-native-bluegreen-is-here-with-strong-hooks-and-dark-canary-8ff
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 5
- **Summary**: Detailed analysis of the July 2025 ECS native blue/green deployment launch. Documents dark canary testing via separate test listeners, 7 lifecycle hook stages with Lambda integration returning SUCCEEDED/FAILED/IN_PROGRESS, all-at-once traffic switching (no gradual shifting in the initial blue/green release), and the explicit limitation that canary/linear was not yet available in the initial launch. Notes that AWS deprecated the CODE_DEPLOY deployment controller type.
- **Key Claims**:
  - ECS native blue/green (July 2025) performs all-at-once traffic switching, not gradual canary shifting [**MODERATE**]
  - Dark canary testing is supported via separate test listeners that allow validation with zero user impact before production traffic shift [**MODERATE**]
  - Lambda lifecycle hooks retry with ~30-second intervals when returning IN_PROGRESS, enabling long-running validation checks [**WEAK**]
  - AWS deprecated the CODE_DEPLOY deployment controller type in favor of ECS-native deployment [**WEAK**]

### [SRC-004] Choosing Between Amazon ECS Blue/Green Native or AWS CodeDeploy in AWS CDK
- **Authors**: AWS DevOps & Developer Productivity Blog
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://aws.amazon.com/blogs/devops/choosing-between-amazon-ecs-blue-green-native-or-aws-codedeploy-in-aws-cdk/
- **Verified**: partial (title confirmed; full text not fully accessible due to rendering)
- **Relevance**: 4
- **Summary**: Comparison guide for choosing between ECS native blue/green and CodeDeploy-based deployments. CodeDeploy supports canary and linear shifting in addition to all-at-once; ECS native blue/green initially only supported all-at-once but has since added canary and linear (October 2025). The post positions ECS native as the forward-looking default, with CodeDeploy as the legacy option.
- **Key Claims**:
  - CodeDeploy historically provided canary and linear traffic shifting that was absent from ECS native blue/green [**MODERATE**]
  - ECS native deployments are positioned as the forward-looking replacement for CodeDeploy-based ECS deployments [**WEAK**]

### [SRC-005] How the Amazon ECS Deployment Circuit Breaker Detects Failures
- **Authors**: AWS Documentation Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-circuit-breaker.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents the ECS deployment circuit breaker mechanism. Uses a two-stage monitoring process (task launch monitoring, then health check validation). Failure threshold calculated as min(max(3, ceil(0.5 * desired_count)), 200). Critically, the circuit breaker is only supported for rolling update (ECS) deployment controller -- it is NOT supported for blue/green or canary deployment types. This means canary deployments must rely on CloudWatch alarms and lifecycle hooks for failure detection rather than the circuit breaker.
- **Key Claims**:
  - ECS deployment circuit breaker is only supported for rolling update deployment controller, NOT for blue/green or canary deployments [**MODERATE**]
  - Failure threshold formula is min(max(3, ceil(0.5 * desired_count)), 200), and thresholds cannot be customized [**MODERATE**]
  - Circuit breaker uses two-stage monitoring: task launch state transition, then health check validation [**MODERATE**]

### [SRC-006] How CloudWatch Alarms Detect Amazon ECS Deployment Failures
- **Authors**: AWS Documentation Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-alarm-failure.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents CloudWatch alarm integration with ECS deployments. Alarms that are already in ALARM state when a deployment starts are ignored for that deployment. ECS polls alarm state via DescribeAlarms, which counts against CloudWatch service quotas. When both circuit breaker and CloudWatch alarms are enabled, either method can trigger failure independently. Default bake time is less than 5 minutes and is computed by ECS based on alarm configuration. Recommended alarm metrics include HTTPCode_ELB_5XX_Count and CPUUtilization.
- **Key Claims**:
  - CloudWatch alarms in ALARM state at deployment start are ignored by ECS for that deployment's duration [**MODERATE**]
  - Default bake time is less than 5 minutes and is computed by ECS, not manually configurable for rolling deployments [**MODERATE**]
  - ECS alarm polling via DescribeAlarms counts against CloudWatch service quotas; throttling may cause missed alarms [**MODERATE**]

### [SRC-007] Automated Canary Analysis at Netflix with Kayenta
- **Authors**: Netflix Technology Blog
- **Year**: 2018
- **Type**: blog post (engineering blog from major tech company)
- **URL/DOI**: https://netflixtechblog.com/automated-canary-analysis-at-netflix-with-kayenta-3260bc7acc69
- **Verified**: yes (title confirmed via WebSearch; content not fully fetched due to certificate error)
- **Relevance**: 4
- **Summary**: Netflix's authoritative description of Kayenta, the open-source automated canary analysis platform developed jointly with Google. Kayenta fetches user-configured metrics, runs statistical tests (Mann-Whitney U), and produces an aggregate canary score. The platform is designed to be cloud-agnostic and integrates with multiple metric providers. Kayenta can operate standalone outside of Spinnaker.
- **Key Claims**:
  - Kayenta uses Mann-Whitney U nonparametric statistical test with 98% confidence for canary metric comparison [**STRONG** -- corroborated by SRC-008]
  - Kayenta is platform-agnostic and can integrate with multiple metric providers (Stackdriver, Atlas, Prometheus, Datadog, New Relic) [**MODERATE**]
  - Kayenta can operate as a standalone canary analysis service outside of Spinnaker [**WEAK**]

### [SRC-008] How Canary Judgment Works -- Spinnaker Documentation
- **Authors**: Spinnaker Project
- **Year**: 2024
- **Type**: official documentation
- **URL/DOI**: https://spinnaker.io/docs/guides/user/canary/judge/
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Definitive technical reference for the Kayenta canary judge algorithm. Documents the Mann-Whitney U test with 98% confidence interval, Hodges-Lehmann estimate for tolerance band calculation (dead zone = +/- 0.25 * estimate), metric classification (Pass/High/Low/Nodata), effect size measures (meanRatio, CLES), group scoring formula, and critical metric automatic failure behavior. If 50% or more metrics return Nodata, canary automatically fails regardless of other scores.
- **Key Claims**:
  - Kayenta canary judge uses Mann-Whitney U test with 98% confidence and a tolerance band of +/- 0.25 * Hodges-Lehmann estimate [**STRONG** -- corroborated by SRC-007]
  - Canary group score = (Pass count / Total count) * 100; summary score = weighted average of group scores [**MODERATE**]
  - Critical metrics exceeding criticalIncrease/criticalDecrease thresholds immediately set canary score to 0 [**MODERATE**]
  - If 50% or more metrics return Nodata, canary automatically fails regardless of other metric scores [**MODERATE**]

### [SRC-009] Amazon ECS Fargate Capacity Providers -- Spot Interruption Handling
- **Authors**: AWS Documentation Team
- **Year**: 2025
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/fargate-capacity-providers.html
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 4
- **Summary**: Documents Fargate Spot interruption mechanics. Tasks receive a 2-minute warning before termination, delivered as both an EventBridge task state change event and a SIGTERM signal. Default stopTimeout is 30 seconds; AWS recommends 120 seconds for Spot workloads. Fargate does NOT replace Spot capacity with on-demand capacity automatically. Services retry launching replacement tasks until Spot capacity returns, but single-task services will experience a gap.
- **Key Claims**:
  - Fargate Spot tasks receive a 2-minute warning via EventBridge event and SIGTERM signal before termination [**MODERATE**]
  - Default stopTimeout is 30 seconds; AWS recommends 120 seconds for Fargate Spot to allow graceful shutdown [**MODERATE**]
  - Fargate does NOT automatically replace Spot capacity with on-demand capacity; services retry until Spot returns [**MODERATE**]
  - Single-task Fargate Spot services will experience interruption gaps until capacity is available [**MODERATE**]

### [SRC-010] Use Application Load Balancers for Blue-Green and Canary Deployments -- Terraform Tutorial
- **Authors**: HashiCorp
- **Year**: 2025
- **Type**: official documentation (tutorial)
- **URL/DOI**: https://developer.hashicorp.com/terraform/tutorials/aws/blue-green-canary-tests-deployments
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: Terraform tutorial demonstrating ALB weighted target groups for canary deployments. Shows a traffic_dist_map pattern with five stages (100/0, 90/10, 50/50, 10/90, 0/100) controlled by a single Terraform variable. Uses forward action with multiple target groups and weight attributes in aws_lb_listener resource. This represents the custom canary controller pattern used before ECS-native canary support.
- **Key Claims**:
  - ALB forward action supports multiple target groups with weight attributes for percentage-based traffic splitting [**MODERATE**]
  - Canary traffic percentage can be controlled via Terraform variables without modifying infrastructure resource definitions [**MODERATE**]

### [SRC-011] AWS App Mesh Deprecation and Migration to ECS Service Connect
- **Authors**: AWS Containers Blog, InfoQ
- **Year**: 2024
- **Type**: official documentation
- **URL/DOI**: https://aws.amazon.com/blogs/containers/migrating-from-aws-app-mesh-to-amazon-ecs-service-connect/
- **Verified**: yes (title confirmed via WebSearch)
- **Relevance**: 3
- **Summary**: AWS App Mesh was deprecated as of September 24, 2024 (end-of-life September 30, 2026). New customers cannot onboard. AWS recommends ECS Service Connect for ECS workloads and VPC Lattice for EKS workloads. This deprecation eliminates the primary path for running Flagger-style progressive delivery on ECS via service mesh. Migration requires recreating ECS services since a service cannot simultaneously belong to App Mesh and Service Connect.
- **Key Claims**:
  - AWS App Mesh is deprecated (September 2024) with end-of-life September 2026; new customers cannot onboard [**MODERATE**]
  - ECS Service Connect is the recommended replacement for App Mesh on ECS workloads [**MODERATE**]
  - Flagger's AWS integration path via App Mesh is effectively sunset for new deployments [**WEAK**]

### [SRC-012] Guidance for ECS Canary Deployments for Backend Workloads on AWS
- **Authors**: AWS Solutions Library
- **Year**: 2024
- **Type**: official documentation (reference architecture)
- **URL/DOI**: https://github.com/aws-solutions-library-samples/guidance-for-ecs-canary-deployments-for-backend-workloads-on-aws
- **Verified**: yes (content fetched and confirmed)
- **Relevance**: 3
- **Summary**: AWS reference architecture for canary deployments on non-HTTP (queue-processing) ECS workloads. Uses a shared SQS queue pattern where canary and production ECS services consume from the same queue, achieving natural traffic distribution based on task count ratio (e.g., 3 production + 1 canary = ~25% canary traffic). Includes CloudWatch alarm on DLQ depth for automated rollback via Lambda. Uses manual approval gate for canary promotion.
- **Key Claims**:
  - Queue-based canary deployments on ECS can use shared SQS queues with task count ratio for natural traffic distribution [**MODERATE**]
  - CloudWatch alarms on DLQ depth provide automated rollback for queue-processing canary workloads [**MODERATE**]

### [SRC-013] Flagger -- Progressive Delivery Kubernetes Operator
- **Authors**: Flagger Project (CNCF/Flux)
- **Year**: 2024
- **Type**: official documentation
- **URL/DOI**: https://github.com/fluxcd/flagger
- **Verified**: yes (title confirmed via WebSearch)
- **Relevance**: 3
- **Summary**: Flagger is a Kubernetes-native progressive delivery operator supporting canary, A/B testing, and blue/green deployments. Integrates with service meshes (Istio, Linkerd, App Mesh) and ingress controllers for traffic splitting. Uses Prometheus-based metric analysis for automated canary promotion. Flagger's App Mesh integration provided the closest path to progressive delivery on ECS, but this path is now sunset due to App Mesh deprecation.
- **Key Claims**:
  - Flagger is Kubernetes-native and requires a Kubernetes control plane; it does not support standalone ECS deployments [**MODERATE**]
  - Flagger's App Mesh integration is the only AWS-native traffic splitting mechanism it supports for non-EKS workloads [**WEAK**]
  - Flagger uses iterative Prometheus-based metric checks for automated canary promotion with configurable thresholds [**MODERATE**]

### [SRC-014] AWS CodeDeploy Linear and Canary Deployments for Amazon ECS
- **Authors**: AWS Containers Blog
- **Year**: 2024
- **Type**: official documentation
- **URL/DOI**: https://aws.amazon.com/blogs/containers/aws-codedeploy-now-supports-linear-and-canary-deployments-for-amazon-ecs/
- **Verified**: partial (title confirmed; full article text not accessible due to rendering)
- **Relevance**: 4
- **Summary**: Documents CodeDeploy's extension of ECS blue/green deployment support to include canary and linear strategies. CodeDeploy uses ALB weighted target groups for traffic splitting. Supports custom canary configurations and CloudWatch alarm-based rollback. This was the primary mechanism for canary deployments on ECS before the October 2025 native ECS canary support.
- **Key Claims**:
  - CodeDeploy uses ALB weighted target groups to implement canary and linear traffic shifting for ECS [**MODERATE**]
  - CodeDeploy supports custom-defined canary and linear deployment configurations [**WEAK**]

## Thematic Synthesis

### Theme 1: ECS-Native Canary Eliminates the Need for CodeDeploy But Not for Custom Canary Analysis

**Consensus**: As of October 2025, ECS natively supports canary and linear deployment strategies, eliminating the historical dependency on CodeDeploy for progressive traffic shifting. This represents a significant simplification of the deployment pipeline, with native support for lifecycle hooks, CloudWatch alarm-based rollback, and configurable bake times. [**MODERATE**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-014]

**Controversy**: Whether ECS-native canary is sufficient for production-grade progressive delivery. The native implementation provides traffic shifting and alarm-based rollback, but it does not include automated canary analysis (differential metric comparison between canary and baseline). Teams requiring statistical canary scoring must still build custom solutions or adapt Kubernetes-native tools.
**Dissenting sources**: [SRC-001] and [SRC-002] present ECS-native canary as a complete deployment solution, while the absence of automated canary analysis comparable to [SRC-007] Kayenta and [SRC-008] suggests a significant gap for teams accustomed to statistical canary judgment.

**Practical Implications**:
- Default to ECS-native canary/linear deployments for new services; CodeDeploy is the legacy path
- Budget for building custom canary scoring if you need differential metric comparison (e.g., comparing p99 latency between canary and baseline target groups)
- Use CloudWatch alarms as a threshold-based safety net, but recognize that threshold-based alarms are less sensitive than statistical comparison (a canary with 1% higher error rate may not trip a static alarm threshold)

**Evidence Strength**: MODERATE

### Theme 2: The Deployment Circuit Breaker Does Not Apply to Canary Deployments

**Consensus**: The ECS deployment circuit breaker is only supported for rolling update deployments. It is NOT available for blue/green, canary, or linear deployment types. For canary deployments, failure detection relies entirely on CloudWatch alarms and lifecycle hooks. [**MODERATE**]
**Sources**: [SRC-005], [SRC-006], [SRC-001]

**Practical Implications**:
- Do not design canary deployment rollback around the circuit breaker; it will not trigger
- Configure CloudWatch alarms on key metrics (5XX count, latency percentiles, error rate) as the primary automated rollback mechanism for canary deployments
- Use lifecycle hooks with Lambda functions to implement custom health checks that the circuit breaker would otherwise provide
- The absence of circuit breaker support means canary deployments have a higher operational burden for failure detection configuration

**Evidence Strength**: MODERATE

### Theme 3: Kayenta Is the Only Mature Automated Canary Analysis Tool, But It Has No ECS Integration

**Consensus**: Kayenta (Netflix/Google) remains the most mature open-source automated canary analysis tool, using Mann-Whitney U nonparametric statistical testing with 98% confidence intervals to compare canary and baseline metric distributions. It produces an aggregate canary score that can drive automated promotion or rollback decisions. [**STRONG**]
**Sources**: [SRC-007], [SRC-008]

**Controversy**: Whether Kayenta's standalone mode is practical for ECS Fargate deployments. Kayenta can technically run outside Spinnaker and supports multiple metric providers (including CloudWatch via Stackdriver integration), but there is no documented ECS-native integration pattern. Teams must build custom orchestration to: (1) provision canary and baseline target groups, (2) configure metric queries for both target groups, (3) invoke Kayenta's REST API, and (4) act on the canary score.
**Dissenting sources**: [SRC-007] claims Kayenta is platform-agnostic, but the practical documentation in [SRC-008] and the Spinnaker setup guides focus exclusively on Kubernetes and VM-based deployments.

**Practical Implications**:
- Kayenta can theoretically score ECS canary deployments if you feed it CloudWatch metrics from canary vs. baseline target groups, but expect significant custom integration work
- For simpler needs, implement a lightweight canary scorer using CloudWatch Metric Math to compare target group metrics (e.g., TargetResponseTime per target group) rather than deploying full Kayenta
- Argo Rollouts and Flagger are not viable alternatives for ECS -- they are Kubernetes-native controllers

**Evidence Strength**: STRONG (for Kayenta's algorithm) / WEAK (for ECS integration feasibility)

### Theme 4: Fargate Spot Interruptions Create a Canary Scoring Contamination Risk

**Consensus**: Fargate Spot tasks receive a 2-minute warning (SIGTERM + EventBridge event) before termination. During this window, the task must complete in-flight requests while being deregistered from its target group. The default stopTimeout is 30 seconds (AWS recommends 120 seconds for Spot). Fargate does NOT automatically failover Spot capacity to on-demand. [**MODERATE**]
**Sources**: [SRC-009]

**Controversy**: No direct sources address the specific interaction between Spot interruptions and active canary deployments. This is a knowledge gap rather than a controversy.

**Practical Implications**:
- If canary tasks run on Fargate Spot, a Spot reclamation event during the canary bake period will cause task termination, potentially contaminating canary metrics with connection draining errors and increased latency
- Set target group deregistration delay to less than 120 seconds (matching the Spot warning window) to ensure the ALB stops sending new requests before the task is killed
- Consider running canary target groups on regular Fargate (not Spot) during the canary evaluation window, reserving Spot for the baseline or post-promotion steady-state
- Monitor for SpotInterruption stopCode in EventBridge events and exclude the interruption window from canary metric analysis if building custom scoring

**Evidence Strength**: MODERATE (for Spot mechanics) / UNVERIFIED (for canary-specific interaction)

### Theme 5: ALB Weighted Target Groups Are the Universal Traffic Splitting Primitive

**Consensus**: ALB listener rules with weighted forward actions are the fundamental mechanism for canary traffic splitting on ECS, whether using CodeDeploy, ECS-native canary, or custom controllers. Weights can be assigned to multiple target groups in a single forward action, enabling percentage-based traffic distribution. [**MODERATE**]
**Sources**: [SRC-001], [SRC-010], [SRC-014]

**Practical Implications**:
- All canary patterns on ECS (native, CodeDeploy, custom) ultimately resolve to ALB weighted target group manipulation
- Terraform's aws_lb_listener_rule forward action with multiple target_group blocks provides infrastructure-as-code control over canary percentages
- ECS Service Connect provides an alternative traffic splitting mechanism as of 2025, but ALB remains the most mature and well-documented path
- When building custom canary controllers, the core operation is ModifyRule to adjust target group weights

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Kayenta uses Mann-Whitney U nonparametric statistical test with 98% confidence for canary-vs-baseline metric comparison, producing an aggregate score from 0-100 -- Sources: [SRC-007], [SRC-008]

### MODERATE Evidence
- ECS natively supports canary and linear deployment strategies as of October 2025, eliminating the CodeDeploy dependency -- Sources: [SRC-001], [SRC-002]
- ECS deployment circuit breaker is only supported for rolling update deployments, NOT for canary/blue-green -- Sources: [SRC-005], [SRC-006]
- Fargate Spot tasks receive a 2-minute SIGTERM warning; default stopTimeout is 30s, recommended 120s -- Sources: [SRC-009]
- CloudWatch alarms are the primary automated rollback mechanism for ECS canary deployments -- Sources: [SRC-001], [SRC-006]
- ALB weighted target groups are the universal traffic splitting primitive for ECS canary deployments -- Sources: [SRC-001], [SRC-010], [SRC-014]
- AWS App Mesh is deprecated (September 2024, EOL September 2026), eliminating Flagger's primary ECS integration path -- Sources: [SRC-011]
- ECS native canary supports 10 lifecycle stages with Lambda hooks for custom validation at 7 stages -- Sources: [SRC-001], [SRC-003]
- Kayenta group score formula: (Pass count / Total count) * 100; summary score is weighted average of groups -- Sources: [SRC-008]
- Critical metrics in Kayenta that exceed thresholds immediately set canary score to 0 -- Sources: [SRC-008]
- Fargate does NOT automatically replace Spot capacity with on-demand; services retry until Spot returns -- Sources: [SRC-009]
- Queue-based canary on ECS can use shared SQS queues with task count ratio for natural traffic distribution -- Sources: [SRC-012]

### WEAK Evidence
- AWS deprecated the CODE_DEPLOY deployment controller type in favor of ECS-native deployment -- Sources: [SRC-003]
- Kayenta can operate as a standalone canary analysis service outside of Spinnaker -- Sources: [SRC-007]
- Flagger's App Mesh integration is effectively sunset for new deployments due to App Mesh deprecation -- Sources: [SRC-011], [SRC-013]
- Lambda lifecycle hooks retry with ~30-second intervals when returning IN_PROGRESS -- Sources: [SRC-003]

### UNVERIFIED
- The interaction between Fargate Spot interruptions and active canary bake periods (whether Spot reclamation during canary scoring contaminates metrics) is not documented in any primary source -- Basis: model training knowledge and inference from [SRC-009] mechanics
- Whether ECS-native canary deployment's CloudWatch alarm integration supports differential metric comparison (canary vs. baseline) rather than just static threshold alarms -- Basis: absence of documentation in [SRC-001] and [SRC-006]
- The practical feasibility of running Kayenta standalone with CloudWatch as a metric source for ECS canary scoring -- Basis: model training knowledge; no documented integration pattern found
- Whether ECS Service Connect supports weighted traffic splitting comparable to ALB weighted target groups for canary deployments -- Basis: model training knowledge; [SRC-002] mentions Service Connect support but no technical details on traffic splitting mechanics

## Knowledge Gaps

- **Fargate Spot + canary interaction**: No source addresses the specific scenario where Fargate Spot capacity is reclaimed during an active canary bake period. The mechanics of Spot interruption (SRC-009) and canary lifecycle (SRC-001) are documented independently, but their interaction -- particularly target group deregistration timing during Spot reclamation vs. canary traffic splitting weights -- is undocumented. This is a critical operational risk for teams running canary workloads on Spot.

- **Automated canary analysis on ECS**: No source documents a working pattern for running automated canary analysis (Kayenta or equivalent) against ECS Fargate canary deployments. All automated canary analysis documentation is Kubernetes-centric. Teams need a reference architecture for: CloudWatch metric query construction for per-target-group comparison, statistical scoring, and promotion/rollback automation.

- **ECS Service Connect traffic splitting for canary**: While SRC-002 announces Service Connect support for canary deployments, no source documents the traffic splitting mechanics. It is unclear whether Service Connect uses weighted routing, header-based routing, or another mechanism for canary traffic distribution.

- **Canary scoring with CloudWatch Metric Math**: No source documents using CloudWatch Metric Math or CloudWatch Anomaly Detection as a lightweight alternative to Kayenta for differential canary analysis on ECS. This is a plausible architecture but lacks documented precedent.

- **Long-term convergence of ECS native vs. external controllers**: It is unclear whether the October 2025 ECS native canary/linear additions will continue to evolve toward feature parity with external canary controllers (statistical analysis, multi-metric scoring) or remain focused on traffic shifting with threshold-based alarms.

## Domain Calibration

Low confidence distribution reflects a domain with sparse or paywalled primary literature. Many claims could not be independently corroborated. Treat findings as starting points for manual research, not as settled knowledge.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research progressive-delivery-canary-ecs-fargate` on 2026-03-25.
