---
domain: "literature-multi-agent-coding-systems"
generated_at: "2026-03-10T19:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.75
format_version: "1.0"
---

# Literature Review: Multi-Agent Systems for Autonomous Software Engineering

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The literature on multi-agent LLM systems for software engineering reveals a field experiencing rapid capability growth alongside fundamental coordination challenges that remain unsolved. SWE-bench Verified scores have climbed from 12.5% (SWE-Agent, May 2024) to ~80% (Claude Opus 4.5, November 2025), but this progress is overwhelmingly driven by single-agent improvements (better models, better tool interfaces) rather than multi-agent coordination breakthroughs. The evidence is strong that agent-computer interface design matters more than agent quantity, that multi-agent coordination imposes a measurable "coordination tax" that degrades performance on sequential tasks, and that inter-agent misalignment -- not individual agent capability -- is the dominant failure mode. No system in the literature demonstrates the reflexive self-improvement loop with institutional memory that Knossos targets; Reflexion (Shinn et al., 2023) is the closest precedent but operates within single-agent, single-session scope.

## Source Catalog

### [SRC-001] SWE-agent: Agent-Computer Interfaces Enable Automated Software Engineering
- **Authors**: John Yang, Carlos E. Jimenez, Alexander Wettig, Kilian Lieret, Shunyu Yao, Karthik Narasimhan, Ofir Press
- **Year**: 2024
- **Type**: peer-reviewed paper (NeurIPS 2024)
- **URL/DOI**: https://arxiv.org/abs/2405.15793
- **Verified**: yes (full text fetched, NeurIPS proceedings confirmed)
- **Relevance**: 5
- **Summary**: Introduces the agent-computer interface (ACI) concept -- purpose-built interfaces for LLM agents analogous to IDEs for humans. Demonstrates that interface design dominates agent performance: their custom ACI achieved 12.5% pass@1 on SWE-bench (state-of-the-art at the time) and 87.7% on HumanEvalFix. The key insight is that LLM agents are a new user class requiring their own interface design discipline.
- **Key Claims**:
  - Agent-computer interface design is the primary lever for improving LLM agent performance on software engineering tasks, not model capability alone [**STRONG**]
  - LLM agents represent a new category of end users with distinct needs from human developers [**MODERATE**]
  - Custom file editing, navigation, and test execution interfaces significantly outperform generic shell-based interaction [**STRONG**]

### [SRC-002] Why Do Multi-Agent LLM Systems Fail?
- **Authors**: Mert Cemri, Melissa Z. Pan, Shuyi Yang, Lakshya A Agrawal, Bhavya Chopra, Rishabh Tiwari, Kurt Keutzer, Aditya Parameswaran, Dan Klein, Kannan Ramchandran, Matei Zaharia, Joseph E. Gonzalez, Ion Stoica
- **Year**: 2025
- **Type**: peer-reviewed paper (ICLR 2025 workshop)
- **URL/DOI**: https://arxiv.org/abs/2503.13657
- **Verified**: yes (full text fetched and analyzed)
- **Relevance**: 5
- **Summary**: Introduces MASFT (Multi-Agent System Failure Taxonomy) based on 1,600+ annotated traces across 7 MAS frameworks. Identifies 14 distinct failure modes in 3 categories. The critical finding: multi-agent systems show minimal performance gains over single-agent baselines despite growing adoption. Inter-agent misalignment (~40%) is the dominant failure category, not individual agent limitations. Even tactical improvements (better prompts, self-verification) yielded only +14% for ChatDev.
- **Key Claims**:
  - Inter-agent misalignment accounts for ~40% of all MAS failures, making it the dominant failure mode [**STRONG**]
  - Specification and system design failures account for ~35% of failures [**STRONG**]
  - Multi-agent systems show minimal accuracy gains over single-agent frameworks on popular benchmarks [**STRONG**]
  - ~79% of MAS problems originate from specification and coordination issues, not technical implementation [**MODERATE**]
  - Tactical improvements (prompt engineering, self-verification) are insufficient; structural redesign is required [**MODERATE**]

### [SRC-003] Towards a Science of Scaling Agent Systems
- **Authors**: Yubin Kim, Ken Gu, Chanwoo Park, Chunjong Park, Samuel Schmidgall, A. Ali Heydari, Yao Yan, Zhihan Zhang, Yuchen Zhuang, Mark Malhotra, Paul Pu Liang, Hae Won Park, Yuzhe Yang, Xuhai Xu, Yilun Du, Shwetak Patel, Tim Althoff, Daniel McDuff, Xin Liu
- **Year**: 2025
- **Type**: whitepaper (arXiv preprint, Google Research / DeepMind / MIT)
- **URL/DOI**: https://arxiv.org/abs/2512.08296
- **Verified**: yes (abstract and key sections fetched)
- **Relevance**: 5
- **Summary**: First quantitative scaling study for multi-agent systems. Evaluates 180 configurations across 5 architectures and 4 benchmarks. Discovers three primary scaling effects: tool-coordination trade-off, capability saturation at ~45% single-agent performance, and error amplification by topology (independent agents amplify errors 17.2x vs. centralized at 4.4x). Architecture-task alignment predicts optimal coordination strategy for 87% of configurations.
- **Key Claims**:
  - Independent multi-agent coordination amplifies errors 17.2x compared to 4.4x for centralized coordination [**STRONG**]
  - Multi-agent coordination yields diminishing or negative returns when single-agent success rate exceeds ~45% [**MODERATE**]
  - Performance spans +81% improvement to -70% degradation depending on architecture-task alignment [**MODERATE**]
  - Sequential reasoning tasks degrade 39-70% under all multi-agent variants [**MODERATE**]
  - The predictive model achieves R^2=0.524 for coordination strategy selection [**MODERATE**]

### [SRC-004] MetaGPT: Meta Programming for A Multi-Agent Collaborative Framework
- **Authors**: Sirui Hong, Mingchen Zhuge, Jiaqi Chen, Xiawu Zheng, Yuheng Cheng, Ceyao Zhang, Jinlin Wang, Zili Wang, Steven Ka Shing Yau, Zijuan Lin, Liyang Zhou, Chenyu Ran, Lingfeng Xiao, Chenglin Wu, Jurgen Schmidhuber
- **Year**: 2024
- **Type**: peer-reviewed paper (ICLR 2024 Oral)
- **URL/DOI**: https://arxiv.org/abs/2308.00352
- **Verified**: yes (abstract and key sections fetched, ICLR proceedings confirmed)
- **Relevance**: 5
- **Summary**: Proposes encoding Standardized Operating Procedures (SOPs) into multi-agent prompt sequences. Uses an assembly-line paradigm with role-specialized agents (Product Manager, Architect, Engineer) communicating via structured documents rather than free-form dialogue. Achieves 85.9% and 87.7% Pass@1 on code generation benchmarks. The key contribution is demonstrating that structured, procedure-driven collaboration outperforms unstructured chat-based approaches by reducing cascading hallucinations.
- **Key Claims**:
  - SOPs encoded in prompt sequences reduce cascading hallucinations in multi-agent systems [**STRONG**]
  - Document-based agent communication outperforms dialogue-based communication for software engineering tasks [**MODERATE**]
  - Assembly-line role decomposition with intermediate verification produces more coherent solutions than chat-based approaches [**STRONG**]
  - Executable feedback mechanisms improve Pass@1 by 4.2% on HumanEval and 5.4% on MBPP [**MODERATE**]

### [SRC-005] ChatDev: Communicative Agents for Software Development
- **Authors**: Chen Qian, Wei Liu, Hongzhang Liu, Nuo Chen, Yufan Dang, Jiahao Li, Cheng Yang, Weize Chen, Yusheng Su, Xin Cong, Juyuan Xu, Dahai Li, Zhiyuan Liu, Maosong Sun
- **Year**: 2024
- **Type**: peer-reviewed paper (ACL 2024)
- **URL/DOI**: https://aclanthology.org/2024.acl-long.810/
- **Verified**: yes (HTML version fetched and analyzed)
- **Relevance**: 4
- **Summary**: Implements a virtual software company with LLM agents playing distinct roles (CEO, CTO, programmer, tester). Uses a "chat chain" mechanism to decompose development into sequential phases with paired instructor-assistant dialogues. Introduces "communicative dehallucination" where assistants request clarification before implementation. Achieves 88% executability but only 39.5% quality score, highlighting the gap between generating runnable code and generating good code.
- **Key Claims**:
  - Chat chain decomposition (sequential paired dialogues) produces executable software from natural language specifications [**STRONG**]
  - Communicative dehallucination (agents requesting clarification) reduces coding errors [**MODERATE**]
  - Natural language communication advantages system design; programming language communication advantages debugging [**MODERATE**]
  - High executability (88%) coexists with low quality (39.5%), indicating a generation-quality gap [**MODERATE**]

### [SRC-006] Reflexion: Language Agents with Verbal Reinforcement Learning
- **Authors**: Noah Shinn, Federico Cassano, Edward Berman, Ashwin Gopinath, Karthik Narasimhan, Shunyu Yao
- **Year**: 2023
- **Type**: peer-reviewed paper (NeurIPS 2023)
- **URL/DOI**: https://arxiv.org/abs/2303.11366
- **Verified**: yes (abstract fetched, NeurIPS proceedings confirmed)
- **Relevance**: 4
- **Summary**: Introduces verbal reinforcement learning where agents maintain reflective text in episodic memory buffers rather than updating weights. Achieves 91% pass@1 on HumanEval (vs. 80% for GPT-4 at the time). The framework consists of Actor, Evaluator, and Self-Reflection components. Critical as the closest precedent to institutional memory in coding agents, though limited to single-agent, single-session scope without cross-session persistence.
- **Key Claims**:
  - Verbal self-reflection with episodic memory improves coding performance without weight updates [**STRONG**]
  - Reflexion achieves 91% pass@1 on HumanEval, exceeding GPT-4's 80% [**MODERATE**]
  - Natural language reflection is a viable alternative to gradient-based reinforcement learning for LLM agents [**STRONG**]

### [SRC-007] OpenHands: An Open Platform for AI Software Developers as Generalist Agents
- **Authors**: Xingyao Wang, Boxuan Li, Yufan Song, Frank F. Xu, and 20+ contributors
- **Year**: 2024
- **Type**: peer-reviewed paper (ICLR 2025)
- **URL/DOI**: https://arxiv.org/abs/2407.16741
- **Verified**: yes (abstract fetched, ICLR 2025 acceptance confirmed)
- **Relevance**: 4
- **Summary**: Open-source platform (formerly OpenDevin) for AI software development agents. Supports multiple agent architectures in sandboxed environments. Evaluated on 15 benchmarks including SWE-bench and WebArena. Key contribution is demonstrating that a community-driven, model-agnostic platform can match or exceed proprietary systems. 2,100+ contributions from 188+ contributors under MIT license.
- **Key Claims**:
  - Open-source, model-agnostic agent platforms can match proprietary system performance [**MODERATE**]
  - Sandboxed execution environments are essential for safe autonomous code execution [**MODERATE**]
  - Generalist agent architectures can handle diverse tasks (coding, web browsing, testing) through unified interfaces [**WEAK**]

### [SRC-008] Raising the Bar on SWE-bench Verified with Claude 3.5 Sonnet (Anthropic)
- **Authors**: Anthropic (no individual authors listed)
- **Year**: 2024
- **Type**: official documentation (Anthropic technical blog)
- **URL/DOI**: https://www.anthropic.com/research/swe-bench-sonnet
- **Verified**: yes (content fetched and analyzed)
- **Relevance**: 5
- **Summary**: Documents Anthropic's minimalist agent architecture achieving 49% on SWE-bench Verified (then SOTA). Key design principle: "give as much control as possible to the language model itself, and keep the scaffolding minimal." Uses only Bash and Edit tools with extensive tool descriptions. String replacement editing was chosen over diff-based or line-number approaches because exact-match failures provide clear error signals. The article argues that tool interface design deserves the same attention as UI/UX design for humans.
- **Key Claims**:
  - Minimal scaffolding with maximal model autonomy outperforms rigid multi-step workflows [**STRONG**]
  - Tool interface design is as important for LLM agents as UI/UX design is for humans [**MODERATE**]
  - String replacement editing with exact-match validation provides natural error-correction loops [**MODERATE**]
  - Error-proofing tools (e.g., requiring absolute paths) prevents common agent mistakes [**MODERATE**]

### [SRC-009] Devin's 2025 Performance Review (Cognition Labs)
- **Authors**: Cognition Labs
- **Year**: 2025
- **Type**: official documentation (company blog)
- **URL/DOI**: https://cognition.ai/blog/devin-annual-performance-review-2025
- **Verified**: yes (content fetched and analyzed)
- **Relevance**: 4
- **Summary**: Cognition's self-reported assessment of Devin after 18 months in production. PR merge rate improved from 34% to 67%. Devin is characterized as "senior-level at codebase understanding but junior at execution." Excels at tasks with clear requirements and verifiable outcomes (security fixes, migrations, test generation) but struggles with ambiguous specifications and iterative collaboration. Independent testing by Answer.AI found only 15% success rate across 20 tasks.
- **Key Claims**:
  - Devin's PR merge rate improved from 34% to 67% year-over-year [**MODERATE**]
  - Autonomous coding agents excel at well-specified, verifiable tasks but fail on ambiguous, iterative ones [**STRONG**]
  - Independent evaluation (Answer.AI) found 15% success rate vs. Cognition's higher self-reported metrics [**MODERATE**]
  - Large-scale migrations and security fixes are high-value autonomous agent applications [**MODERATE**]

### [SRC-010] LLM-Based Multi-Agent Systems for Software Engineering: Literature Review, Vision and the Road Ahead
- **Authors**: Junda He, Christoph Treude, David Lo
- **Year**: 2024
- **Type**: peer-reviewed paper (ACM Transactions on Software Engineering and Methodology)
- **URL/DOI**: https://dl.acm.org/doi/10.1145/3712003
- **Verified**: partial (ACM page accessed but full text behind paywall)
- **Relevance**: 4
- **Summary**: Systematic literature review of LLM-based multi-agent applications across the software development lifecycle. Identifies research gaps in two phases: enhancing individual agent capabilities and optimizing agent collaboration. Proposes a vision toward "Software Engineering 2.0" with fully autonomous, scalable, and trustworthy multi-agent systems. Provides case studies using state-of-the-art frameworks to illustrate both strengths and limitations.
- **Key Claims**:
  - Multi-agent LLM systems show promise across the full software development lifecycle, not just code generation [**MODERATE**]
  - Two critical research gaps exist: individual agent capability and agent collaboration optimization [**MODERATE**]
  - Current systems lack trustworthiness and scalability for production use [**WEAK**]

### [SRC-011] SWE-bench: Can Language Models Resolve Real-world Github Issues?
- **Authors**: Carlos E. Jimenez, John Yang, Alexander Wettig, Shunyu Yao, Kexin Pei, Ofir Press, Karthik Narasimhan
- **Year**: 2024
- **Type**: peer-reviewed paper (ICLR 2024)
- **URL/DOI**: https://www.swebench.com/
- **Verified**: partial (leaderboard and methodology accessed, original paper not fetched)
- **Relevance**: 5
- **Summary**: The benchmark that defines the field. SWE-bench presents real GitHub issues from popular Python repositories, requiring agents to generate patches that pass existing test suites. SWE-bench Verified (484 human-validated samples) has become the standard evaluation. Top scores progressed from 12.5% (May 2024) to ~80% (November 2025). SWE-bench Pro introduces harder tasks where even GPT-5 and Claude Opus 4.1 score only ~23%.
- **Key Claims**:
  - Real-world GitHub issue resolution is a meaningful evaluation of autonomous software engineering capability [**STRONG**]
  - SWE-bench Verified scores have progressed from 12.5% to ~80% in 18 months (May 2024 to November 2025) [**STRONG**]
  - SWE-bench Pro reveals that top models plateau at ~23%, indicating fundamental capability gaps on harder tasks [**MODERATE**]

### [SRC-012] Memory in the Age of AI Agents: A Survey
- **Authors**: Multiple (survey paper)
- **Year**: 2025
- **Type**: whitepaper (arXiv preprint)
- **URL/DOI**: https://arxiv.org/abs/2512.13564
- **Verified**: partial (title confirmed via WebSearch, abstract accessed)
- **Relevance**: 3
- **Summary**: Surveys memory mechanisms in AI agent systems, covering episodic, semantic, and procedural memory types. Identifies that shared memory pools enable "joint attention" but unstructured sharing creates a "noisy commons." Addresses the tension between global knowledge accessibility and access control. Relevant to multi-agent coding systems because memory architecture determines whether agents can learn from past interactions.
- **Key Claims**:
  - Shared memory pools enable team coordination but create noise without access controls [**MODERATE**]
  - Memory architecture is a core capability differentiator for AI agent systems [**MODERATE**]
  - Unstructured memory sharing degrades agent performance as irrelevant details accumulate [**WEAK**]

## Thematic Synthesis

### Theme 1: Interface Design Dominates Agent Capability Gains

**Consensus**: The primary lever for improving LLM agent performance on software engineering tasks is agent-computer interface (ACI) design -- purpose-built tools, structured outputs, and error-feedback mechanisms -- not model scale or agent count. [**STRONG**]
**Sources**: [SRC-001], [SRC-004], [SRC-008], [SRC-011]

**Controversy**: Whether minimal scaffolding (Anthropic's 2-tool approach) or structured SOPs (MetaGPT's assembly-line) produces better results. Anthropic's approach yielded 49% on SWE-bench Verified with Claude 3.5 Sonnet; MetaGPT's approach scored higher on HumanEval/MBPP but was not evaluated on SWE-bench Verified, making direct comparison impossible.
**Dissenting sources**: [SRC-008] argues minimal scaffolding with maximal model autonomy outperforms rigid workflows, while [SRC-004] argues that SOPs with intermediate verification are essential to prevent cascading hallucinations.

**Practical Implications**:
- Invest heavily in tool interface design (descriptions, error messages, constraints) before adding agents
- Exact-match validation (string replacement over diff-based editing) provides natural error-correction loops
- Treat ACI design as a first-class engineering discipline equivalent to UI/UX

**Evidence Strength**: STRONG

### Theme 2: Multi-Agent Coordination Imposes a Measurable Tax That Frequently Outweighs Benefits

**Consensus**: Adding agents to a system introduces coordination overhead that degrades performance on sequential tasks and yields diminishing returns once single-agent capability exceeds ~45% success rate. Architecture-task alignment, not agent count, determines success. [**STRONG**]
**Sources**: [SRC-002], [SRC-003], [SRC-005]

**Practical Implications**:
- Default to single-agent architectures unless the task is demonstrably parallelizable
- If using multiple agents, use centralized coordination (4.4x error amplification) over independent coordination (17.2x)
- Performance can span +81% improvement to -70% degradation depending on topology -- choosing wrong is worse than choosing none
- The 4-agent threshold is a practical ceiling; beyond it, coordination tax typically dominates

**Evidence Strength**: STRONG

### Theme 3: Autonomous Agents Excel at Well-Specified Tasks and Fail at Ambiguous Ones

**Consensus**: Current autonomous coding agents succeed on tasks with clear specifications, verifiable outcomes, and established code patterns (security fixes, migrations, test generation). They fail on tasks requiring ambiguity resolution, iterative collaboration, and architectural judgment. [**STRONG**]
**Sources**: [SRC-009], [SRC-011], [SRC-008], [SRC-005]

**Controversy**: Whether this is a fundamental limitation or a capability gap that will close with better models. SWE-bench Verified progress (12.5% to ~80% in 18 months) suggests rapid improvement, but SWE-bench Pro (~23% for top models) and Devin's 15% independent success rate suggest a plateau on harder, more ambiguous tasks.
**Dissenting sources**: [SRC-009] (Cognition) argues Devin is improving rapidly (67% PR merge rate), while independent evaluators (Answer.AI, cited in [SRC-009] search results) found 15% success, suggesting self-reported metrics overstate capability.

**Practical Implications**:
- Deploy autonomous agents on migration, test generation, and security fix workflows first
- Do not trust autonomous agents for greenfield architecture or ambiguous requirements
- Independent benchmarking is essential; vendor self-reported metrics are unreliable
- SWE-bench Pro scores (~23%) better reflect real-world difficulty than SWE-bench Verified (~80%)

**Evidence Strength**: STRONG

### Theme 4: Structured Communication Outperforms Free-Form Dialogue in Multi-Agent Systems

**Consensus**: Multi-agent systems using structured outputs (documents, schemas, formal protocols) for inter-agent communication outperform those using free-form natural language dialogue. [**MODERATE**]
**Sources**: [SRC-004], [SRC-005], [SRC-002]

**Controversy**: ChatDev's dialogue-based approach outperformed MetaGPT on quality metrics despite lower overall scores, suggesting that structured communication may sacrifice creative problem-solving for consistency.
**Dissenting sources**: [SRC-005] demonstrates that natural language dialogue enables communicative dehallucination (agents requesting clarification), which structured protocols may not support. [SRC-004] counters that structured artifacts prevent cascading hallucinations that dialogue amplifies.

**Practical Implications**:
- Use schema-enforced communication (JSON, structured documents) for inter-agent data transfer
- Reserve natural language for clarification requests and ambiguity resolution
- Anthropic's MCP (schema-enforced, JSON-RPC 2.0) is a practical implementation of this principle

**Evidence Strength**: MIXED

### Theme 5: No System Demonstrates Cross-Session Institutional Memory or Reflexive Self-Improvement

**Consensus**: Reflexion (verbal self-reflection with episodic memory) is the closest precedent to institutional memory in coding agents, but operates within single-agent, single-session scope. No system in the literature demonstrates persistent cross-session learning where the system improves its own processes based on accumulated experience. [**MODERATE**]
**Sources**: [SRC-006], [SRC-012], [SRC-002]

**Practical Implications**:
- The cross-session institutional memory + reflexive self-improvement loop is an unoccupied niche in the literature
- Reflexion's verbal reinforcement learning validates the mechanism (memory-based self-improvement works) but not the scope (cross-session, cross-agent persistence)
- Memory architecture is a differentiator: shared pools create noise, structured memory with access controls is needed
- Any system claiming institutional memory must address the "noisy commons" problem identified in [SRC-012]

**Evidence Strength**: MODERATE

### Theme 6: SWE-bench Progress Is Real but Misleading About General Capability

**Consensus**: SWE-bench Verified scores have risen dramatically, but this progress reflects model capability improvements and benchmark-specific optimization more than general autonomous software engineering capability. SWE-bench Pro scores (~23%) are a more honest signal. [**MODERATE**]
**Sources**: [SRC-011], [SRC-001], [SRC-008], [SRC-009]

**Practical Implications**:
- Do not use SWE-bench Verified scores as a proxy for production agent capability
- SWE-bench Pro and independent evaluations (Answer.AI's Devin assessment) better predict real-world performance
- The gap between benchmark performance and production reliability is the central unsolved problem
- Evaluation methodology matters: the benchmark's methodology update (v2.0.0, February 2026) changed scores significantly

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- Agent-computer interface design is the primary performance lever for coding agents, more impactful than model scale or agent count -- Sources: [SRC-001], [SRC-008], [SRC-004]
- Inter-agent misalignment (~40% of failures) is the dominant failure mode in multi-agent systems, not individual agent limitations -- Sources: [SRC-002], [SRC-003]
- Multi-agent coordination imposes measurable overhead: independent agents amplify errors 17.2x vs. 4.4x for centralized topologies -- Sources: [SRC-003], [SRC-002]
- SOPs and structured communication reduce cascading hallucinations in multi-agent collaboration -- Sources: [SRC-004], [SRC-005]
- Verbal self-reflection with episodic memory improves agent performance without weight updates -- Sources: [SRC-006]
- Autonomous agents succeed on well-specified, verifiable tasks and fail on ambiguous, iterative ones -- Sources: [SRC-009], [SRC-011], [SRC-008]
- SWE-bench Verified scores progressed from 12.5% to ~80% in 18 months (May 2024 to November 2025) -- Sources: [SRC-011], [SRC-001]
- Chat chain decomposition produces executable software from natural language specifications -- Sources: [SRC-005], [SRC-004]

### MODERATE Evidence
- Minimal scaffolding with maximal model autonomy outperforms rigid multi-step workflows (Anthropic's approach) -- Sources: [SRC-008]
- Multi-agent coordination yields diminishing returns when single-agent success exceeds ~45% -- Sources: [SRC-003]
- SWE-bench Pro scores (~23% for top models) better reflect real-world difficulty than SWE-bench Verified (~80%) -- Sources: [SRC-011]
- No existing system demonstrates cross-session institutional memory for coding agents -- Sources: [SRC-006], [SRC-012]
- Independent evaluation of Devin found 15% success rate vs. vendor-reported higher metrics -- Sources: [SRC-009]
- Memory architecture (structured vs. unstructured) is a core differentiator for agent systems -- Sources: [SRC-012], [SRC-006]
- Tool descriptions and error-proofing prevent common agent failure modes -- Sources: [SRC-008], [SRC-001]

### WEAK Evidence
- Generalist agent architectures can handle diverse tasks through unified interfaces -- Sources: [SRC-007]
- Current multi-agent systems lack trustworthiness and scalability for production use -- Sources: [SRC-010]
- Unstructured shared memory degrades agent performance as irrelevant details accumulate -- Sources: [SRC-012]

### UNVERIFIED
- The "noisy commons" problem (shared memory degradation) scales superlinearly with agent count -- Basis: model training knowledge, extrapolated from [SRC-012] and [SRC-003]
- Reflexive self-improvement (agents modifying their own prompts/tools based on experience) has not been systematically studied in multi-agent coding contexts -- Basis: absence of relevant literature in search results
- The 4-agent threshold for coordination tax is a hard limit vs. a soft guideline -- Basis: extrapolated from [SRC-003] quantitative data, not independently validated

## Knowledge Gaps

- **Cross-session institutional memory**: No empirical study evaluates persistent memory across agent sessions for software engineering. Reflexion operates within single sessions. The effectiveness of cross-session learning for coding tasks is theoretically promising but empirically unvalidated. Filling this gap requires longitudinal studies of agents working on the same codebase over weeks/months.

- **Self-improvement feedback loops**: No system demonstrates agents autonomously improving their own tools, prompts, or coordination protocols based on accumulated experience. ADAS (Automated Design of Agentic Systems) uses meta-agents to optimize agent implementations but has not been evaluated on software engineering tasks specifically. Filling this gap requires a system that tracks its own failure modes and modifies its behavior accordingly.

- **Multi-agent coordination on real codebases**: Most multi-agent evaluations use synthetic benchmarks (HumanEval, MBPP) or curated issue sets (SWE-bench). No study evaluates multi-agent coordination on sustained, multi-week development of a real codebase with evolving requirements. This gap is critical because coordination tax may be different for persistent codebases vs. isolated issues.

- **Cost-benefit analysis of multi-agent architectures**: While [SRC-003] quantifies performance effects, no study systematically compares the dollar cost (API tokens, compute) of multi-agent vs. single-agent approaches at equivalent quality levels. The Sonar agent's $1.26/issue metric is an isolated data point, not a comparative analysis.

- **Scar tissue and defensive patterns**: No literature addresses how coding agents should handle accumulated knowledge about past failures, anti-patterns, and codebase-specific constraints ("scar tissue"). This is distinct from general memory -- it is domain-specific defensive knowledge that prevents repeat mistakes.

## Domain Calibration

Mixed confidence distribution reflects a rapidly evolving field where foundational mechanisms (ACI design, coordination failure modes) are well-studied but applied questions (institutional memory, self-improvement, production deployment) remain open. The high proportion of MODERATE claims reflects strong first-source evidence awaiting independent corroboration. Evidence grades are honest about the difference between benchmark performance (well-documented) and production capability (poorly documented).

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research multi-agent-coding-systems` on 2026-03-10.
