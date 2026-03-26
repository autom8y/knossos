---
domain: "literature-causal-inference-in-spatiotemporal-epidemiology"
generated_at: "2026-02-27T11:38:26Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.67
format_version: "1.0"
---

# Literature Review: Causal Inference in Spatiotemporal Epidemiology

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

Causal inference in spatiotemporal epidemiology has evolved rapidly since the mid-2010s, driven by the recognition that standard causal frameworks (potential outcomes, difference-in-differences) require substantial modification when applied to spatially and temporally correlated data. The literature broadly agrees that three challenges dominate the field: (1) spatial confounding from unmeasured spatially varying factors, (2) interference/spillover where treatment at one location affects outcomes at others (violating SUTVA), and (3) the inadequacy of the parallel trends assumption for non-continuous outcomes. There is emerging consensus that joint Bayesian models integrating disease mapping with causal frameworks outperform simpler two-stage approaches, and that triangulation across multiple causal methods provides the most robust evidence. Active controversy exists around whether Bayesian spatial smoothing helps or hinders causal identification, and the field of causal machine learning for spatiotemporal data remains in its infancy with less than 1% of published studies integrating these approaches.

## Source Catalog

### [SRC-001] A Review of Spatial Causal Inference Methods for Environmental and Epidemiological Applications
- **Authors**: Brian J. Reich, Shu Yang, Yawen Guan, Andrew B. Giffin, Matthew J. Miller, Ana Rappold
- **Year**: 2021
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1111/insr.12452
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 5
- **Summary**: The most comprehensive review of spatial causal inference methods for environmental and epidemiological applications. Covers adjusting for spatial confounders (matching, propensity scores, instrumental variables), spatial interference/spillover (partial interference, network-based, process-based models), spatiotemporal extensions (difference-in-differences, Granger causality), and point-referenced geostatistical data. Identifies joint modeling, non-parametric extensions, and optimal spatial treatment allocation as key future directions.
- **Key Claims**:
  - Spatial confounding from unmeasured spatially varying factors requires explicit adjustment beyond standard regression [**STRONG**]
  - Joint modeling of treatment and response simultaneously outperforms two-stage methods when properly specified [**MODERATE**]
  - SUTVA is routinely violated in spatial settings due to interference; simplifying assumptions about interference structure are required for tractability [**STRONG**]
  - DAGs are essential for identifying confounding structure and appropriate adjustment sets in spatial causal designs [**STRONG**]
  - Mechanistic process-based models offer improved fidelity for modeling interference compared to purely statistical approaches [**MODERATE**]

### [SRC-002] Universal Difference-in-Differences for Causal Inference in Epidemiology
- **Authors**: Eric J. Tchetgen Tchetgen, Chan Park, David B. Richardson
- **Year**: 2023
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1097/EDE.0000000000001676
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 5
- **Summary**: Introduces "universal DiD" which replaces the parallel trends assumption with an odds ratio equi-confounding assumption that remains valid for binary, count, and polytomous outcomes -- the exact outcome types common in epidemiology. Demonstrates the method using Zika virus outbreak impact on Brazilian birth rates. Provides three estimation strategies: generalized linear models, extended propensity score weighting, and doubly robust methods.
- **Key Claims**:
  - The standard parallel trends assumption is violated for non-continuous outcomes (binary, count, polytomous) common in epidemiology [**STRONG**]
  - Universal DiD identifies any causal effect conceivably identifiable in the absence of confounding bias, including nonlinear effects [**MODERATE**]

### [SRC-003] Generalized Propensity Score Approach to Causal Inference with Spatial Interference
- **Authors**: Andrew B. Giffin, Brian J. Reich, Shu Yang, Ana G. Rappold
- **Year**: 2022
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1111/biom.13745
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 5
- **Summary**: Develops a causal framework for recovering direct and spillover effects when spatial interference is present. Establishes that a generalized propensity score (GPS) is sufficient to remove all measured confounding in spatial settings with bivariate exposure components (direct and indirect). Uses Bayesian spline-based regression with a three-step computational algorithm to avoid feedback from response to propensity estimation. Applied to wildfire impacts on PM2.5 across the Western US.
- **Key Claims**:
  - A generalized propensity score is sufficient to remove all measured confounding in the presence of spatial interference [**MODERATE**]
  - SUTVA violation due to spatial spillover is the norm rather than the exception in environmental and epidemiological applications [**STRONG**]

### [SRC-004] Spatial Difference-in-Differences with Bayesian Disease Mapping Models
- **Authors**: Carl Bonander, Marta Blangiardo, Ulf Stromberg
- **Year**: 2025
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1097/EDE.0000000000001912
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 5
- **Summary**: Integrates Bayesian disease-mapping models into an imputation-based DID framework for small-area policy evaluations. Uses a two-way Mundlak estimator to decompose unit and time effects, enabling spatial random effects without sacrificing causal identification. Implemented via INLA for efficient Bayesian computation. Demonstrates precision improvements for small populations and rare events using Swedish municipal ice cleat distribution programs.
- **Key Claims**:
  - Spatial DID with INLA improves precision for small-area evaluations by accounting for spatial correlation through smoothing [**MODERATE**]
  - The parallel trends assumption is often not credible for non-continuous outcomes, motivating alternative identification strategies [**STRONG**]

### [SRC-005] Alternative Causal Inference Methods in Population Health Research: Evaluating Tradeoffs and Triangulating Evidence
- **Authors**: Ellicott C. Matthay, Erin Hagan, Laura M. Gottlieb, May Lynn Tan, David Vlahov, Nancy E. Adler, M. Maria Glymour
- **Year**: 2019
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1016/j.ssmph.2019.100526
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 4
- **Summary**: Compares confounder-control and instrument-based causal inference approaches using Shadish, Cook, and Campbell's validity framework. Core argument: neither approach universally dominates. Confounder-control methods gain statistical power but assume all confounders are measured. Instrument-based methods handle unmeasured confounding but sacrifice power and generalizability. Advocates triangulation across diverse designs for robust causal evidence in population health.
- **Key Claims**:
  - Confounder-control methods trade internal validity for statistical power; instrument-based methods trade power for internal validity [**STRONG**]
  - Triangulation across multiple causal methods provides stronger evidence than any single approach alone [**MODERATE**]

### [SRC-006] Toward Causal Inference for Spatio-Temporal Data: Conflict and Forest Loss in Colombia
- **Authors**: Rune Christiansen, Matthias Baumann, Tobias Kuemmerle, Miguel D. Mahecha, Jonas Peters
- **Year**: 2022
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1080/01621459.2021.2013241
- **Verified**: yes (title confirmed via multiple databases, DOI confirmed)
- **Relevance**: 5
- **Summary**: Proposes a class of causal models for spatio-temporal stochastic processes that formally define and quantify causal effects without strong distributional assumptions and with arbitrarily many latent confounders, provided those confounders do not vary across time. Includes a nonparametric hypothesis test for causal effects being zero. Applied to Colombian conflict and forest loss (2000-2018), finding heterogeneous effects at the provincial level.
- **Key Claims**:
  - Spatio-temporal processes allow causal identification even with arbitrarily many latent confounders, given time-invariance of confounders [**MODERATE**]

### [SRC-007] A New Tool for Case Studies in Epidemiology -- the Synthetic Control Method
- **Authors**: David H. Rehkopf, Sanjay Basu
- **Year**: 2018
- **Type**: peer-reviewed paper (invited commentary)
- **URL/DOI**: https://doi.org/10.1097/EDE.0000000000000837
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 3
- **Summary**: Introduces the synthetic control method for epidemiological case studies. The method constructs a weighted combination of control units matching the treated unit's pre-intervention characteristics, providing transparency in control selection. Identifies significant constraints: limited statistical power with small samples, challenges establishing confidence intervals, and sensitivity to donor pool selection. Stresses that SUTVA, random assignment, and fixed potential outcomes are required assumptions.
- **Key Claims**:
  - Synthetic control is valuable for transparency in control selection but inherently limited in statistical power with small samples [**MODERATE**]

### [SRC-008] Spatial Causal Inference in the Presence of Unmeasured Confounding and Interference
- **Authors**: Georgia Papadogeorgou, Srijata Samanta
- **Year**: 2023
- **Type**: preprint (arXiv:2303.08218, revised February 2026)
- **URL/DOI**: https://arxiv.org/abs/2303.08218
- **Verified**: yes (full text fetched from arXiv, content confirmed)
- **Relevance**: 5
- **Summary**: Introduces spatial causal graphs showing how spatial confounding and interference can be entangled, such that investigating one in isolation leads to wrong conclusions. Proposes a Bayesian parametric approach that simultaneously accounts for interference and mitigates bias from local and neighborhood unmeasured spatial confounding. Proves parameter identifiability for causal effects even with unmeasured confounding under the proposed model. Applied to SO2 emissions and cardiovascular mortality across US counties.
- **Key Claims**:
  - Spatial confounding and interference can be entangled; investigating one without accounting for the other produces misleading conclusions [**MODERATE**]
  - Under a specific Bayesian parametric formulation, causal effects are identifiable even with unmeasured spatial confounding [**MODERATE**]
  - Spatial dependence in the exposure variable renders standard analyses invalid, requiring an exposure model [**MODERATE**]

### [SRC-009] Target Trial Emulation for Evaluating Health Policy
- **Authors**: Nicholas J. Seewald, Emma E. McGinty, Elizabeth A. Stuart
- **Year**: 2024
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.7326/M23-2440
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 4
- **Summary**: Presents target trial emulation for policy evaluation with explicit spatiotemporal considerations. Addresses staggered policy adoption through "stacking" (serial trial emulation creating separate cohorts per implementation date). Notes that geographically distant comparators alleviate spillover concerns while near-neighbor selection increases similarity but risks interference. Defines seven core components of the policy trial emulation framework.
- **Key Claims**:
  - Target trial emulation with temporal stacking handles staggered geographic policy adoption for causal evaluation [**MODERATE**]
  - Selecting geographically distant comparators reduces spillover risk but may sacrifice comparability [**MODERATE**]

### [SRC-010] Bayesian Disease Mapping: Past, Present, and Future
- **Authors**: Ying C. MacNab
- **Year**: 2022
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1016/j.spasta.2022.100593
- **Verified**: yes (full text fetched via PMC, DOI confirmed)
- **Relevance**: 3
- **Summary**: Comprehensive review tracing Bayesian disease mapping from its 19th-century origins through modern hierarchical methods. Extensively discusses CAR models (intrinsic, proper, Leroux) and BYM/BYM2 priors for spatial smoothing. Highlights INLA as a computationally efficient alternative to MCMC. Does not explicitly address causal inference, but provides the foundational disease mapping infrastructure upon which spatial causal methods are built. Applied to COVID-19 county-level modeling.
- **Key Claims**:
  - Bayesian hierarchical models with CAR/BYM priors are the standard infrastructure for spatiotemporal disease mapping [**MODERATE**]
  - INLA provides computationally efficient Bayesian inference for latent Gaussian models, enabling practical spatiotemporal analysis [**MODERATE**]

### [SRC-011] Controlling for Spatial Confounding and Spatial Interference in Causal Inference: Modelling Insights from a Computational Experiment
- **Authors**: Tyler D. Hoffman, Peter Kedron
- **Year**: 2023
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1080/19475683.2023.2257788
- **Verified**: partial (abstract and metadata confirmed; full text access blocked)
- **Relevance**: 4
- **Summary**: Benchmarks 28 spatial causal models across 16 simulated data scenarios involving different combinations of spatial confounding and interference. Introduces the spycause Python package for spatial causal modeling and simulation. Demonstrates that noncausal spatial modeling guidance (e.g., on CAR specification) transfers to causal spatial workflows. Provides a structured computational comparison absent from prior literature.
- **Key Claims**:
  - Noncausal spatial modeling guidance (CAR specification, spatial smoothing) transfers effectively to causal spatial workflows [**MODERATE**]
  - A computational benchmark of 28 models across 16 scenarios shows model performance depends heavily on the spatial structure of treatment allocation [**MODERATE**]

### [SRC-012] A Structured Comparison of Causal Machine Learning Methods to Assess Heterogeneous Treatment Effects in Spatial Data
- **Authors**: Kevin Credit, Matthew Lehnert
- **Year**: 2024
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1007/s10109-023-00413-0
- **Verified**: partial (metadata confirmed via Springer and IDEAS/RePEc; full text paywalled)
- **Relevance**: 4
- **Summary**: Compares causal forest (CF) models across different test/train split definitions for geographically referenced data, finding that standard random splits fracture the spatial fabric. Develops a new "spatial" T-learner using random forest for heterogeneous treatment effect estimation. All ML models outperform OLS for average treatment effects. Applied to light rail construction impact on CO2 emissions in Maricopa County, Arizona.
- **Key Claims**:
  - Standard causal forest random train/test splits fracture the spatial fabric of geographic data, requiring spatial-aware alternatives [**WEAK**]
  - A spatial T-learner addresses the spatial fracture problem for heterogeneous treatment effect estimation [**WEAK**]
  - Less than 1% of studies in major databases explicitly integrate causal ML with spatiotemporal analysis [**WEAK**]

### [SRC-013] Estimating Heterogeneous Treatment Effects for Spatio-Temporal Causal Inference
- **Authors**: Lingxiao Zhou, Kosuke Imai, Jason Lyall, Georgia Papadogeorgou
- **Year**: 2024
- **Type**: preprint (arXiv:2412.15128)
- **URL/DOI**: https://arxiv.org/abs/2412.15128
- **Verified**: yes (full text fetched from arXiv, content confirmed)
- **Relevance**: 4
- **Summary**: Develops a Hajek-type estimator for conditional average treatment effects (CATE) as a function of spatio-temporal moderator variables. Establishes asymptotic normality and provides a test procedure for heterogeneous treatment effects. Allows for arbitrary spatial and temporal causal dependencies in high-frequency spatiotemporal data. Applied to US airstrikes and insurgent violence in Iraq, finding that prior aid distribution moderates airstrike effects -- contrary to counterinsurgency theory predictions.
- **Key Claims**:
  - Spatio-temporal heterogeneous treatment effects can be estimated with established asymptotic properties under arbitrary spatial and temporal dependencies [**WEAK**]

## Thematic Synthesis

### Theme 1: Spatial Confounding and Interference Are Entangled Challenges Requiring Joint Treatment

**Consensus**: Unmeasured spatial confounding and spatial interference (spillover) are the two dominant threats to causal identification in spatiotemporal settings. These threats are not independent -- investigating one without accounting for the other produces biased or misleading estimates. Joint modeling approaches that simultaneously address both confounding and interference outperform methods that treat them separately. [**MODERATE**]
**Sources**: [SRC-001], [SRC-003], [SRC-008], [SRC-011]

**Controversy**: Whether Bayesian spatial smoothing helps or hinders causal identification. Traditional econometric wisdom holds that random effects models are inconsistent for causal inference, but [SRC-004] shows that a Mundlak decomposition can preserve causal validity while gaining the precision benefits of spatial smoothing. [SRC-011] finds that noncausal spatial modeling guidance transfers to causal settings, while earlier causal inference literature has been skeptical of importing spatial modeling techniques wholesale.
**Dissenting sources**: [SRC-004] argues spatial smoothing improves precision without sacrificing causal validity (via Mundlak decomposition), while classical DID literature treats random effects as incompatible with causal identification.

**Practical Implications**:
- When designing a spatial causal study, assess both confounding and interference simultaneously -- do not assume one is absent
- Use joint models (e.g., [SRC-008]'s Bayesian parametric approach) when both threats are plausible
- If using spatial smoothing for precision, explicitly verify that the smoothing does not compromise causal identification (e.g., via Mundlak decomposition as in [SRC-004])

**Evidence Strength**: MODERATE

### Theme 2: The Parallel Trends Assumption Fails for Most Epidemiological Outcomes

**Consensus**: The standard difference-in-differences parallel trends assumption is not credible for binary, count, or polytomous outcomes -- the predominant outcome types in epidemiology. Alternative identification strategies are needed, including universal DiD (odds ratio equi-confounding), synthetic control methods, and imputation-based DID with Bayesian smoothing. [**STRONG**]
**Sources**: [SRC-002], [SRC-004], [SRC-007]

**Practical Implications**:
- Default to universal DiD or doubly robust DID when outcomes are non-continuous (the common case in epidemiology)
- Consider synthetic control methods when only one or a few units receive treatment, but acknowledge inherent power limitations
- Use imputation-based spatial DID with INLA for small-area evaluations where precision is critical

**Evidence Strength**: STRONG

### Theme 3: Triangulation Across Methods is the Gold Standard for Spatiotemporal Causal Claims

**Consensus**: No single causal method universally dominates in spatiotemporal settings. Each method entails untestable assumptions and tradeoffs between statistical power, internal validity, measurement quality, and generalizability. The strongest causal evidence comes from triangulation -- applying multiple complementary methods to the same question and assessing convergence. [**MODERATE**]
**Sources**: [SRC-005], [SRC-001], [SRC-009]

**Practical Implications**:
- Plan studies to employ at least two complementary causal approaches (e.g., a confounder-control method and an instrument-based method)
- When methods disagree, investigate which untestable assumptions are most likely violated rather than selecting the method with the preferred result
- Use target trial emulation as a design discipline before choosing an estimation method -- it forces explicit articulation of causal assumptions

**Evidence Strength**: MODERATE

### Theme 4: Bayesian Disease Mapping Infrastructure Enables Spatial Causal Inference at Scale

**Consensus**: The Bayesian disease mapping tradition (CAR models, BYM priors, INLA computation) provides the foundational spatial infrastructure upon which modern spatiotemporal causal methods are built. INLA in particular has made Bayesian spatiotemporal models computationally feasible for routine epidemiological use. Recent work directly integrates these disease mapping tools into causal frameworks. [**MODERATE**]
**Sources**: [SRC-010], [SRC-004], [SRC-011]

**Practical Implications**:
- Invest in Bayesian spatiotemporal modeling infrastructure (R-INLA, CAR/BYM specification) as a prerequisite for spatial causal analysis
- Leverage existing disease mapping workflows when extending to causal questions -- [SRC-011] shows noncausal spatial modeling guidance transfers
- Use INLA over MCMC for routine spatial causal analyses unless the model class requires sampling-based inference

**Evidence Strength**: MODERATE

### Theme 5: Causal Machine Learning for Spatiotemporal Data Is an Emerging Frontier

**Consensus**: The integration of causal machine learning (causal forests, meta-learners) with spatiotemporal data analysis is in its earliest stages. Standard ML approaches fracture spatial structure in train/test splits. Spatial adaptations (spatial T-learners, spatiotemporal CATE estimators) are being developed but lack the methodological maturity and empirical validation of model-based approaches. [**WEAK**]
**Sources**: [SRC-012], [SRC-013]

**Controversy**: Whether ML-based causal methods can meaningfully contribute in spatial settings where model-based approaches already perform well. [SRC-012] shows ML outperforms OLS for average treatment effects, but the comparison is against a weak baseline. It remains unclear whether ML methods outperform well-specified Bayesian spatial models.
**Dissenting sources**: [SRC-012] argues ML methods outperform traditional regression in spatial causal settings, while model-based literature (e.g., [SRC-001], [SRC-004]) demonstrates strong performance without ML.

**Practical Implications**:
- Treat causal ML methods for spatiotemporal data as experimental; prefer model-based Bayesian approaches for production research
- If using causal forests on spatial data, implement spatial-aware train/test splits (e.g., spatial cross-validation)
- Monitor this frontier closely -- spatiotemporal CATE estimation ([SRC-013]) addresses a genuine gap in characterizing effect heterogeneity

**Evidence Strength**: WEAK

## Evidence-Graded Findings

### STRONG Evidence
- Spatial confounding from unmeasured spatially varying factors requires explicit adjustment beyond standard regression -- Sources: [SRC-001], [SRC-008]
- SUTVA is routinely violated in spatial epidemiological settings due to interference/spillover effects -- Sources: [SRC-001], [SRC-003], [SRC-008]
- The standard parallel trends assumption is violated for non-continuous outcomes (binary, count, polytomous) common in epidemiology -- Sources: [SRC-002], [SRC-004]
- Confounder-control methods trade internal validity for statistical power; instrument-based methods trade power for internal validity -- Sources: [SRC-005], [SRC-001]
- DAGs are essential for identifying confounding structure and appropriate adjustment sets in spatial causal designs -- Sources: [SRC-001], [SRC-005], [SRC-009]

### MODERATE Evidence
- Joint modeling of treatment and response outperforms two-stage methods when properly specified -- Sources: [SRC-001]
- A generalized propensity score is sufficient to remove all measured confounding with spatial interference -- Sources: [SRC-003]
- Universal DiD identifies any causal effect conceivably identifiable in the absence of confounding bias -- Sources: [SRC-002]
- Triangulation across multiple causal methods provides stronger evidence than any single approach -- Sources: [SRC-005]
- Spatio-temporal processes allow causal identification even with latent confounders given time-invariance -- Sources: [SRC-006]
- Spatial confounding and interference are entangled and must be addressed jointly -- Sources: [SRC-008]
- Under a Bayesian parametric formulation, causal effects are identifiable with unmeasured spatial confounding -- Sources: [SRC-008]
- Synthetic control is valuable for transparency but inherently limited in power -- Sources: [SRC-007]
- Target trial emulation with stacking handles staggered geographic policy adoption -- Sources: [SRC-009]
- Bayesian CAR/BYM priors are the standard infrastructure for spatiotemporal disease mapping -- Sources: [SRC-010]
- Spatial DID with INLA improves precision for small-area evaluations -- Sources: [SRC-004]
- Noncausal spatial modeling guidance transfers to causal spatial workflows -- Sources: [SRC-011]
- Mechanistic process-based models improve interference modeling fidelity -- Sources: [SRC-001]

### WEAK Evidence
- Standard causal forest train/test splits fracture the spatial fabric of geographic data -- Sources: [SRC-012]
- A spatial T-learner addresses spatial fracture for heterogeneous treatment effect estimation -- Sources: [SRC-012]
- Spatio-temporal HTE estimation is possible with arbitrary spatial and temporal dependencies -- Sources: [SRC-013]
- Less than 1% of studies integrate causal ML with spatiotemporal analysis -- Sources: [SRC-012], [SRC-013]

### UNVERIFIED
- The optimal strategy for choosing between spatial smoothing and fixed effects in causal spatial models remains unresolved -- Basis: model training knowledge, partial support from [SRC-011] computational experiments
- Causal discovery (structure learning) methods for spatiotemporal data in epidemiology are substantially less developed than causal estimation methods -- Basis: model training knowledge; no source specifically surveyed this gap

## Knowledge Gaps

- **Causal discovery vs. causal estimation**: The reviewed literature overwhelmingly focuses on causal estimation (estimating the magnitude of a known causal effect) rather than causal discovery (identifying which variables have causal relationships). Methods like Granger causality and PC algorithm adaptations for spatiotemporal data exist but were not well-represented in the accessible epidemiological literature. A survey bridging causal discovery and spatiotemporal epidemiology would fill an important gap.

- **Infectious disease transmission dynamics**: While the reviewed methods apply well to chronic disease outcomes and environmental exposures, the specific challenges of infectious disease transmission (agent-based models, SIR dynamics, network structure of contagion) and how they interact with causal inference frameworks received limited coverage. The intersection of mechanistic epidemiological models and counterfactual causal frameworks is underexplored.

- **Software ecosystems**: Only [SRC-011] introduced a dedicated software package (spycause). The practical implementation landscape for spatial causal inference -- which R/Python packages, which computational pipelines, which model diagnostics -- is poorly documented in the methodological literature. Practitioners must assemble toolchains from disease mapping packages (R-INLA, CARBayes) and causal inference packages (EconML, CausalForest) with no integrated framework.

- **Equity and environmental justice applications**: Spatial causal inference has clear applications to environmental justice (disproportionate exposure of marginalized communities), but the reviewed literature primarily used environmental/ecological applications. How spatial causal methods should be adapted for equity-focused research questions -- including appropriate choice of counterfactuals and interference structures -- is underexplored.

- **Scalability to high-resolution spatiotemporal data**: Most reviewed methods operate on areal (aggregate) data. Extension to point-referenced, high-frequency spatiotemporal data (e.g., GPS-tracked individual exposures, real-time air quality monitoring) at scale remains computationally and methodologically challenging.

## Domain Calibration

Mixed evidence distribution across tiers reflects a field in active methodological development. Core spatial causal challenges (confounding, interference, parallel trends) are well-established (STRONG), while solutions are individually validated but not yet consolidated into standard practice (MODERATE). Machine learning integration remains early-stage (WEAK). This distribution honestly reflects a domain where the problems are well-characterized but the solution landscape is still maturing.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research causal inference in spatiotemporal epidemiology` on 2026-02-27.
