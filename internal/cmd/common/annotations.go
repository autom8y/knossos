// Package common provides shared types and utilities for CLI commands.
package common

import "github.com/spf13/cobra"

// Annotation keys for command metadata.
const (
	// AnnotationNeedsProject indicates whether a command requires a project context.
	// Value should be "true" or "false".
	AnnotationNeedsProject = "needsProject"
)

// SetNeedsProject sets the needsProject annotation on a command.
// When recursive is true, also sets the annotation on all subcommands.
func SetNeedsProject(cmd *cobra.Command, needs bool, recursive bool) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	value := "false"
	if needs {
		value = "true"
	}
	cmd.Annotations[AnnotationNeedsProject] = value

	if recursive {
		for _, sub := range cmd.Commands() {
			SetNeedsProject(sub, needs, true)
		}
	}
}

// NeedsProject checks if a command requires a project context.
// It checks the command's annotation first, then parent's annotation,
// then returns the default value (true).
func NeedsProject(cmd *cobra.Command) bool {
	// Check command's own annotation
	if cmd.Annotations != nil {
		if val, ok := cmd.Annotations[AnnotationNeedsProject]; ok {
			return val == "true"
		}
	}

	// Check parent's annotation
	if cmd.Parent() != nil && cmd.Parent().Annotations != nil {
		if val, ok := cmd.Parent().Annotations[AnnotationNeedsProject]; ok {
			return val == "true"
		}
	}

	// Default: require project
	return true
}
