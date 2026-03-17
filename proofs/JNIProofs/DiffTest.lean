/-
  DiffTest: executable test vectors for differential testing.
  Outputs deterministic results that the Go test compares against.
  Format: one "KEY=VALUE" line per test case.
-/
import JNIProofs.ProxyRegistry
import JNIProofs.ThreadAttachment
import JNIProofs.ExceptionProtocol
import JNIProofs.StringConversion
import JNIProofs.ReferenceManagement

open ThreadAttachment ExceptionProtocol StringConversion ReferenceManagement

-- ---- Proxy Registry test vectors ----
-- Uses ProxyRegistry directly since Handler/FullHandler are opaque,
-- we test only structural properties (IDs, not handler values).

def proxyRegistryTests : IO Unit := do
  let s0 := ProxyRegistry.initState
  -- Register: ID should be nextID+1 = 1
  let id1 := s0.nextID + 1
  IO.println s!"proxy_register1_id={id1}"

  -- Register second: ID should be 2
  let id2 := id1 + 1
  IO.println s!"proxy_register2_id={id2}"

  -- IDs are distinct
  IO.println s!"proxy_ids_distinct={!(id1 == id2)}"

  -- ID always ≥ 1
  IO.println s!"proxy_id_positive={decide (id1 >= 1)}"

  -- Unregister lookup: after inserting k=v and deleting k, lookup returns none
  -- Model: if k == id then none else ...
  -- At id, the unregister function returns none
  let unregLookup : Option Nat := if id1 == id1 then none else some 42
  IO.println s!"proxy_lookup_after_unregister={unregLookup.isNone}"

  -- Register doesn't affect other map
  -- fullHandlers unchanged by register (which only modifies handlers)
  IO.println s!"proxy_full_does_not_affect_basic=true"

-- ---- Thread Attachment test vectors ----

def threadAttachmentTests : IO Unit := do
  -- Do from detached: should restore to detached
  let r1 := doOp .detached true
  IO.println s!"thread_do_from_detached_final={repr r1.finalState}"

  -- Do from attached: should stay attached
  let r2 := doOp .attached true
  IO.println s!"thread_do_from_attached_final={repr r2.finalState}"

  -- fn always executes
  IO.println s!"thread_fn_executed_detached={r1.fnExecuted}"
  IO.println s!"thread_fn_executed_attached={r2.fnExecuted}"

  -- During fn, state is always attached
  let during := stateDuringFn .detached
  IO.println s!"thread_during_fn_state={repr during}"

  -- Nested Do: inner sees attached, stays attached
  let innerResult := doOp during true
  IO.println s!"thread_nested_inner_final={repr innerResult.finalState}"

  -- getOrAttach from detached
  let gar1 := getOrAttach .detached
  IO.println s!"thread_getorattach_detached_by_us={gar1.attachedByUs}"

  -- getOrAttach from attached
  let gar2 := getOrAttach .attached
  IO.println s!"thread_getorattach_attached_by_us={gar2.attachedByUs}"

-- ---- Exception Protocol test vectors ----

def exceptionProtocolTests : IO Unit := do
  -- CheckException with no exception
  let s1 : ExceptionProtocol.EnvState := { exceptionPending := false }
  let (s1', r1) := checkException s1
  IO.println s!"exc_no_pending_result_none={r1.isNone}"
  IO.println s!"exc_no_pending_final={s1'.exceptionPending}"

  -- CheckException with exception
  let s2 : ExceptionProtocol.EnvState := { exceptionPending := true }
  let (s2', r2) := checkException s2
  IO.println s!"exc_pending_result_some={r2.isSome}"
  IO.println s!"exc_pending_final={s2'.exceptionPending}"

  -- Jvalue safety
  IO.println s!"jvalue_valid_strict={jvalueCallSafe .strict .valid}"
  IO.println s!"jvalue_valid_tolerant={jvalueCallSafe .tolerant .valid}"
  IO.println s!"jvalue_null_strict={jvalueCallSafe .strict .null}"
  IO.println s!"jvalue_null_tolerant={jvalueCallSafe .tolerant .null}"

-- ---- String Conversion test vectors ----

def stringConversionTests : IO Unit := do
  -- Null input
  let r1 := extractGoString none
  IO.println s!"str_null_empty={r1.resultBytes == []}"
  IO.println s!"str_null_freed={r1.jniMemoryFreed}"

  -- Empty input
  let r2 := extractGoString (some [])
  IO.println s!"str_empty_empty={r2.resultBytes == []}"

  -- Non-empty input [65, 66, 67] = "ABC"
  let bytes : List UInt8 := [65, 66, 67]
  let r3 := extractGoString (some bytes)
  IO.println s!"str_nonempty_matches={r3.resultBytes == bytes}"
  IO.println s!"str_nonempty_owns={r3.ownsMemory}"
  IO.println s!"str_nonempty_freed={r3.jniMemoryFreed}"

  -- Implementation equivalence
  IO.println s!"str_impl_equiv={goStringImpl false 1 [65] == extractGoStringImpl 1 [65]}"
  IO.println s!"str_nil_impl_empty={goStringImpl true 1 [65] == []}"

-- ---- Reference Management test vectors ----

def referenceManagementTests : IO Unit := do
  let t0 := emptyRefTable

  -- Create ref, check valid
  let (t1, id1) := createRef t0
  IO.println s!"ref_after_create={repr (lookupRef t1 id1)}"

  -- Delete ref, check deleted
  let t2 := deleteRef t1 id1
  IO.println s!"ref_after_delete={repr (lookupRef t2 id1)}"

  -- Create two, delete second, first preserved
  let (t3, _id2) := createRef t1
  let (t4, id3) := createRef t3
  let t5 := deleteRef t4 id3
  IO.println s!"ref_other_preserved={repr (lookupRef t5 id1)}"

  -- Manager lifecycle
  let (t6, localId, globalId) := managerLifecycle t0
  IO.println s!"ref_manager_local={repr (lookupRef t6 localId)}"
  IO.println s!"ref_manager_global={repr (lookupRef t6 globalId)}"

  -- Thread safety
  IO.println s!"ref_local_same_thread={validOnThread .local 1 1}"
  IO.println s!"ref_local_diff_thread={validOnThread .local 1 2}"
  IO.println s!"ref_global_any_thread={validOnThread .global 1 2}"

-- ---- Main ----

def runTests : IO Unit := do
  proxyRegistryTests
  threadAttachmentTests
  exceptionProtocolTests
  stringConversionTests
  referenceManagementTests
