/-
  ThreadAttachment: formal model and correctness proofs for the
  VM.Do() thread attachment state machine (thread.go).

  The Go implementation:
    1. LockOSThread (deferred UnlockOSThread)
    2. getOrAttachCurrentThread:
       - GetEnv → if ok, return (env, false)  -- already attached
       - GetEnv → if EDETACHED, AttachCurrentThread → return (env, true)
    3. If we attached (true), defer detachCurrentThread
    4. Execute fn(env)

  Key properties proven:
    1. Do never detaches a thread it didn't attach
    2. If Do attaches, it always detaches (even on fn error)
    3. Nested Do is safe: inner Do finds thread attached, doesn't re-attach
    4. After Do returns, attachment state is restored to pre-Do state
-/
import JNIProofs.Basic

namespace ThreadAttachment

-- Thread attachment state.
inductive AttachState where
  | detached
  | attached
  deriving DecidableEq, Repr

-- Result of getOrAttachCurrentThread.
structure GetOrAttachResult where
  state : AttachState        -- resulting state (always attached on success)
  attachedByUs : Bool        -- true if we performed the attach
  success : Bool             -- false on error

-- Model getOrAttachCurrentThread.
-- Mirrors: try GetEnv; if EDETACHED, AttachCurrentThread.
def getOrAttach (s : AttachState) : GetOrAttachResult :=
  match s with
  | .attached => { state := .attached, attachedByUs := false, success := true }
  | .detached => { state := .attached, attachedByUs := true, success := true }

-- Model the full Do operation.
-- Returns (fn result succeeded, final attachment state).
-- fnSucceeds models whether the user's callback returns nil or error.
structure DoResult where
  finalState : AttachState
  fnExecuted : Bool

def doOp (initialState : AttachState) (_fnSucceeds : Bool) : DoResult :=
  let gar := getOrAttach initialState
  if ¬gar.success then
    { finalState := initialState, fnExecuted := false }
  else
    -- fn is executed (regardless of _fnSucceeds, the cleanup still happens)
    let postFnState :=
      if gar.attachedByUs then
        AttachState.detached  -- defer vm.detachCurrentThread()
      else
        gar.state             -- no detach needed
    { finalState := postFnState, fnExecuted := true }

-- --------------------------------------------------------------
-- Property 1: Do never detaches a thread it didn't attach.
-- If the thread was already attached, it remains attached after Do.
-- --------------------------------------------------------------

theorem do_preserves_preattached (fnSucceeds : Bool) :
    (doOp .attached fnSucceeds).finalState = .attached := by
  simp [doOp, getOrAttach]

-- --------------------------------------------------------------
-- Property 2: If Do attaches, it always detaches (cleanup guarantee).
-- If the thread was detached, it returns to detached after Do.
-- --------------------------------------------------------------

theorem do_detaches_what_it_attached (fnSucceeds : Bool) :
    (doOp .detached fnSucceeds).finalState = .detached := by
  simp [doOp, getOrAttach]

-- --------------------------------------------------------------
-- Property 3: Do always restores the original attachment state.
-- Combining properties 1 and 2.
-- --------------------------------------------------------------

theorem do_restores_state (s : AttachState) (fnSucceeds : Bool) :
    (doOp s fnSucceeds).finalState = s := by
  cases s <;> simp [doOp, getOrAttach]

-- --------------------------------------------------------------
-- Property 4: fn is always executed (getOrAttach always succeeds).
-- --------------------------------------------------------------

theorem do_always_executes_fn (s : AttachState) (fnSucceeds : Bool) :
    (doOp s fnSucceeds).fnExecuted = true := by
  cases s <;> simp [doOp, getOrAttach]

-- --------------------------------------------------------------
-- Property 5: Nested Do is safe.
-- Inner Do sees "attached" state, so attachedByUs=false, no detach.
-- --------------------------------------------------------------

-- The state during fn execution (after attach, before detach).
def stateDuringFn (initialState : AttachState) : AttachState :=
  (getOrAttach initialState).state

-- During fn execution, the thread is always attached.
theorem during_fn_always_attached (s : AttachState) :
    stateDuringFn s = .attached := by
  cases s <;> simp [stateDuringFn, getOrAttach]

-- Inner getOrAttach sees "attached", so attachedByUs = false.
theorem nested_does_not_reattach (s : AttachState) :
    (getOrAttach (stateDuringFn s)).attachedByUs = false := by
  cases s <;> simp [stateDuringFn, getOrAttach]

-- Nested Do preserves the attached state (inner Do is a no-op
-- regarding attachment).
theorem nested_do_preserves_attached (s : AttachState) (fnSucceeds : Bool) :
    (doOp (stateDuringFn s) fnSucceeds).finalState = .attached := by
  cases s <;> simp [doOp, stateDuringFn, getOrAttach]

-- Full nesting correctness: outer Do(inner Do) restores original state.
-- Model: outer Do attaches if needed, inner Do is no-op, outer detaches.
theorem nested_do_full_correctness (s : AttachState) (fn1 fn2 : Bool) :
    let outerDuring := stateDuringFn s
    let innerResult := doOp outerDuring fn2
    -- Inner Do leaves the thread attached (for outer's fn to continue).
    innerResult.finalState = .attached ∧
    -- Outer Do restores original state.
    (doOp s fn1).finalState = s := by
  cases s <;> simp [doOp, stateDuringFn, getOrAttach]

end ThreadAttachment
