package rite

import (
	"time"

	"github.com/autom8y/ariadne/internal/errors"
	"github.com/autom8y/ariadne/internal/paths"
)

// InvokeOptions configures the invoke operation.
type InvokeOptions struct {
	TargetRite    string // Rite to invoke from
	Component     string // "skills", "agents", or "" for all
	DryRun        bool   // Preview only
	NoInscription bool   // Skip CLAUDE.md updates (Phase 2)
}

// InvokeResult contains the result of an invoke operation.
type InvokeResult struct {
	InvokedRite        string        `json:"invoked_rite"`
	Component          string        `json:"component"`
	InvocationID       string        `json:"invocation_id"`
	BorrowedSkills     []string      `json:"borrowed_skills"`
	BorrowedAgents     []InvokedAgent `json:"borrowed_agents"`
	InscriptionUpdated bool          `json:"inscription_updated"`
	EstimatedTokens    int           `json:"estimated_tokens"`
	DryRun             bool          `json:"dry_run,omitempty"`
}

// ReleaseOptions configures the release operation.
type ReleaseOptions struct {
	Target string // Rite name or invocation ID
	All    bool   // Release all invocations
	DryRun bool   // Preview only
}

// ReleaseResult contains the result of a release operation.
type ReleaseResult struct {
	ReleasedRites      []string `json:"released_rites"`
	ReleasedSkills     []string `json:"released_skills"`
	ReleasedAgents     []string `json:"released_agents"`
	InvocationCount    int      `json:"invocation_count"`
	TokensFreed        int      `json:"tokens_freed"`
	InscriptionUpdated bool     `json:"inscription_updated"`
	DryRun             bool     `json:"dry_run,omitempty"`
}

// Invoker handles rite invoke and release operations.
type Invoker struct {
	resolver     *paths.Resolver
	discovery    *Discovery
	stateManager *StateManager
	budget       *BudgetCalculator
}

// NewInvoker creates a new invoker.
func NewInvoker(resolver *paths.Resolver) *Invoker {
	return &Invoker{
		resolver:     resolver,
		discovery:    NewDiscovery(resolver),
		stateManager: NewStateManager(resolver),
		budget:       NewBudgetCalculator(),
	}
}

// Invoke adds components from another rite without switching context.
func (i *Invoker) Invoke(opts InvokeOptions) (*InvokeResult, error) {
	// 1. Load target rite manifest
	targetRite, err := i.discovery.GetManifest(opts.TargetRite)
	if err != nil {
		return nil, err
	}

	// 2. Validate component request matches rite form
	if err := i.validateComponentRequest(targetRite, opts.Component); err != nil {
		return nil, err
	}

	// 3. Load current invocation state
	state, err := i.stateManager.Load()
	if err != nil {
		return nil, err
	}

	// 4. Check for conflicts (same agent borrowed from different rite)
	conflicts := i.detectConflicts(state, targetRite, opts.Component)
	if len(conflicts) > 0 {
		return nil, errors.ErrBorrowConflict(conflicts)
	}

	// 5. Determine what to borrow based on component filter
	borrowed := i.selectComponents(targetRite, opts.Component)

	// 6. Calculate budget impact
	estimatedTokens := i.budget.CalculateInvocationCost(borrowed)
	if state.IsBudgetExceeded(estimatedTokens) {
		return nil, errors.ErrBudgetExceeded(state.Budget.TotalTokens, estimatedTokens, state.Budget.BudgetLimit)
	}

	// 7. Build result
	result := &InvokeResult{
		InvokedRite:     opts.TargetRite,
		Component:       opts.Component,
		BorrowedSkills:  borrowed.Skills,
		BorrowedAgents:  borrowed.Agents,
		EstimatedTokens: estimatedTokens,
		DryRun:          opts.DryRun,
	}

	// Dry run stops here
	if opts.DryRun {
		return result, nil
	}

	// 8. Generate invocation ID
	invocationID := GenerateInvocationID()
	result.InvocationID = invocationID

	// 9. Create invocation record
	invocation := Invocation{
		ID:        invocationID,
		RiteName:  opts.TargetRite,
		Component: opts.Component,
		Skills:    borrowed.Skills,
		Agents:    borrowed.Agents,
		InvokedAt: time.Now().UTC(),
	}

	// 10. Update state
	state.AddInvocation(invocation)
	state.UpdateBudget(state.Budget.NativeTokens, state.Budget.BorrowedTokens+estimatedTokens)

	// 11. Save state
	if err := i.stateManager.Save(state); err != nil {
		return nil, err
	}

	// Note: CLAUDE.md injection is Phase 2
	result.InscriptionUpdated = false

	return result, nil
}

// Release removes borrowed components from a previous invocation.
func (i *Invoker) Release(opts ReleaseOptions) (*ReleaseResult, error) {
	// 1. Load current state
	state, err := i.stateManager.Load()
	if err != nil {
		return nil, err
	}

	result := &ReleaseResult{
		DryRun: opts.DryRun,
	}

	var toRelease []Invocation

	// 2. Determine what to release
	if opts.All {
		toRelease = state.Invocations
	} else if opts.Target != "" {
		// Check if target is an invocation ID
		if inv := state.FindByID(opts.Target); inv != nil {
			toRelease = []Invocation{*inv}
		} else {
			// Treat as rite name
			toRelease = state.FindByRite(opts.Target)
		}
	}

	if len(toRelease) == 0 {
		if opts.Target != "" {
			return nil, errors.ErrInvocationNotFound(opts.Target)
		}
		// Nothing to release
		return result, nil
	}

	// 3. Calculate what will be released
	riteSet := make(map[string]bool)
	for _, inv := range toRelease {
		riteSet[inv.RiteName] = true
		result.ReleasedSkills = append(result.ReleasedSkills, inv.Skills...)
		for _, agent := range inv.Agents {
			result.ReleasedAgents = append(result.ReleasedAgents, agent.Name)
		}
		result.TokensFreed += i.budget.CalculateInvocationCost(&BorrowedComponents{
			Skills: inv.Skills,
			Agents: inv.Agents,
		})
	}

	for rite := range riteSet {
		result.ReleasedRites = append(result.ReleasedRites, rite)
	}
	result.InvocationCount = len(toRelease)

	// Dry run stops here
	if opts.DryRun {
		return result, nil
	}

	// 4. Remove invocations from state
	if opts.All {
		state.RemoveAll()
	} else if opts.Target != "" {
		if state.FindByID(opts.Target) != nil {
			state.RemoveInvocation(opts.Target)
		} else {
			state.RemoveByRite(opts.Target)
		}
	}

	// 5. Update budget
	state.UpdateBudget(state.Budget.NativeTokens, state.Budget.BorrowedTokens-result.TokensFreed)

	// 6. Save state
	if err := i.stateManager.Save(state); err != nil {
		return nil, err
	}

	// Note: CLAUDE.md cleanup is Phase 2
	result.InscriptionUpdated = false

	return result, nil
}

// BorrowedComponents holds the selected components to borrow.
type BorrowedComponents struct {
	Skills []string
	Agents []InvokedAgent
}

// validateComponentRequest validates that the rite can provide the requested components.
func (i *Invoker) validateComponentRequest(rite *RiteManifest, component string) error {
	switch component {
	case "skills":
		if !rite.HasSkills() {
			return errors.ErrInvalidRiteForm(string(rite.Form), "skills")
		}
	case "agents":
		if !rite.HasAgents() {
			return errors.ErrInvalidRiteForm(string(rite.Form), "agents")
		}
	case "": // All components - no validation needed
	default:
		return errors.New(errors.CodeUsageError, "component must be 'skills', 'agents', or empty for all")
	}
	return nil
}

// detectConflicts checks for borrowing conflicts.
func (i *Invoker) detectConflicts(state *InvocationState, rite *RiteManifest, component string) []string {
	var conflicts []string

	borrowedAgents := make(map[string]string) // agent name -> source rite
	for _, inv := range state.Invocations {
		for _, agent := range inv.Agents {
			borrowedAgents[agent.Name] = inv.RiteName
		}
	}

	// Check if we're trying to borrow agents
	if component == "" || component == "agents" {
		for _, agent := range rite.Agents {
			if sourceRite, exists := borrowedAgents[agent.Name]; exists {
				if sourceRite != rite.Name {
					conflicts = append(conflicts, agent.Name+" (already borrowed from "+sourceRite+")")
				}
			}
		}
	}

	return conflicts
}

// selectComponents selects which components to borrow based on the filter.
func (i *Invoker) selectComponents(rite *RiteManifest, component string) *BorrowedComponents {
	borrowed := &BorrowedComponents{}

	switch component {
	case "skills":
		borrowed.Skills = rite.SkillRefs()
	case "agents":
		borrowed.Agents = i.agentRefsToInvoked(rite.Agents)
	default: // All
		borrowed.Skills = rite.SkillRefs()
		borrowed.Agents = i.agentRefsToInvoked(rite.Agents)
	}

	return borrowed
}

// agentRefsToInvoked converts AgentRef slice to InvokedAgent slice.
func (i *Invoker) agentRefsToInvoked(refs []AgentRef) []InvokedAgent {
	agents := make([]InvokedAgent, len(refs))
	for j, ref := range refs {
		agents[j] = InvokedAgent{
			Name: ref.Name,
			File: ref.File,
		}
	}
	return agents
}

// GetCurrentState returns the current invocation state.
func (i *Invoker) GetCurrentState() (*InvocationState, error) {
	return i.stateManager.Load()
}

// SetBudgetLimit sets the budget limit in the state.
func (i *Invoker) SetBudgetLimit(limit int) error {
	state, err := i.stateManager.Load()
	if err != nil {
		return err
	}
	state.SetBudgetLimit(limit)
	return i.stateManager.Save(state)
}
