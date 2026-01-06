package team

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscovery_List(t *testing.T) {
	// Get the testdata path relative to this test file
	testdataPath := filepath.Join("..", "..", "testdata", "rites")

	// Make path absolute
	absPath, err := filepath.Abs(testdataPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Verify testdata exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "")

	rites, err := d.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should find at least valid-rite and minimal-rite
	if len(rites) < 2 {
		t.Errorf("List() returned %d rites, want at least 2", len(rites))
	}

	// Find valid-rite
	var validRite *Rite
	for _, rite := range rites {
		if rite.Name == "valid-rite" {
			validRite = &rite
			break
		}
	}

	if validRite == nil {
		t.Fatal("valid-rite not found in list")
	}

	if validRite.AgentCount != 2 {
		t.Errorf("valid-rite.AgentCount = %d, want 2", validRite.AgentCount)
	}

	if validRite.EntryPoint != "agent-a" {
		t.Errorf("valid-rite.EntryPoint = %q, want %q", validRite.EntryPoint, "agent-a")
	}
}

func TestDiscovery_Get(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "rites")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "")

	tests := []struct {
		name      string
		riteName  string
		wantErr   bool
		wantAgent int
	}{
		{
			name:      "valid rite",
			riteName:  "valid-rite",
			wantErr:   false,
			wantAgent: 2,
		},
		{
			name:      "minimal rite",
			riteName:  "minimal-rite",
			wantErr:   false,
			wantAgent: 1,
		},
		{
			name:     "non-existent rite",
			riteName: "does-not-exist",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rite, err := d.Get(tt.riteName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get(%q) error = %v, wantErr %v", tt.riteName, err, tt.wantErr)
				return
			}
			if !tt.wantErr && rite.AgentCount != tt.wantAgent {
				t.Errorf("Get(%q).AgentCount = %d, want %d", tt.riteName, rite.AgentCount, tt.wantAgent)
			}
		})
	}
}

func TestDiscovery_Exists(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "rites")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "")

	if !d.Exists("valid-rite") {
		t.Error("Exists(valid-rite) = false, want true")
	}

	if d.Exists("non-existent") {
		t.Error("Exists(non-existent) = true, want false")
	}
}

func TestDiscovery_ActiveRite(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "rites")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "valid-rite")

	rites, _ := d.List()

	// Check that valid-rite is marked active
	for _, rite := range rites {
		if rite.Name == "valid-rite" && !rite.Active {
			t.Error("valid-rite.Active = false, want true")
		}
		if rite.Name == "minimal-rite" && rite.Active {
			t.Error("minimal-rite.Active = true, want false")
		}
	}

	if d.ActiveRiteName() != "valid-rite" {
		t.Errorf("ActiveRiteName() = %q, want %q", d.ActiveRiteName(), "valid-rite")
	}
}
