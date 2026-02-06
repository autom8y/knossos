package materialize

import (
	"encoding/json"
	"os"
)

// mergeMCPServers merges MCP server declarations from a rite manifest into existing settings.
// Uses union merge semantics:
// - Rite manifest servers are added/updated
// - Existing satellite servers not in manifest are preserved
// - Output follows Claude Code's mcpServers format:
//   {
//     "mcpServers": {
//       "server-name": {
//         "command": "cmd",
//         "args": ["arg1"],
//         "env": {"KEY": "value"}
//       }
//     }
//   }
func mergeMCPServers(existingSettings map[string]any, mcpServers []MCPServer) map[string]any {
	// Ensure existingSettings has mcpServers key
	if existingSettings["mcpServers"] == nil {
		existingSettings["mcpServers"] = make(map[string]any)
	}

	// Get existing mcpServers (may be empty map)
	mcpServersMap, ok := existingSettings["mcpServers"].(map[string]any)
	if !ok {
		// If not a map, create new empty map
		mcpServersMap = make(map[string]any)
		existingSettings["mcpServers"] = mcpServersMap
	}

	// Merge rite manifest servers (add/update)
	for _, server := range mcpServers {
		serverConfig := make(map[string]any)
		serverConfig["command"] = server.Command

		if len(server.Args) > 0 {
			serverConfig["args"] = server.Args
		}

		if len(server.Env) > 0 {
			serverConfig["env"] = server.Env
		}

		mcpServersMap[server.Name] = serverConfig
	}

	return existingSettings
}

// loadExistingSettings reads settings.local.json if it exists.
// Returns empty map if file doesn't exist or is invalid JSON.
func loadExistingSettings(settingsPath string) (map[string]any, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		// Invalid JSON - return empty map (will be overwritten)
		return make(map[string]any), nil
	}

	return settings, nil
}

// saveSettings writes settings to settings.local.json with pretty formatting.
func saveSettings(settingsPath string, settings map[string]any) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}
