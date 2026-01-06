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
	os.Unsetenv("ROSTER_HOME")
	os.Setenv("KNOSSOS_SUPPRESS_DEPRECATION", "1")
	defer os.Unsetenv("KNOSSOS_SUPPRESS_DEPRECATION")

	// Set primary variable
	os.Setenv("KNOSSOS_HOME", "/custom/knossos")
	defer os.Unsetenv("KNOSSOS_HOME")

	home := KnossosHome()
	if home != "/custom/knossos" {
		t.Errorf("KnossosHome() = %q, want /custom/knossos", home)
	}
}

func TestKnossosHome_Fallback(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	// Clean environment
	os.Unsetenv("KNOSSOS_HOME")
	os.Setenv("KNOSSOS_SUPPRESS_DEPRECATION", "1")
	defer os.Unsetenv("KNOSSOS_SUPPRESS_DEPRECATION")

	// Set fallback variable
	os.Setenv("ROSTER_HOME", "/legacy/roster")
	defer os.Unsetenv("ROSTER_HOME")

	home := KnossosHome()
	if home != "/legacy/roster" {
		t.Errorf("KnossosHome() = %q, want /legacy/roster", home)
	}
}

func TestKnossosHome_Default(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	// Clean environment
	os.Unsetenv("KNOSSOS_HOME")
	os.Unsetenv("ROSTER_HOME")
	os.Setenv("KNOSSOS_SUPPRESS_DEPRECATION", "1")
	defer os.Unsetenv("KNOSSOS_SUPPRESS_DEPRECATION")

	home := KnossosHome()
	expected := filepath.Join(os.Getenv("HOME"), "Code", "roster")
	if home != expected {
		t.Errorf("KnossosHome() = %q, want %q", home, expected)
	}
}

func TestKnossosHome_Precedence(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	// Set both variables
	os.Setenv("KNOSSOS_HOME", "/primary")
	os.Setenv("ROSTER_HOME", "/fallback")
	os.Setenv("KNOSSOS_SUPPRESS_DEPRECATION", "1")
	defer os.Unsetenv("KNOSSOS_HOME")
	defer os.Unsetenv("ROSTER_HOME")
	defer os.Unsetenv("KNOSSOS_SUPPRESS_DEPRECATION")

	home := KnossosHome()
	if home != "/primary" {
		t.Errorf("KnossosHome() = %q, want /primary (KNOSSOS_HOME takes precedence)", home)
	}
}

func TestKnossosHome_Caching(t *testing.T) {
	ResetKnossosHome()
	defer ResetKnossosHome()

	os.Setenv("KNOSSOS_HOME", "/first")
	os.Setenv("KNOSSOS_SUPPRESS_DEPRECATION", "1")
	defer os.Unsetenv("KNOSSOS_HOME")
	defer os.Unsetenv("KNOSSOS_SUPPRESS_DEPRECATION")

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
