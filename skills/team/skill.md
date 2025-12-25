# Team Pack Swapper

Swap Claude Code agent team packs for different workflows.

## Usage

```
/team                 # Show current active team
/team <pack-name>     # Switch to specified team pack
/team --list          # List all available team packs
```

## Available Team Packs

- **dev-pack**: Development workflow (Architect, Principal Engineer, QA Adversary)
- **doc-pack**: Documentation workflow (Technical Writer, API Documenter, README Author)
- **hygiene-pack**: Code quality workflow (Refactorer, Linter Advisor, Dependency Auditor)

## Implementation

This skill invokes the swap-team.sh script located at ~/Code/roster/.

## Instructions

When the user invokes `/team`, execute the swap-team.sh script with the provided arguments:

1. If no arguments provided: Query current team
2. If `--list` or `-l`: List available teams
3. If `<pack-name>` provided: Swap to that team pack

Always use the Bash tool to execute:
```bash
~/Code/roster/swap-team.sh [args]
```

Display the script output to the user. The script handles all validation, backup, and swap operations.

## Error Handling

The script returns appropriate exit codes:
- 0: Success
- 1: Invalid arguments
- 2: Validation failure (pack doesn't exist)
- 3: Backup failure
- 4: Swap failure

If the script fails, display the error message and suggest next steps based on the output.
