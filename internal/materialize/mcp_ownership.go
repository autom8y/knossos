package materialize

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/materialize/hooks"
)

// MCPOwnership tracks which MCP servers were materialized by which rite.
// Persisted at .knossos/mcp-ownership.json.
type MCPOwnership struct {
	Rite    string   `json:"rite"`
	Servers []string `json:"servers"`
}

const mcpOwnershipFile = "mcp-ownership.json"

// writeMCPOwnership writes the ownership file tracking which servers belong to the current rite.
func writeMCPOwnership(knossosDir, riteName string, serverNames []string) error {
	ownership := MCPOwnership{
		Rite:    riteName,
		Servers: serverNames,
	}
	// Sort for deterministic output (idempotency)
	sort.Strings(ownership.Servers)

	data, err := json.MarshalIndent(ownership, "", "  ")
	if err != nil {
		return err
	}

	ownershipPath := filepath.Join(knossosDir, mcpOwnershipFile)
	_, err = fileutil.WriteIfChanged(ownershipPath, data, 0644)
	return err
}

// loadMCPOwnership loads the ownership file. Returns nil if the file doesn't exist.
func loadMCPOwnership(knossosDir string) *MCPOwnership {
	ownershipPath := filepath.Join(knossosDir, mcpOwnershipFile)
	data, err := os.ReadFile(ownershipPath)
	if err != nil {
		return nil
	}

	var ownership MCPOwnership
	if err := json.Unmarshal(data, &ownership); err != nil {
		return nil
	}

	return &ownership
}

// pruneStaleMCPServers removes MCP servers owned by the previous rite from .mcp.json.
// Called on rite transition. Returns the number of servers pruned.
// Satellite-added servers (not in ownership) are never touched.
func (m *Materializer) pruneStaleMCPServers(projectRoot string) int {
	knossosDir := m.resolver.KnossosDir()
	prev := loadMCPOwnership(knossosDir)
	if prev == nil || len(prev.Servers) == 0 {
		return 0
	}

	mcpJsonPath := filepath.Join(projectRoot, ".mcp.json")
	existing, err := hooks.LoadExistingSettings(mcpJsonPath)
	if err != nil {
		slog.Warn("failed to load .mcp.json for MCP pruning", "error", err)
		return 0
	}

	mcpServersMap, ok := existing["mcpServers"].(map[string]any)
	if !ok || len(mcpServersMap) == 0 {
		return 0
	}

	pruned := 0
	for _, serverName := range prev.Servers {
		if _, exists := mcpServersMap[serverName]; exists {
			delete(mcpServersMap, serverName)
			pruned++
		}
	}

	if pruned > 0 {
		if err := hooks.SaveSettings(mcpJsonPath, existing); err != nil {
			slog.Warn("failed to write pruned .mcp.json", "error", err)
			return 0
		}
		slog.Info("pruned stale MCP servers on rite switch",
			"pruned", pruned,
			"previous_rite", prev.Rite,
		)
	}

	return pruned
}
