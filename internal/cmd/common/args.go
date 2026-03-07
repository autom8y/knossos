package common

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ExactArgs returns a positional args validator requiring exactly n arguments.
// Unlike cobra.ExactArgs, the error includes the command's usage line.
func ExactArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d\n\nUsage:\n  %s", n, len(args), cmd.UseLine())
		}
		return nil
	}
}

// RangeArgs returns a positional args validator requiring between min and max arguments.
// Unlike cobra.RangeArgs, the error includes the command's usage line.
func RangeArgs(min, max int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min || len(args) > max {
			return fmt.Errorf("accepts between %d and %d arg(s), received %d\n\nUsage:\n  %s", min, max, len(args), cmd.UseLine())
		}
		return nil
	}
}

// MaximumNArgs returns a positional args validator allowing at most n arguments.
// Unlike cobra.MaximumNArgs, the error includes the command's usage line.
func MaximumNArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > n {
			return fmt.Errorf("accepts at most %d arg(s), received %d\n\nUsage:\n  %s", n, len(args), cmd.UseLine())
		}
		return nil
	}
}

// NoArgs validates that no arguments are provided.
// Unlike cobra.NoArgs, the error includes the command's usage line.
func NoArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("accepts no arguments, received %d\n\nUsage:\n  %s", len(args), cmd.UseLine())
	}
	return nil
}
