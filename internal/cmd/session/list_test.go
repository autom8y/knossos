package session

import (
	"os"
	"testing"
	"time"
)

// --- Stale Session Threshold Tests ---
//
// staleSessionThreshold() reads ARI_STALE_SESSION_DAYS env var to
// determine when parked sessions are considered stale. It defaults to
// 2 days and falls back to the default for invalid/zero/negative values.

func TestStaleSessionThreshold_Default(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 2 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() = %v, want %v (2 days default)", threshold, expected)
	}
}

func TestStaleSessionThreshold_EnvOverride(t *testing.T) {
	os.Setenv("ARI_STALE_SESSION_DAYS", "5")
	defer os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 5 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() = %v, want %v (5 days from env)", threshold, expected)
	}
}

func TestStaleSessionThreshold_InvalidEnv(t *testing.T) {
	// Non-numeric string should fall back to default
	os.Setenv("ARI_STALE_SESSION_DAYS", "abc")
	defer os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 2 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() with invalid env = %v, want %v (default fallback)", threshold, expected)
	}
}

func TestStaleSessionThreshold_Zero(t *testing.T) {
	// Zero should fall back to default (d > 0 check in implementation)
	os.Setenv("ARI_STALE_SESSION_DAYS", "0")
	defer os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 2 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() with zero = %v, want %v (default fallback)", threshold, expected)
	}
}

func TestStaleSessionThreshold_Negative(t *testing.T) {
	// Negative should fall back to default (d > 0 check in implementation)
	os.Setenv("ARI_STALE_SESSION_DAYS", "-3")
	defer os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 2 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() with negative = %v, want %v (default fallback)", threshold, expected)
	}
}

func TestStaleSessionThreshold_LargeValue(t *testing.T) {
	// Large valid value should work
	os.Setenv("ARI_STALE_SESSION_DAYS", "365")
	defer os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 365 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() = %v, want %v (365 days)", threshold, expected)
	}
}

func TestStaleSessionThreshold_EmptyString(t *testing.T) {
	// Empty string env var should fall back to default
	os.Setenv("ARI_STALE_SESSION_DAYS", "")
	defer os.Unsetenv("ARI_STALE_SESSION_DAYS")

	threshold := staleSessionThreshold()
	expected := 2 * 24 * time.Hour

	if threshold != expected {
		t.Errorf("staleSessionThreshold() with empty string = %v, want %v (default fallback)", threshold, expected)
	}
}
