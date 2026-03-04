// Package hook implements the ari hook commands.
package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autom8y/knossos/internal/registry"
)

// BudgetOutput represents the output of the budget hook.
type BudgetOutput struct {
	Count    int    `json:"count"`
	Warn     int    `json:"warn_threshold,omitempty"`
	Park     int    `json:"park_threshold,omitempty"`
	Severity string `json:"severity,omitempty"` // "warn", "park", or empty
	Message  string `json:"message,omitempty"`
}

// Text implements output.Textable for text output.
func (b BudgetOutput) Text() string {
	if b.Message != "" {
		return b.Message
	}
	return fmt.Sprintf("count=%d", b.Count)
}

// Budget configuration from environment.
const (
	envBudgetDisable = "ARI_BUDGET_DISABLE"
	envMsgWarn       = "ARI_MSG_WARN"
	envMsgPark       = "ARI_MSG_PARK"
	envSessionKey    = "ARI_SESSION_KEY"
	defaultWarn      = 250
)

// newBudgetCmd creates the budget hook subcommand.
func newBudgetCmd(ctx *cmdContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Track tool use count and warn on cognitive budget thresholds",
		Long: `Maintains per-session message count and warns when thresholds are breached.

This hook is triggered on PostToolUse events (all tools). It:
- Increments a per-session counter in a temp file
- Warns (once) when warn threshold is crossed
- Alerts (once) when park threshold is crossed

Environment Variables:
  ARI_MSG_WARN       Warning threshold (default: 250)
  ARI_MSG_PARK       Park suggestion threshold (default: disabled)
  ARI_BUDGET_DISABLE Set to 1 to disable budget tracking
  ARI_SESSION_KEY    Explicit session key (for testing)

Performance: <5ms target execution time (file I/O only).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.withTimeout(func() error {
				return runBudget(ctx)
			})
		},
	}

	return cmd
}

func runBudget(ctx *cmdContext) error {
	printer := ctx.getPrinter()

	// Fast path: disabled via env
	if os.Getenv(envBudgetDisable) == "1" {
		return printer.Print(BudgetOutput{Message: "budget tracking disabled"})
	}

	// Resolve thresholds
	warnThreshold := defaultWarn
	if v := os.Getenv(envMsgWarn); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			warnThreshold = n
		}
	}

	parkThreshold := 0
	if v := os.Getenv(envMsgPark); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			parkThreshold = n
		}
	}

	// Resolve state file path
	stateFile := resolveStateFile(ctx)

	// Increment counter atomically
	count, err := incrementCounter(stateFile)
	if err != nil {
		// Fail open — never block tool execution
		return printer.Print(BudgetOutput{Count: 0, Message: "counter error (fail-open)"})
	}

	out := BudgetOutput{
		Count: count,
		Warn:  warnThreshold,
	}
	if parkThreshold > 0 {
		out.Park = parkThreshold
	}

	// Check warn threshold (one-shot)
	if count >= warnThreshold {
		warnMarker := stateFile + ".warned"
		if _, err := os.Stat(warnMarker); os.IsNotExist(err) {
			out.Severity = "warn"
			out.Message = fmt.Sprintf(
				"Tool use count (%d) reached warning threshold (%d). Consider using %s to preserve session state.",
				count, warnThreshold, registry.Ref(registry.DromenaPark))
			// Write marker (best-effort)
			_ = os.WriteFile(warnMarker, []byte("1"), 0644)
			fmt.Fprintf(os.Stderr, "[cognitive-budget] Warning: %s\n", out.Message)
		}
	}

	// Check park threshold (one-shot, if configured)
	if parkThreshold > 0 && count >= parkThreshold {
		parkMarker := stateFile + ".park-warned"
		if _, err := os.Stat(parkMarker); os.IsNotExist(err) {
			out.Severity = "park"
			out.Message = fmt.Sprintf(
				"Tool use count (%d) reached park threshold (%d). Recommend %s now to preserve session state and avoid context degradation.",
				count, parkThreshold, registry.Ref(registry.DromenaPark))
			// Write marker (best-effort)
			_ = os.WriteFile(parkMarker, []byte("1"), 0644)
			fmt.Fprintf(os.Stderr, "[cognitive-budget] Alert: %s\n", out.Message)
		}
	}

	return printer.Print(out)
}

// resolveStateFile determines the temp file path for counter state.
// Key resolution: ARI_SESSION_KEY > CLAUDE_SESSION_ID > ppid-{PPID}
func resolveStateFile(ctx *cmdContext) string {
	var key string

	if v := os.Getenv(envSessionKey); v != "" {
		key = v
	} else {
		hookEnv := ctx.getHookEnv()
		if hookEnv.SessionID != "" {
			key = hookEnv.SessionID
		} else {
			key = fmt.Sprintf("ppid-%d", os.Getppid())
		}
	}

	// Sanitize key for use as filename
	key = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, key)

	return filepath.Join(os.TempDir(), "ari-msg-count-"+key)
}

// incrementCounter atomically reads, increments, and writes the counter.
func incrementCounter(stateFile string) (int, error) {
	// Read current count
	count := 0
	data, err := os.ReadFile(stateFile)
	if err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			count = n
		}
	}

	count++

	// Atomic write: temp file + rename
	dir := filepath.Dir(stateFile)
	tmpFile, err := os.CreateTemp(dir, "ari-budget-*")
	if err != nil {
		// Fallback: direct write
		return count, os.WriteFile(stateFile, []byte(strconv.Itoa(count)), 0644)
	}

	if _, err := fmt.Fprintf(tmpFile, "%d", count); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return count, os.WriteFile(stateFile, []byte(strconv.Itoa(count)), 0644)
	}
	_ = tmpFile.Close()

	if err := os.Rename(tmpFile.Name(), stateFile); err != nil {
		_ = os.Remove(tmpFile.Name())
		return count, os.WriteFile(stateFile, []byte(strconv.Itoa(count)), 0644)
	}

	return count, nil
}
