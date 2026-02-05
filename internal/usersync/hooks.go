package usersync

import (
	"os"
	"path/filepath"
	"strings"
)

// Common executable extensions
var executableExtensions = map[string]bool{
	".sh":   true,
	".bash": true,
	".zsh":  true,
	".py":   true,
	".rb":   true,
	".pl":   true,
}

// isExecutable checks if a file path is an executable script.
func isExecutable(path string) bool {
	// Check by extension
	ext := strings.ToLower(filepath.Ext(path))
	if executableExtensions[ext] {
		return true
	}

	// Check if in lib/ directory (all lib scripts should be executable)
	if strings.Contains(path, string(filepath.Separator)+"lib"+string(filepath.Separator)) {
		return true
	}

	// Also check if file basename suggests a script
	base := filepath.Base(path)
	if strings.HasPrefix(base, "hook-") || strings.HasPrefix(base, "pre-") || strings.HasPrefix(base, "post-") {
		return true
	}

	// Check if file has executable bit (Unix)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode()&0111 != 0
}

// isHookConfigFile checks if a file is a hook configuration file.
func isHookConfigFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

// HooksSyncer provides hooks-specific sync functionality.
type HooksSyncer struct {
	*Syncer
}

// NewHooksSyncer creates a syncer specifically for hooks.
func NewHooksSyncer() (*HooksSyncer, error) {
	syncer, err := NewSyncer(ResourceHooks)
	if err != nil {
		return nil, err
	}
	return &HooksSyncer{Syncer: syncer}, nil
}

// HooksAnalysis provides insight into hooks structure.
type HooksAnalysis struct {
	ConfigFiles  []string // .yaml/.yml files
	LibScripts   []string // Files in lib/ directory
	OtherScripts []string // Other executable scripts
	TotalFiles   int
}

// AnalyzeSource analyzes the source hooks directory structure.
func (h *HooksSyncer) AnalyzeSource() (*HooksAnalysis, error) {
	analysis := &HooksAnalysis{
		ConfigFiles:  []string{},
		LibScripts:   []string{},
		OtherScripts: []string{},
	}

	err := filepath.WalkDir(h.sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(h.sourceDir, path)
		if err != nil {
			return err
		}

		analysis.TotalFiles++

		if isHookConfigFile(path) {
			analysis.ConfigFiles = append(analysis.ConfigFiles, relPath)
		} else if strings.HasPrefix(relPath, "lib"+string(filepath.Separator)) {
			analysis.LibScripts = append(analysis.LibScripts, relPath)
		} else if isExecutable(path) {
			analysis.OtherScripts = append(analysis.OtherScripts, relPath)
		}

		return nil
	})

	return analysis, err
}

// EnsureExecutable ensures all scripts in the target hooks directory have execute permission.
func (h *HooksSyncer) EnsureExecutable() error {
	return filepath.WalkDir(h.targetDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if isExecutable(path) {
			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			// Add execute permission if missing
			if info.Mode()&0111 == 0 {
				newMode := info.Mode() | 0755
				if err := os.Chmod(path, newMode); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
