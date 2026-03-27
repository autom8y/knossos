---
domain: "literature-progressive-delivery"
generated_at: "2026-03-06T18:45:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.75
format_version: "1.0"
---

# Literature Review: Progressive Delivery, Canary Analysis, ECS & Lambda

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

Progressive delivery -- the practice of gradually exposing new software versions to increasing fractions of production traffic while continuously evaluating health metrics -- has become the dominant deployment safety paradigm. The literature converges on the Mann-Whitney U test (as implemented in Netflix/Google's Kayenta) as the standard statistical method for automated canary scoring, with a 98% confidence threshold. AWS now provides native progressive delivery primitives for both ECS (built-in canary/linear since October 2025) and Lambda (alias traffic shifting with CodeDeploy), eliminating the need for third-party controllers in AWS-native stacks. Key controversy persists around rollback timing: industry practice ranges from 90 seconds (Flagger with aggressive thresholds) to 30+ minutes (AWS CodeDeploy canary configurations), with the optimal window depending on traffic volume and metric pipeline latency. False positive rates in automated rollback systems are a documented problem, with one production case study showing 80% reduction (5/week to 1/week) by switching from aggregate error rates to canary-vs-baseline differential metrics.

## Source Catalog

### [SRC-001] Google SRE Workbook: Canarying Releases
- **Authors**: Google SRE Team (Betsy Beyer et al.)
- **Year**: 2018
- **Type**: official documentation (Google SRE Workbook, Chapter 16)
- **URL/DOI**: https://sre.google/workbook/canarying-releases/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Defines canarying as "a partial and time-limited deployment of a change in a service and its evaluation." Provides the foundational framework for canary population sizing, metric selection, false positive/negative trade-offs, and error budget impact. Demonstrates that a 5% canary population with 20% error rate yields only 1% overall error rate, preserving error budgets proportionally.
- **Key Claims**:
  - A canary population of 5% experiencing 20% errors results in only 1% overall error rate -- blast radius is proportional to canary size [**STRONG**]
  - Detection and rollback take approximately the same time whether using naive deployment or canary deployment, but canaries provide the information at much lower cost to the system [**STRONG**]
  - Overly strict acceptance criteria produce false positives; overly loose criteria allow bad deployments through undetected [**STRONG**]
  - Before/after metric evaluation is "risky" because time is "one of the biggest sources of change in observed metrics" -- control vs. canary comparison is preferable [**STRONG**]
  - Effective canary metrics should number "perhaps no more than a dozen" to prevent diminishing returns [**MODERATE**]
  - Only one canary deployment should run at a time to avoid signal contamination [**MODERATE**]
  - Canary duration must span at least "the duration of a single work unit" in asynchronous systems [**MODERATE**]

### [SRC-002] Spinnaker Canary Judge (NetflixACAJudge) Documentation
- **Authors**: Netflix / Spinnaker Contributors
- **Year**: 2018-2025 (continuously maintained)
- **Type**: official documentation
- **URL/DOI**: https://spinnaker.io/docs/guides/user/canary/judge/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Documents the statistical engine behind Netflix's automated canary analysis. The judge uses the Mann-Whitney U test at 98% confidence with a tolerance band of 0.25x the Hodges-Lehmann estimate. Metrics are classified as Pass/High/Low/Nodata/Error, and group scores are computed as (Pass count / Total count) x 100. Critical metrics that breach thresholds immediately set the canary score to 0.
- **Key Claims**:
  - The Mann-Whitney U test at 98% confidence is the core statistical method for canary metric comparison [**STRONG**]
  - Effect size thresholds (meanRatio or CLES) are secondary gates -- statistical significance must be established first [**STRONG**]
  - Critical metrics that exceed effect size thresholds immediately set the canary score to 0, bypassing weighted scoring [**MODERATE**]
  - The "50% NODATA rule" triggers automatic failure when half or more metrics lack data [**MODERATE**]
  - IQR-based outlier removal using Tukey fences with a default factor of 3.0 plus 1st/99th percentile bounds is applied before comparison [**MODERATE**]
  - Tiny Gaussian noise is added to degenerate distributions (single unique value) to enable statistical testing [**WEAK**]

### [SRC-003] Amazon ECS Built-in Linear and Canary Deployments
- **Authors**: AWS (Kevin Gibbs, Mike Rizzo -- re:Invent 2025 presenters)
- **Year**: 2025
- **Type**: official documentation + conference talk (re:Invent 2025 CNS315)
- **URL/DOI**: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/canary-deployment.html
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Documents ECS-native canary deployments announced October 30, 2025. Replaces the previous CodeDeploy dependency with built-in deployment controller. Supports weighted target group routing via ALB and Service Connect, lifecycle hooks at 8 validation points, CloudWatch alarm-triggered rollback, and configurable bake times. Canary percentage configurable from 0.1% to 99.9%.
- **Key Claims**:
  - ECS now supports native canary deployments without CodeDeploy, using weighted ALB listener rules for traffic shifting [**STRONG**]
  - Lifecycle hooks are available at 8 stages (PRE_SCALE_UP through POST_PRODUCTION_TRAFFIC_SHIFT), each supporting Lambda-based validation [**STRONG**]
  - CloudWatch alarms automatically trigger rollback during canary and bake time phases [**STRONG**]
  - Advanced deployments maintain double task count during deployment, enabling faster rollback than rolling deployments [**MODERATE**]
  - NLB is limited to blue-green only -- no canary/linear support due to Layer 4 operation preventing path-based routing [**MODERATE**]
  - Service Connect identifies test traffic via custom headers (default: X-Amazon-ECS-blue-green-test) [**MODERATE**]
  - Each lifecycle stage can last up to 24 hours; entire CloudFormation deployment limited to 36 hours [**WEAK**]

### [SRC-004] AWS Lambda Alias Traffic Shifting Documentation
- **Authors**: AWS
- **Year**: 2024-2025 (continuously maintained)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/lambda/latest/dg/configuring-alias-routing.html
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Documents Lambda's weighted alias mechanism for canary deployments. An alias can point to a maximum of two published function versions with configurable traffic weights. Lambda uses a probabilistic model for distribution, meaning low-traffic functions may see significant variance from configured percentages. Integrates with CodeDeploy via SAM's AutoPublishAlias and DeploymentPreference for automated progressive delivery.
- **Key Claims**:
  - Lambda aliases support weighted routing across exactly two published function versions [**STRONG**]
  - Lambda uses a probabilistic model for traffic distribution -- at low traffic levels, actual percentages may vary significantly from configured weights [**STRONG**]
  - Both versions must share the same execution role and DLQ configuration; $LATEST cannot be used [**MODERATE**]
  - SAM's AutoPublishAlias + DeploymentPreference automates CodeDeploy integration for Lambda canary/linear deployments [**MODERATE**]
  - Pre/post-traffic hook functions enable validation before and after traffic shifts, with automatic rollback on failure [**MODERATE**]

### [SRC-005] AWS CodeDeploy Deployment Configurations
- **Authors**: AWS
- **Year**: 2024-2025 (continuously maintained)
- **Type**: official documentation
- **URL/DOI**: https://docs.aws.amazon.com/codedeploy/latest/userguide/deployment-configurations.html
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Enumerates all predefined CodeDeploy deployment configurations for Lambda and ECS. Lambda offers canary windows of 5, 10, 15, and 30 minutes at 10% initial traffic. ECS offers canary windows of 5 and 15 minutes at 10%. Linear options shift 10% per interval (1-10 minutes). Custom configurations can be created for non-standard windows.
- **Key Claims**:
  - Lambda canary configurations support 10% initial traffic with evaluation windows of 5, 10, 15, or 30 minutes before full shift [**STRONG**]
  - ECS canary configurations support 10% initial traffic with 5 or 15 minute evaluation windows [**STRONG**]
  - Linear configurations shift 10% every 1, 2, 3, or 10 minutes for Lambda; every 1 or 3 minutes for ECS [**STRONG**]
  - NLB-backed ECS deployments only support AllAtOnce -- no canary or linear [**MODERATE**]
  - Custom deployment configurations can be created for non-standard canary/linear percentages and intervals [**MODERATE**]

### [SRC-006] Flagger: Deployment Strategies and Metrics Analysis
- **Authors**: Flagger / Flux CD Contributors
- **Year**: 2020-2026 (continuously maintained)
- **Type**: official documentation
- **URL/DOI**: https://docs.flagger.app/usage/deployment-strategies / https://docs.flagger.app/usage/metrics
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Documents Flagger's progressive delivery control loop. Provides two builtin metrics (request success rate, request P99 duration) with configurable thresholds. The analysis loop runs at a configurable interval; rollback occurs within interval x threshold duration. Supports 13+ metric providers via MetricTemplate CRDs. Non-linear weight progression via stepWeights arrays enables custom traffic ramp profiles. SkipAnalysis mode available for emergency deployments.
- **Key Claims**:
  - Flagger's minimum validation duration is interval x (maxWeight / stepWeight); rollback timeframe is interval x threshold [**STRONG**]
  - Builtin metrics are HTTP request success rate (non-5xx percentage) and P99 request duration [**STRONG**]
  - MetricTemplate CRDs support Prometheus, Datadog, CloudWatch, New Relic, Graphite, Stackdriver, InfluxDB, Dynatrace, Splunk, and Keptn [**STRONG**]
  - Non-linear weight progression via stepWeights array (e.g., [1, 2, 10, 80]) enables custom traffic ramp profiles [**MODERATE**]
  - Session affinity via cookies ensures users consistently route to their assigned version during canary [**MODERATE**]
  - skipAnalysis mode bypasses metrics validation for emergency promotions [**WEAK**]

### [SRC-007] Argo Rollouts: Analysis & Progressive Delivery
- **Authors**: Argo Project Contributors
- **Year**: 2020-2026 (continuously maintained)
- **Type**: official documentation
- **URL/DOI**: https://argo-rollouts.readthedocs.io/en/stable/features/analysis/
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Documents Argo Rollouts' AnalysisTemplate system for metric-driven deployment decisions. Supports background analysis (non-blocking, runs alongside traffic shifts) and inline analysis (blocking, gates progression). Metrics can be evaluated with successCondition/failureCondition expressions, with consecutiveSuccessLimit (since v1.8) for requiring sustained metric health. Dry-run mode enables observational analysis without affecting rollout outcomes.
- **Key Claims**:
  - Background analysis runs continuously alongside canary steps without blocking progression; inline analysis blocks until completion [**STRONG**]
  - AnalysisTemplate supports successCondition and failureCondition expressions for flexible metric evaluation [**STRONG**]
  - consecutiveSuccessLimit (v1.8+) requires N consecutive successes for overall analysis success [**MODERATE**]
  - Dry-run mode executes metrics without affecting rollout outcomes -- useful for validating new metrics before enforcing them [**MODERATE**]
  - Post-promotion analysis failure in blue-green mode triggers automatic rollback to the previous version [**MODERATE**]

### [SRC-008] Canary Deployment with Automated Rollback at Headout
- **Authors**: Headout Engineering Team
- **Year**: 2024
- **Type**: blog post (engineering blog)
- **URL/DOI**: https://www.headout.studio/canary-deployment-with-automated-rollback/
- **Verified**: yes
- **Relevance**: 5
- **Summary**: Production case study documenting the transition from aggregate error rate monitoring to canary-vs-baseline error differential. Initial approach using total error rate produced 5 false positive rollbacks per week due to confounding factors (spam attacks, misconfigured logs). Switching to error difference between canary and stable pods reduced false positives by 80% to 1 per week. Uses Argo Rollouts with New Relic integration and 2-minute pause between routing stages.
- **Key Claims**:
  - Aggregate error rate monitoring produced 5 false positive rollbacks per week in production [**MODERATE**]
  - Switching to canary-vs-stable error differential reduced false positives by 80% (5/week to 1/week) [**MODERATE**]
  - Total error rate was confounded by spam attacks, misconfigured error logs, and infrastructure issues unrelated to code changes [**MODERATE**]
  - 2-minute pause between routing stages is needed to collect sufficient metric data in New Relic [**WEAK**]
  - Uniform traffic distribution across pods is a prerequisite for reliable error comparison [**WEAK**]

### [SRC-009] Canary Metrics Analysis: Building Statistical Analysis Systems
- **Authors**: OneUptime Engineering Team
- **Year**: 2026
- **Type**: blog post (technical guide)
- **URL/DOI**: https://oneuptime.com/blog/post/2026-01-30-canary-metrics-analysis/view
- **Verified**: yes
- **Relevance**: 4
- **Summary**: Technical guide for building canary metric analysis systems from scratch. Recommends the Mann-Whitney U test for metric comparison combined with Cohen's d for effect size measurement. Advocates comparing P50/P95/P99 latency percentiles rather than averages, and using weighted scoring across multiple metric categories with configurable pass thresholds. Emphasizes warmup periods, minimum sample sizes, and consecutive failure thresholds as production safeguards.
- **Key Claims**:
  - The Mann-Whitney U test combined with Cohen's d effect size provides robust canary metric comparison without normality assumptions [**MODERATE**]
  - Latency analysis should use P50/P95/P99 percentiles rather than averages to detect tail latency regressions [**MODERATE**]
  - Warmup periods, minimum sample size requirements, and consecutive failure thresholds are essential production safeguards against premature rollback decisions [**MODERATE**]
  - Weighted scoring across metric categories (latency, errors, throughput, resources) with configurable pass thresholds produces reliable composite scores [**WEAK**]

### [SRC-010] Kayenta: Open Automated Canary Analysis Tool (Google Cloud Blog)
- **Authors**: Google Cloud / Netflix (joint announcement)
- **Year**: 2018
- **Type**: blog post (official product blog)
- **URL/DOI**: https://cloud.google.com/blog/products/gcp/introducing-kayenta-an-open-automated-canary-analysis-tool-from-google-and-netflix
- **Verified**: partial (page metadata confirmed, article body not fully extracted due to rendering)
- **Relevance**: 4
- **Summary**: Announces Kayenta as a joint Google-Netflix open-source project for automated canary analysis. Kayenta fetches user-configured metrics from their sources, runs statistical tests, and provides an aggregate score (0-100). Supports Stackdriver, Prometheus, Datadog, and Netflix Atlas. Designed to reduce risk from production changes by replacing manual monitoring graph inspection with automated statistical comparison.
- **Key Claims**:
  - Kayenta produces an aggregate canary score from 0 to 100, classified as success, marginal, or failure [**STRONG**]
  - Manual inspection of monitoring graphs is insufficient to reliably detect performance regressions -- automated statistical comparison is necessary [**MODERATE**]
  - Kayenta supports multiple metric backends including Stackdriver, Prometheus, Datadog, and Netflix Atlas [**MODERATE**]

### [SRC-011] AWS re:Invent 2025 -- Accelerate Software Delivery with Amazon ECS (CNS315)
- **Authors**: Kevin Gibbs, Mike Rizzo (AWS)
- **Year**: 2025
- **Type**: conference talk (AWS re:Invent 2025)
- **URL/DOI**: https://dev.to/kazuya_dev/aws-reinvent-2025-accelerate-software-delivery-with-amazon-ecs-cns315-2c2o
- **Verified**: yes (session notes)
- **Relevance**: 4
- **Summary**: re:Invent session demonstrating ECS advanced deployment strategies. Covers four strategies (rolling, blue-green, canary, linear) across four traffic shifting mechanisms (ALB, Service Connect, NLB, headless). Documents three failure detection approaches: circuit breaker, CloudWatch alarms, and custom Lambda hook tests. Covers migration path from CodeDeploy to ECS-native deployments.
- **Key Claims**:
  - Three failure detection mechanisms: circuit breaker (task health), CloudWatch alarms (metric thresholds), and custom Lambda hooks (domain-specific validation) [**MODERATE**]
  - NLB includes a 10-minute delay for routing synchronization, unsuitable for rapid canary iteration [**WEAK**]
  - Migration from CodeDeploy to ECS-native deployments supports in-place service updates or parallel service creation [**WEAK**]

### [SRC-012] Progressive Delivery Using AWS App Mesh and Flagger
- **Authors**: AWS Containers Team
- **Year**: 2020
- **Type**: blog post (AWS official blog)
- **URL/DOI**: https://aws.amazon.com/blogs/containers/progressive-delivery-using-aws-app-mesh-and-flagger/
- **Verified**: partial (title and URL confirmed via search)
- **Relevance**: 3
- **Summary**: Demonstrates Flagger integration with AWS App Mesh for progressive delivery on EKS. Shows how Flagger's control loop can leverage App Mesh virtual routers for traffic shifting in AWS-native Kubernetes environments. Bridges the gap between Kubernetes-native progressive delivery tools and AWS service mesh infrastructure.
- **Key Claims**:
  - Flagger can leverage AWS App Mesh virtual routers for traffic shifting, providing Kubernetes-native progressive delivery on AWS [**MODERATE**]
  - The combination of Flagger + App Mesh enables automated canary analysis with metric-gated promotion in AWS EKS environments [**WEAK**]

## Thematic Synthesis

### Theme 1: The Mann-Whitney U Test Is the De Facto Standard for Automated Canary Scoring

**Consensus**: The Mann-Whitney U test, a nonparametric statistical test that does not assume normal distribution of metrics, is the primary statistical method for automated canary analysis across major implementations. Netflix/Google's Kayenta uses it at 98% confidence with a 0.25x Hodges-Lehmann tolerance band. Effect size (Cohen's d or meanRatio) serves as a secondary gate after statistical significance is established. [**STRONG**]
**Sources**: [SRC-002], [SRC-009], [SRC-010]

**Controversy**: Whether effect size thresholds should be primary or secondary to statistical significance. Kayenta/Spinnaker treats effect size as a secondary gate (significance first), while some practitioners argue that practically insignificant but statistically significant differences should not trigger rollback.
**Dissenting sources**: [SRC-002] enforces significance-first, while [SRC-009] presents Cohen's d as a co-equal measure alongside the U test.

**Practical Implications**:
- Use the Mann-Whitney U test (not t-tests or simple threshold comparison) for canary metric comparison
- Set confidence at 95-98% to balance detection sensitivity with false positive rate
- Apply effect size as a secondary filter to avoid rolling back on statistically significant but operationally trivial differences
- Ensure minimum sample sizes before running statistical tests -- low-traffic canaries produce unreliable results

**Evidence Strength**: STRONG

### Theme 2: Differential Metrics Dramatically Reduce False Positive Rollbacks

**Consensus**: Comparing canary error rates against baseline (control) error rates -- rather than against absolute thresholds -- substantially reduces false positive automated rollbacks. Confounding factors (traffic spikes, spam, infrastructure noise) affect both populations equally and cancel out in differential comparison. [**STRONG**]
**Sources**: [SRC-001], [SRC-008], [SRC-002]

**Controversy**: None significant. The literature uniformly advocates differential comparison over absolute thresholds.

**Practical Implications**:
- Never use aggregate error rate alone as a rollback trigger -- always compare canary against a contemporaneous baseline
- Ensure uniform traffic distribution across canary and baseline pods so differential metrics are comparable
- Budget 2+ minutes of metric collection between traffic shifting stages for metric pipelines to stabilize
- Expect approximately 1 false positive rollback per week even with differential metrics in high-deployment-velocity environments (per Headout's production data)

**Evidence Strength**: STRONG

### Theme 3: AWS Provides Native Progressive Delivery for Both ECS and Lambda Without Third-Party Controllers

**Consensus**: As of October 2025, AWS offers built-in canary and linear deployment strategies for ECS (via native deployment controller) and Lambda (via alias traffic shifting + CodeDeploy). These eliminate the need for Kubernetes-based controllers (Flagger, Argo Rollouts) or external deployment services for AWS-native workloads. ECS supports canary percentages from 0.1% to 99.9% with lifecycle hooks at 8 stages. Lambda supports weighted aliases across exactly two published versions. [**STRONG**]
**Sources**: [SRC-003], [SRC-004], [SRC-005], [SRC-011]

**Controversy**: Whether ECS-native deployments are mature enough to replace CodeDeploy-based deployments. The feature was announced October 2025, and early adopters may encounter edge cases not covered in documentation.
**Dissenting sources**: [SRC-011] presents migration from CodeDeploy as straightforward, but the feature's recency (< 6 months at time of review) means limited production battle-testing data.

**Practical Implications**:
- For new ECS services, prefer built-in canary/linear over CodeDeploy -- simpler architecture, fewer moving parts
- Lambda's probabilistic traffic model means low-traffic functions will see high variance from configured weights -- plan canary evaluation windows accordingly
- NLB-backed ECS services cannot use canary/linear -- only blue-green via AllAtOnce
- Use lifecycle hooks for pre-flight checks (e.g., trusted image verification at PRE_SCALE_UP) and post-shift validation
- Configure CloudWatch alarms on application-level metrics (error rate, P99 latency) for automated rollback

**Evidence Strength**: STRONG

### Theme 4: Rollback Timing Is Configuration-Dependent and Ranges from 90 Seconds to 30+ Minutes

**Consensus**: There is no universal "right" rollback timing. The optimal window depends on traffic volume (determines when statistical significance is achievable), metric pipeline latency (determines when metrics become available for comparison), and business criticality (determines acceptable blast radius duration). Industry practice ranges from Flagger's minimum of interval x threshold (e.g., 30s x 3 = 90 seconds) to AWS CodeDeploy's canary windows of 5-30 minutes. [**MODERATE**]
**Sources**: [SRC-001], [SRC-005], [SRC-006], [SRC-008]

**Controversy**: Whether faster detection is always better. Faster detection reduces blast radius but increases false positive rates due to insufficient sample sizes.
**Dissenting sources**: [SRC-001] warns that canary duration must span "the duration of a single work unit" -- too-short canaries miss slow-burning issues. [SRC-006] enables sub-minute rollback via aggressive threshold/interval settings.

**Practical Implications**:
- High-traffic services (>1000 RPS): 2-5 minute canary intervals with 3-failure threshold = 6-15 minute rollback window
- Low-traffic services (<100 RPS): 5-10 minute intervals needed for statistical significance, yielding 15-30 minute rollback windows
- Lambda canaries at low invocation rates should use 15-30 minute CodeDeploy windows due to probabilistic traffic distribution variance
- Asynchronous workloads (queue consumers, batch processors) need canary duration >= single work unit duration
- Bake time after full traffic shift (ECS's deployment bake time) provides a final safety net before old version termination

**Evidence Strength**: MIXED (consensus on the range, controversy on optimal values)

### Theme 5: Metric Selection Should Prioritize SLIs Over Infrastructure Metrics

**Consensus**: Effective canary metrics should reflect user-perceivable service quality (SLIs) rather than infrastructure signals. HTTP success rate and P99 latency are the most commonly recommended builtin metrics. The Google SRE Workbook recommends no more than "a dozen" metrics to prevent signal dilution. Metrics must be clearly attributable to the canary change and breakable by population (canary vs. control). [**MODERATE**]
**Sources**: [SRC-001], [SRC-002], [SRC-006], [SRC-009]

**Controversy**: Whether custom business metrics (conversion rates, user engagement) should be included in automated canary scoring or left to separate A/B testing frameworks.

**Practical Implications**:
- Start with two metrics: request success rate and P99 latency -- these catch most regressions
- Add 3-5 domain-specific metrics (e.g., queue depth, cache hit rate) as second-tier signals
- Mark critical business metrics (e.g., payment success rate) as critical in Kayenta/Argo Rollouts so a single breach triggers immediate score=0
- Avoid high-variance infrastructure metrics (CPU utilization, memory) as primary rollback triggers -- they produce false positives due to instance heterogeneity
- Use percentile-based latency (P95/P99) rather than averages to detect tail latency regressions

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- A canary population of 5% experiencing 20% errors results in only 1% overall error rate -- blast radius is directly proportional to canary population size -- Sources: [SRC-001]
- The Mann-Whitney U test at 98% confidence is the industry-standard statistical method for automated canary metric comparison, as implemented in Netflix/Google's Kayenta -- Sources: [SRC-002], [SRC-009], [SRC-010]
- Control-vs-canary metric comparison is superior to before/after evaluation because time is "one of the biggest sources of change in observed metrics" -- Sources: [SRC-001]
- ECS now supports native canary/linear deployments (October 2025) with weighted ALB routing, 8 lifecycle hooks, and CloudWatch alarm-triggered rollback -- Sources: [SRC-003], [SRC-011]
- Lambda alias traffic shifting supports weighted routing across exactly two published versions using a probabilistic distribution model -- Sources: [SRC-004]
- CodeDeploy provides predefined canary configurations: 10% initial traffic with 5/10/15/30 minute evaluation windows for Lambda; 5/15 minutes for ECS -- Sources: [SRC-005]
- Flagger's rollback timeframe equals interval x threshold; minimum validation duration equals interval x (maxWeight / stepWeight) -- Sources: [SRC-006]
- Argo Rollouts supports both background analysis (non-blocking) and inline analysis (blocking) for metric-driven canary decisions -- Sources: [SRC-007]

### MODERATE Evidence
- Switching from aggregate error rate to canary-vs-baseline error differential reduced false positive rollbacks by 80% (5/week to 1/week) in production -- Sources: [SRC-008]
- Effective canary metrics should number no more than about a dozen to prevent diminishing returns from signal dilution -- Sources: [SRC-001]
- Effect size thresholds (meanRatio, Cohen's d) serve as secondary gates after statistical significance in Kayenta -- effect size alone should not trigger rollback -- Sources: [SRC-002], [SRC-009]
- Manual monitoring graph inspection is insufficient to reliably detect canary regressions -- automated statistical analysis is necessary at deployment velocity -- Sources: [SRC-010]
- Kayenta's 50% NODATA rule automatically fails canaries when half or more metrics lack data -- Sources: [SRC-002]
- Warmup periods, minimum sample sizes, and consecutive failure thresholds are essential safeguards against premature rollback -- Sources: [SRC-009]
- Lambda's probabilistic traffic model produces significant variance from configured weights at low invocation volumes -- Sources: [SRC-004]
- NLB-backed ECS services are limited to blue-green (AllAtOnce) -- no canary or linear support -- Sources: [SRC-003], [SRC-005]
- Three ECS failure detection mechanisms work in concert: circuit breaker (task health), CloudWatch alarms (metric thresholds), and custom Lambda hooks -- Sources: [SRC-011]

### WEAK Evidence
- 2-minute pause between canary routing stages is sufficient for New Relic metric collection to stabilize -- Sources: [SRC-008]
- NLB includes a 10-minute delay for routing synchronization, making it unsuitable for rapid canary iteration -- Sources: [SRC-011]
- Each ECS lifecycle stage can last up to 24 hours; CloudFormation deployment limited to 36 hours total -- Sources: [SRC-003]
- Flagger's skipAnalysis mode bypasses all metric validation for emergency promotions -- Sources: [SRC-006]
- Tiny Gaussian noise is added to degenerate metric distributions to enable Mann-Whitney U testing in Kayenta -- Sources: [SRC-002]

### UNVERIFIED
- Google published a "Safe Rollout" paper in 2023 with automated canary scoring methodology -- Basis: mentioned in user request; no paper with this exact title found via web search. Google's canary methodology is documented in the SRE Workbook (2018) and Kayenta (2018), but a 2023-specific paper could not be verified.
- Optimal false negative rates (missed bad deployments) for automated canary systems in production -- Basis: no quantitative study of false negative rates was found in the accessible literature. Industry focus is heavily on false positive reduction; false negative measurement requires tracking post-deployment incidents attributable to canary analysis misses, which few organizations publish.

## Knowledge Gaps

- **False negative rate quantification**: While false positive rates are documented (e.g., Headout's 5/week to 1/week), no accessible study quantifies false negative rates (bad deployments that pass canary analysis). This is a critical gap because false negatives cause production incidents. Filling this gap would require correlating post-deployment incident data with canary analysis pass decisions across many deployments.

- **Google "Safe Rollout" 2023 paper**: The user referenced a Google 2023 paper on automated canary scoring. Extensive search did not locate a paper with this exact title. Google's published canary methodology appears in the SRE Workbook (2018) and Kayenta (2018). A 2023 paper may exist under a different title, be internal-only, or be behind institutional access. Manual search of Google Research publications or ACM Digital Library would be needed.

- **ECS native canary at scale**: The ECS built-in canary feature (October 2025) is less than 6 months old at time of review. No independent production case studies or failure mode analyses are available yet. Early adopter reports and failure mode documentation would strengthen confidence in this finding.

- **Cross-service canary orchestration**: How to coordinate canary analysis across coupled ECS services and Lambda functions (e.g., a canary API Gateway + canary Lambda + canary downstream ECS service) is not addressed in any source. Each AWS service has independent canary mechanisms with no built-in cross-service coordination.

- **Cost impact of progressive delivery**: The resource cost of maintaining double task sets (ECS) or provisioned concurrency for two versions (Lambda) during canary windows is not quantified in the literature. This is operationally significant for cost-sensitive workloads.

## Domain Calibration

This review covers a well-practiced domain with strong primary documentation from Google SRE, Netflix/Spinnaker, and AWS. The majority of claims are backed by official documentation or established tooling, yielding a moderate-to-high confidence distribution. The gap between academic research (sparse) and practitioner documentation (abundant) reflects that progressive delivery is an engineering practice rather than an academic research area. Evidence grades reflect the quality of accessible documentation, not independent verification of the claims within that documentation.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best. The Netflix Tech Blog Kayenta article could not be fetched due to a TLS certificate error; claims about Kayenta are corroborated via the Spinnaker judge documentation instead.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.
5. **Google "Safe Rollout" paper**: A specific 2023 Google paper on automated canary scoring referenced in the research request could not be located. This may indicate the paper is internal, behind institutional access, or referenced by an approximate title.

Generated by `/research progressive delivery canary analysis ECS Lambda` on 2026-03-06.
