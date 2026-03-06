---
name: dion
description: "[DEPRECATED] Use /land instead. Alias redirects to /land. Will be removed in a future release."
argument-hint: "[--domain=DOMAIN] [--force]"
allowed-tools: Bash, Read
model: opus
---

# /dion (Deprecated)

This command has been replaced by `/land`.

`/land` provides the same Dionysus synthesis functionality plus automatic .know/ refresh guidance.

## Migration Guide

| Old command | New equivalent |
|-------------|----------------|
| `/dion` | `/land` (full pipeline) |
| `/dion --domain=scar-tissue` | `/land --domain=scar-tissue` |
| `/dion --force` | `/land --force` |
| `/dion` (synthesis only, no .know/) | `/land --skip-know` |

## Action Required

**Please invoke `/land` now.** This alias will be removed in a future release.

If you want the old behavior (Dionysus synthesis only, without .know/ refresh guidance):

```
/land --skip-know
```

To run the full pipeline including .know/ refresh guidance:

```
/land
```

This alias exists only to prevent breakage for users with `/dion` in their muscle memory or documentation. No synthesis or tool invocations are performed here.
