# Getting Started with Knossos

Knossos adds structured, multi-agent workflows to [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Instead of one AI assistant doing everything, Knossos coordinates specialized agents — each with clear responsibilities, handoff criteria, and quality gates.

## Install

```bash
go install github.com/autom8y/knossos/cmd/ari@latest
```

Verify the installation:

```bash
ari version
```

> **Requirements**: Go 1.22+, Claude Code CLI installed and authenticated.

## Initialize a Project

Navigate to any git repository and run:

```bash
ari init --rite review
```

This creates a `.claude/` directory with:

| Directory | Contents |
|-----------|----------|
| `.claude/agents/` | Agent prompts — each agent has a defined role and tools |
| `.claude/skills/` | Reference knowledge agents can load on-demand |
| `.claude/commands/` | Slash commands you can type in Claude Code |
| `.claude/CLAUDE.md` | Project instructions, always loaded into context |
| `.claude/settings.json` | Hook configuration for safety enforcement |

The `review` rite is a good starting point — it's a 3-phase code review workflow that works on any codebase.

## Run Your First Session

1. Open the project in Claude Code:
   ```bash
   claude
   ```

2. Type `/go` to start a session. The system will:
   - Detect the active rite (review)
   - Ask what you want to do
   - Create a session to track progress

3. Describe your task:
   ```
   Review this codebase for quality issues
   ```

4. The orchestrator (Pythia) coordinates the workflow:
   - **Scanner** reads the codebase and identifies areas of concern
   - **Assessor** evaluates findings and prioritizes by impact
   - **Reporter** produces a structured review document

5. Find the results in `.claude/wip/review/`.

## What Just Happened

When you typed `/go`, Knossos activated an **orchestrated workflow**:

1. **Orchestrator** (Pythia) received your request and decided which specialist to invoke first
2. **Scanner** ran — it used Glob, Grep, and Read to scan your codebase structure, producing a scan-findings document
3. Pythia reviewed the scan findings and routed to the **Assessor**
4. **Assessor** evaluated each finding for severity and added recommendations
5. Pythia routed to the **Reporter** for final synthesis
6. **Reporter** produced the review document with executive summary and prioritized findings

Each agent only has access to the tools it needs. The scanner can read files but not edit them. The orchestrator can only read — it coordinates but never executes.

## Available Rites

A **rite** is a workflow definition — it specifies which agents exist, what phases they work through, and how work flows between them.

| Rite | Purpose | Agents |
|------|---------|--------|
| `review` | Language-agnostic code review | 4 (scanner, assessor, reporter + orchestrator) |
| `slop-chop` | AI-generated code quality gate | 6 (detect, analyze, decay, remediate, verdict + orchestrator) |
| `10x-dev` | Full development lifecycle | 5 (requirements, design, build, test + orchestrator) |

Switch rites anytime:

```bash
ari sync --rite slop-chop
```

## Key Commands

These slash commands work in Claude Code after initialization:

| Command | What it does |
|---------|-------------|
| `/go` | Start or resume a session |
| `/start` | Begin a new initiative with full tracking |
| `/park` | Save session state for later |
| `/continue` | Resume a parked session |
| `/wrap` | Complete a session with summary |
| `/commit` | Create a well-structured git commit |
| `/consult` | Get guidance on which rite or workflow to use |

## Project Structure

After `ari init`, your project has two layers:

```
your-project/
  src/                    # Your code (unchanged)
  .claude/                # Knossos workspace
    agents/               # Agent prompts (who does what)
    skills/               # Reference knowledge (loaded on-demand)
    commands/             # Slash commands (user-invoked actions)
    CLAUDE.md             # Project instructions (always in context)
    ACTIVE_RITE           # Currently active workflow
    ACTIVE_WORKFLOW.yaml  # Workflow definition (phases, routing)
    settings.json         # Hook configuration
```

Knossos never modifies your source code during initialization. It only creates the `.claude/` directory.

## Next Steps

- **Switch rites**: `ari sync --rite slop-chop` to try the AI code quality gate
- **Explore commands**: Type `/` in Claude Code to see all available slash commands
- **Learn concepts**: See [Concepts](concepts.md) for how rites, agents, and phases work together
- **Build your own rite**: Look at `rites/review/` as a template — manifest + workflow + agents + skills
