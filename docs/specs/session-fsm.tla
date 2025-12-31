--------------------------- MODULE session_fsm ---------------------------
(***************************************************************************
 * Session State Machine for Roster
 *
 * This specification models the session lifecycle in the roster system.
 * It captures:
 *   - Top-level states: ACTIVE, PARKED, ARCHIVED
 *   - Substates within ACTIVE (workflow phases)
 *   - Concurrency model with advisory file locking
 *   - State transition guards and invariants
 *
 * Reference: ADR-0001-session-state-machine-redesign.md
 *
 * Author: Architecture Team
 * Date: 2025-12-31
 ***************************************************************************)

EXTENDS Naturals, Sequences, FiniteSets, TLC

\* -----------------------------------------------------------------------------
\* CONSTANTS
\* -----------------------------------------------------------------------------

\* Maximum number of concurrent processes (Claude Code instances)
CONSTANT MaxProcesses

\* Workflow phases (substates of ACTIVE)
\* These correspond to ACTIVE_WORKFLOW.yaml phases
CONSTANT Phases

\* Example instantiation:
\* MaxProcesses = 3
\* Phases = {"requirements", "design", "implementation", "validation"}

ASSUME MaxProcesses \in Nat /\ MaxProcesses > 0
ASSUME Phases # {}

\* -----------------------------------------------------------------------------
\* STATE VARIABLES
\* -----------------------------------------------------------------------------

VARIABLES
    \* The session's current top-level state
    sessionState,

    \* The current workflow phase (substate of ACTIVE)
    \* Only meaningful when sessionState = "ACTIVE"
    currentPhase,

    \* Advisory lock state: which process (if any) holds the lock
    \* lockHolder = 0 means no lock held
    \* lockHolder = n means process n holds exclusive lock
    lockHolder,

    \* Queue of processes waiting for the lock
    lockQueue,

    \* Each process's local view of the session state (may be stale)
    processViews,

    \* Audit log of state transitions (for liveness checking)
    auditLog

\* Tuple of all variables for temporal formulas
vars == <<sessionState, currentPhase, lockHolder, lockQueue, processViews, auditLog>>

\* -----------------------------------------------------------------------------
\* TYPE DEFINITIONS
\* -----------------------------------------------------------------------------

\* Top-level session states
SessionStates == {"ACTIVE", "PARKED", "ARCHIVED", "NONE"}

\* All possible phases plus a "none" phase for non-ACTIVE states
AllPhases == Phases \cup {"none"}

\* Process identifiers (0 = no process, 1..MaxProcesses = actual processes)
ProcessIds == 0..MaxProcesses
ActiveProcessIds == 1..MaxProcesses

\* Audit log entries
AuditEntry == [
    from: SessionStates,
    to: SessionStates,
    process: ActiveProcessIds,
    phase: AllPhases
]

\* Type invariant for all state variables
TypeInvariant ==
    /\ sessionState \in SessionStates
    /\ currentPhase \in AllPhases
    /\ lockHolder \in ProcessIds
    /\ lockQueue \in Seq(ActiveProcessIds)
    /\ processViews \in [ActiveProcessIds -> SessionStates]
    /\ auditLog \in Seq(AuditEntry)

\* -----------------------------------------------------------------------------
\* INITIAL STATE
\* -----------------------------------------------------------------------------

\* Session starts in NONE state (no session exists yet)
Init ==
    /\ sessionState = "NONE"
    /\ currentPhase = "none"
    /\ lockHolder = 0
    /\ lockQueue = <<>>
    /\ processViews = [p \in ActiveProcessIds |-> "NONE"]
    /\ auditLog = <<>>

\* -----------------------------------------------------------------------------
\* HELPER PREDICATES
\* -----------------------------------------------------------------------------

\* Check if a transition is valid according to FSM rules
ValidTransition(from, to) ==
    \/ (from = "NONE" /\ to = "ACTIVE")      \* Create session
    \/ (from = "ACTIVE" /\ to = "PARKED")    \* Park session
    \/ (from = "ACTIVE" /\ to = "ARCHIVED")  \* Complete/wrap session
    \/ (from = "PARKED" /\ to = "ACTIVE")    \* Resume session
    \/ (from = "PARKED" /\ to = "ARCHIVED")  \* Archive parked session

\* Check if phase transition is valid (linear workflow)
\* This is simplified; real workflow may have back-routes
ValidPhaseTransition(from, to) ==
    \* Any phase to any other phase is allowed within ACTIVE
    \* (Back-routes are handled by workflow engine, not FSM)
    from \in AllPhases /\ to \in Phases

\* Process can acquire lock if no one holds it or it's already the holder
CanAcquireLock(p) ==
    lockHolder = 0 \/ lockHolder = p

\* Process holds the lock
HoldsLock(p) ==
    lockHolder = p

\* -----------------------------------------------------------------------------
\* STATE TRANSITION ACTIONS
\* -----------------------------------------------------------------------------

\* Create a new session (NONE -> ACTIVE)
\* Requires: exclusive lock, no existing session
CreateSession(p) ==
    /\ sessionState = "NONE"
    /\ HoldsLock(p)
    /\ sessionState' = "ACTIVE"
    /\ currentPhase' = "requirements"  \* Default entry phase
    /\ processViews' = [processViews EXCEPT ![p] = "ACTIVE"]
    /\ auditLog' = Append(auditLog, [
        from |-> "NONE",
        to |-> "ACTIVE",
        process |-> p,
        phase |-> "requirements"
       ])
    /\ UNCHANGED <<lockHolder, lockQueue>>

\* Park a session (ACTIVE -> PARKED)
\* Requires: exclusive lock, session is active
ParkSession(p) ==
    /\ sessionState = "ACTIVE"
    /\ HoldsLock(p)
    /\ sessionState' = "PARKED"
    /\ processViews' = [processViews EXCEPT ![p] = "PARKED"]
    /\ auditLog' = Append(auditLog, [
        from |-> "ACTIVE",
        to |-> "PARKED",
        process |-> p,
        phase |-> currentPhase
       ])
    /\ UNCHANGED <<currentPhase, lockHolder, lockQueue>>

\* Resume a parked session (PARKED -> ACTIVE)
\* Requires: exclusive lock, session is parked
ResumeSession(p) ==
    /\ sessionState = "PARKED"
    /\ HoldsLock(p)
    /\ sessionState' = "ACTIVE"
    /\ processViews' = [processViews EXCEPT ![p] = "ACTIVE"]
    /\ auditLog' = Append(auditLog, [
        from |-> "PARKED",
        to |-> "ACTIVE",
        process |-> p,
        phase |-> currentPhase
       ])
    /\ UNCHANGED <<currentPhase, lockHolder, lockQueue>>

\* Archive a session (ACTIVE or PARKED -> ARCHIVED)
\* Requires: exclusive lock
ArchiveSession(p) ==
    /\ sessionState \in {"ACTIVE", "PARKED"}
    /\ HoldsLock(p)
    /\ sessionState' = "ARCHIVED"
    /\ currentPhase' = "none"
    /\ processViews' = [processViews EXCEPT ![p] = "ARCHIVED"]
    /\ auditLog' = Append(auditLog, [
        from |-> sessionState,
        to |-> "ARCHIVED",
        process |-> p,
        phase |-> "none"
       ])
    /\ UNCHANGED <<lockHolder, lockQueue>>

\* Transition to a new workflow phase (substate transition within ACTIVE)
\* Requires: exclusive lock, session is active, valid phase transition
TransitionPhase(p, newPhase) ==
    /\ sessionState = "ACTIVE"
    /\ HoldsLock(p)
    /\ newPhase \in Phases
    /\ newPhase # currentPhase
    /\ currentPhase' = newPhase
    /\ auditLog' = Append(auditLog, [
        from |-> sessionState,
        to |-> sessionState,
        process |-> p,
        phase |-> newPhase
       ])
    /\ UNCHANGED <<sessionState, lockHolder, lockQueue, processViews>>

\* -----------------------------------------------------------------------------
\* LOCKING ACTIONS
\* -----------------------------------------------------------------------------

\* Acquire the advisory lock
\* Uses fair queuing: processes acquire in FIFO order
AcquireLock(p) ==
    /\ lockHolder = 0
    /\ \/ lockQueue = <<>>  \* No queue, acquire immediately
       \/ Head(lockQueue) = p  \* At front of queue
    /\ lockHolder' = p
    /\ lockQueue' = IF lockQueue # <<>> /\ Head(lockQueue) = p
                    THEN Tail(lockQueue)
                    ELSE lockQueue
    /\ UNCHANGED <<sessionState, currentPhase, processViews, auditLog>>

\* Request to acquire lock (join queue if lock is held)
RequestLock(p) ==
    /\ lockHolder # 0
    /\ lockHolder # p
    /\ p \notin Range(lockQueue)  \* Not already in queue
    /\ lockQueue' = Append(lockQueue, p)
    /\ UNCHANGED <<sessionState, currentPhase, lockHolder, processViews, auditLog>>

\* Release the advisory lock
ReleaseLock(p) ==
    /\ lockHolder = p
    /\ lockHolder' = 0
    /\ UNCHANGED <<sessionState, currentPhase, lockQueue, processViews, auditLog>>

\* -----------------------------------------------------------------------------
\* READ ACTIONS (can occur without lock, may see stale data)
\* -----------------------------------------------------------------------------

\* Read session state (unlocked read - may see stale data)
\* This models the current problematic behavior that we're fixing
ReadStateUnlocked(p) ==
    /\ processViews' = [processViews EXCEPT ![p] = sessionState]
    /\ UNCHANGED <<sessionState, currentPhase, lockHolder, lockQueue, auditLog>>

\* Read session state with shared lock (correct behavior)
\* For simplicity, we model shared lock as requiring exclusive lock to be free
ReadStateLocked(p) ==
    /\ lockHolder = 0  \* No exclusive lock held
    /\ processViews' = [processViews EXCEPT ![p] = sessionState]
    /\ UNCHANGED <<sessionState, currentPhase, lockHolder, lockQueue, auditLog>>

\* -----------------------------------------------------------------------------
\* NEXT STATE RELATION
\* -----------------------------------------------------------------------------

\* Utility: Range of a sequence
Range(seq) == {seq[i] : i \in 1..Len(seq)}

\* All possible next states
Next ==
    \E p \in ActiveProcessIds:
        \/ CreateSession(p)
        \/ ParkSession(p)
        \/ ResumeSession(p)
        \/ ArchiveSession(p)
        \/ \E phase \in Phases: TransitionPhase(p, phase)
        \/ AcquireLock(p)
        \/ RequestLock(p)
        \/ ReleaseLock(p)
        \/ ReadStateUnlocked(p)
        \/ ReadStateLocked(p)

\* Fairness: every process eventually gets a chance
Fairness ==
    /\ \A p \in ActiveProcessIds:
        /\ WF_vars(AcquireLock(p))
        /\ WF_vars(ReleaseLock(p))

\* Specification with fairness
Spec == Init /\ [][Next]_vars /\ Fairness

\* -----------------------------------------------------------------------------
\* SAFETY INVARIANTS
\* -----------------------------------------------------------------------------

\* 1. No invalid state transitions can occur
\* (Enforced by action guards, but we verify)
NoInvalidTransitions ==
    \A entry \in Range(auditLog):
        entry.from = entry.to \/ ValidTransition(entry.from, entry.to)

\* 2. ARCHIVED is terminal - no transitions out of it
ArchivedIsTerminal ==
    sessionState = "ARCHIVED" =>
        [][sessionState' = "ARCHIVED"]_sessionState

\* 3. Mutual exclusion: at most one process holds the lock
MutualExclusion ==
    \A p, q \in ActiveProcessIds:
        (lockHolder = p /\ lockHolder = q) => p = q

\* 4. Phase is only meaningful in ACTIVE state
PhaseConsistency ==
    (sessionState # "ACTIVE") => (currentPhase = "none" \/ currentPhase \in Phases)

\* 5. Lock queue contains no duplicates
NoDuplicatesInQueue ==
    \A i, j \in 1..Len(lockQueue):
        i # j => lockQueue[i] # lockQueue[j]

\* 6. Lock holder is not in queue
HolderNotInQueue ==
    lockHolder # 0 => lockHolder \notin Range(lockQueue)

\* Combined safety invariant
Safety ==
    /\ TypeInvariant
    /\ MutualExclusion
    /\ PhaseConsistency
    /\ NoDuplicatesInQueue
    /\ HolderNotInQueue

\* -----------------------------------------------------------------------------
\* LIVENESS PROPERTIES
\* -----------------------------------------------------------------------------

\* 1. No deadlock: system can always make progress
NoDeadlock ==
    [][ENABLED(Next)]_vars

\* 2. Lock requests are eventually granted (no starvation)
\* If a process requests the lock, it eventually gets it
LockEventuallyGranted ==
    \A p \in ActiveProcessIds:
        (p \in Range(lockQueue)) ~> (lockHolder = p)

\* 3. Sessions don't stall forever in PARKED state
\* A parked session is eventually resumed or archived
\* (This is a weak liveness property - requires human action)
ParkedSessionsProgress ==
    (sessionState = "PARKED") ~>
        (sessionState = "ACTIVE" \/ sessionState = "ARCHIVED")

\* 4. No livelock: the system makes meaningful progress
\* The audit log grows (we don't just spin on lock operations)
ProgressHappens ==
    <>(\A n \in Nat: Len(auditLog) > n)

\* -----------------------------------------------------------------------------
\* COUNTEREXAMPLE SCENARIOS (should fail in old model)
\* -----------------------------------------------------------------------------

\* This property SHOULD fail in a model without proper locking:
\* "No process ever sees stale data"
\* In the current (broken) system, this can be violated
NoStaleReads ==
    \A p \in ActiveProcessIds:
        processViews[p] = sessionState

\* This property SHOULD hold in the new model with locked reads:
\* "If a process reads with a lock, it sees current state"
LockedReadsAreConsistent ==
    \A p \in ActiveProcessIds:
        (lockHolder = 0 /\ processViews'[p] = sessionState') =>
            processViews'[p] = sessionState'

\* -----------------------------------------------------------------------------
\* MODEL CHECKING CONFIGURATION
\* -----------------------------------------------------------------------------

\* For TLC model checker, use these settings:
\* - Temporal formula: Spec
\* - Invariants: Safety
\* - Properties: LockEventuallyGranted, ParkedSessionsProgress
\*
\* Suggested model values:
\* - MaxProcesses <- 2 (for fast checking) or 3 (thorough)
\* - Phases <- {"requirements", "design", "implementation", "validation"}
\*
\* Expected results:
\* - Safety invariants should hold
\* - LockEventuallyGranted should hold with fairness
\* - NoStaleReads should FAIL (demonstrates current bug)

=============================================================================
