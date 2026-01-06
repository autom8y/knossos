package team

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscovery_List(t *testing.T) {
	// Get the testdata path relative to this test file
	testdataPath := filepath.Join("..", "..", "testdata", "teams")

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

	teams, err := d.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should find at least valid-team and minimal-team
	if len(teams) < 2 {
		t.Errorf("List() returned %d teams, want at least 2", len(teams))
	}

	// Find valid-team
	var validRite *Rite
	for _,  rite := range teams {
		if rite.Name == "valid-team" {
			validRite = &rite
			break
		}
	}

	if validRite == nil {
		t.Fatal("valid-team not found in list")
	}

	if validRite.AgentCount != 2 {
		t.Errorf("valid-team.AgentCount = %d, want 2", validRite.AgentCount)
	}

	if validRite.EntryPoint != "agent-a" {
		t.Errorf("valid-team.EntryPoint = %q, want %q", validRite.EntryPoint, "agent-a")
	}
}

func TestDiscovery_Get(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "teams")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "")

	tests := []struct {
		name      string
		teamName  string
		wantErr   bool
		wantAgent int
	}{
		{
			name:      "valid team",
			teamName:  "valid-team",
			wantErr:   false,
			wantAgent: 2,
		},
		{
			name:      "minimal team",
			teamName:  "minimal-team",
			wantErr:   false,
			wantAgent: 1,
		},
		{
			name:     "non-existent team",
			teamName: "does-not-exist",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := d.Get(tt.teamName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get(%q) error = %v, wantErr %v", tt.teamName, err, tt.wantErr)
				return
			}
			if !tt.wantErr && team.AgentCount != tt.wantAgent {
				t.Errorf("Get(%q).AgentCount = %d, want %d", tt.teamName, team.AgentCount, tt.wantAgent)
			}
		})
	}
}

func TestDiscovery_Exists(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "teams")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "")

	if !d.Exists("valid-team") {
		t.Error("Exists(valid-team) = false, want true")
	}

	if d.Exists("non-existent") {
		t.Error("Exists(non-existent) = true, want false")
	}
}

func TestDiscovery_ActiveTeam(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "teams")
	absPath, _ := filepath.Abs(testdataPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("testdata not found at %s", absPath)
	}

	d := NewDiscoveryWithPaths(absPath, "", "valid-team")

	teams, _ := d.List()

	// Check that valid-team is marked active
	for _,  rite := range teams {
		if rite.Name == "valid-team" && !rite.Active {
			t.Error("valid-rite.Active = false, want true")
		}
		if rite.Name == "minimal-team" && rite.Active {
			t.Error("minimal-rite.Active = true, want false")
		}
	}

	if d.ActiveRiteName() != "valid-team" {
		t.Errorf("ActiveRiteName() = %q, want %q", d.ActiveRiteName(), "valid-team")
	}
}
