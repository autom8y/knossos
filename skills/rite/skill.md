# Rite Swapper

Swap Claude Code agent rites for different workflows.

## Usage

```
/team                 # Show current active rite
/team <rite-name>     # Switch to specified rite
/team --list          # List all available rites
```

## Available Rites

- **dev-pack**: Development workflow (Architect, Principal Engineer, QA Adversary)
- **doc-pack**: Documentation workflow (Technical Writer, API Documenter, README Author)
- **hygiene**: Code quality workflow (Refactorer, Linter Advisor, Dependency Auditor)

## Implementation

This skill invokes the swap-team.sh script located at $ROSTER_HOME/.

## Instructions

When the user invokes `/team`, execute the swap-team.sh script with the provided arguments:

1. If no arguments provided: Query current rite
2. If `--list` or `-l`: List available rites
3. If `<rite-name>` provided: Swap to that rite

Always use the Bash tool to execute:
```bash
$ROSTER_HOME/swap-team.sh [args]
```

Display the script output to the user. The script handles all validation, backup, and swap operations.

## Error Handling

The script returns appropriate exit codes:
- 0: Success
- 1: Invalid arguments
- 2: Validation failure (rite doesn't exist)
- 3: Backup failure
- 4: Swap failure

If the script fails, display the error message and suggest next steps based on the output.
