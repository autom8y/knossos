// Package testutil provides test utilities for hook testing.
package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GoldenFile manages golden file comparison for hook output testing.
type GoldenFile struct {
	t    *testing.T
	dir  string
	name string
}

// Golden creates a new golden file handler.
// Golden files are stored in the testdata directory next to the test file.
func Golden(t *testing.T, name string) *GoldenFile {
	t.Helper()

	// Find testdata directory relative to test file
	dir := filepath.Join("testdata", "golden")

	return &GoldenFile{
		t:    t,
		dir:  dir,
		name: name,
	}
}

// GoldenWithDir creates a golden file handler with a specific directory.
func GoldenWithDir(t *testing.T, dir, name string) *GoldenFile {
	t.Helper()
	return &GoldenFile{
		t:    t,
		dir:  dir,
		name: name,
	}
}

// Path returns the full path to the golden file.
func (g *GoldenFile) Path() string {
	return filepath.Join(g.dir, g.name+".golden")
}

// Update returns true if UPDATE_GOLDEN=1 is set.
func (g *GoldenFile) Update() bool {
	return os.Getenv("UPDATE_GOLDEN") == "1"
}

// Assert compares actual output against the golden file.
// If UPDATE_GOLDEN=1, it updates the golden file instead.
func (g *GoldenFile) Assert(actual []byte) {
	g.t.Helper()

	path := g.Path()

	if g.Update() {
		g.write(path, actual)
		return
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			g.t.Fatalf("Golden file not found: %s\nRun with UPDATE_GOLDEN=1 to create it.\nActual output:\n%s", path, string(actual))
		}
		g.t.Fatalf("Failed to read golden file %s: %v", path, err)
	}

	if !bytesEqual(expected, actual) {
		g.t.Errorf("Golden file mismatch: %s\nExpected:\n%s\nActual:\n%s\nRun with UPDATE_GOLDEN=1 to update.", path, string(expected), string(actual))
	}
}

// AssertString compares a string against the golden file.
func (g *GoldenFile) AssertString(actual string) {
	g.Assert([]byte(actual))
}

// AssertJSON compares JSON output against the golden file.
// The actual value is marshaled and pretty-printed before comparison.
func (g *GoldenFile) AssertJSON(actual any) {
	g.t.Helper()

	data, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		g.t.Fatalf("Failed to marshal actual value to JSON: %v", err)
	}

	// Add trailing newline for consistency
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	g.Assert(data)
}

// AssertJSONString compares a JSON string against the golden file.
// The string is parsed and re-formatted for consistent comparison.
func (g *GoldenFile) AssertJSONString(actual string) {
	g.t.Helper()

	var v any
	if err := json.Unmarshal([]byte(actual), &v); err != nil {
		g.t.Fatalf("Failed to parse actual JSON: %v\nInput: %s", err, actual)
	}

	g.AssertJSON(v)
}

// Read returns the golden file content.
func (g *GoldenFile) Read() ([]byte, error) {
	return os.ReadFile(g.Path())
}

// MustRead returns the golden file content or fails the test.
func (g *GoldenFile) MustRead() []byte {
	g.t.Helper()
	data, err := g.Read()
	if err != nil {
		g.t.Fatalf("Failed to read golden file %s: %v", g.Path(), err)
	}
	return data
}

func (g *GoldenFile) write(path string, data []byte) {
	g.t.Helper()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		g.t.Fatalf("Failed to create golden directory %s: %v", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		g.t.Fatalf("Failed to write golden file %s: %v", path, err)
	}

	g.t.Logf("Updated golden file: %s", path)
}

// bytesEqual compares two byte slices, normalizing line endings.
func bytesEqual(a, b []byte) bool {
	// Normalize line endings
	aNorm := strings.ReplaceAll(string(a), "\r\n", "\n")
	bNorm := strings.ReplaceAll(string(b), "\r\n", "\n")

	// Trim trailing whitespace for comparison
	aNorm = strings.TrimRight(aNorm, " \t\n")
	bNorm = strings.TrimRight(bNorm, " \t\n")

	return aNorm == bNorm
}

// FixtureDir returns the path to the fixtures directory.
func FixtureDir(t *testing.T) string {
	t.Helper()
	// Go up from test package to find fixtures
	return filepath.Join("..", "fixtures")
}

// LoadFixture loads a fixture file by name.
func LoadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(FixtureDir(t), name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", path, err)
	}
	return data
}

// LoadFixtureString loads a fixture file as a string.
func LoadFixtureString(t *testing.T, name string) string {
	return string(LoadFixture(t, name))
}
