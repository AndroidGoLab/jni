/-
  ProxyRegistry: formal model and correctness proofs for the proxy
  handler registry (proxy.go).

  Key properties proven:
    1. IDs are always positive (≥ 1)
    2. IDs are strictly monotonically increasing (unique)
    3. After register, lookup succeeds with the registered handler
    4. After unregister, lookup fails
    5. Register/unregister of one ID does not affect other IDs
    6. The two handler maps (basic and full) are independent
    7. Dispatch priority — full handlers take precedence
-/
import JNIProofs.Basic

namespace ProxyRegistry

opaque Handler : Type := Nat
opaque FullHandler : Type := Nat

structure State where
  nextID : Nat
  handlers : Nat → Option Handler
  fullHandlers : Nat → Option FullHandler

def initState : State where
  nextID := 0
  handlers := fun _ => none
  fullHandlers := fun _ => none

def register (s : State) (h : Handler) : State × Nat :=
  let id := s.nextID + 1
  ({ s with
    nextID := id
    handlers := fun k => if k == id then some h else s.handlers k }, id)

def registerFull (s : State) (h : FullHandler) : State × Nat :=
  let id := s.nextID + 1
  ({ s with
    nextID := id
    fullHandlers := fun k => if k == id then some h else s.fullHandlers k }, id)

def lookup (s : State) (id : Nat) : Option Handler := s.handlers id
def lookupFull (s : State) (id : Nat) : Option FullHandler := s.fullHandlers id

def unregister (s : State) (id : Nat) : State :=
  { s with
    handlers := fun k => if k == id then none else s.handlers k
    fullHandlers := fun k => if k == id then none else s.fullHandlers k }

def WellFormed (s : State) : Prop :=
  ∀ k, k > s.nextID → s.handlers k = none ∧ s.fullHandlers k = none

theorem initState_wellFormed : WellFormed initState := by
  intro k _; exact ⟨rfl, rfl⟩

theorem register_preserves_wellFormed {s : State} {h : Handler}
    (wf : WellFormed s) : WellFormed (register s h).1 := by
  intro k hk
  simp only [register] at hk ⊢
  constructor
  · simp only [beq_iff_eq]
    split
    case isTrue heq => subst heq; omega
    case isFalse _ => exact (wf k (by omega)).1
  · exact (wf k (by omega)).2

theorem registerFull_preserves_wellFormed {s : State} {h : FullHandler}
    (wf : WellFormed s) : WellFormed (registerFull s h).1 := by
  intro k hk
  simp only [registerFull] at hk ⊢
  constructor
  · exact (wf k (by omega)).1
  · simp only [beq_iff_eq]
    split
    case isTrue heq => subst heq; omega
    case isFalse _ => exact (wf k (by omega)).2

-- Property 1: IDs are always positive (≥ 1).
theorem register_id_positive (s : State) (h : Handler) :
    (register s h).2 ≥ 1 := by
  simp [register]

theorem registerFull_id_positive (s : State) (h : FullHandler) :
    (registerFull s h).2 ≥ 1 := by
  simp [registerFull]

-- Property 2: Successive IDs are strictly increasing.
theorem register_id_strictly_increasing (s : State) (h1 h2 : Handler) :
    let (s', id1) := register s h1
    let (_, id2) := register s' h2
    id2 > id1 := by
  simp [register]

theorem register_nextID_increases (s : State) (h : Handler) :
    (register s h).1.nextID > s.nextID := by
  simp [register]

theorem register_ids_distinct (s : State) (h1 h2 : Handler) :
    let (s', id1) := register s h1
    let (_, id2) := register s' h2
    id1 ≠ id2 := by
  simp [register]

-- Property 3: After register, lookup succeeds.
theorem register_then_lookup (s : State) (h : Handler) :
    let (s', id) := register s h
    lookup s' id = some h := by
  simp [register, lookup]

theorem registerFull_then_lookupFull (s : State) (h : FullHandler) :
    let (s', id) := registerFull s h
    lookupFull s' id = some h := by
  simp [registerFull, lookupFull]

-- Property 4: After unregister, lookup fails.
theorem unregister_then_lookup (s : State) (id : Nat) :
    lookup (unregister s id) id = none := by
  simp [unregister, lookup]

theorem unregister_then_lookupFull (s : State) (id : Nat) :
    lookupFull (unregister s id) id = none := by
  simp [unregister, lookupFull]

-- Property 5: Register/unregister does not affect other IDs.
theorem register_preserves_other_handlers (s : State) (h : Handler) (k : Nat)
    (hk : k ≠ (register s h).2) :
    lookup (register s h).1 k = lookup s k := by
  simp only [register, lookup, beq_iff_eq] at *
  split
  case isTrue heq => exact absurd heq hk
  case isFalse _ => rfl

theorem unregister_preserves_other_handlers (s : State) (id k : Nat)
    (hk : k ≠ id) :
    lookup (unregister s id) k = lookup s k := by
  simp only [unregister, lookup, beq_iff_eq]
  split
  case isTrue heq => exact absurd heq hk
  case isFalse _ => rfl

-- Property 6: Basic and full handler maps are independent.
theorem register_does_not_affect_full (s : State) (h : Handler) (k : Nat) :
    lookupFull (register s h).1 k = lookupFull s k := by
  simp [register, lookupFull]

theorem registerFull_does_not_affect_basic (s : State) (h : FullHandler) (k : Nat) :
    lookup (registerFull s h).1 k = lookup s k := by
  simp [registerFull, lookup]

-- Property 7: Dispatch priority — full handlers take precedence.
inductive DispatchResult where
  | full (h : FullHandler)
  | basic (h : Handler)
  | notFound

def dispatch (s : State) (id : Nat) : DispatchResult :=
  match lookupFull s id with
  | some h => .full h
  | none =>
    match lookup s id with
    | some h => .basic h
    | none => .notFound

theorem dispatch_full_priority (s : State) (h : FullHandler) :
    let (s', id) := registerFull s h
    dispatch s' id = .full h := by
  simp [registerFull, dispatch, lookupFull]

theorem dispatch_basic_when_no_full (s : State) (h : Handler)
    (wf : WellFormed s) :
    let (s', id) := register s h
    dispatch s' id = .basic h := by
  simp only [register, dispatch, lookupFull, lookup]
  have hfull := (wf (s.nextID + 1) (by omega)).2
  simp [hfull]

theorem dispatch_after_unregister (s : State) (id : Nat) :
    dispatch (unregister s id) id = .notFound := by
  simp [dispatch, unregister, lookupFull, lookup]

end ProxyRegistry
