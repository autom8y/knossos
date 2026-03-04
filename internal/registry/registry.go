// Package registry provides a unified denial-recovery registry for knossos
// platform references. It maps stable keys to concrete values (agent names,
// skill paths, CLI commands) and optional recovery hints.
//
// This is a LEAF package — it imports only stdlib. No internal/ imports.
package registry

import (
	"fmt"
	"strings"
)

// RefCategory classifies the kind of platform reference.
type RefCategory int

const (
	// CategorySkill identifies a legomena (skill) reference.
	CategorySkill RefCategory = iota
	// CategoryAgent identifies an agent reference.
	CategoryAgent
	// CategoryCLI identifies a CLI command reference.
	CategoryCLI
	// CategoryDromena identifies a dromena (slash command) reference.
	CategoryDromena
)

// RefKey is a typed string constant used to look up registry entries.
type RefKey string

// Registry keys for platform references.
const (
	// Skills
	SkillConventions      RefKey = "skill.conventions"
	SkillCommitBehavior   RefKey = "skill.commit-behavior"
	SkillAttributionGuard RefKey = "skill.attribution-guard"

	// Agents
	AgentPythia          RefKey = "agent.pythia"
	AgentMoirai          RefKey = "agent.moirai"
	AgentConsultant      RefKey = "agent.consultant"
	AgentContextEngineer RefKey = "agent.context-engineer"

	// CLI commands
	CLISessionFieldSet RefKey = "cli.session-field-set"
	CLISessionLog      RefKey = "cli.session-log"
	CLISessionWrap     RefKey = "cli.session-wrap"

	// Dromena (slash commands)
	DromenaPark RefKey = "dromena.park"
)

// RefEntry holds the concrete value and optional recovery hint for a registry key.
type RefEntry struct {
	// Category classifies the type of reference.
	Category RefCategory
	// Value is the concrete platform reference (agent name, skill path, CLI command).
	Value string
	// Recovery is an optional human-readable hint for resolving denial errors.
	// Empty string means no recovery hint is available.
	Recovery string
}

// entries is the authoritative registry map. Named "entries" to avoid shadowing
// the package name.
var entries = map[RefKey]RefEntry{
	SkillConventions: {
		Category: CategorySkill,
		Value:    "conventions",
		Recovery: "Load skill conventions for safe git workflow guidance.",
	},
	SkillCommitBehavior: {
		Category: CategorySkill,
		Value:    "commit:behavior",
		Recovery: "Load skill commit:behavior for full specification.",
	},
	SkillAttributionGuard: {
		Category: CategorySkill,
		Value:    "conventions",
		Recovery: "Commits use user-only attribution. Load skill conventions for policy details.",
	},
	AgentPythia: {
		Category: CategoryAgent,
		Value:    "pythia",
		Recovery: "",
	},
	AgentMoirai: {
		Category: CategoryAgent,
		Value:    "moirai",
		Recovery: `Task(moirai, "<operation>")`,
	},
	AgentConsultant: {
		Category: CategoryAgent,
		Value:    "consultant",
		Recovery: "",
	},
	AgentContextEngineer: {
		Category: CategoryAgent,
		Value:    "context-engineer",
		Recovery: "",
	},
	CLISessionFieldSet: {
		Category: CategoryCLI,
		Value:    "ari session field-set",
		Recovery: "",
	},
	CLISessionLog: {
		Category: CategoryCLI,
		Value:    "ari session log",
		Recovery: "",
	},
	CLISessionWrap: {
		Category: CategoryCLI,
		Value:    "ari session wrap",
		Recovery: "",
	},
	DromenaPark: {
		Category: CategoryDromena,
		Value:    "/park",
		Recovery: "",
	},
}

// throughlineKeys lists agents that participate in throughline (resume) protocol.
// These agents maintain conversation history across workflow phases.
var throughlineKeys = []RefKey{
	AgentPythia,
	AgentMoirai,
	AgentConsultant,
	AgentContextEngineer,
}

// Ref returns the concrete value for a registry key.
// Panics with a clear message if the key is not registered.
func Ref(key RefKey) string {
	entry, ok := entries[key]
	if !ok {
		panic(fmt.Sprintf("registry: unknown key %q — check registry.go for valid RefKey constants", key))
	}
	return entry.Value
}

// Recovery returns the recovery hint for a registry key.
// Returns an empty string if no recovery hint is registered.
// Panics if the key is not registered.
func Recovery(key RefKey) string {
	entry, ok := entries[key]
	if !ok {
		panic(fmt.Sprintf("registry: unknown key %q — check registry.go for valid RefKey constants", key))
	}
	return entry.Recovery
}

// TaskDelegation returns a formatted Task delegation string for an agent key.
// With no ops: `Task(<agent>, "<operation>")`.
// With ops: `Task(<agent>, "<operation>") -- operations: op1, op2`.
func TaskDelegation(agent RefKey, ops ...string) string {
	agentValue := Ref(agent)
	base := fmt.Sprintf("Task(%s, \"<operation>\")", agentValue)
	if len(ops) == 0 {
		return base
	}
	return fmt.Sprintf("%s -- operations: %s", base, strings.Join(ops, ", "))
}

// ThroughlineAgents returns a map of agent value strings that participate in
// the throughline (resume) protocol. Keys are agent values (e.g. "pythia"),
// values are always true.
func ThroughlineAgents() map[string]bool {
	result := make(map[string]bool, len(throughlineKeys))
	for _, key := range throughlineKeys {
		result[Ref(key)] = true
	}
	return result
}

// EntriesByCategory returns all registry entries matching the given category.
// Order is not guaranteed.
func EntriesByCategory(cat RefCategory) []RefEntry {
	var result []RefEntry
	for _, entry := range entries {
		if entry.Category == cat {
			result = append(result, entry)
		}
	}
	return result
}
