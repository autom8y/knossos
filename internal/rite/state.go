package rite

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/fileutil"
	"github.com/autom8y/knossos/internal/paths"
)

// InvocationState tracks active rite invocations.
type InvocationState struct {
	SchemaVersion string       `yaml:"schema_version" json:"schema_version"`
	CurrentRite   string       `yaml:"current_rite" json:"current_rite"`
	LastUpdated   time.Time    `yaml:"last_updated" json:"last_updated"`
	Invocations   []Invocation `yaml:"invocations" json:"invocations"`
	Budget        StateBudget  `yaml:"budget" json:"budget"`
}

// Invocation represents a single rite invocation.
type Invocation struct {
	ID        string     `yaml:"id" json:"id"`
	RiteName  string     `yaml:"rite_name" json:"rite_name"`
	Component string     `yaml:"component" json:"component"` // "skills", "agents", or "" for all
	Skills    []string   `yaml:"skills,omitempty" json:"skills,omitempty"`
	Agents    []InvokedAgent `yaml:"agents,omitempty" json:"agents,omitempty"`
	InvokedAt time.Time  `yaml:"invoked_at" json:"invoked_at"`
	ExpiresAt *time.Time `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// InvokedAgent represents an agent borrowed from another rite.
type InvokedAgent struct {
	Name string `yaml:"name" json:"name"`
	File string `yaml:"file" json:"file"`
}

// StateBudget tracks token budget across invocations.
type StateBudget struct {
	NativeTokens   int `yaml:"native_tokens" json:"native_tokens"`
	BorrowedTokens int `yaml:"borrowed_tokens" json:"borrowed_tokens"`
	TotalTokens    int `yaml:"total_tokens" json:"total_tokens"`
	BudgetLimit    int `yaml:"budget_limit" json:"budget_limit"`
}

// StateManager handles invocation state persistence.
type StateManager struct {
	resolver *paths.Resolver
}

// NewStateManager creates a new state manager.
func NewStateManager(resolver *paths.Resolver) *StateManager {
	return &StateManager{
		resolver: resolver,
	}
}

// Load reads the invocation state from disk.
func (m *StateManager) Load() (*InvocationState, error) {
	path := m.resolver.InvocationStateFile()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return m.NewState(), nil
		}
		return nil, errors.Wrap(errors.CodeFileNotFound, "failed to read invocation state", err)
	}

	var state InvocationState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, errors.ErrParseError(path, "yaml", err)
	}

	return &state, nil
}

// Save writes the invocation state to disk.
func (m *StateManager) Save(state *InvocationState) error {
	path := m.resolver.InvocationStateFile()

	state.LastUpdated = time.Now().UTC()

	data, err := yaml.Marshal(state)
	if err != nil {
		return errors.Wrap(errors.CodeGeneralError, "failed to marshal invocation state", err)
	}

	if err := fileutil.AtomicWriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.CodePermissionDenied, "failed to write invocation state", err)
	}

	return nil
}

// NewState creates a new empty invocation state.
func (m *StateManager) NewState() *InvocationState {
	return &InvocationState{
		SchemaVersion: "1.0",
		LastUpdated:   time.Now().UTC(),
		Invocations:   []Invocation{},
		Budget: StateBudget{
			BudgetLimit: DefaultBudgetLimit,
		},
	}
}

// DefaultBudgetLimit is the default context budget limit.
const DefaultBudgetLimit = 50000

// AddInvocation adds a new invocation to the state.
func (s *InvocationState) AddInvocation(inv Invocation) {
	s.Invocations = append(s.Invocations, inv)
	s.LastUpdated = time.Now().UTC()
}

// RemoveInvocation removes an invocation by ID and returns it.
func (s *InvocationState) RemoveInvocation(id string) *Invocation {
	for i, inv := range s.Invocations {
		if inv.ID == id {
			removed := s.Invocations[i]
			s.Invocations = append(s.Invocations[:i], s.Invocations[i+1:]...)
			s.LastUpdated = time.Now().UTC()
			return &removed
		}
	}
	return nil
}

// RemoveByRite removes all invocations for a rite and returns them.
func (s *InvocationState) RemoveByRite(riteName string) []Invocation {
	var removed []Invocation
	var remaining []Invocation

	for _, inv := range s.Invocations {
		if inv.RiteName == riteName {
			removed = append(removed, inv)
		} else {
			remaining = append(remaining, inv)
		}
	}

	s.Invocations = remaining
	s.LastUpdated = time.Now().UTC()
	return removed
}

// RemoveAll removes all invocations.
func (s *InvocationState) RemoveAll() []Invocation {
	removed := s.Invocations
	s.Invocations = []Invocation{}
	s.LastUpdated = time.Now().UTC()
	return removed
}

// FindByID finds an invocation by ID.
func (s *InvocationState) FindByID(id string) *Invocation {
	for i := range s.Invocations {
		if s.Invocations[i].ID == id {
			return &s.Invocations[i]
		}
	}
	return nil
}

// FindByRite finds all invocations for a rite.
func (s *InvocationState) FindByRite(riteName string) []Invocation {
	var result []Invocation
	for _, inv := range s.Invocations {
		if inv.RiteName == riteName {
			result = append(result, inv)
		}
	}
	return result
}

// HasInvocations returns true if there are active invocations.
func (s *InvocationState) HasInvocations() bool {
	return len(s.Invocations) > 0
}

// InvocationCount returns the number of active invocations.
func (s *InvocationState) InvocationCount() int {
	return len(s.Invocations)
}

// GetBorrowedSkills returns all borrowed skill refs.
func (s *InvocationState) GetBorrowedSkills() []string {
	var skills []string
	for _, inv := range s.Invocations {
		skills = append(skills, inv.Skills...)
	}
	return skills
}

// GetBorrowedAgents returns all borrowed agents.
func (s *InvocationState) GetBorrowedAgents() []InvokedAgent {
	var agents []InvokedAgent
	for _, inv := range s.Invocations {
		agents = append(agents, inv.Agents...)
	}
	return agents
}

// IsRiteInvoked returns true if the rite has any active invocations.
func (s *InvocationState) IsRiteInvoked(riteName string) bool {
	for _, inv := range s.Invocations {
		if inv.RiteName == riteName {
			return true
		}
	}
	return false
}

// UpdateBudget recalculates the budget totals.
func (s *InvocationState) UpdateBudget(nativeTokens, borrowedTokens int) {
	s.Budget.NativeTokens = nativeTokens
	s.Budget.BorrowedTokens = borrowedTokens
	s.Budget.TotalTokens = nativeTokens + borrowedTokens
}

// SetBudgetLimit sets the budget limit.
func (s *InvocationState) SetBudgetLimit(limit int) {
	s.Budget.BudgetLimit = limit
}

// IsBudgetExceeded returns true if the budget would be exceeded by adding tokens.
func (s *InvocationState) IsBudgetExceeded(additionalTokens int) bool {
	return s.Budget.TotalTokens+additionalTokens > s.Budget.BudgetLimit
}

// BudgetRemaining returns the remaining budget.
func (s *InvocationState) BudgetRemaining() int {
	remaining := s.Budget.BudgetLimit - s.Budget.TotalTokens
	if remaining < 0 {
		return 0
	}
	return remaining
}

// BudgetUsagePercent returns the budget usage as a percentage.
func (s *InvocationState) BudgetUsagePercent() float64 {
	if s.Budget.BudgetLimit == 0 {
		return 0
	}
	return float64(s.Budget.TotalTokens) / float64(s.Budget.BudgetLimit) * 100
}

// GenerateInvocationID creates a unique invocation ID.
func GenerateInvocationID() string {
	b := make([]byte, 6)
	rand.Read(b)
	timestamp := time.Now().Format("20060102")
	return "inv-" + timestamp + "-" + hex.EncodeToString(b)
}

// CleanExpired removes expired invocations.
func (s *InvocationState) CleanExpired() []Invocation {
	now := time.Now()
	var removed []Invocation
	var remaining []Invocation

	for _, inv := range s.Invocations {
		if inv.ExpiresAt != nil && inv.ExpiresAt.Before(now) {
			removed = append(removed, inv)
		} else {
			remaining = append(remaining, inv)
		}
	}

	if len(removed) > 0 {
		s.Invocations = remaining
		s.LastUpdated = time.Now().UTC()
	}

	return removed
}
