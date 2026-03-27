---
domain: "literature-agent-memory-architecture-progressive-api-discovery"
generated_at: "2026-03-26T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
generator: bibliotheca
confidence: 0.65
format_version: "1.0"
---

# Literature Review: Agent Memory Architecture and Progressive API Discovery

> LLM-synthesized literature review. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

The intersection of agent memory architecture and progressive API/tool discovery is an active research area spanning academic publications and industrial practice. The literature converges on a core insight: LLM agents require tiered memory hierarchies (analogous to OS virtual memory) to manage persistent knowledge, and progressive disclosure of tool/API schemas is essential to avoid context window saturation. Strong evidence supports the OS-inspired memory hierarchy pattern (MemGPT/Letta), while moderate-to-strong evidence supports progressive disclosure as the dominant pattern for scalable tool exposure (Anthropic Agent Skills, MCP meta-tool pattern, Speakeasy dynamic toolsets). The field is evolving rapidly, with 2025-2026 producing the first comprehensive surveys and production implementations, though rigorous controlled studies remain sparse.

## Source Catalog

### [SRC-001] MemGPT: Towards LLMs as Operating Systems
- **Authors**: Charles Packer, Sarah Wooders, Kevin Lin, Vivian Fang, Shishir G. Patil, Ion Stoica, Joseph E. Gonzalez
- **Year**: 2023
- **Type**: peer-reviewed paper (ICLR 2024, originally arXiv October 2023)
- **URL/DOI**: https://arxiv.org/abs/2310.08560
- **Verified**: yes (full text fetched from arXiv, widely cited)
- **Relevance**: 5
- **Summary**: Introduces virtual context management for LLM agents, directly inspired by OS memory hierarchies. Proposes a two-tier architecture with main context (in-context, analogous to RAM) and external context (out-of-context, analogous to disk), with the agent using function calls to move data between tiers. Evaluated on document analysis exceeding context windows and multi-session chat with persistent memory. The system uses interrupts to manage control flow.
- **Key Claims**:
  - LLM context windows can be extended via OS-inspired virtual memory management with tiered storage [**STRONG**]
  - Agents can self-manage their own memory through function calls that page data between tiers [**STRONG**]
  - Multi-session conversational agents require persistent memory to maintain coherent long-term interactions [**MODERATE**]

### [SRC-002] Memory in the Age of AI Agents: A Survey
- **Authors**: Yuyang Hu, Shichun Liu, Yanwei Yue, Guibin Zhang, Boyang Liu, et al. (47 co-authors)
- **Year**: 2025 (December 2025, revised January 2026)
- **Type**: peer-reviewed paper (arXiv preprint, under review)
- **URL/DOI**: https://arxiv.org/abs/2512.13564
- **Verified**: yes (abstract and taxonomy fetched from arXiv)
- **Relevance**: 5
- **Summary**: The most comprehensive survey on agent memory to date. Proposes a three-dimensional taxonomy distinguishing memory by forms (token-level, parametric, latent), functions (factual, experiential, working), and dynamics (formation, evolution, retrieval). Argues that existing long/short-term memory frameworks are insufficient for capturing contemporary agent memory diversity. Catalogs benchmarks and open-source frameworks.
- **Key Claims**:
  - Agent memory should be taxonomized along three dimensions: forms, functions, and dynamics [**MODERATE**]
  - Token-level, parametric, and latent memory represent the three dominant realization forms [**MODERATE**]
  - Factual, experiential, and working memory are the three functional categories [**MODERATE**]
  - Existing long/short-term frameworks are insufficient for modern agent memory systems [**MODERATE**]

### [SRC-003] A Survey on the Memory Mechanism of Large Language Model based Agents
- **Authors**: Zeyu Zhang, Xiaohe Bo, Chen Ma, Rui Li, Xu Chen, Quanyu Dai, Jieming Zhu, Zhenhua Dong, Ji-Rong Wen
- **Year**: 2024
- **Type**: peer-reviewed paper (ACM Transactions on Information Systems, arXiv April 2024)
- **URL/DOI**: https://arxiv.org/abs/2404.13501
- **Verified**: partial (abstract fetched, full taxonomy not accessible in fetch)
- **Relevance**: 4
- **Summary**: Systematic survey examining how memory systems enable LLM agents to evolve and interact with their environments over extended periods. Reviews design and evaluation methodologies for memory modules, covering real-world applications and limitations. Published in ACM TOIS, providing peer-reviewed validation of the field's core concepts.
- **Key Claims**:
  - Memory is the foundational component supporting meaningful agent-environment interactions [**STRONG**]
  - Memory modules require systematic design and evaluation methodologies [**MODERATE**]

### [SRC-004] A-MEM: Agentic Memory for LLM Agents
- **Authors**: Wujiang Xu, Zujie Liang, Kai Mei, Hang Gao, Juntao Tan, Yongfeng Zhang
- **Year**: 2025 (NeurIPS 2025)
- **URL/DOI**: https://arxiv.org/abs/2502.12110
- **Type**: peer-reviewed paper (NeurIPS 2025)
- **Verified**: yes (abstract fetched, NeurIPS acceptance confirmed)
- **Relevance**: 4
- **Summary**: Introduces a Zettelkasten-inspired agentic memory system where each memory unit is enriched with LLM-generated keywords, tags, and contextual descriptions. The system dynamically creates interconnected knowledge networks through indexing and linking, with memory evolution triggered by new memories updating existing entries. Achieves superior performance over baselines across six foundation models.
- **Key Claims**:
  - Zettelkasten-style interconnected note structures outperform flat memory stores for LLM agents [**MODERATE**]
  - Memory evolution (new memories updating existing ones) is critical for knowledge network quality [**MODERATE**]
  - Agent-driven dynamic indexing and linking creates more adaptive memory than static organization [**MODERATE**]

### [SRC-005] Mem0: Building Production-Ready AI Agents with Scalable Long-Term Memory
- **Authors**: Prateek Chhikara, Dev Khant, Saket Aryan, Taranjeet Singh, Deshraj Yadav
- **Year**: 2025 (ECAI 2025)
- **Type**: peer-reviewed paper (European Conference on AI)
- **URL/DOI**: https://arxiv.org/abs/2504.19413
- **Verified**: yes (abstract and claims fetched from arXiv and mem0.ai/research)
- **Relevance**: 4
- **Summary**: Presents a two-phase memory pipeline (extraction + update) for production agent memory. Base Mem0 uses vector similarity; enhanced Mem0-g adds graph-based stores for relational structure. Reports 26% accuracy improvement over OpenAI's memory on LOCOMO benchmark (66.9% vs 52.9%), 91% latency reduction, and 90% token savings versus full-context approaches.
- **Key Claims**:
  - Dynamic memory extraction and consolidation across sessions outperforms full-context approaches [**STRONG**]
  - Graph-based memory representations capture richer multi-session relationships than flat vector stores [**MODERATE**]
  - Memory systems achieve 90%+ token savings compared to full conversation replay [**MODERATE**]

### [SRC-006] Effective Context Engineering for AI Agents
- **Authors**: Prithvi Rajasekaran, Ethan Dixon, Carly Ryan, Jeremy Hadfield (Anthropic Applied AI)
- **Year**: 2025 (September 29, 2025)
- **Type**: official documentation (Anthropic engineering blog)
- **URL/DOI**: https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
- **Verified**: yes (full content fetched)
- **Relevance**: 5
- **Summary**: Defines context engineering as finding the smallest set of high-signal tokens that maximize desired outcomes. Introduces three techniques for long-horizon tasks: compaction (summarizing conversation histories), structured note-taking (agentic memory with external notes), and sub-agent architectures. Advocates just-in-time context strategies and progressive disclosure where agents incrementally discover relevant context through exploration.
- **Key Claims**:
  - Context is a finite resource with diminishing returns requiring curation over exhaustive inclusion [**MODERATE**]
  - Just-in-time context loading outperforms pre-loading for tool and knowledge discovery [**MODERATE**]
  - Three complementary strategies exist for long-horizon context management: compaction, structured notes, and sub-agents [**MODERATE**]
  - Progressive disclosure enables agents to incrementally discover relevant context through exploration [**MODERATE**]

### [SRC-007] Equipping Agents for the Real World with Agent Skills
- **Authors**: Barry Zhang, Keith Lazuka, Mahesh Murag (Anthropic)
- **Year**: 2025 (October 16, 2025; open standard December 18, 2025)
- **Type**: official documentation (Anthropic engineering blog + open standard)
- **URL/DOI**: https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills
- **Verified**: yes (content fetched via redirect)
- **Relevance**: 5
- **Summary**: Introduces the three-tier progressive disclosure architecture for agent skills: (1) Metadata layer with YAML frontmatter (~100 tokens per skill) preloaded at startup, (2) Core instructions layer (full SKILL.md) loaded on relevance determination, (3) Reference and executable layers loaded on-demand during execution. Skills are effectively unbounded since not all material loads simultaneously. Published as an open standard.
- **Key Claims**:
  - Three-tier progressive disclosure (metadata/instructions/references) is the optimal architecture for agent skill systems [**MODERATE**]
  - Metadata-first loading (~100 tokens per skill) enables scalable skill discovery without context bloat [**MODERATE**]
  - Skills content is effectively unbounded when using progressive on-demand loading [**MODERATE**]

### [SRC-008] The Meta-Tool Pattern: Progressive Disclosure for MCP (Bounded Context Packs)
- **Authors**: Professor Synapse (SynapticLabs)
- **Year**: 2025
- **Type**: blog post (technical architecture)
- **URL/DOI**: https://blog.synapticlabs.ai/bounded-context-packs-meta-tool-pattern
- **Verified**: partial (URL confirmed, full text extraction failed due to page structure)
- **Relevance**: 4
- **Summary**: Proposes the meta-tool pattern for MCP where two registered tools (discovery + execution) replace loading all tool schemas at startup. Uses bounded context packs to organize tools by domain. Reports 85-95% token overhead reduction. Identifies tool bloat as the core problem: a 128K context window loses 7K+ to unused tool schemas before any conversation begins.
- **Key Claims**:
  - Two meta-tools (discovery + execution) can replace loading all tool schemas, reducing token overhead by 85-95% [**WEAK**]
  - Tool schema bloat is the primary token waste in multi-tool agent systems [**MODERATE**]
  - Bounded context packs (domain-grouped tool sets) are the natural organization unit for progressive tool discovery [**WEAK**]

### [SRC-009] Comparing Progressive Discovery and Semantic Search for Powering Dynamic MCP
- **Authors**: Speakeasy engineering team
- **Year**: 2025-2026
- **Type**: blog post (engineering benchmarks)
- **URL/DOI**: https://www.speakeasy.com/blog/100x-token-reduction-dynamic-toolsets
- **Verified**: yes (content fetched, benchmark data extracted)
- **Relevance**: 4
- **Summary**: Empirically compares two approaches to dynamic tool discovery for MCP: progressive search (hierarchical navigation with prefix-based lookup) and semantic search (embedding-based natural language discovery). Reports 100x+ token reduction with both approaches (e.g., 6,000 vs 405,000 tokens for a 400-tool server). Progressive search offers complete visibility but more tool calls; semantic search is faster but may miss tools.
- **Key Claims**:
  - Dynamic tool discovery achieves 100x+ token reduction compared to static tool loading [**WEAK**]
  - Progressive search provides complete tool visibility at the cost of more discovery calls [**WEAK**]
  - Semantic search is faster for tool discovery but may have incomplete coverage [**WEAK**]
  - Both approaches maintain consistent performance as toolsets scale from 40 to 400 tools [**WEAK**]

### [SRC-010] Agent Skills: Progressive Disclosure as a System Design Pattern
- **Authors**: Aurimas Griciūnas
- **Year**: 2026 (March 11, 2026)
- **Type**: blog post (SwirlAI Newsletter, technical analysis)
- **URL/DOI**: https://www.newsletter.swirlai.com/p/agent-skills-progressive-disclosure
- **Verified**: yes (content fetched)
- **Relevance**: 4
- **Summary**: Independent empirical analysis of Anthropic's Agent Skills progressive disclosure. Measured Layer 1 (discovery) at ~80 tokens/skill median, with all 17 official Anthropic skills totaling ~1,700 tokens. Layer 2 (activation) ranges from ~275 to ~8,000 tokens with ~2,000 median. Identifies the "lost-in-the-middle" phenomenon as the key motivation: important information buried in lengthy contexts gets missed, making selective loading critical for accuracy.
- **Key Claims**:
  - Metadata-first skill loading costs ~80 tokens per skill at the discovery layer [**WEAK**]
  - The "lost-in-the-middle" phenomenon makes selective context loading an accuracy concern, not just an efficiency concern [**MODERATE**]
  - 17 skills at the discovery layer cost ~1,700 tokens total, less than a single activated skill [**WEAK**]

### [SRC-011] Writing a Good CLAUDE.md
- **Authors**: HumanLayer team
- **Year**: 2025
- **Type**: blog post (practitioner guidance)
- **URL/DOI**: https://www.humanlayer.dev/blog/writing-a-good-claude-md
- **Verified**: yes (content fetched)
- **Relevance**: 4
- **Summary**: Practitioner analysis of CLAUDE.md as a persistent knowledge bootstrapping mechanism. Key finding: frontier LLMs can follow ~150-200 instructions with reasonable consistency, and Claude Code's system prompt already uses ~50, leaving limited capacity. Advocates organizing knowledge into WHY/WHAT/HOW categories, keeping CLAUDE.md under 300 lines, and using progressive disclosure via separate files with pointers rather than inlined content. Emphasizes "prefer pointers to copies" as a core principle.
- **Key Claims**:
  - Frontier LLMs can reliably follow ~150-200 instructions, constraining how much knowledge can be seeded [**WEAK**]
  - Memory seeding files should use pointers to detailed files rather than inlining content [**MODERATE**]
  - CLAUDE.md content should be under 300 lines with only universally applicable instructions [**WEAK**]
  - Progressive disclosure via directory-organized markdown files is the recommended knowledge architecture [**MODERATE**]

### [SRC-012] MACLA: Learning Hierarchical Procedural Memory for LLM Agents through Bayesian Selection and Contrastive Refinement
- **Authors**: Saman Forouzandeh, Wei Peng, Parham Moradi, Xinghuo Yu, Mahdi Jalili
- **Year**: 2025 (accepted at AAMAS 2026)
- **Type**: peer-reviewed paper
- **URL/DOI**: https://arxiv.org/abs/2512.18950
- **Verified**: yes (abstract and results fetched from arXiv)
- **Relevance**: 3
- **Summary**: Demonstrates that hierarchical procedural memory (learned from trajectories) can be maintained externally while keeping the LLM frozen. The system extracts 187 reusable procedures from 2,851 trajectories, tracks reliability via Bayesian posteriors, and achieves 78.1% average performance across ALFWorld, WebShop, TravelPlanner, and InterCodeSQL -- constructing memory in 56 seconds (2,800x faster than parameter-training baselines).
- **Key Claims**:
  - External hierarchical procedural memory enables agent learning without LLM parameter modification [**STRONG**]
  - Bayesian selection over learned procedures outperforms raw LLM inference for repeated task types [**MODERATE**]
  - Procedural memory can compress thousands of trajectories into hundreds of reusable procedures [**MODERATE**]

### [SRC-013] Schema First Tool APIs for LLM Agents: A Controlled Study of Tool Misuse, Recovery, and Budgeted Performance
- **Authors**: Akshey Sigdel, Rista Baral
- **Year**: 2026
- **Type**: peer-reviewed paper (arXiv preprint, March 2026)
- **URL/DOI**: https://arxiv.org/abs/2603.13404
- **Verified**: yes (full text fetched from arXiv HTML)
- **Relevance**: 3
- **Summary**: Controlled study comparing three tool interface conditions: natural language documentation, JSON Schema specifications, and JSON Schema with structured diagnostics. Finds that schema conditions reduce interface misuse (malformed calls) but not semantic misuse (wrong tool selection). Task success remained zero across all conditions in the pilot, suggesting schema improvements address contract adherence but not planning quality.
- **Key Claims**:
  - JSON Schema specifications reduce interface misuse (malformed calls) compared to prose documentation [**MODERATE**]
  - Schema-based interfaces do not reduce semantic misuse (wrong tool selection) [**MODERATE**]
  - Interface rigor improves technical compliance without necessarily improving task success [**MODERATE**]

### [SRC-014] Claude Agent Skills: A First Principles Deep Dive
- **Authors**: Lee Han Chung
- **Year**: 2025 (October 26, 2025)
- **Type**: blog post (technical analysis)
- **URL/DOI**: https://leehanchung.github.io/blogs/2025/10/26/claude-skills-deep-dive/
- **Verified**: yes (content fetched)
- **Relevance**: 4
- **Summary**: Independent technical analysis of Claude Agent Skills' progressive disclosure lifecycle. Documents the three-tier loading mechanism: Tier 1 (frontmatter, minimal) appears in the Skill tool description; Tier 2 (SKILL.md, comprehensive) loads on selection; Tier 3 (supporting resources) loads on-demand during execution. Notes skill descriptions budget is limited to 15,000 characters, SKILL.md recommended under 5,000 words. Discovery uses natural language reasoning -- no algorithmic matching or embeddings.
- **Key Claims**:
  - Progressive skill loading uses natural language reasoning for discovery, not embeddings or algorithms [**WEAK**]
  - Per-skill invocation overhead is ~1,500+ tokens due to injected metadata [**WEAK**]
  - Skill description budget (15,000 chars) and SKILL.md size (under 5,000 words) are the practical constraints [**WEAK**]

## Thematic Synthesis

### Theme 1: OS-Inspired Memory Hierarchies Are the Dominant Agent Memory Architecture

**Consensus**: LLM agent memory should be organized in tiers analogous to operating system memory hierarchies, with fast/small in-context memory and slow/large external storage connected by explicit data movement operations. [**STRONG**]
**Sources**: [SRC-001], [SRC-002], [SRC-003], [SRC-004], [SRC-005], [SRC-012]

**Controversy**: Whether the hierarchy should be two-tier (MemGPT: main context + external) or three-tier (Letta: core + recall + archival) remains unsettled. The three-dimensional taxonomy of [SRC-002] (forms x functions x dynamics) suggests the two/three-tier models are oversimplifications.
**Dissenting sources**: [SRC-002] argues existing tier models are "insufficient for capturing contemporary agent memory diversity," while [SRC-001] and [SRC-005] demonstrate practical success with simpler two-tier models.

**Practical Implications**:
- Default to a minimum two-tier hierarchy: in-context working memory + persistent external storage
- Use function calls (not implicit mechanisms) to give agents explicit control over memory movement between tiers
- Consider graph-based storage for the external tier when relational structure matters (Mem0-g achieves ~2% improvement over flat vector stores)

**Evidence Strength**: STRONG

### Theme 2: Progressive Disclosure Is the Scaling Solution for Tool/API Discovery

**Consensus**: Pre-loading all tool schemas into agent context is unsustainable past ~50-100 tools. Progressive disclosure -- revealing tool capabilities incrementally through metadata-first discovery -- is the emerging standard for scalable tool exposure. [**MODERATE**]
**Sources**: [SRC-006], [SRC-007], [SRC-008], [SRC-009], [SRC-010], [SRC-014]

**Controversy**: The optimal discovery mechanism is contested. Anthropic's Agent Skills uses natural language reasoning over metadata ([SRC-007], [SRC-014]); MCP meta-tools use hierarchical prefix-based navigation ([SRC-008]); Speakeasy offers both progressive and semantic search ([SRC-009]). No controlled comparison exists across these approaches.
**Dissenting sources**: [SRC-009] argues semantic search is faster than progressive discovery but may miss tools, while [SRC-008] argues hierarchical navigation provides complete visibility.

**Practical Implications**:
- Implement a three-tier loading pattern: metadata (~100 tokens/skill) always present, full instructions on relevance match, reference material on-demand
- Budget ~1,700 tokens for 15-20 skill metadata entries at the discovery layer
- For systems with 100+ tools, implement dynamic discovery (meta-tools or semantic search) rather than static schema loading
- The "lost-in-the-middle" effect makes progressive disclosure an accuracy concern, not merely an efficiency optimization

**Evidence Strength**: MODERATE

### Theme 3: Memory Seeding and Knowledge Bootstrapping Require Explicit Architecture

**Consensus**: LLM agents are stateless by default; persistent knowledge requires deliberate seeding mechanisms and careful management of what enters context when. The quality of initial memory seeding significantly affects agent performance. [**MODERATE**]
**Sources**: [SRC-005], [SRC-006], [SRC-011], [SRC-004]

**Practical Implications**:
- Seed memory files should contain pointers to detailed content rather than inlined knowledge ("prefer pointers to copies")
- Keep always-loaded seed content under 300 lines / ~150-200 instructions to stay within reliable instruction-following capacity
- Organize seed knowledge into WHY/WHAT/HOW categories with progressive depth
- Use YAML frontmatter for machine-parseable metadata that enables automated freshness checking

**Evidence Strength**: MODERATE

### Theme 4: Self-Editing Memory Enables Agent Autonomy

**Consensus**: Agents that can modify their own memory (adding, updating, linking, and deleting entries) perform better than those limited to append-only or read-only memory access. [**MODERATE**]
**Sources**: [SRC-001], [SRC-004], [SRC-005]

**Controversy**: The degree of autonomy is debated. MemGPT/Letta gives agents full CRUD over their memory blocks. A-Mem ([SRC-004]) adds interconnected linking inspired by Zettelkasten. Mem0 ([SRC-005]) uses a background extraction-update pipeline with ADD/UPDATE/DELETE/NOOP operations and conflict detection. The risk of memory corruption from autonomous editing is acknowledged but not well-studied.
**Dissenting sources**: [SRC-005] uses a structured pipeline with explicit conflict detection, suggesting that unconstrained self-editing may introduce errors.

**Practical Implications**:
- Give agents explicit memory-editing tools (not just read/append) as part of their function-call repertoire
- Implement conflict detection when agents can overwrite existing memories
- Consider async "sleep-time" memory refinement for non-urgent knowledge consolidation
- The Zettelkasten pattern (tagging, linking, evolving notes) is a promising organizational structure for self-managed memory

**Evidence Strength**: MODERATE

### Theme 5: Schema-Based Tool Interfaces Improve Contract Adherence but Not Planning

**Consensus**: Structured schemas (JSON Schema, OpenAPI) reduce malformed tool calls compared to prose documentation, but do not help agents select the right tool for a task. Tool discovery and tool invocation are orthogonal problems requiring different solutions. [**MODERATE**]
**Sources**: [SRC-013], [SRC-006], [SRC-008]

**Practical Implications**:
- Use JSON Schema for tool interface contracts to reduce malformed calls
- Invest separately in tool discovery mechanisms (metadata, descriptions, hierarchical organization) to address semantic misuse
- Rich, context-aware descriptions in tool metadata matter more than schema precision for tool selection
- Error documentation in API schemas is critical for agent recovery

**Evidence Strength**: MODERATE

## Evidence-Graded Findings

### STRONG Evidence
- OS-inspired tiered memory hierarchies (main context + external storage) are the foundational architecture for persistent agent memory -- Sources: [SRC-001], [SRC-003], [SRC-005]
- Agents that self-manage memory through explicit function calls outperform static context approaches for multi-session tasks -- Sources: [SRC-001], [SRC-005]
- External hierarchical procedural memory enables agent learning without LLM parameter modification -- Sources: [SRC-012], [SRC-001]
- Dynamic memory extraction and consolidation across sessions outperforms full-context approaches (26% accuracy improvement, 90% token savings) -- Sources: [SRC-005]

### MODERATE Evidence
- Progressive disclosure with three-tier loading (metadata/instructions/references) is the optimal architecture for scalable tool exposure -- Sources: [SRC-007], [SRC-010], [SRC-014]
- The "lost-in-the-middle" phenomenon makes selective context loading an accuracy concern, not just an efficiency concern -- Sources: [SRC-010], [SRC-006]
- JSON Schema specifications reduce interface misuse but not semantic misuse in tool calling -- Sources: [SRC-013]
- Agent memory should be taxonomized by forms (token/parametric/latent), functions (factual/experiential/working), and dynamics -- Sources: [SRC-002]
- Memory seeding files should use pointers to detailed content rather than inlining -- Sources: [SRC-011], [SRC-006]
- Graph-based memory representations capture richer multi-session relationships than flat vector stores -- Sources: [SRC-005]
- Bayesian selection over learned procedures outperforms raw LLM inference for repeated task types -- Sources: [SRC-012]
- Zettelkasten-style interconnected note structures outperform flat memory stores for LLM agents -- Sources: [SRC-004]

### WEAK Evidence
- Metadata-first skill loading costs ~80 tokens per skill at the discovery layer -- Sources: [SRC-010]
- Dynamic tool discovery achieves 100x+ token reduction compared to static tool loading -- Sources: [SRC-009]
- Frontier LLMs can reliably follow ~150-200 instructions, constraining knowledge seeding capacity -- Sources: [SRC-011]
- Progressive search provides complete tool visibility at the cost of more discovery calls; semantic search trades visibility for speed -- Sources: [SRC-009]
- Two meta-tools (discovery + execution) can replace loading all tool schemas, reducing overhead by 85-95% -- Sources: [SRC-008]

### UNVERIFIED
- Cursor's .cursor/rules directory split (activated per-file-match) represents an independent convergent evolution toward progressive disclosure -- Basis: model training knowledge + search results without fetched content
- OpenAI's ChatGPT memory feature stores memories as natural language notes extracted from conversations, with per-GPT memory isolation -- Basis: search results and community discussions; official documentation was not fetchable (403)

## Knowledge Gaps

- **Controlled comparisons of progressive disclosure strategies**: No published study compares metadata-based discovery (Anthropic Skills), hierarchical navigation (MCP meta-tools), and semantic search (Speakeasy) under identical conditions. The field lacks standard benchmarks for tool discovery efficiency.

- **Memory seeding optimal strategies**: While practitioner guidance exists (CLAUDE.md best practices, .cursorrules patterns), no empirical study has measured the impact of different seeding strategies (content-rich vs. structure-only vs. skip) on downstream agent performance.

- **Memory corruption and drift**: Self-editing memory systems risk gradual quality degradation over time. No longitudinal study examines how agent-managed memories evolve across hundreds of sessions or whether conflict detection mechanisms are sufficient to prevent drift.

- **Cross-system transferability**: Whether memory architectures designed for one LLM (e.g., GPT-4) transfer effectively to others (Claude, Gemini, open-source models) is unstudied. The taxonomy in [SRC-002] is model-agnostic in theory but untested in practice.

- **Multi-agent memory sharing**: How multiple agents should share or partition persistent memory is an emerging topic flagged by [SRC-002] but not yet well-covered in the literature. Production systems like Letta support it but without published evaluation.

## Domain Calibration

Mixed evidence distribution reflects an active, rapidly evolving domain where production implementations are outpacing academic evaluation. The combination of strong evidence for memory hierarchies and moderate evidence for progressive disclosure reflects the field's maturation trajectory: foundational patterns are established, but optimization strategies are still emerging. Treat progressive disclosure findings as strong practitioner consensus pending formal evaluation.

## Methodology Note

This literature review was generated by an LLM (Claude) using WebSearch for source discovery and WebFetch for source verification. Limitations:

1. **Training data cutoff**: Model knowledge is frozen at the training cutoff date. Recent publications may be missing despite web search augmentation.
2. **Source access**: Paywalled content could not be fully verified. Claims from paywalled sources are graded MODERATE at best.
3. **Citation accuracy**: Paper titles and authors were verified via web search where possible, but some citations may contain inaccuracies. DOIs are included only when confirmed.
4. **No domain expertise**: This review reflects what the literature says, not expert judgment about what the literature gets right. Use as a structured research draft, not as authoritative assessment.

Generated by `/research agent-memory-architecture-progressive-api-discovery` on 2026-03-26.
