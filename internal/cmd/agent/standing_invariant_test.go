package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/autom8y/knossos/internal/registry"
)

// TestStandingAgentRegistryInvariant validates that every entry in the
// standingAgents map has:
//  1. A corresponding agents/{name}.md file in the repository
//  2. A corresponding Agent{Name} registry constant with matching value
//
// This test prevents ghost agent propagation: the pattern where a name
// is added to the standingAgents map without creating the backing file
// and registry constant, causing downstream help text, error messages,
// and roster output to reference a nonexistent agent.
//
// See ADR: Naming Contract for Materialization Pipeline.
func TestStandingAgentRegistryInvariant(t *testing.T) {
	t.Parallel()

	// 1. Locate repo root via runtime.Caller
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// This file lives at internal/cmd/agent/standing_invariant_test.go.
	// filepath.Dir gives internal/cmd/agent/; three levels up reaches the repo root.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")

	// Sanity: verify repo root contains go.mod
	if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err != nil {
		t.Fatalf("repo root sanity check failed: %s does not contain go.mod", repoRoot)
	}

	// 2. Build set of registered agent names from registry
	registeredAgents := make(map[string]bool)
	for _, entry := range registry.EntriesByCategory(registry.CategoryAgent) {
		registeredAgents[entry.Value] = true
	}

	// 3. Validate each standing agent
	for name := range standingAgents {
		t.Run(name, func(t *testing.T) {
			// Invariant 1: agents/{name}.md must exist
			agentFile := filepath.Join(repoRoot, "agents", name+".md")
			if _, err := os.Stat(agentFile); os.IsNotExist(err) {
				t.Errorf("standing agent %q has no agents/%s.md file; "+
					"either create the file or remove %q from the "+
					"standingAgents map in summon.go", name, name, name)
			}

			// Invariant 2: registry must have a CategoryAgent entry with Value == name
			if !registeredAgents[name] {
				t.Errorf("standing agent %q has no Agent{Name} registry "+
					"constant in internal/registry/registry.go; add a "+
					"constant with Category: CategoryAgent and Value: %q",
					name, name)
			}
		})
	}
}
