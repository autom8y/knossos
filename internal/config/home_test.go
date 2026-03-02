package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKnossosHome_Primary(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	// Clean environment
	os.Unsetenv("KNOSSOS_HOME")

	// Set primary variable
	os.Setenv("KNOSSOS_HOME", "/custom/knossos")
	defer os.Unsetenv("KNOSSOS_HOME")

	home := KnossosHome()
	if home != "/custom/knossos" {
		t.Errorf("KnossosHome() = %q, want /custom/knossos", home)
	}
}

func TestKnossosHome_Default(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	// Clean environment
	os.Unsetenv("KNOSSOS_HOME")

	home := KnossosHome()
	expected := filepath.Join(os.Getenv("HOME"), "Code", "knossos")
	if home != expected {
		t.Errorf("KnossosHome() = %q, want %q", home, expected)
	}
}

func TestKnossosHome_Precedence(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	// Set KNOSSOS_HOME
	os.Setenv("KNOSSOS_HOME", "/primary")
	defer os.Unsetenv("KNOSSOS_HOME")

	home := KnossosHome()
	if home != "/primary" {
		t.Errorf("KnossosHome() = %q, want /primary", home)
	}
}

func TestKnossosHome_Caching(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	os.Setenv("KNOSSOS_HOME", "/first")
	defer os.Unsetenv("KNOSSOS_HOME")

	// First call
	home1 := KnossosHome()

	// Change environment (should not affect cached value)
	os.Setenv("KNOSSOS_HOME", "/second")

	// Second call should return cached value
	home2 := KnossosHome()

	if home1 != home2 {
		t.Errorf("KnossosHome() not cached: first=%q, second=%q", home1, home2)
	}
	if home1 != "/first" {
		t.Errorf("KnossosHome() = %q, want /first", home1)
	}
}

// TestKnossosHome_ResetCleanupPattern verifies the canonical test-safety pattern
// documented in ResetKnossosHome.
//
// This is the required pattern for any test that needs a custom KNOSSOS_HOME:
//
//	config.ResetKnossosHome()
//	t.Setenv("KNOSSOS_HOME", customDir)
//	t.Cleanup(config.ResetKnossosHome)
func TestKnossosHome_ResetCleanupPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Reset before setting so no stale cache from prior tests.
	ResetKnossosHome()
	// Step 2: Use t.Setenv so the env var is restored after the test.
	t.Setenv("KNOSSOS_HOME", tmpDir)
	// Step 3: Register cleanup so subsequent tests see a fresh cache.
	t.Cleanup(ResetKnossosHome)

	got := KnossosHome()
	if got != tmpDir {
		t.Errorf("KnossosHome() = %q, want %q", got, tmpDir)
	}
}

// TestKnossosHome_CachePoisonRegression ensures that a reset before calling
// KnossosHome prevents using a stale default path when KNOSSOS_HOME is set.
func TestKnossosHome_CachePoisonRegression(t *testing.T) {
	// Simulate a prior test having called KnossosHome without cleanup.
	// First call caches the default value.
	ResetKnossosHome()
	_ = KnossosHome() // primes the cache with whatever is currently set

	// Now apply the correct reset-before-set sequence.
	tmpDir := t.TempDir()
	ResetKnossosHome()
	t.Setenv("KNOSSOS_HOME", tmpDir)
	t.Cleanup(ResetKnossosHome)

	got := KnossosHome()
	if got != tmpDir {
		t.Errorf("cache poison: KnossosHome() = %q, want %q (reset did not work)", got, tmpDir)
	}
}
