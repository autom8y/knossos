# Roster - Agent Team Pack Management

## Scripts

- `swap-team.sh` - Switch active team pack
- `generate-team-context.sh` - Output team routing table (used by session hooks)
- `load-workflow.sh` - Load workflow.yaml for a team
- `get-workflow-field.sh` - Extract specific workflow fields

## Usage

### Generate Team Context

```bash
# For active team
./generate-team-context.sh

# For specific team
./generate-team-context.sh 10x-dev-pack
```

Output: Markdown table of phase→agent mappings for session hook injection.
