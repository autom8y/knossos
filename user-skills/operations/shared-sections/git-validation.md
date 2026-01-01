# Git State Validation Pattern

> Validate git repository state before operations.

## When to Apply

- `/commit`: Pre-commit checks
- `/pr`: Pre-PR checks

## Validation Checks

| Check | Command | Pass | Fail |
|-------|---------|------|------|
| In repo | `git rev-parse --git-dir` | Exists | Error: Not a repo |
| Not main | `git rev-parse --abbrev-ref HEAD` | != main/master | Error: Switch branch |
| No conflict | `git status` | No "Unmerged" | Error: Resolve conflicts |
| Clean state | `git status --porcelain` | Empty or staged | Prompt or error |
| Has remote | `git remote -v` | Origin exists | Error: Add remote |

## Validation Sequence

### Basic Checks (All Commands)

```bash
# Verify in git repository
git rev-parse --git-dir 2>/dev/null
if [ $? -ne 0 ]; then
  echo "Error: Not a git repository"
  echo "Run 'git init' or navigate to repo"
  exit 1
fi

# Check for merge conflicts
if git status | grep -q "Unmerged paths"; then
  echo "Error: Unresolved merge conflict"
  echo "Resolve conflicts first with: git status"
  exit 1
fi
```

### Branch Validation (PR Only)

```bash
# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Ensure not on main/master
if [[ "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "master" ]]; then
  echo "Error: Cannot {verb} from main branch"
  echo "Switch to feature branch first:"
  echo "  git checkout -b feature/my-feature"
  exit 1
fi
```

### Remote Validation (PR Only)

```bash
# Check if remote exists
if ! git remote -v | grep -q origin; then
  echo "Error: No remote configured"
  echo "Add remote repository:"
  echo "  git remote add origin https://github.com/user/repo.git"
  exit 1
fi

# Check if branch tracks remote (optional)
if ! git rev-parse --abbrev-ref @{upstream} 2>/dev/null; then
  echo "Note: Branch not tracking remote (will push with -u)"
fi
```

### Clean State Validation

```bash
# Check for uncommitted changes
UNCOMMITTED=$(git status --porcelain)

if [ -n "$UNCOMMITTED" ]; then
  # For /commit: Offer staging
  # For /pr: Error and exit
  echo "Uncommitted changes detected:"
  echo "$UNCOMMITTED"

  # Command-specific behavior
  if [[ "$VERB" == "create PR" ]]; then
    echo "Commit changes first: /commit"
    exit 1
  elif [[ "$VERB" == "commit" ]]; then
    # Enter smart staging flow
    smart_staging
  fi
fi
```

## Error Messages

| Condition | Message Template |
|-----------|------------------|
| Not a repo | `"Not a git repository. Run 'git init' or navigate to repo."` |
| On main | `"Cannot {verb} from main branch. Switch to feature branch."` |
| Merge conflict | `"Unresolved merge conflict. Resolve conflicts first."` |
| No remote | `"No remote configured. Run 'git remote add origin URL'."` |
| Uncommitted (PR) | `"Uncommitted changes detected. Commit first: /commit"` |
| Detached HEAD | `"You are in detached HEAD state. Create branch: git checkout -b <name>"` |

## Customization Points

| Parameter | Description | Commands |
|-----------|-------------|----------|
| `verb` | Action description | commit ("commit"), pr ("create PR") |
| `require_clean` | Strict cleanliness | pr (yes), commit (no) |
| `require_remote` | Remote must exist | pr (yes), commit (no) |
| `require_branch` | Not on main/master | pr (yes), commit (no) |

## Usage in behavior.md

**From commit-ref/behavior.md**:

```markdown
### 1. Validate Git State

Apply [Git Validation Pattern](../shared-sections/git-validation.md):
- Requirement: In git repository
- Verb: "commit"

[Command-specific validation details...]
```

**From pr-ref/behavior.md**:

```markdown
### 1. Validate Git State

Apply [Git Validation Pattern](../shared-sections/git-validation.md):
- Requirement: On feature branch (not main/master)
- Verb: "create PR"

[Command-specific validation details...]
```

## Design Rationale

### Why Separate Pattern?

Both `/commit` and `/pr` need git validation, but with different requirements:

| Validation | /commit | /pr | Reason |
|------------|---------|-----|--------|
| In repo | Required | Required | Both need git |
| No conflicts | Required | Required | Both blocked by conflicts |
| Not on main | No | Yes | PRs can't target same branch |
| Has remote | No | Yes | PRs need remote |
| Clean state | Optional | Required | Commits can stage; PRs can't have uncommitted |

### Error vs Warning

| Condition | /commit | /pr |
|-----------|---------|-----|
| Uncommitted changes | Offer staging | **Error** (must commit) |
| No remote | Warning | **Error** (need for PR) |
| Detached HEAD | Warning | Warning |
| On main | Warning | **Error** (can't PR) |

## Edge Cases

### Detached HEAD

Both commands warn but allow proceeding:

```bash
if [ "$CURRENT_BRANCH" == "HEAD" ]; then
  echo "Warning: You are in detached HEAD state"
  echo "Current commit: $(git rev-parse --short HEAD)"
  echo ""
  echo "Recommendation: Create a branch first"
  echo "  git checkout -b <branch-name>"
  echo ""
  read -p "Proceed anyway? [y/N]: " CONFIRM
  if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi
```

### Partial Staging

For `/commit` only:

```bash
# Check if some files staged, some not
STAGED=$(git diff --staged --name-only)
UNSTAGED=$(git diff --name-only)

if [ -n "$STAGED" ] && [ -n "$UNSTAGED" ]; then
  echo "Note: Some files staged, some unstaged"
  echo "Staged files will be committed:"
  echo "$STAGED"
  echo ""
  echo "Unstaged files will NOT be committed:"
  echo "$UNSTAGED"
fi
```

## Cross-Reference

- [commit-ref/behavior.md](../commit-ref/behavior.md#1-validate-git-state) - Commit validation
- [pr-ref/behavior.md](../pr-ref/behavior.md#1-validate-git-state) - PR validation
