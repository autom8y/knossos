// Package envload resolves configuration from a dotenv file at an org's
// XDG data path, layered under process environment variables and CLI flags.
package envload

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// Parse reads dotenv-format content from an io.Reader and returns key-value pairs.
//
// Format rules:
//   - Lines starting with # are comments
//   - Empty lines are ignored
//   - Format is KEY=value (no export prefix)
//   - Values may be unquoted, single-quoted, or double-quoted
//   - Double-quoted values support \n escape sequences
//   - Single-quoted values are literal (no escape processing)
//   - No variable interpolation ($VAR is treated as literal)
//   - Trailing comments on value lines are NOT supported
//   - Duplicate keys log a warning; last value wins
func Parse(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Find the first = separator.
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			return nil, fmt.Errorf("line %d: malformed line (expected KEY=value): %q", lineNum, line)
		}

		key := strings.TrimSpace(line[:idx])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}

		raw := line[idx+1:]
		value, err := parseValue(raw)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if _, exists := result[key]; exists {
			slog.Warn("duplicate key in env file, using last value", "key", key, "line", lineNum)
		}
		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading env file: %w", err)
	}

	return result, nil
}

// parseValue extracts the value from the right side of KEY=value.
func parseValue(raw string) (string, error) {
	raw = strings.TrimLeft(raw, " \t")

	if len(raw) == 0 {
		return "", nil
	}

	switch raw[0] {
	case '"':
		return parseDoubleQuoted(raw)
	case '\'':
		return parseSingleQuoted(raw)
	default:
		return strings.TrimRight(raw, " \t"), nil
	}
}

// parseDoubleQuoted extracts a double-quoted value, expanding \n to newline.
func parseDoubleQuoted(raw string) (string, error) {
	if len(raw) < 2 || raw[len(raw)-1] != '"' {
		return "", fmt.Errorf("unterminated double quote")
	}

	inner := raw[1 : len(raw)-1]
	// Expand escape sequences. Order matters: \\ must be replaced first
	// to avoid double-processing (e.g., \\n should become \n literal, not backslash+newline).
	// Use a placeholder to prevent re-processing.
	const placeholder = "\x00BACKSLASH\x00"
	inner = strings.ReplaceAll(inner, `\\`, placeholder)
	inner = strings.ReplaceAll(inner, `\"`, `"`)
	inner = strings.ReplaceAll(inner, `\n`, "\n")
	inner = strings.ReplaceAll(inner, `\t`, "\t")
	inner = strings.ReplaceAll(inner, placeholder, `\`)

	return inner, nil
}

// parseSingleQuoted extracts a single-quoted value (literal, no escapes).
func parseSingleQuoted(raw string) (string, error) {
	if len(raw) < 2 || raw[len(raw)-1] != '\'' {
		return "", fmt.Errorf("unterminated single quote")
	}

	return raw[1 : len(raw)-1], nil
}
