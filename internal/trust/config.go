package trust

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TrustConfig is the top-level configuration for the trust package.
// Loadable from YAML. Every field has a validated default.
type TrustConfig struct {
	// Decay configures the exponential decay model.
	Decay DecayConfig `yaml:"decay"`

	// Thresholds configures the confidence tier boundaries.
	Thresholds TierThresholds `yaml:"thresholds"`

	// Weights configures the composite scoring weights.
	Weights ScoringWeights `yaml:"weights"`
}

// TierThresholds defines the boundaries between confidence tiers.
// A score >= HighThreshold is HIGH, >= LowThreshold is MEDIUM, else LOW.
//
// INVARIANT: 0.0 <= LowThreshold < HighThreshold <= 1.0
//
// Default values are CONSERVATIVE per Decision #14:
//
//	HighThreshold: 0.7 (only very fresh + well-covered domains qualify)
//	LowThreshold: 0.4 (anything below this triggers refusal)
type TierThresholds struct {
	// HighThreshold: composite score >= this value produces TierHigh.
	HighThreshold float64 `yaml:"high_threshold"`

	// LowThreshold: composite score < this value produces TierLow.
	// Scores between LowThreshold and HighThreshold produce TierMedium.
	LowThreshold float64 `yaml:"low_threshold"`
}

// ScoringWeights configures the relative importance of each input signal.
// Used in the weighted geometric mean: Overall = (F^wf * R^wr * C^wc)^(1/(wf+wr+wc))
//
// Default values emphasize freshness as the dominant signal:
//
//	Freshness:  0.45 (temporal decay is the strongest trust signal)
//	Retrieval:  0.25 (search relevance matters but sprint-2 data is preliminary)
//	Coverage:   0.30 (domain coverage is a strong indicator of answer completeness)
//
// INVARIANT: All weights must be > 0.0.
type ScoringWeights struct {
	Freshness float64 `yaml:"freshness"`
	Retrieval float64 `yaml:"retrieval"`
	Coverage  float64 `yaml:"coverage"`
}

// DefaultConfig returns the production default configuration.
// All values are conservative per Decision #14.
func DefaultConfig() TrustConfig {
	return TrustConfig{
		Decay: DecayConfig{
			DefaultHalfLifeDays: 7.0, // empirical default from Sprint-2 parameter sweep
			HalfLives: map[DomainType]float64{
				DomainArchitecture:      14.0,  // was 30.0; empirical from 187-doc corpus
				DomainConventions:       7.0,   // was 21.0; practices evolve each sprint
				DomainDesignConstraints: 14.0,  // was 30.0; architectural constraints persist
				DomainScarTissue:        10.0,  // was 60.0; lessons age as code changes
				DomainTestCoverage:      5.0,   // was 7.0; coverage changes with every PR
				DomainFeat:              10.0,  // was 14.0; features change at sprint cadence
				DomainRelease:           3.0,   // was 30.0; release info is time-sensitive
				DomainLiterature:        90.0,  // was 180.0; literature reviews persist
			},
		},
		Thresholds: TierThresholds{
			HighThreshold: 0.7,
			LowThreshold:  0.4,
		},
		Weights: ScoringWeights{
			Freshness: 0.45,
			Retrieval: 0.25,
			Coverage:  0.30,
		},
	}
}

// LoadConfig reads a TrustConfig from a YAML file. Returns DefaultConfig() values
// for any fields not present in the file. Returns DefaultConfig() if path is empty
// or the file does not exist.
func LoadConfig(path string) (TrustConfig, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // missing file -> defaults
		}
		return cfg, fmt.Errorf("read trust config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("parse trust config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return DefaultConfig(), fmt.Errorf("invalid trust config: %w", err)
	}

	return cfg, nil
}

// Validate checks that the configuration is internally consistent.
// Returns an error describing the first violation found.
func (c *TrustConfig) Validate() error {
	// Tier threshold invariants
	if c.Thresholds.LowThreshold < 0.0 || c.Thresholds.LowThreshold > 1.0 {
		return fmt.Errorf("low_threshold must be in [0.0, 1.0], got %f", c.Thresholds.LowThreshold)
	}
	if c.Thresholds.HighThreshold < 0.0 || c.Thresholds.HighThreshold > 1.0 {
		return fmt.Errorf("high_threshold must be in [0.0, 1.0], got %f", c.Thresholds.HighThreshold)
	}
	if c.Thresholds.LowThreshold >= c.Thresholds.HighThreshold {
		return fmt.Errorf("low_threshold (%f) must be less than high_threshold (%f)",
			c.Thresholds.LowThreshold, c.Thresholds.HighThreshold)
	}

	// Weight invariants
	if c.Weights.Freshness <= 0 {
		return fmt.Errorf("freshness weight must be > 0, got %f", c.Weights.Freshness)
	}
	if c.Weights.Retrieval <= 0 {
		return fmt.Errorf("retrieval weight must be > 0, got %f", c.Weights.Retrieval)
	}
	if c.Weights.Coverage <= 0 {
		return fmt.Errorf("coverage weight must be > 0, got %f", c.Weights.Coverage)
	}

	// Decay config invariants
	if c.Decay.DefaultHalfLifeDays <= 0 {
		return fmt.Errorf("default_half_life_days must be > 0, got %f", c.Decay.DefaultHalfLifeDays)
	}
	for dt, hl := range c.Decay.HalfLives {
		if hl <= 0 {
			return fmt.Errorf("half_life for %s must be > 0, got %f", dt, hl)
		}
	}

	return nil
}
