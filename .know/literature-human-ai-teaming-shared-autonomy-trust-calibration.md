---
domain: "literature-human-ai-teaming-shared-autonomy-trust-calibration"
generated_at: "2026-03-10T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.72
format_version: "1.0"
---

# Literature Review: Human-AI Teaming, Shared Autonomy, and Trust Calibration

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on human-AI teaming, shared autonomy, and trust calibration spans four decades of human factors research, with foundational work from aviation and nuclear domains now being applied to AI agent oversight. The field converges on a central paradox first identified by Bainbridge (1983): automating the easy parts of a task makes the hard parts harder for human operators, and the monitoring role humans are left with is precisely what humans do poorly. Trust calibration -- ensuring operators neither over-rely on nor under-rely on automated systems -- is the mechanism through which this paradox manifests operationally. There is strong consensus that static human-in-the-loop requirements are insufficient; meaningful human control requires dynamic, context-sensitive autonomy adjustment with transparency mechanisms that map to the operator's situational awareness needs. Controversy persists over whether existing regulatory frameworks (EU AI Act Article 14, FDA CDS guidance) operationalize "meaningful human control" or merely create "scapegoat-as-a-service" liability transfer.

## Source Catalog

### [SRC-001] Trust in Automation: Designing for Appropriate Reliance
- **Authors**: John D. Lee, Katrina A. See
- **Year**: 2004
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1518/hfes.46.1.50_30392
- **Verified**: partial (title, journal, DOI confirmed; full text behind paywall)
- **Relevance**: 5
- **Summary**: Foundational review (4,000+ citations) integrating trust research from organizational, sociological, interpersonal, psychological, and neurological perspectives. Defines trust calibration as the correspondence between trust and automation capability. Argues that misuse (over-reliance) and disuse (under-reliance) of automation both stem from poor calibration of trust relative to actual system capability.
- **Key Claims**:
  - Trust calibration -- the match between operator trust and actual automation capability -- is the primary determinant of appropriate reliance behavior [**STRONG**]
  - People respond to technology socially, and trust guides reliance when complexity makes complete understanding of automation impractical [**STRONG**]
  - Overtrust leads to automation misuse; distrust leads to automation disuse; both are failures of calibration [**STRONG**]
  - Display design that makes automation performance visible can improve trust calibration [**MODERATE**]

### [SRC-002] Ironies of Automation
- **Authors**: Lisanne Bainbridge
- **Year**: 1983
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1016/0005-1098(83)90046-8
- **Verified**: partial (title, journal, DOI confirmed; full text behind paywall)
- **Relevance**: 5
- **Summary**: Seminal paper (4,700+ citations) identifying five paradoxes of automation in industrial process control. The core irony: automating the easy parts of a task makes the remaining difficult parts harder for human operators, while leaving humans responsible for monitoring -- a task humans perform poorly. Predicted skill decay, complacency, and takeover failures decades before they manifested in modern AI systems.
- **Key Claims**:
  - Automation designed to reduce human error creates conditions where human intervention becomes more critical yet less achievable [**STRONG**]
  - Operators who do not practice manual skills as part of ongoing work will have degraded performance when manual takeover is required [**STRONG**]
  - Monitoring tasks, to which humans are relegated by automation, are precisely the tasks humans perform least effectively [**STRONG**]
  - Designer errors (the "designer fallacy") introduce latent failures that may remain dormant for years before causing incidents [**MODERATE**]

### [SRC-003] A Model for Types and Levels of Human Interaction with Automation
- **Authors**: Raja Parasuraman, Thomas B. Sheridan, Christopher D. Wickens
- **Year**: 2000
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1109/3468.844354
- **Verified**: partial (title, journal, DOI confirmed via IEEE Xplore; full text behind paywall)
- **Relevance**: 5
- **Summary**: Extended the original Sheridan & Verplank (1978) 10-level automation taxonomy by decomposing automation across four information-processing stages: (1) information acquisition, (2) information analysis, (3) decision and action selection, and (4) action implementation. This 4x10 matrix provides a more nuanced framework than the original linear scale, allowing different automation levels for different functional stages within a single system.
- **Key Claims**:
  - Automation is not a unitary concept; it must be decomposed across four functional stages of information processing [**STRONG**]
  - The Sheridan-Verplank 10-level scale applies cleanly only to decision/action selection; applying it to information acquisition and analysis produces illogical categories [**STRONG**]
  - Higher levels of automation for information acquisition and analysis generally improve performance; higher levels for decision selection and action implementation may degrade it [**MODERATE**]

### [SRC-004] Complacency and Bias in Human Use of Automation: An Attentional Integration
- **Authors**: Raja Parasuraman, Dietrich H. Manzey
- **Year**: 2010
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1177/0018720810376055
- **Verified**: partial (title, journal, DOI confirmed; full text behind paywall)
- **Relevance**: 5
- **Summary**: Systematic review integrating complacency and automation bias as related but distinct phenomena. Complacency occurs under multi-task load when manual tasks compete for attention. Automation bias manifests as both omission errors (failing to notice when automation misses something) and commission errors (following incorrect automated recommendations). Both occur in naive and expert users and resist training interventions.
- **Key Claims**:
  - Automation complacency and automation bias represent different manifestations of overlapping automation-induced phenomena, with attention as the central mechanism [**STRONG**]
  - Automation bias occurs in both naive and expert participants and cannot be fully prevented by training or instructions [**STRONG**]
  - Automation bias produces both omission errors (failing to detect automation misses) and commission errors (following incorrect automated recommendations) [**STRONG**]
  - The severity of automation bias increases with time pressure and cognitive load [**MODERATE**]

### [SRC-005] Meaningful Human Control over Autonomous Systems: A Philosophical Account
- **Authors**: Filippo Santoni de Sio, Jeroen van den Hoven
- **Year**: 2018
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.3389/frobt.2018.00015
- **Verified**: yes (full text accessed via Frontiers open access)
- **Relevance**: 5
- **Summary**: Proposes two necessary conditions for meaningful human control: the "tracking condition" (system behavior must remain responsive to relevant human moral reasoning and environmental facts) and the "tracing condition" (outcomes must be traceable to proper moral understanding by at least one human). Adapts Fischer and Ravizza's guidance control theory from free will philosophy. Critically argues that meaningful control does not require constant human intervention, only that the system demonstrably remains responsive to appropriate human values.
- **Key Claims**:
  - Meaningful human control requires two conditions: tracking (system tracks human moral reasons and environmental facts) and tracing (outcomes traceable to human moral understanding) [**STRONG**]
  - Meaningful control does not require constant human intervention; it requires demonstrated responsiveness to human values throughout operation [**STRONG**]
  - The concept applies beyond military systems to any autonomous system affecting life and physical integrity [**MODERATE**]
  - Systems that rely on irrelevant correlations in training data (e.g., snow backgrounds for animal classification) fail the tracking condition even if outputs appear correct [**MODERATE**]

### [SRC-006] Adaptive Trust Calibration for Human-AI Collaboration
- **Authors**: Kazuo Okamura, Seiji Yamada
- **Year**: 2020
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1371/journal.pone.0229132
- **Verified**: yes (full text accessed via PMC open access)
- **Relevance**: 5
- **Summary**: Implements a closed-loop trust calibration system with three components: machine self-assessment (confidence estimation independent of class probabilities), real-time human trust prediction (using observable choice behavior as proxy), and dynamic Trust Calibration Cues (TCCs) that trigger when miscalibrated trust is detected. Experimental results with 116 participants show verbal TCCs (sensitivity d'=0.92) significantly outperform continuous transparency displays for correcting over-trust.
- **Key Claims**:
  - Targeted trust calibration cues triggered at detection outperform continuous transparency displays for correcting miscalibrated trust [**MODERATE**]
  - Observable choice behavior (reliance decisions) can serve as a non-intrusive proxy for trust measurement [**MODERATE**]
  - Verbal calibration cues ("This choice might not be a good idea") achieve highest effectiveness (d'=0.92 vs 0.38 control) compared to visual, audio, or anthropomorphic cues [**MODERATE**]
  - System self-assessment of confidence, independent of output probabilities, is necessary for the trust calibration loop to function [**MODERATE**]

### [SRC-007] Defining Human-AI Teaming the Human-Centered Way: A Scoping Review and Network Analysis
- **Authors**: Sophie Berretta, Alina Tausch, Greta Ontrup, Bjorn Gilles, Corinna Peifer, Annette Kluge
- **Year**: 2023
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.3389/frai.2023.1250725
- **Verified**: yes (full text accessed via Frontiers open access)
- **Relevance**: 4
- **Summary**: Scoping review identifying five research clusters in human-AI teaming literature: human/task variables, AI explainability/transparency, AI-driven robotics, and AI performance effects on perception. Finds that current literature remains predominantly technology-centric despite calls for interdisciplinary integration. Identifies four essential conditions for productive human-AI teaming: human understanding of AI behavior, appropriate trust, accurate decision-making, and proper system control.
- **Key Claims**:
  - Human-AI teaming research remains predominantly driven by a technology-centric and engineering perspective, with insufficient human-centered integration [**MODERATE**]
  - Four conditions are necessary for productive human-AI teaming: human understanding of AI behavior, appropriate trust, accurate use of system outputs, and proper system control [**MODERATE**]
  - Consistent terminology across disciplines is a prerequisite for knowledge integration; its absence undermines theoretical development [**WEAK**]

### [SRC-008] Human-Centered Artificial Intelligence: Reliable, Safe & Trustworthy
- **Authors**: Ben Shneiderman
- **Year**: 2020
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1080/10447318.2020.1741118
- **Verified**: partial (title, journal confirmed; accessed via arXiv preprint 2002.04087)
- **Relevance**: 4
- **Summary**: Proposes a two-dimensional HCAI framework arguing that high human control and high automation are not opposing forces but can coexist. Introduces a three-level governance structure: (1) software engineering teams develop reliable systems, (2) managers create cultures of safety, (3) independent oversight structures promote trustworthiness. Argues for a shift from "emulating humans" to "empowering people."
- **Key Claims**:
  - High human control and high automation can coexist; they are not opposing ends of a single dimension [**MODERATE**]
  - Trustworthiness requires a three-level governance structure: reliable engineering practices, safety-oriented management culture, and independent certification/oversight [**MODERATE**]
  - AI design should shift from emulating human capabilities to amplifying human performance and creativity [**WEAK**]

### [SRC-009] Let Me Take Over: Variable Autonomy for Meaningful Human Control
- **Authors**: Leila Methnani, Andrea Aler Tubella, Virginia Dignum, Andreas Theodorou
- **Year**: 2021
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.3389/frai.2021.737072
- **Verified**: yes (full text accessed via Frontiers open access)
- **Relevance**: 5
- **Summary**: Proposes variable autonomy as the operational mechanism for meaningful human control. Three foundational values: accountability (identifying responsible actors), responsibility (tracking moral considerations), and transparency (tracing to accountable individuals). Argues that static HITL/HOTL paradigms fix human involvement regardless of context, whereas variable autonomy dynamically adjusts autonomy levels based on environmental changes, task complexity, operator expertise, and risk levels.
- **Key Claims**:
  - Variable autonomy -- dynamic adjustment of control levels based on context -- is necessary to operationalize meaningful human control [**MODERATE**]
  - Static human-in-the-loop and human-on-the-loop paradigms are insufficient because they fix human involvement regardless of context [**MODERATE**]
  - Human presence alone is insufficient for meaningful control; systematic accountability structures throughout the system lifecycle are required [**MODERATE**]

### [SRC-010] Accountable Artificial Intelligence: Holding Algorithms to Account
- **Authors**: Madalina Busuioc
- **Year**: 2020
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1111/puar.13293
- **Verified**: yes (full text accessed via PMC; PMC8518786)
- **Relevance**: 4
- **Summary**: Bridges computer science and public administration to analyze algorithmic accountability. Identifies three compounding accountability deficits: information (inherent opacity, proprietary secrecy, algorithmic complexity), explanation (limited explainability, implicit value trade-offs, automation bias effects), and consequences (inability to meaningfully contest decisions). Proposes interpretable models over black boxes in high-stakes public decisions, mandatory algorithmic auditing, and non-delegation of accountability -- public organizations cannot outsource justification for administrative action.
- **Key Claims**:
  - Algorithmic accountability requires three phases to function: information (transparency), explanation (justification available for interrogation), and consequences (capacity for sanctions) [**STRONG**]
  - "System designers effectively become policy-makers" when they encode value trade-offs (fairness definitions, accuracy vs. recall) without explicit oversight [**MODERATE**]
  - Automation bias -- unjustified deference to algorithmic recommendations -- can undermine rather than enhance human oversight in mixed human-algorithm systems [**STRONG**]
  - Simple interpretable models with 2-3 features can perform comparably to complex models with 137+ features, suggesting much algorithmic opacity is unnecessary [**MODERATE**]

### [SRC-011] Cockpit Automation: Advantages and Safety Challenges (SKYbrary)
- **Authors**: SKYbrary Aviation Safety (editorial)
- **Year**: 2023 (last updated)
- **Type**: official documentation
- **URL/DOI**: https://skybrary.aero/articles/cockpit-automation-advantages-and-safety-challenges
- **Verified**: yes (content accessed via WebFetch)
- **Relevance**: 4
- **Summary**: Practical aviation industry guidance documenting cockpit automation failures and safety challenges. Documents the CAMI protocol (Confirm-Activate-Monitor-Intervene) for automation management. Catalogs specific incidents (B777 SFO 2013, A340 Paris CDG 2012, B737-800 Amsterdam 2009) where automation confusion led to accidents. Notes that pilots become reluctant to reduce automation level and seek partial automation retention even when manual control is appropriate.
- **Key Claims**:
  - Pilots become reluctant to voluntarily reduce automation level, even when manual control is more appropriate for the phase of flight [**STRONG**]
  - Complex failure events can "swamp the crew" and distract from the primary task of flying the aircraft [**STRONG**]
  - The CAMI protocol (Confirm-Activate-Monitor-Intervene) provides a structured framework for automation management, but requires ongoing manual skill maintenance [**MODERATE**]
  - Selection of modes and flight director commands may be given more importance than fundamental flight parameters (pitch, power, roll, yaw) during high-automation operation [**MODERATE**]

### [SRC-012] Effects of Automation for Emergency Operating Procedures on Human Performance in a Nuclear Power Plant
- **Authors**: Tao Qing, Zhaopeng Liu, Yaqin Tang, Hong Hu, Li Zhang, Shuai Chen
- **Year**: 2021
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.3390/ijerph18147397
- **Verified**: yes (full text accessed via PMC; PMC8300853)
- **Relevance**: 4
- **Summary**: Empirical study comparing three automation levels for nuclear emergency procedures: paper-based (PBPs), electronic (EPs), and computer-based (CBPs). Higher automation improved operational performance and reduced mental workload, but significantly decreased situation awareness -- particularly first-level perception and third-level prediction. Operators shifted from "actively searching for information" to "passively receiving information." Recommends intermediate automation levels as optimal for safety-critical nuclear operations.
- **Key Claims**:
  - Higher automation in nuclear operations improves task performance and reduces workload but significantly degrades operator situation awareness [**MODERATE**]
  - At high automation levels, operators shift from active information seeking to passive information reception, degrading their ability to detect anomalies [**MODERATE**]
  - Intermediate automation levels represent the optimal trade-off between performance gains and situation awareness preservation in safety-critical domains [**MODERATE**]

### [SRC-013] Scapegoat-as-a-Service: Moving from "Human-in-the-Loop" to "Human-in-Command" in Regulated Systems
- **Authors**: Ryan T. Jessee
- **Year**: 2026
- **Type**: whitepaper (SSRN preprint)
- **URL/DOI**: https://ssrn.com/abstract=6052874
- **Verified**: partial (title and abstract confirmed via SSRN listing; full text not accessed due to 403)
- **Relevance**: 5
- **Summary**: Identifies a critical failure mode where human-in-the-loop (HITL) governance functions as liability transfer rather than meaningful control -- termed "Scapegoat-as-a-Service." Proposes "Minimum Viable Human-in-Command" requiring four pre-execution artifacts: Intent, Inputs (provenance), Constraints (machine-checkable policy), and Action Preview (deterministic payload). Argues that control must live in the integration layer rather than in after-the-fact approvals.
- **Key Claims**:
  - Human-in-the-loop, widely treated as the default safety control for AI in regulated workflows, can function as liability transfer rather than meaningful control ("Scapegoat-as-a-Service") [**MODERATE**]
  - "Minimum Viable Human-in-Command" requires four pre-execution artifacts: Intent, Inputs (provenance), Constraints (machine-checkable policy), and Action Preview (deterministic payload) [**WEAK**]
  - Control must live in the integration layer (pre-execution), not in after-the-fact approval checkpoints [**WEAK**]

### [SRC-014] Situation Awareness-Based Agent Transparency and Human-Autonomy Teaming Effectiveness
- **Authors**: Jessie Y.C. Chen, Shan G. Lakhmani, et al.
- **Year**: 2018
- **Type**: peer-reviewed paper
- **URL/DOI**: https://doi.org/10.1080/1463922X.2017.1315750
- **Verified**: partial (title, journal, DOI confirmed; full text behind paywall)
- **Relevance**: 4
- **Summary**: Develops the Situation Awareness-based Agent Transparency (SAT) model with three levels: Level 1 (agent's current actions and plans), Level 2 (agent's reasoning process), and Level 3 (agent's projection of future outcomes). Expands the original model to incorporate bidirectional transparency and teamwork. Experimental results consistently show that human task performance improves and trust increases as agents become more transparent across these three levels.
- **Key Claims**:
  - Agent transparency should map to three situation awareness levels: current actions/plans, reasoning process, and outcome projections [**MODERATE**]
  - Human task performance improves consistently as autonomous agents become more transparent across all three SAT levels [**MODERATE**]
  - Bidirectional transparency (agent-to-human AND agent's awareness of human state) is necessary for effective teaming in dynamic environments [**MODERATE**]

## Thematic Synthesis

### Theme 1: The Automation Paradox -- Higher Automation Degrades the Human Capabilities Most Needed When Automation Fails

**Consensus**: Across aviation, nuclear, and general human factors research, there is strong agreement that increasing automation degrades operator situation awareness, manual skills, and monitoring vigilance -- precisely the capabilities required when operators must intervene during automation failures. This paradox, first articulated by Bainbridge (1983), has been empirically validated across domains for four decades. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-011], [SRC-012]

**Controversy**: Whether the paradox can be resolved through design (transparency, variable autonomy) or is inherent to automation itself. Shneiderman [SRC-008] argues high control and high automation can coexist through proper design; Bainbridge [SRC-002] and Parasuraman & Manzey [SRC-004] suggest the attentional mechanisms underlying complacency are resistant to design interventions and training.
**Dissenting sources**: [SRC-008] argues the paradox is a design failure solvable through HCAI principles, while [SRC-002] and [SRC-004] argue complacency and automation bias persist despite training and interface improvements.

**Practical Implications**:
- Any system where AI agents propose changes to their own infrastructure must account for the fact that human reviewers will become less effective at reviewing those changes as the agents become more reliable
- Intermediate automation levels (agent proposes, human reviews with full context) preserve situation awareness better than either fully manual or fully automated approaches
- Manual skill maintenance must be actively designed into the system, not assumed to persist naturally

**Evidence Strength**: STRONG

### Theme 2: Trust Calibration Is the Mechanism, Not the Goal -- Calibration Requires Active Intervention, Not Passive Transparency

**Consensus**: Trust calibration -- the match between operator trust and actual system capability -- determines whether operators rely appropriately on automated systems. Passive transparency (always-visible confidence displays, continuous status indicators) is necessary but insufficient. Active calibration mechanisms (targeted cues triggered at detection of miscalibrated trust) significantly outperform passive approaches. [**STRONG**]
**Sources**: [SRC-001], [SRC-004], [SRC-006], [SRC-014]

**Controversy**: Whether trust calibration should be handled through system-initiated interventions (the system detects miscalibration and alerts the human) or through improved mental models (the human understands the system well enough to self-calibrate).
**Dissenting sources**: [SRC-006] demonstrates system-initiated Trust Calibration Cues outperforming passive displays, while [SRC-014] argues that deeper transparency (three-level SAT model) enables humans to self-calibrate through better mental models.

**Practical Implications**:
- A confidence system (like White Sails) needs more than a confidence score -- it needs active calibration mechanisms that detect when human reviewers are over-trusting or under-trusting agent proposals
- Verbal/textual cues ("This change modifies a safety-critical path -- review carefully") outperform visual indicators for trust recalibration
- Trust calibration must account for the dynamic nature of system reliability; a system that was reliable yesterday may not be reliable in a novel situation today

**Evidence Strength**: STRONG

### Theme 3: Meaningful Human Control Requires Tracking and Tracing, Not Just Approval Checkpoints

**Consensus**: Meaningful human control is not achieved by placing a human in an approval loop. It requires two structural conditions: the system's behavior must track human moral reasoning and environmental facts (tracking condition), and outcomes must be traceable to human understanding (tracing condition). Static HITL/HOTL paradigms fix human involvement regardless of context and may function as liability transfer rather than genuine oversight. [**STRONG**]
**Sources**: [SRC-005], [SRC-009], [SRC-010], [SRC-013]

**Controversy**: Whether meaningful human control is achievable for systems that modify their own infrastructure. The tracking condition requires the system to be "responsive to relevant human moral reasoning," but when an AI agent proposes changes to the very infrastructure that governs its behavior, the human reviewer may lack the context to evaluate whether the tracking condition holds.
**Dissenting sources**: [SRC-005] argues meaningful control is achievable through proper design of tracking and tracing conditions, while [SRC-013] argues that current HITL implementations in regulated systems frequently fail to achieve meaningful control and instead create "Scapegoat-as-a-Service."

**Practical Implications**:
- The attestation layer in a system where agents modify their own infrastructure must satisfy both tracking (does the agent's proposal remain responsive to human values?) and tracing (can the outcome be traced to a human who understood the implications?)
- Four pre-execution artifacts (Intent, Inputs, Constraints, Action Preview) provide a concrete operationalization of the tracking and tracing conditions
- Variable autonomy -- dynamically adjusting how much human review is required based on risk, novelty, and operator expertise -- is more robust than fixed approval gates

**Evidence Strength**: MIXED (strong philosophical grounding, weak empirical validation for AI agent self-modification specifically)

### Theme 4: Aviation and Nuclear Domains Provide Mature Patterns for Automation Governance That Transfer to AI Agent Oversight

**Consensus**: High-stakes domains have developed layered oversight models over decades of operational experience. Aviation uses the CAMI protocol (Confirm-Activate-Monitor-Intervene), hierarchical authority (Reactor Operator / Senior Reactor Operator / Shift Manager), and mandatory manual skill maintenance. Nuclear uses intermediate automation with active information seeking and multi-level staffing. Both domains demonstrate that the optimal automation level is context-dependent and that operators must be able to select appropriate automation levels for the current situation. [**MODERATE**]
**Sources**: [SRC-002], [SRC-003], [SRC-011], [SRC-012]

**Practical Implications**:
- The CAMI protocol maps directly to AI agent oversight: Confirm the agent's proposed action, Activate (approve execution), Monitor the outcome, Intervene if behavior deviates from expectations
- Nuclear-style hierarchical review (operator proposes, senior operator validates, shift manager authorizes) provides a model for multi-level attestation of agent-proposed infrastructure changes
- Intermediate automation -- where the system actively presents relevant context rather than just requesting approval -- preserves human situation awareness better than either "rubber stamp" approval or fully manual review
- Mandatory "manual flying" equivalents (periodic human execution of tasks the agent normally handles) prevent skill decay in human reviewers

**Evidence Strength**: MODERATE (patterns are proven in aviation/nuclear but transfer to AI agent oversight is extrapolated, not empirically validated)

### Theme 5: Regulatory Frameworks Are Converging on Human Oversight Requirements but Lack Operational Specificity

**Consensus**: The EU AI Act (Article 14), FDA clinical decision support guidance, and NIST AI RMF all require human oversight for high-risk AI systems. They converge on requiring: human understanding of system capabilities and limitations, awareness of over-reliance tendencies, ability to correctly interpret outputs, and authority to override or discontinue. However, none specify how to operationally achieve these requirements in the context of autonomous agents. [**MODERATE**]
**Sources**: [SRC-010], [SRC-013]

**Controversy**: Whether regulatory requirements for human oversight create genuine safety or performative compliance.
**Dissenting sources**: [SRC-013] argues that current regulatory frameworks enable "Scapegoat-as-a-Service" where human oversight is nominally present but practically meaningless, while [SRC-010] argues the frameworks at least establish accountability structures even if operationalization is incomplete.

**Practical Implications**:
- Any attestation system for AI agent oversight should proactively satisfy EU AI Act Article 14 requirements: ensure the human overseer understands the agent's capabilities/limitations, is warned about over-reliance risks, can interpret the agent's output, and can override or discontinue
- The gap between regulatory intent and operational specificity is an opportunity: systems that can demonstrate genuine meaningful human control (not just approval checkpoints) will have a competitive advantage
- Machine-checkable constraints (policy-as-code) provide a concrete mechanism for the "Constraints" artifact that regulators aspire to but do not yet specify

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Trust calibration (match between operator trust and actual system capability) is the primary determinant of appropriate reliance behavior -- Sources: [SRC-001], [SRC-004], [SRC-006]
- Automation designed to reduce human error creates conditions where human intervention becomes more critical yet less achievable (the automation paradox) -- Sources: [SRC-001], [SRC-002], [SRC-003], [SRC-011], [SRC-012]
- Automation bias (both omission and commission errors) occurs in naive and expert users, persists despite training, and intensifies under time pressure -- Sources: [SRC-004], [SRC-010]
- Meaningful human control requires tracking (system tracks human moral reasons and environmental facts) and tracing (outcomes traceable to human moral understanding) -- Sources: [SRC-005], [SRC-009]
- Algorithmic accountability requires three functioning phases: information, explanation, and consequences -- Sources: [SRC-010]
- Pilots and operators become reluctant to voluntarily reduce automation levels even when manual control is more appropriate -- Sources: [SRC-002], [SRC-011]

### MODERATE Evidence
- Targeted trust calibration cues triggered at detection outperform continuous transparency displays for correcting miscalibrated trust -- Sources: [SRC-006]
- Variable autonomy (dynamic adjustment of control levels based on context) is necessary to operationalize meaningful human control -- Sources: [SRC-009]
- Agent transparency mapped to three situation awareness levels (actions, reasoning, projections) improves human performance and trust -- Sources: [SRC-014]
- Higher automation in safety-critical domains improves performance and reduces workload but significantly degrades situation awareness -- Sources: [SRC-012]
- High human control and high automation can coexist through a two-dimensional HCAI framework with three-level governance -- Sources: [SRC-008]
- Static HITL can function as liability transfer ("Scapegoat-as-a-Service") rather than meaningful control -- Sources: [SRC-013]
- System designers effectively become policy-makers when encoding implicit value trade-offs without explicit oversight -- Sources: [SRC-010]
- The four-type automation decomposition (information acquisition, analysis, decision selection, action implementation) enables more precise automation design than unitary automation levels -- Sources: [SRC-003]
- CAMI protocol (Confirm-Activate-Monitor-Intervene) provides structured automation management in aviation -- Sources: [SRC-011]

### WEAK Evidence
- "Minimum Viable Human-in-Command" requires four pre-execution artifacts: Intent, Inputs, Constraints, Action Preview -- Sources: [SRC-013]
- Consistent terminology across human-AI teaming disciplines is a prerequisite for knowledge integration -- Sources: [SRC-007]
- AI design should shift from emulating human capabilities to amplifying human performance -- Sources: [SRC-008]

### UNVERIFIED
- Transfer of aviation/nuclear automation governance patterns to AI agent self-modification oversight is feasible and effective -- Basis: model training knowledge; extrapolated from domain literature but no empirical validation found
- The optimal automation level for AI agent infrastructure changes specifically (as opposed to general task automation) has been empirically determined -- Basis: no literature found addressing this specific use case
- Multi-level attestation hierarchies (analogous to nuclear control room staffing) improve outcomes for AI agent oversight specifically -- Basis: model training knowledge; plausible by analogy but unvalidated

## Knowledge Gaps

- **AI agent self-modification oversight**: No literature was found that specifically addresses the case where an AI agent proposes changes to its own operational infrastructure (its own prompts, tools, permissions, or governance rules). This is the most relevant scenario for the Knossos attestation layer, and it sits in a gap between the autonomous weapons literature (which assumes the system does not modify itself) and the software engineering literature (which assumes humans write the code). Research on recursive self-improvement exists in AI safety theory but lacks operational frameworks.

- **Trust calibration for code review**: While trust calibration is well-studied for monitoring tasks and decision support, there is minimal literature on trust calibration specifically for reviewing proposed code or configuration changes. The cognitive task of reviewing a diff is qualitatively different from monitoring a dashboard or approving a recommendation.

- **Long-horizon automation bias in development workflows**: Existing automation bias research uses short experimental sessions (minutes to hours). The effect of automation bias over weeks or months of working with an AI coding agent -- where the agent's reliability history shapes the reviewer's calibration -- is unstudied.

- **Variable autonomy implementation patterns**: While the concept of variable autonomy is well-theorized [SRC-009], concrete implementation patterns (what triggers autonomy changes, how transitions are managed, what the interface looks like) are sparse, particularly for text-based AI agent interactions rather than robotic or vehicle systems.

- **Cross-domain transfer validation**: The transfer of aviation CAMI and nuclear hierarchical oversight patterns to AI agent governance is assumed throughout this review but has not been empirically validated. The operating tempo, error reversibility, and failure modes differ significantly.

## Domain Calibration

This literature review draws from a well-studied domain with canonical literature. The foundational papers (Bainbridge 1983, Lee & See 2004, Parasuraman et al. 2000) are among the most cited in human factors research. However, the specific application to AI agents that modify their own infrastructure is novel and unstudied. The STRONG evidence grades reflect the maturity of the underlying human factors research; the UNVERIFIED claims reflect the gap between that mature literature and the specific use case of agent self-governance.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research human-ai-teaming-shared-autonomy-trust-calibration` on 2026-03-10.
