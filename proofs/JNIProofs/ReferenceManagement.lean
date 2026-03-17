/-
  ReferenceManagement: formal model and correctness proofs for JNI
  reference lifecycle management.

  Key properties proven:
    1. A reference is valid after creation
    2. After deletion, reference is in deleted state
    3. Deletion of one ref does not affect other refs
    4. The manager lifecycle is leak-free
    5. Global refs survive thread transitions
-/
import JNIProofs.Basic

namespace ReferenceManagement

inductive RefState where
  | valid
  | deleted
  deriving DecidableEq, Repr

inductive RefType where
  | local
  | global
  | weakGlobal
  deriving DecidableEq, Repr

structure RefTable where
  refMap : Nat → Option RefState
  nextId : Nat

def emptyRefTable : RefTable :=
  { refMap := fun _ => none, nextId := 1 }

def createRef (table : RefTable) : RefTable × Nat :=
  let id := table.nextId
  ({ refMap := fun k => if k == id then some .valid else table.refMap k
     nextId := id + 1 }, id)

def deleteRef (table : RefTable) (id : Nat) : RefTable :=
  { table with refMap := fun k => if k == id then some .deleted else table.refMap k }

def lookupRef (table : RefTable) (id : Nat) : Option RefState :=
  table.refMap id

-- Property 1: A reference is valid after creation.
theorem ref_valid_after_create (table : RefTable) :
    let (table', id) := createRef table
    lookupRef table' id = some .valid := by
  simp [createRef, lookupRef]

-- Property 2: After deletion, reference is in deleted state.
theorem ref_deleted_after_delete (table : RefTable) :
    let (table', id) := createRef table
    let table'' := deleteRef table' id
    lookupRef table'' id = some .deleted := by
  simp [createRef, deleteRef, lookupRef]

-- Property 3: Deletion does not affect other references.
theorem delete_preserves_other (table : RefTable) :
    let (t1, id1) := createRef table
    let (t2, id2) := createRef t1
    let t3 := deleteRef t2 id2
    lookupRef t3 id1 = some .valid := by
  simp [createRef, deleteRef, lookupRef]

-- Property 4: Manager lifecycle is leak-free.
def managerLifecycle (table : RefTable) : RefTable × Nat × Nat :=
  let (t1, localId) := createRef table
  let (t2, globalId) := createRef t1
  let t3 := deleteRef t2 localId
  (t3, localId, globalId)

theorem manager_lifecycle_no_local_leak (table : RefTable) :
    let (t, localId, globalId) := managerLifecycle table
    lookupRef t localId = some .deleted ∧
    lookupRef t globalId = some .valid := by
  simp [managerLifecycle, createRef, deleteRef, lookupRef]

theorem manager_close_frees_global (table : RefTable) :
    let (t, _, globalId) := managerLifecycle table
    let tClosed := deleteRef t globalId
    lookupRef tClosed globalId = some .deleted := by
  simp [managerLifecycle, createRef, deleteRef, lookupRef]

theorem manager_full_cleanup (table : RefTable) :
    let (t, localId, globalId) := managerLifecycle table
    let tClosed := deleteRef t globalId
    lookupRef tClosed localId = some .deleted ∧
    lookupRef tClosed globalId = some .deleted := by
  simp [managerLifecycle, createRef, deleteRef, lookupRef]

-- Property 5: Thread safety of reference types.
def validOnThread (ty : RefType) (creatingThread currentThread : Nat) : Bool :=
  match ty with
  | .local => creatingThread == currentThread
  | .global => true
  | .weakGlobal => true

theorem local_ref_invalid_on_other_thread (t1 t2 : Nat) (h : t1 ≠ t2) :
    validOnThread .local t1 t2 = false := by
  simp [validOnThread]; omega

theorem global_ref_valid_on_any_thread (t1 t2 : Nat) :
    validOnThread .global t1 t2 = true := rfl

theorem manager_ref_valid_anywhere (t1 t2 : Nat) :
    validOnThread .global t1 t2 = true := rfl

end ReferenceManagement
