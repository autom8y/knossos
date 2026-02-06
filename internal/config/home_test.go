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
