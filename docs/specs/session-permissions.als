/**
 * Session Permissions Model for Roster
 *
 * This Alloy specification models the permission and capability system
 * for session state management. It captures:
 *   - Agent roles and their capabilities
 *   - State-dependent operation permissions
 *   - Role hierarchy and delegation
 *   - Separation of duties constraints
 *
 * Reference: ADR-0001-session-state-machine-redesign.md
 *
 * Author: Architecture Team
 * Date: 2025-12-31
 */

// =============================================================================
// BASIC SIGNATURES
// =============================================================================

/**
 * Session states (top-level FSM states)
 */
abstract sig SessionState {}
one sig NONE, ACTIVE, PARKED, ARCHIVED extends SessionState {}

/**
 * Workflow phases (substates of ACTIVE)
 * These correspond to ACTIVE_WORKFLOW.yaml phases
 */
abstract sig Phase {}
one sig Requirements, Design, Implementation, Validation extends Phase {}

/**
 * Operations that can be performed on sessions
 */
abstract sig Operation {}

// State transition operations
one sig CreateSession, ParkSession, ResumeSession, ArchiveSession extends Operation {}

// Phase transition operations
one sig TransitionPhase extends Operation {}

// Read operations
one sig ReadState, ReadContext extends Operation {}

// Mutation operations
one sig UpdateContext, AddArtifact, RecordHandoff extends Operation {}

// Administrative operations
one sig MigrateSession, ValidateSchema, EmitEvent extends Operation {}

/**
 * Agents in the roster system
 */
abstract sig Agent {
    // Roles this agent has
    roles: set Role
}

// Workflow agents (from 10x-dev-pack)
one sig RequirementsAnalyst, Architect, PrincipalEngineer, QAAdversary extends Agent {}

// Special agents
one sig Moirai, Orchestrator, Hook extends Agent {}

// System actor (for automated operations)
one sig System extends Agent {}

/**
 * Roles define sets of capabilities
 */
abstract sig Role {
    // Operations this role can perform
    canPerform: set Operation,
    // States in which this role's operations are valid
    validInStates: set SessionState,
    // Phases in which this role's operations are valid (when state is ACTIVE)
    validInPhases: set Phase
}

// Role hierarchy
one sig WorkflowParticipant extends Role {}
one sig StateManager extends Role {}
one sig PhaseOwner extends Role {}
one sig SystemAdmin extends Role {}
one sig Observer extends Role {}

// =============================================================================
// ROLE DEFINITIONS
// =============================================================================

/**
 * Workflow participants can read and contribute artifacts
 */
fact WorkflowParticipantCapabilities {
    WorkflowParticipant.canPerform = ReadState + ReadContext + AddArtifact + RecordHandoff
    WorkflowParticipant.validInStates = ACTIVE
    WorkflowParticipant.validInPhases = Phase
}

/**
 * State managers can perform state transitions
 * This is primarily moirai
 */
fact StateManagerCapabilities {
    StateManager.canPerform = CreateSession + ParkSession + ResumeSession + ArchiveSession +
                              UpdateContext + ValidateSchema + EmitEvent
    StateManager.validInStates = SessionState - ARCHIVED
    StateManager.validInPhases = Phase
}

/**
 * Phase owners can transition between phases
 */
fact PhaseOwnerCapabilities {
    PhaseOwner.canPerform = TransitionPhase + UpdateContext
    PhaseOwner.validInStates = ACTIVE
    PhaseOwner.validInPhases = Phase
}

/**
 * System admin can perform administrative operations
 */
fact SystemAdminCapabilities {
    SystemAdmin.canPerform = MigrateSession + ValidateSchema + EmitEvent + ArchiveSession
    SystemAdmin.validInStates = SessionState
    SystemAdmin.validInPhases = Phase
}

/**
 * Observers can only read
 */
fact ObserverCapabilities {
    Observer.canPerform = ReadState + ReadContext
    Observer.validInStates = SessionState
    Observer.validInPhases = Phase
}

// =============================================================================
// AGENT-ROLE ASSIGNMENTS
// =============================================================================

/**
 * Workflow agents have participant and phase owner roles
 */
fact WorkflowAgentRoles {
    RequirementsAnalyst.roles = WorkflowParticipant + PhaseOwner
    Architect.roles = WorkflowParticipant + PhaseOwner
    PrincipalEngineer.roles = WorkflowParticipant + PhaseOwner
    QAAdversary.roles = WorkflowParticipant + PhaseOwner
}

/**
 * moirai is the primary state manager
 */
fact MoiraiRoles {
    Moirai.roles = StateManager + PhaseOwner + Observer
}

/**
 * Orchestrator coordinates but doesn't directly manage state
 */
fact OrchestratorRoles {
    Orchestrator.roles = WorkflowParticipant + PhaseOwner + Observer
}

/**
 * Hooks have limited state management for fast-path operations
 */
fact HookRoles {
    // Hooks can read and emit events, but complex mutations go to moirai
    Hook.roles = Observer
}

/**
 * System has admin capabilities for maintenance operations
 */
fact SystemRoles {
    System.roles = SystemAdmin + Observer
}

// =============================================================================
// PERMISSION PREDICATES
// =============================================================================

/**
 * Check if an agent can perform an operation in a given state
 */
pred canPerformOperation[a: Agent, op: Operation, state: SessionState] {
    some r: a.roles | op in r.canPerform and state in r.validInStates
}

/**
 * Check if an agent can perform an operation in a given state and phase
 */
pred canPerformInPhase[a: Agent, op: Operation, state: SessionState, phase: Phase] {
    state = ACTIVE implies {
        some r: a.roles |
            op in r.canPerform and
            state in r.validInStates and
            phase in r.validInPhases
    }
    state != ACTIVE implies canPerformOperation[a, op, state]
}

/**
 * Phase ownership: which agent owns which phase
 */
pred ownsPhase[a: Agent, p: Phase] {
    (p = Requirements and a = RequirementsAnalyst) or
    (p = Design and a = Architect) or
    (p = Implementation and a = PrincipalEngineer) or
    (p = Validation and a = QAAdversary)
}

// =============================================================================
// SEPARATION OF DUTIES CONSTRAINTS
// =============================================================================

/**
 * Only moirai should perform state transitions in production
 * (Hooks may have fallback capability but prefer delegation)
 */
fact StateTransitionAuthority {
    all a: Agent |
        (CreateSession + ParkSession + ResumeSession + ArchiveSession) & (a.roles.canPerform) != none
        implies a in (Moirai + System)
}

/**
 * Phase transitions should be initiated by phase owners or moirai
 */
fact PhaseTransitionAuthority {
    all a: Agent |
        TransitionPhase in a.roles.canPerform implies
        (PhaseOwner in a.roles or StateManager in a.roles)
}

/**
 * Migration operations are system-only
 */
fact MigrationAuthority {
    all a: Agent |
        MigrateSession in a.roles.canPerform implies a = System
}

// =============================================================================
// INVARIANTS
// =============================================================================

/**
 * No agent has empty roles
 */
assert NoEmptyRoles {
    all a: Agent | some a.roles
}

/**
 * All roles have at least one permitted operation
 */
assert RolesHaveOperations {
    all r: Role | some r.canPerform
}

/**
 * Observers cannot modify state
 */
assert ObserversReadOnly {
    all a: Agent, op: Operation |
        (a.roles = Observer and op in a.roles.canPerform) implies
        op in (ReadState + ReadContext)
}

/**
 * ARCHIVED state permits no mutations
 */
assert ArchivedImmutable {
    all a: Agent, op: Operation |
        canPerformOperation[a, op, ARCHIVED] implies op in (ReadState + ReadContext + EmitEvent)
}

/**
 * moirai can perform all necessary state management operations
 */
assert MoiraiComplete {
    CreateSession + ParkSession + ResumeSession + ArchiveSession in Moirai.roles.canPerform
}

/**
 * Each phase has exactly one owner
 */
assert UniquePhaseOwnership {
    all p: Phase | one a: Agent | ownsPhase[a, p]
}

// =============================================================================
// STATE-OPERATION MATRIX
// =============================================================================

/**
 * Define valid operations per state
 * This is the authoritative permission matrix
 */
sig StateOperationMatrix {
    allowed: SessionState -> Operation
}

fact StateOperationRules {
    all m: StateOperationMatrix | {
        // NONE state: only create
        m.allowed[NONE] = CreateSession + ReadState

        // ACTIVE state: all workflow operations
        m.allowed[ACTIVE] = ReadState + ReadContext + UpdateContext + AddArtifact +
                           RecordHandoff + TransitionPhase + ParkSession + ArchiveSession +
                           ValidateSchema + EmitEvent

        // PARKED state: resume, archive, or read
        m.allowed[PARKED] = ReadState + ReadContext + ResumeSession + ArchiveSession +
                           ValidateSchema + EmitEvent

        // ARCHIVED state: read-only plus emit
        m.allowed[ARCHIVED] = ReadState + ReadContext + EmitEvent + MigrateSession
    }
}

/**
 * Operations must be in the allowed set for their state
 */
assert OperationsRespectState {
    all m: StateOperationMatrix, s: SessionState, op: Operation |
        some a: Agent | canPerformOperation[a, op, s] implies op in m.allowed[s]
}

// =============================================================================
// DELEGATION MODEL
// =============================================================================

/**
 * Delegation: hooks can delegate to moirai
 */
sig DelegationRequest {
    from: Agent,
    to: Agent,
    operation: Operation,
    state: SessionState
}

/**
 * Valid delegation: from must be unable, to must be able
 */
fact ValidDelegation {
    all d: DelegationRequest | {
        // Delegator cannot perform the operation directly
        not canPerformOperation[d.from, d.operation, d.state]
        // Delegate can perform the operation
        canPerformOperation[d.to, d.operation, d.state]
    }
}

/**
 * Hooks delegate complex operations to moirai
 */
assert HooksDelegateToMoirai {
    all d: DelegationRequest |
        d.from = Hook and d.operation in (CreateSession + ParkSession + ResumeSession + ArchiveSession)
        implies d.to = Moirai
}

// =============================================================================
// WORKFLOW PHASE TRANSITIONS
// =============================================================================

/**
 * Model phase transition requests
 */
sig PhaseTransitionRequest {
    initiator: Agent,
    fromPhase: Phase,
    toPhase: Phase
}

/**
 * Valid phase transitions follow workflow rules
 * (This is simplified; real workflow has back-routes)
 */
pred validPhaseTransition[r: PhaseTransitionRequest] {
    // Forward transitions
    (r.fromPhase = Requirements and r.toPhase = Design) or
    (r.fromPhase = Design and r.toPhase = Implementation) or
    (r.fromPhase = Implementation and r.toPhase = Validation) or
    // Back-routes (simplified)
    (r.fromPhase = Validation and r.toPhase = Requirements) or
    (r.fromPhase = Validation and r.toPhase = Implementation) or
    (r.fromPhase = Implementation and r.toPhase = Design)
}

/**
 * Initiator must own the source phase or be an orchestrator
 */
pred canInitiateTransition[r: PhaseTransitionRequest] {
    ownsPhase[r.initiator, r.fromPhase] or
    r.initiator in (Orchestrator + Moirai)
}

assert PhaseTransitionsAuthorized {
    all r: PhaseTransitionRequest |
        validPhaseTransition[r] implies canInitiateTransition[r]
}

// =============================================================================
// CHECK COMMANDS
// =============================================================================

// Run checks (should all pass)
check NoEmptyRoles for 10
check RolesHaveOperations for 10
check ObserversReadOnly for 10
check ArchivedImmutable for 10
check MoiraiComplete for 10
check UniquePhaseOwnership for 10
check OperationsRespectState for 10
check HooksDelegateToMoirai for 10

// Find example configurations
run {} for 5

// =============================================================================
// COUNTEREXAMPLES (should fail to demonstrate constraints)
// =============================================================================

/**
 * This predicate should be UNSATISFIABLE
 * It tries to find a case where Hook modifies state directly
 */
pred hookModifiesStateDirectly {
    some s: SessionState |
        canPerformOperation[Hook, CreateSession, s] or
        canPerformOperation[Hook, ParkSession, s] or
        canPerformOperation[Hook, ResumeSession, s] or
        canPerformOperation[Hook, ArchiveSession, s]
}

// This should find no instances
run hookModifiesStateDirectly for 5

/**
 * This predicate should be UNSATISFIABLE
 * It tries to find a case where ARCHIVED state is mutated
 */
pred archivedStateMutated {
    some a: Agent |
        canPerformOperation[a, UpdateContext, ARCHIVED] or
        canPerformOperation[a, AddArtifact, ARCHIVED] or
        canPerformOperation[a, TransitionPhase, ARCHIVED]
}

// This should find no instances
run archivedStateMutated for 5
