---
name: session-common
description: "Shared session lifecycle schemas and patterns. Use when: understanding session context structure, validation rules, phase transitions, anti-patterns. Reference module (not invoked directly)."
---

# Session Common

> **Reference Module** - Shared documentation for session lifecycle

## Not a Command Skill

This module provides canonical schemas used by session commands.
It is not invoked directly via slash commands.

## Contents

- [session-context-schema.md](session-context-schema.md) - Field definitions for session context structure
- [session-phases.md](session-phases.md) - Phase transitions and lifecycle stages
- [session-validation.md](session-validation.md) - Pre-flight checks and validation rules
- [anti-patterns.md](anti-patterns.md) - Common mistakes and anti-patterns

## Used By

Skills that reference this module:
- start-ref
- wrap-ref
- park-ref
- resume
- handoff-ref

## Pattern: Reference Module

session-common follows the "routing hub" pattern - it contains shared documentation
referenced by other skills but is not invoked as a command itself. This pattern
establishes a canonical source for session lifecycle schemas across all session
management commands.
