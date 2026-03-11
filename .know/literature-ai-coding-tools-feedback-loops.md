---
domain: "literature-ai-coding-tools-feedback-loops"
generated_at: "2026-03-11T12:00:00Z"
expires_after: "180d"
source_scope:
  - "external-literature"
  - "product-documentation"
  - "technical-blogs"
generator: bibliotheca
confidence: 0.75
format_version: "1.0"
---

# Literature Review: AI Coding Tools That Learn from Usage Patterns

> LLM-synthesized literature review from web research conducted 2026-03-11. Citations should be independently verified before use in production decisions or published work.

## Executive Summary

This review examines production AI coding tools and agent frameworks that demonstrably learn from operational feedback, distinguishing between **actual model-level learning**, **context engineering** (persistent prompt injection), and **marketing claims**. The critical finding across all systems is that **no production AI coding tool performs per-user model fine-tuning in real time**. Instead, "learning" falls into three distinct categories:

1. **Aggregate model improvement**: Acceptance/rejection telemetry feeds into periodic global model retraining (GitHub Copilot, Amazon Q Developer). The user's individual patterns influence the next model version alongside millions of other signals.

2. **Context engineering**: Persistent files (CLAUDE.md, .cursorrules, copilot-instructions.md) inject project-specific and user-specific instructions into every prompt. The model itself is unchanged; the context it receives is curated. This is the dominant mechanism in Claude Code, Cursor, and GitHub Copilot's custom instructions.

3. **Self-modifying agent systems**: Research prototypes (Darwin Godel Machine, SICA) that rewrite their own code/prompts based on benchmark performance. These demonstrate real self-improvement but operate in sandboxed research settings, not production coding workflows.

The practical implication: the most effective "learning" in production today is context engineering -- curating what the model sees, not changing the model itself.

---

## 1. GitHub Copilot

### 1.1 Feedback Signal: What Is Collected

GitHub Copilot collects **user engagement data** including:
- Pseudonymous identifiers on interactions (accepted/dismissed completions)
- Error messages and system logs
- Product usage metrics (suggestion frequency, latency, completion length)
- Real-time user feedback (thumbs up/down, optional comments)
- Code snippets surrounding suggestions (for Free/Pro users who opt in)

**What is NOT collected** for Business/Enterprise tiers:
- No prompt telemetry is collected
- Zero data retention for code snippets and usage telemetry
- Organization administrators control all data policies

For Free/Pro users: prompts and suggestions are shared with GitHub for product improvement, but users can opt out via "Allow GitHub to use my code snippets from the code editor for product improvements" setting.

### 1.2 What Gets Modified

**GitHub does NOT train on individual user code.** The model improvement pathway is:

1. **Aggregate telemetry** from millions of users feeds into training data curation
2. **Synthetic data distillation** from larger models trains smaller, specialized models
3. **LLM-based graders** filter out ambiguous or low-signal training samples
4. **Dozens of model candidates** are trained per month; the best-performing version ships
5. **Reinforcement learning** applied to improve suggestion quality based on aggregate acceptance patterns

The custom model for **Next Edit Suggestions (NES)** is particularly notable: GitHub trains a specialized model that predicts the next logical edit based on editing behavior. No existing dataset captured real-time editing behavior, so GitHub had to construct synthetic training pipelines. GitHub is exploring adaptive behavior where NES adjusts to individual editing style over time (accepting, dismissing, or ignoring suggestions).

### 1.3 Privacy Architecture

| Tier | Code Snippets Collected | Used for Training | Retention |
|------|------------------------|-------------------|-----------|
| Free/Pro (opt-in) | Yes | Aggregate model improvement | Standard |
| Free/Pro (opt-out) | No | No | None |
| Business | No | No | Zero |
| Enterprise | No | No | Zero |

Access to collected data is restricted to GitHub Copilot team personnel, Microsoft Azure teams, and OpenAI employees working on Copilot.

### 1.4 Safety Bounds

- Suggestion acceptance as a reward signal can **reduce suggestion quality** (perverse incentive discovered in 2024 research)
- A retrospective evaluation on 535 programmers showed the system could avoid displaying suggestions that would have been rejected, but optimizing purely for acceptance rate degraded output quality
- Code review shifted from "thoroughness" to "high-signal feedback" after learning developers value speed over completeness

### 1.5 Agentic Memory System (2025-2026, Public Preview)

GitHub recently launched a **persistent, repository-level memory** system:
- Copilot develops knowledge of codebases that persists across sessions
- Memory is shared across surfaces: if the coding agent discovers how a repo handles database connections, code review can apply that knowledge to spot inconsistencies in PRs
- Before applying any memory, the agent verifies accuracy by checking cited code locations
- Enabled by default for Pro/Pro+ users; disabled by default for Enterprise (opt-in by admins)

### 1.6 Custom Instructions (.github/copilot-instructions.md)

- Markdown file providing persistent project context
- Describes build/test/validation procedures for the coding agent
- Instructions that explain **why** a convention exists produce better edge-case handling
- Can be supplemented with `.instructions.md` files for file-type-specific rules
- `AGENTS.md` for multi-agent workspace coordination

**Assessment**: Copilot's "learning" is primarily aggregate model improvement from telemetry plus context injection via instructions files. The new agentic memory system is the closest thing to per-project learning, but it operates at the context level, not the model weights level.

---

## 2. Cursor / AI Code Editors

### 2.1 The Technical Truth About Per-User Learning

**Cursor does NOT fine-tune or update the underlying model per-user.** Large language models don't retain memory between completions. All "personalization" in Cursor operates through context injection:

1. **Rules System** (`.cursor/rules/*.mdc`): Markdown Component files with metadata (description, glob patterns, alwaysApply flag). When applied, rule contents are included at the start of model context.
2. **System Prompt Injection**: Cursor injects role prompts like "You are a powerful agentic AI coding assistant, powered by [model name]."
3. **Codebase Indexing**: Semantic search over the indexed codebase retrieves relevant context. This is retrieval, not learning.

### 2.2 Rules Evolution

The rules system has evolved through three generations:
- **Gen 1**: `.cursorrules` file in project root (now deprecated)
- **Gen 2**: `.cursor/rules/*.mdc` files -- version-controlled, scoped by glob pattern, selectively activated
- **Gen 3**: Rule files can be "always on," "auto-detected" based on relevance, or "manually invoked"

Token optimization: splitting rules into scoped files reduces token usage by only activating relevant rules per interaction.

### 2.3 What Cursor Does Well (Context Engineering)

- **Codebase-aware suggestions**: Semantic indexing means suggestions reference actual project patterns
- **Multi-agent architecture** (Cursor 2.0, October 2025): Composer model optimized for context-aware edits, Multi-Agent Judging, Debug Mode
- **Visual Editor** (December 2025): Browser-based rendering for front-end development

### 2.4 What Cursor Does NOT Do

- No per-user model fine-tuning
- No persistent learning across sessions (beyond rule files)
- No feedback loop that modifies model behavior based on acceptance/rejection
- The underlying models (Claude, GPT, or Cursor's own models) remain unchanged across users

**Assessment**: Cursor's "learning" is entirely context engineering. The rules system is well-designed for its purpose, but claims of the system "learning your preferences" refer to semantic indexing and context injection, not model adaptation.

---

## 3. Claude Code / Anthropic's Approach

### 3.1 Memory Architecture (CLAUDE.md)

Claude Code's memory operates through a hierarchy of files loaded into the system prompt:

| File | Scope | Loaded When |
|------|-------|-------------|
| `~/.claude/CLAUDE.md` | User global | Every session |
| `project/CLAUDE.md` | Project | Every session in that project |
| `~/.claude/projects/<path>/CLAUDE.md` | User-project | Every session in that project |
| `MEMORY.md` (auto-generated) | Per-project | Every session in that project |

**Effectiveness data**: Files under 200 lines achieve rule application rate above 92%, compared to 71% beyond 400 lines. 15 imperative rules produce compliant code in 94% of cases; descriptive style drops to 73%.

### 3.2 Auto-Memory as Self-Improvement

Claude Code's auto-memory creates a genuine feedback loop:
1. Claude encounters a correction or learns a project convention during work
2. Claude writes what it learned to MEMORY.md
3. Future sessions load MEMORY.md, incorporating the lesson
4. Over dozens of iterations, effectiveness measurably increases

This is **context engineering that creates a feedback loop**, not model-level learning. The model weights are unchanged; what changes is the curated context the model receives.

### 3.3 Hooks as Deterministic Guardrails

Hooks provide a complementary mechanism:
- Shell commands that execute at specific lifecycle points (PreToolUse, PostToolUse, etc.)
- **Deterministic** "must-do" rules vs. CLAUDE.md's "should-do" suggestions
- Can enforce conventions that the model might otherwise drift from
- Example: pre-commit validation, automatic formatting, tool access control

### 3.4 Context Engineering vs. Model Learning

This is the critical distinction:

| Dimension | Context Engineering | Model Learning |
|-----------|-------------------|----------------|
| What changes | Prompt content | Model weights |
| Persistence | Files on disk | Weight updates |
| Scope | Per-project/user | Global model |
| Reliability | High (deterministic loading) | Variable (training dynamics) |
| Degradation risk | Context window overflow | Model collapse, capability loss |
| Human oversight | Full (files are readable) | Opaque (weights are not inspectable) |

Claude Code's approach is **entirely context engineering**. The model never changes. What changes is the carefully curated set of instructions, memories, and context injected into every session. This has a key advantage: full transparency and human editability.

### 3.5 Knossos as Meta-Framework

The Knossos platform (this codebase) represents a sophisticated implementation of context engineering as a self-improvement mechanism:
- **Rites** define workflow patterns materialized into `.claude/` configuration
- **Mena** (skills/commands) provide reusable context chunks
- **Hooks** enforce deterministic behavior constraints
- **Session system** (.sos/) tracks operational state across interactions
- **Knowledge base** (.know/) accumulates persistent project understanding

This is context engineering elevated to a system architecture -- the "Rails for Claude Code" approach.

**Assessment**: Claude Code's approach is honest about what it is: context engineering, not model learning. The auto-memory mechanism creates a genuine feedback loop at the context level. Knossos extends this into a full framework. The advantage over model-level learning is transparency, editability, and human control.

---

## 4. Other Notable Systems

### 4.1 Amazon Q Developer (formerly CodeWhisperer)

**Feedback Signal**: Acceptance/rejection telemetry, interaction pattern analysis (49% of early interactions are incremental single-character deletions and partial insertions).

**What Gets Modified**:
- Global model refined via RLHF (Reinforcement Learning from Human Feedback) using crowd-sourced evaluations
- **Customizations feature** (Pro tier): Customers point Q Developer at private repositories; after a training run, suggestions follow the organization's naming conventions, helper libraries, and API wrappers
- This is the closest any production tool comes to per-organization model fine-tuning

**Retention and complexity patterns**: Comment-guided suggestion retention rises from 38% to 88% as task complexity increases.

**Assessment**: Q Developer's Customizations feature is genuinely per-organization model fine-tuning, making it the most aggressive production "learning" system among coding assistants.

### 4.2 Tabnine

**Feedback Signal**: Local code context, codebase indexing (GitHub/GitLab/Bitbucket), non-code data sources (Jira/Confluence).

**Multi-level personalization**:
1. Local code awareness (IDE context)
2. Codebase-level awareness (repository indexing)
3. Non-code data source awareness
4. Customized rule sets for agents
5. **Model fine-tuning** on customer codebases

**Code Review Agent**: Codifies institutional knowledge, corporate policies, and software development standards using awareness from "golden code repos."

**Assessment**: Tabnine offers genuine per-organization model customization and positions strongly on privacy (named Gartner Visionary in September 2025). The multi-level personalization stack is technically sound.

### 4.3 Google Gemini Code Assist

**Feedback Signal**: Thumbs up/down buttons on chat responses, suggestion acceptance/rejection.

**Adaptation mechanisms**:
- Gemini 3 Pro (JetBrains integration, November 2025): "learns from your code to replicate your project's conventions"
- Retrieval-augmented generation (RAG) rolling out as experiment for improved repository-specific suggestions
- Standard/Enterprise tiers include codebase-aware context injection

**Assessment**: Gemini Code Assist is primarily RAG-based context injection with some experimental per-repository adaptation. The "learns from your code" claim appears to refer to RAG retrieval, not model fine-tuning.

### 4.4 JetBrains AI Assistant

JetBrains integrates multiple models (Gemini 3 Pro as of November 2025) with IDE-native context. The feedback mechanism is primarily thumbs up/down on responses. No evidence of per-user model adaptation; operates through context injection from IDE state and project structure.

---

## 5. Self-Modifying Agent Systems

### 5.1 Darwin Godel Machine (Sakana AI, May 2025)

The most significant research result in self-modifying AI agents:

**What it is**: A system that iteratively modifies its own Python codebase and validates each change against coding benchmarks.

**Results**:
- SWE-bench: 20.0% to 50.0% autonomous improvement
- Polyglot benchmark: 14.2% to 30.7% (surpassing hand-designed agent Aider)

**What gets modified**: The agent's own code -- prompt templates, tool implementations, validation steps, solution ranking logic. Discovered improvements include:
- Patch validation steps
- Better file viewing tools
- Enhanced editing tools
- Multi-solution generation and ranking
- Failure history tracking ("what has been tried before and why it failed")

**Generalizability**: Improvements discovered using Claude 3.5 Sonnet transferred to o3-mini and Claude 3.7 Sonnet, suggesting DGM discovers **fundamental workflow improvements**, not model-specific tricks.

**Safety**: All modifications run in sandboxed environments, under human supervision, with strict limits on web access.

### 5.2 SICA -- Self-Improving Coding Agent (ICLR 2025)

Published as a workshop paper at ICLR 2025 SSI-FM (Robeyns, Szummer, Aitchison):

**What it is**: A fully self-referential agent that edits its entire codebase, including its own prompts.

**Results**: Performance on SWE-bench Verified improved from 17% to 53%.

**Mechanism**: Archive-based evolutionary approach -- the best-performing agent from the archive serves as the meta-agent, instructed to identify and implement improvements based on the full archive history.

**Key distinction**: "Fully self-improving" -- no separation between meta-agent and target agent. The system that does the improving IS the system being improved.

### 5.3 AutoGPT: From Autonomy to Control

**Evolution**: AutoGPT's trajectory is instructive about the limits of self-modifying systems:
- **Original vision** (2023): Fully autonomous agents that self-prompt and chain actions
- **Reality discovered**: Inherent unpredictability of LLMs in production made full autonomy impractical
- **Current state** (2025-2026): Low-code platform that puts users in control of agent construction

**Failure modes**:
- Infinite loops from self-prompting drift
- Costly API spirals without human-in-the-loop checks
- Off-target behavior when self-prompts deviate from original goals

**Lesson**: AutoGPT's retreat from full autonomy to user-controlled workflows is the strongest empirical evidence that unconstrained self-modification in production is currently impractical.

### 5.4 MetaGPT: Structured Workflows

MetaGPT takes the opposite approach from AutoGPT:
- Encodes **Standardized Operating Procedures (SOPs)** into prompt sequences
- Uses an assembly-line paradigm with role-assigned agents
- Self-improvement is constrained to operating within pre-defined workflow structures

**Assessment**: MetaGPT does not self-modify in the DGM/SICA sense. Its "improvement" comes from structured decomposition and role specialization, not recursive self-editing.

### 5.5 CrewAI: Memory-Based Learning

CrewAI provides the most production-ready memory architecture for multi-agent systems:

**Memory components**:
- **Short-term**: ChromaDB + RAG for current session context
- **Long-term**: SQLite3 for cross-session insights and task results
- **Entity memory**: RAG-based capture of people, places, concepts
- **Unified Memory API**: Single class replacing separate memory types, with LLM-analyzed content (infers scope, categories, importance)

**Adaptive scoring**: Composite scoring blends semantic similarity, recency, and importance for memory recall.

**Self-improvement claim**: "Memory makes systems more intelligent with each run" -- this is persistent context accumulation (similar to Claude Code's MEMORY.md), not model-level learning.

### 5.6 Reflexion: Verbal Reinforcement Learning

**Mechanism**: Converts environmental feedback into linguistic self-reflection, provided as context for the next episode.

**Results**: ReAct + Reflexion completed 130/134 AlfWorld tasks; outperformed prior SOTA on MBPP, HumanEval, Leetcode Hard.

**Limitation**: Relies on the agent's ability to accurately evaluate its own performance. Deriving useful reflections from a frozen LLM is challenging -- the model that reflects is the same model that made the mistake.

### 5.7 Taxonomy from Yohei Nakajima (BabyAGI Creator)

Nakajima synthesized NeurIPS 2025 research into five categories of self-improvement:

1. **Self-reflection and in-loop feedback**: Prompt-level improvement without changing weights
2. **Self-generated data and curricula**: Agents create their own training data
3. **Self-adapting models**: Agents that fine-tune or edit themselves
4. **Self-improving code agents**: Agents that modify their own code/policies/architecture
5. **Embodied self-improvement**: Learning by acting in environments

Plus a critical sixth dimension: **Verification, safety, and control** -- keeping self-improvement from going off the rails.

---

## 6. Comparative Analysis

### What Actually Learns vs. What Doesn't

| System | Per-User Model Change | Per-Org Model Change | Context Engineering | Self-Modifying Code |
|--------|----------------------|---------------------|--------------------|--------------------|
| GitHub Copilot | No | No | Yes (instructions, memory) | No |
| Cursor | No | No | Yes (rules, indexing) | No |
| Claude Code | No | No | Yes (CLAUDE.md, memory, hooks) | No |
| Amazon Q Developer | No | **Yes** (Customizations) | Yes | No |
| Tabnine | No | **Yes** (fine-tuning) | Yes | No |
| Gemini Code Assist | No | Experimental (RAG) | Yes | No |
| DGM (research) | N/A | N/A | N/A | **Yes** |
| SICA (research) | N/A | N/A | N/A | **Yes** |
| CrewAI | No | No | Yes (memory system) | No |

### The Feedback Loop Spectrum

**Level 0 -- No feedback**: Static model, static prompts. (No production tool is here anymore.)

**Level 1 -- Aggregate telemetry**: Acceptance/rejection data feeds global model retraining months later. (GitHub Copilot's global model, Gemini Code Assist.)

**Level 2 -- Context accumulation**: System writes memories/rules that persist across sessions. The model is unchanged; what it sees is curated. (Claude Code MEMORY.md, Copilot agentic memory, CrewAI long-term memory.)

**Level 3 -- Per-organization fine-tuning**: Custom model trained on organization's codebase. (Amazon Q Developer Customizations, Tabnine model fine-tuning.)

**Level 4 -- Self-modifying code/prompts**: Agent edits its own implementation based on benchmark performance. (DGM, SICA -- research only, not production.)

**Level 5 -- Recursive self-improvement**: Agent improves its ability to improve itself. (Theoretical/early research. DGM shows hints but bounded by foundation model capabilities.)

### Safety Bounds Across Systems

| System | Primary Safety Mechanism |
|--------|------------------------|
| GitHub Copilot | Opt-out telemetry, zero retention for Enterprise, human-reviewed model candidates |
| Cursor | Rules are human-readable files; no model modification |
| Claude Code | Context files are fully transparent and editable; hooks provide deterministic overrides |
| Amazon Q Developer | Organization-controlled customization scope; RLHF alignment |
| DGM (research) | Sandboxed execution, human supervision, restricted web access |
| SICA (research) | Benchmark-gated improvement; archive maintains history of all variants |
| AutoGPT | Step limits, human-in-the-loop approval (added after discovering failure modes) |
| CrewAI | Memory is queryable and auditable; no model modification |

---

## 7. Key Findings and Implications

### Finding 1: Context Engineering Is the Dominant Paradigm

Every production coding tool uses context engineering as its primary "learning" mechanism. The convergent design: persistent files (CLAUDE.md, .cursorrules, copilot-instructions.md) that inject project knowledge into every prompt. This works because it is transparent, human-editable, and deterministic.

### Finding 2: Per-User Model Adaptation Does Not Exist in Production

No production AI coding tool fine-tunes models per individual user. The closest analogues are per-organization fine-tuning (Amazon Q Developer, Tabnine) and GitHub Copilot's NES adaptive behavior (still exploratory).

### Finding 3: Self-Modifying Systems Work in Research, Not Production

DGM (20% to 50% on SWE-bench) and SICA (17% to 53%) demonstrate that self-modifying agents can achieve significant improvements. But these operate on benchmarks with clear success criteria, in sandboxed environments, with human supervision. The gap between benchmark self-improvement and production self-improvement remains large.

### Finding 4: AutoGPT's Retreat Is the Strongest Empirical Signal

AutoGPT's evolution from "fully autonomous self-prompting agent" to "user-controlled low-code platform" is the most informative data point about production viability of self-modification. Unconstrained self-prompting produced infinite loops, cost spirals, and off-target behavior.

### Finding 5: The Knossos Pattern Is Architecturally Distinct

Knossos represents a pattern not seen in other systems: **meta-framework for context engineering itself**. While Claude Code provides CLAUDE.md and hooks, Knossos provides a system for generating, materializing, and managing the context engineering artifacts across multiple workflow types. This is "learning about how to learn" at the context level -- a form of recursive self-improvement bounded by human oversight (rites, sessions, materialization invariants).

### Finding 6: The Acceptance-Rate Trap

GitHub's research revealed that optimizing for suggestion acceptance rate can **degrade suggestion quality**. This is a general warning for any feedback-loop system: the measurable proxy (acceptance) may diverge from the actual goal (code quality). Systems that optimize for observable feedback signals risk Goodhart's Law effects.

---

## Source Catalog

### [SRC-001] When to Show a Suggestion? Integrating Human Feedback in AI-Assisted Programming
- **URL**: https://arxiv.org/html/2306.04930v3
- **Type**: peer-reviewed paper
- **Relevance**: 5 -- Demonstrates acceptance/rejection feedback mechanisms and the perverse incentive of optimizing for acceptance rate

### [SRC-002] Building an Agentic Memory System for GitHub Copilot
- **URL**: https://github.blog/ai-and-ml/github-copilot/building-an-agentic-memory-system-for-github-copilot/
- **Type**: engineering blog
- **Relevance**: 5 -- Primary source on Copilot's persistent memory architecture

### [SRC-003] The Road to Better Completions: Building a Faster, Smarter GitHub Copilot
- **URL**: https://github.blog/ai-and-ml/github-copilot/the-road-to-better-completions-building-a-faster-smarter-github-copilot-with-a-new-custom-model/
- **Type**: engineering blog
- **Relevance**: 4 -- Details custom model training pipeline for code completions

### [SRC-004] Evolving GitHub Copilot's Next Edit Suggestions Through Custom Model Training
- **URL**: https://github.blog/ai-and-ml/github-copilot/evolving-github-copilots-next-edit-suggestions-through-custom-model-training/
- **Type**: engineering blog
- **Relevance**: 5 -- Primary source on NES model training, synthetic data pipelines, and per-developer adaptive behavior exploration

### [SRC-005] GitHub Copilot Privacy and Data Handling
- **URL**: https://resources.github.com/learn/pathways/copilot/essentials/how-github-copilot-handles-data/
- **Type**: official documentation
- **Relevance**: 5 -- Authoritative source on data collection, retention, and training policies

### [SRC-006] Cursor Rules for AI
- **URL**: https://docs.cursor.com/context/rules-for-ai
- **Type**: official documentation
- **Relevance**: 5 -- Primary source on Cursor's rules system architecture

### [SRC-007] Cursor Rules (New System)
- **URL**: https://docs.cursor.com/context/rules
- **Type**: official documentation
- **Relevance**: 4 -- Documentation of the .mdc-based rules migration

### [SRC-008] How Claude Remembers Your Project
- **URL**: https://code.claude.com/docs/en/memory
- **Type**: official documentation
- **Relevance**: 5 -- Authoritative source on Claude Code's memory system

### [SRC-009] The CLAUDE.md Memory System Deep Dive
- **URL**: https://institute.sfeir.com/en/claude-code/claude-code-memory-system-claude-md/deep-dive/
- **Type**: technical analysis
- **Relevance**: 4 -- Quantitative data on CLAUDE.md effectiveness (92% rule application under 200 lines)

### [SRC-010] Self-Improving Coding Agents (Addy Osmani)
- **URL**: https://addyosmani.com/blog/self-improving-agents/
- **Type**: technical blog
- **Relevance**: 4 -- Practitioner perspective on continuous coding loops and validation patterns

### [SRC-011] Darwin Godel Machine: Open-Ended Evolution of Self-Improving Agents
- **URL**: https://arxiv.org/abs/2505.22954
- **Type**: preprint
- **Relevance**: 5 -- Primary source on DGM, benchmark results (20% to 50% SWE-bench), safety architecture

### [SRC-012] A Self-Improving Coding Agent (ICLR 2025)
- **URL**: https://arxiv.org/abs/2504.15228
- **Type**: workshop paper (ICLR 2025 SSI-FM)
- **Relevance**: 5 -- Primary source on SICA, 17% to 53% improvement on SWE-bench Verified

### [SRC-013] Better Ways to Build Self-Improving AI Agents (Yohei Nakajima)
- **URL**: https://yoheinakajima.com/better-ways-to-build-self-improving-ai-agents/
- **Type**: technical blog
- **Relevance**: 4 -- NeurIPS 2025 synthesis taxonomy of self-improvement approaches

### [SRC-014] CrewAI Memory Documentation
- **URL**: https://docs.crewai.com/en/concepts/memory
- **Type**: official documentation
- **Relevance**: 4 -- Primary source on CrewAI's multi-level memory architecture

### [SRC-015] Reflexion: Language Agents with Verbal Reinforcement Learning
- **URL**: https://www.promptingguide.ai/techniques/reflexion
- **Type**: reference guide / paper summary
- **Relevance**: 4 -- Foundational work on verbal self-reflection as feedback mechanism

### [SRC-016] Tabnine Personalization Announcement
- **URL**: https://www.globenewswire.com/news-release/2024/02/22/2833727/0/en/Tabnine-Launches-New-Capabilities-to-Personalize-AI-Coding-Assistant-to-Any-Development-Team.html
- **Type**: press release
- **Relevance**: 3 -- Details multi-level personalization and model fine-tuning capabilities

### [SRC-017] About Agentic Memory for GitHub Copilot
- **URL**: https://docs.github.com/en/copilot/concepts/agents/copilot-memory
- **Type**: official documentation
- **Relevance**: 5 -- Authoritative source on Copilot's new agentic memory system

### [SRC-018] Adding Repository Custom Instructions for GitHub Copilot
- **URL**: https://docs.github.com/copilot/customizing-copilot/adding-custom-instructions-for-github-copilot
- **Type**: official documentation
- **Relevance**: 4 -- .github/copilot-instructions.md specification

### [SRC-019] Sakana AI Darwin Godel Machine
- **URL**: https://sakana.ai/dgm/
- **Type**: research announcement
- **Relevance**: 5 -- Primary source with safety architecture details and generalizability results

### [SRC-020] Experience with GitHub Copilot for Developer Productivity at Zoominfo
- **URL**: https://arxiv.org/html/2501.13282v1
- **Type**: preprint
- **Relevance**: 3 -- Enterprise deployment data (33% acceptance rate across 400+ developers)
